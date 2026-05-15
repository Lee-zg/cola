// storage 包负责本地 SQLite 持久化，是书签、分类、分析状态和设置的唯一写入点。
package storage

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"cola/internal/bookmark"

	_ "modernc.org/sqlite"
)

type Store struct {
	db      *sql.DB
	dbPath  string
	dataDir string
}

var idSequence atomic.Uint64

func Open(ctx context.Context, dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	// modernc SQLite 在桌面单用户场景下使用单连接更可控，也能避免恢复/关闭时出现多连接文件锁。
	db.SetMaxOpenConns(1)
	store := &Store{db: db, dbPath: dbPath, dataDir: filepath.Dir(dbPath)}
	if err := store.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) DBPath() string {
	return s.dbPath
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) init(ctx context.Context) error {
	// 当前版本直接在启动时幂等建表；后续需要破坏性迁移时应改为版本化 migration。
	statements := []string{
		`PRAGMA journal_mode=WAL;`,
		`PRAGMA foreign_keys=ON;`,
		`CREATE TABLE IF NOT EXISTS categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			parent_id TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0,
			is_system INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS folders (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			parent_id TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS bookmarks (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			folder TEXT NOT NULL DEFAULT 'Unsorted',
			category_id TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '[]',
			keywords TEXT NOT NULL DEFAULT '[]',
			aliases TEXT NOT NULL DEFAULT '[]',
			domain TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			last_analyzed_at TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS bookmark_previews (
			id TEXT PRIMARY KEY,
			bookmark_id TEXT NOT NULL,
			source TEXT NOT NULL,
			file_path TEXT NOT NULL,
			thumb_path TEXT NOT NULL DEFAULT '',
			original_url TEXT NOT NULL DEFAULT '',
			mime TEXT NOT NULL DEFAULT '',
			width INTEGER NOT NULL DEFAULT 0,
			height INTEGER NOT NULL DEFAULT 0,
			size INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			FOREIGN KEY(bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS analysis_jobs (
			bookmark_id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			engine TEXT NOT NULL DEFAULT '',
			error TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL,
			FOREIGN KEY(bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_folder ON bookmarks(folder);`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_domain ON bookmarks(domain);`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_updated ON bookmarks(updated_at);`,
		`CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_previews_bookmark ON bookmark_previews(bookmark_id);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_previews_bookmark_unique ON bookmark_previews(bookmark_id);`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	if err := s.ensureColumn(ctx, "bookmarks", "category_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_bookmarks_category ON bookmarks(category_id);`); err != nil {
		return err
	}
	if err := s.ensureSystemCategories(ctx); err != nil {
		return err
	}
	if err := s.migrateFoldersToCategories(ctx); err != nil {
		return err
	}
	if err := s.repairBookmarkCategories(ctx); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO folders(id, name, sort_order) VALUES('folder_unsorted', ?, 0)`, bookmark.UncategorizedName)
	return err
}

func (s *Store) ensureColumn(ctx context.Context, table, column, definition string) error {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `ALTER TABLE `+table+` ADD COLUMN `+column+` `+definition)
	return err
}

func (s *Store) ensureSystemCategories(ctx context.Context) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO categories
		(id, name, parent_id, sort_order, is_system, created_at, updated_at)
		VALUES(?, ?, '', -1, 1, ?, ?)`, bookmark.UncategorizedID, bookmark.UncategorizedName, now, now)
	return err
}

func (s *Store) migrateFoldersToCategories(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT folder FROM bookmarks WHERE COALESCE(category_id, '') = ''`)
	if err != nil {
		return err
	}
	var folders []string
	for rows.Next() {
		var folder string
		if err := rows.Scan(&folder); err != nil {
			_ = rows.Close()
			return err
		}
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, folder := range folders {
		categoryID, err := s.ensureCategoryPath(ctx, splitCategoryPath(folder))
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE bookmarks SET category_id = ? WHERE COALESCE(category_id, '') = '' AND folder = ?`, categoryID, folder); err != nil {
			return err
		}
	}
	_, err = s.db.ExecContext(ctx, `UPDATE bookmarks SET category_id = ?, folder = ? WHERE COALESCE(category_id, '') = ''`, bookmark.UncategorizedID, bookmark.UncategorizedName)
	return err
}

func (s *Store) repairBookmarkCategories(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT b.id, b.folder
		FROM bookmarks b
		LEFT JOIN categories c ON c.id = b.category_id
		WHERE COALESCE(b.category_id, '') = '' OR c.id IS NULL`)
	if err != nil {
		return err
	}
	type repairItem struct {
		id     string
		folder string
	}
	var items []repairItem
	for rows.Next() {
		var item repairItem
		if err := rows.Scan(&item.id, &item.folder); err != nil {
			_ = rows.Close()
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, item := range items {
		categoryID, err := s.ensureCategoryPath(ctx, splitCategoryPath(item.folder))
		if err != nil {
			return err
		}
		path, err := s.categoryPathString(ctx, categoryID)
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE bookmarks SET category_id = ?, folder = ? WHERE id = ?`, categoryID, path, item.id); err != nil {
			return err
		}
	}
	return nil
}

// CreateBookmark 只创建单条用户输入的书签，URL 唯一冲突转成稳定的业务错误给前端展示。
func (s *Store) CreateBookmark(ctx context.Context, input bookmark.BookmarkInput) (bookmark.Bookmark, error) {
	normalized, domain, err := bookmark.NormalizeInput(input)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	categoryID, folderName, err := s.resolveInputCategory(ctx, normalized)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	now := time.Now().UTC()
	item := bookmark.Bookmark{
		ID:          newID("bm"),
		Title:       normalized.Title,
		URL:         normalized.URL,
		Description: normalized.Description,
		Folder:      folderName,
		CategoryID:  categoryID,
		Tags:        normalized.Tags,
		Keywords:    normalized.Keywords,
		Aliases:     normalized.Aliases,
		Domain:      domain,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO bookmarks
		(id, title, url, description, folder, category_id, tags, keywords, aliases, domain, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Title, item.URL, item.Description, item.Folder, item.CategoryID,
		mustJSON(item.Tags), mustJSON(item.Keywords), mustJSON(item.Aliases),
		item.Domain, formatTime(item.CreatedAt), formatTime(item.UpdatedAt)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return bookmark.Bookmark{}, errors.New("bookmark url already exists")
		}
		return bookmark.Bookmark{}, err
	}
	return s.GetBookmark(ctx, item.ID)
}

// UpsertBookmarks 用一个事务导入批量书签；单条规范化失败只计为跳过，不中断整批导入。
func (s *Store) UpsertBookmarks(ctx context.Context, inputs []bookmark.BookmarkInput) (bookmark.ImportResult, error) {
	result := bookmark.ImportResult{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	now := time.Now().UTC()
	for _, input := range inputs {
		normalized, domain, normErr := bookmark.NormalizeInput(input)
		if normErr != nil {
			result.Skipped++
			result.Errors = append(result.Errors, normErr.Error())
			continue
		}
		categoryID, folderName, catErr := s.resolveInputCategoryTx(ctx, tx, normalized)
		if catErr != nil {
			return result, catErr
		}
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO folders(id, name, sort_order) VALUES(?, ?, 0)`, folderID(folderName), folderName); err != nil {
			return result, err
		}
		res, execErr := tx.ExecContext(ctx, `INSERT OR IGNORE INTO bookmarks
			(id, title, url, description, folder, category_id, tags, keywords, aliases, domain, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			newID("bm"), normalized.Title, normalized.URL, normalized.Description, folderName, categoryID,
			mustJSON(normalized.Tags), mustJSON(normalized.Keywords), mustJSON(normalized.Aliases),
			domain, formatTime(now), formatTime(now))
		if execErr != nil {
			return result, execErr
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			result.Skipped++
		} else {
			result.Imported++
		}
	}
	if err = tx.Commit(); err != nil {
		return result, err
	}
	return result, nil
}

func (s *Store) UpdateBookmark(ctx context.Context, id string, input bookmark.BookmarkInput) (bookmark.Bookmark, error) {
	normalized, domain, err := bookmark.NormalizeInput(input)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	categoryID, folderName, err := s.resolveInputCategory(ctx, normalized)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `UPDATE bookmarks SET
		title = ?, url = ?, description = ?, folder = ?, category_id = ?, tags = ?, keywords = ?, aliases = ?,
		domain = ?, updated_at = ?
		WHERE id = ?`,
		normalized.Title, normalized.URL, normalized.Description, folderName, categoryID,
		mustJSON(normalized.Tags), mustJSON(normalized.Keywords), mustJSON(normalized.Aliases),
		domain, formatTime(now), id)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return bookmark.Bookmark{}, sql.ErrNoRows
	}
	return s.GetBookmark(ctx, id)
}

func (s *Store) DeleteBookmark(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM bookmarks WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GetBookmark(ctx context.Context, id string) (bookmark.Bookmark, error) {
	row := s.db.QueryRowContext(ctx, bookmarkSelectSQL()+` WHERE b.id = ?`, id)
	item, err := scanBookmark(row)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	if err := s.decorateBookmark(ctx, &item); err != nil {
		return bookmark.Bookmark{}, err
	}
	return item, nil
}

func (s *Store) ListBookmarks(ctx context.Context, req bookmark.SearchRequest) (bookmark.SearchResult, error) {
	req = bookmark.NormalizeSearch(req)
	where, args, err := s.buildWhere(ctx, req)
	if err != nil {
		return bookmark.SearchResult{}, err
	}
	orderBy := "updated_at DESC"
	switch req.Sort {
	case "title":
		orderBy = "lower(b.title) ASC"
	case "created":
		orderBy = "b.created_at DESC"
	default:
		orderBy = "b.updated_at DESC"
	}

	countQuery := `SELECT count(*) FROM bookmarks b ` + where
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return bookmark.SearchResult{}, err
	}

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, req.Limit, req.Offset)
	rows, err := s.db.QueryContext(ctx, bookmarkSelectSQL()+` `+where+` ORDER BY `+orderBy+` LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return bookmark.SearchResult{}, err
	}
	items := []bookmark.Bookmark{}
	for rows.Next() {
		item, err := scanBookmark(rows)
		if err != nil {
			_ = rows.Close()
			return bookmark.SearchResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return bookmark.SearchResult{}, err
	}
	if err := rows.Close(); err != nil {
		return bookmark.SearchResult{}, err
	}
	for index := range items {
		if err := s.decorateBookmark(ctx, &items[index]); err != nil {
			return bookmark.SearchResult{}, err
		}
	}
	return bookmark.SearchResult{Items: items, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}

// AllBookmarks 复用分页查询分批读取，避免绕过搜索规范化和排序规则。
func (s *Store) AllBookmarks(ctx context.Context) ([]bookmark.Bookmark, error) {
	result, err := s.ListBookmarks(ctx, bookmark.SearchRequest{Limit: bookmark.MaxLimit, Sort: "title"})
	if err != nil {
		return nil, err
	}
	if result.Total <= len(result.Items) {
		return result.Items, nil
	}
	all := make([]bookmark.Bookmark, 0, result.Total)
	for offset := 0; offset < result.Total; offset += bookmark.MaxLimit {
		page, err := s.ListBookmarks(ctx, bookmark.SearchRequest{Limit: bookmark.MaxLimit, Offset: offset, Sort: "title"})
		if err != nil {
			return nil, err
		}
		all = append(all, page.Items...)
	}
	return all, nil
}

func (s *Store) ListFolders(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name FROM categories WHERE id != ? ORDER BY lower(name)`, bookmark.UncategorizedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	folders := []string{bookmark.UncategorizedName}
	for rows.Next() {
		var folder string
		if err := rows.Scan(&folder); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}
	return folders, rows.Err()
}

func (s *Store) ListTags(ctx context.Context) ([]string, error) {
	items, err := s.AllBookmarks(ctx)
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, item := range items {
		tags = append(tags, item.Tags...)
	}
	return bookmark.NormalizeList(tags), nil
}

func (s *Store) ListCategories(ctx context.Context) ([]bookmark.CategoryNode, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT c.id, c.name, COALESCE(c.parent_id, ''), c.sort_order, c.is_system,
		c.created_at, c.updated_at, COUNT(b.id)
		FROM categories c
		LEFT JOIN bookmarks b ON b.category_id = c.id
		GROUP BY c.id, c.name, c.parent_id, c.sort_order, c.is_system, c.created_at, c.updated_at
		ORDER BY c.sort_order ASC, lower(c.name) ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodes := map[string]*bookmark.CategoryNode{}
	childrenByParent := map[string][]*bookmark.CategoryNode{}
	order := []string{}
	for rows.Next() {
		var node bookmark.CategoryNode
		var isSystem int
		var createdAt, updatedAt string
		if err := rows.Scan(&node.ID, &node.Name, &node.ParentID, &node.SortOrder, &isSystem, &createdAt, &updatedAt, &node.Count); err != nil {
			return nil, err
		}
		node.IsSystem = isSystem == 1
		node.CreatedAt = parseTime(createdAt)
		node.UpdatedAt = parseTime(updatedAt)
		node.Children = []bookmark.CategoryNode{}
		nodes[node.ID] = &node
		order = append(order, node.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, id := range order {
		node := nodes[id]
		childrenByParent[normalizeParentID(node.ParentID)] = append(childrenByParent[normalizeParentID(node.ParentID)], node)
	}
	var buildChildren func(parentID string) []bookmark.CategoryNode
	buildChildren = func(parentID string) []bookmark.CategoryNode {
		children := childrenByParent[parentID]
		out := make([]bookmark.CategoryNode, 0, len(children))
		for _, child := range children {
			node := *child
			node.Children = buildChildren(node.ID)
			for _, grandchild := range node.Children {
				node.Count += grandchild.Count
			}
			out = append(out, node)
		}
		return out
	}
	root := bookmark.CategoryNode{
		ID:       bookmark.RootCategoryID,
		Name:     "全部分类",
		IsSystem: true,
		Children: []bookmark.CategoryNode{},
	}
	root.Children = buildChildren("")
	for _, child := range root.Children {
		root.Count += child.Count
	}
	return []bookmark.CategoryNode{root}, nil
}

func (s *Store) CreateCategory(ctx context.Context, input bookmark.CategoryInput) (bookmark.CategoryNode, error) {
	normalized, err := bookmark.NormalizeCategoryInput(input)
	if err != nil {
		return bookmark.CategoryNode{}, err
	}
	parentID := normalizeParentID(normalized.ParentID)
	if parentID == bookmark.UncategorizedID {
		return bookmark.CategoryNode{}, errors.New("uncategorized category cannot contain child categories")
	}
	if parentID != "" {
		if _, err := s.getCategory(ctx, parentID); err != nil {
			return bookmark.CategoryNode{}, err
		}
	}
	if exists, err := s.categoryNameExists(ctx, parentID, normalized.Name, ""); err != nil {
		return bookmark.CategoryNode{}, err
	} else if exists {
		return bookmark.CategoryNode{}, errors.New("category name already exists")
	}
	now := time.Now().UTC()
	node := bookmark.CategoryNode{
		ID:        newID("cat"),
		Name:      normalized.Name,
		ParentID:  parentID,
		SortOrder: s.nextCategorySortOrder(ctx, parentID),
		CreatedAt: now,
		UpdatedAt: now,
		Children:  []bookmark.CategoryNode{},
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO categories(id, name, parent_id, sort_order, is_system, created_at, updated_at)
		VALUES(?, ?, ?, ?, 0, ?, ?)`, node.ID, node.Name, nullableParent(node.ParentID), node.SortOrder, formatTime(now), formatTime(now))
	if err != nil {
		return bookmark.CategoryNode{}, err
	}
	return node, nil
}

func (s *Store) UpdateCategory(ctx context.Context, id string, input bookmark.CategoryInput) (bookmark.CategoryNode, error) {
	node, err := s.getCategory(ctx, id)
	if err != nil {
		return bookmark.CategoryNode{}, err
	}
	if node.IsSystem {
		return bookmark.CategoryNode{}, errors.New("system category cannot be edited")
	}
	normalized, err := bookmark.NormalizeCategoryInput(input)
	if err != nil {
		return bookmark.CategoryNode{}, err
	}
	if exists, err := s.categoryNameExists(ctx, normalizeParentID(node.ParentID), normalized.Name, id); err != nil {
		return bookmark.CategoryNode{}, err
	} else if exists {
		return bookmark.CategoryNode{}, errors.New("category name already exists")
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `UPDATE categories SET name = ?, updated_at = ? WHERE id = ?`, normalized.Name, formatTime(now), id); err != nil {
		return bookmark.CategoryNode{}, err
	}
	if err := s.syncBookmarkFoldersForCategory(ctx, id); err != nil {
		return bookmark.CategoryNode{}, err
	}
	return s.getCategory(ctx, id)
}

func (s *Store) MoveCategory(ctx context.Context, id string, input bookmark.MoveCategoryInput) (bookmark.CategoryNode, error) {
	node, err := s.getCategory(ctx, id)
	if err != nil {
		return bookmark.CategoryNode{}, err
	}
	if node.IsSystem {
		return bookmark.CategoryNode{}, errors.New("system category cannot be moved")
	}
	parentID := normalizeParentID(input.ParentID)
	if parentID == bookmark.UncategorizedID {
		return bookmark.CategoryNode{}, errors.New("category cannot move into uncategorized")
	}
	if parentID != "" {
		if parentID == id {
			return bookmark.CategoryNode{}, errors.New("category cannot move into itself")
		}
		descendants, err := s.descendantCategoryIDs(ctx, id)
		if err != nil {
			return bookmark.CategoryNode{}, err
		}
		for _, descendantID := range descendants {
			if descendantID == parentID {
				return bookmark.CategoryNode{}, errors.New("category cannot move into its child")
			}
		}
		if _, err := s.getCategory(ctx, parentID); err != nil {
			return bookmark.CategoryNode{}, err
		}
	}
	if exists, err := s.categoryNameExists(ctx, parentID, node.Name, id); err != nil {
		return bookmark.CategoryNode{}, err
	} else if exists {
		return bookmark.CategoryNode{}, errors.New("category name already exists")
	}
	sortOrder := input.SortOrder
	if sortOrder < 0 {
		sortOrder = s.nextCategorySortOrder(ctx, parentID)
	}
	now := formatTime(time.Now().UTC())
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return bookmark.CategoryNode{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `UPDATE categories SET parent_id = ?, sort_order = ?, updated_at = ? WHERE id = ?`, nullableParent(parentID), sortOrder, now, id); err != nil {
		return bookmark.CategoryNode{}, err
	}
	if err = rebalanceCategorySortOrderTx(ctx, tx, id, parentID, sortOrder, node.IsSystem); err != nil {
		return bookmark.CategoryNode{}, err
	}
	if err = tx.Commit(); err != nil {
		return bookmark.CategoryNode{}, err
	}
	if err := s.syncBookmarkFoldersForCategory(ctx, id); err != nil {
		return bookmark.CategoryNode{}, err
	}
	return s.getCategory(ctx, id)
}

func (s *Store) DeleteCategory(ctx context.Context, id string) error {
	return s.DeleteCategoryWithOptions(ctx, id, bookmark.DeleteCategoryInput{})
}

func (s *Store) DeleteCategoryWithOptions(ctx context.Context, id string, input bookmark.DeleteCategoryInput) error {
	node, err := s.getCategory(ctx, id)
	if err != nil {
		return err
	}
	if node.IsSystem {
		return errors.New("system category cannot be deleted")
	}
	descendants, err := s.descendantCategoryIDs(ctx, id)
	if err != nil {
		return err
	}
	deleteIDs := append([]string{id}, descendants...)
	targetID := normalizeParentID(node.ParentID)
	if targetID == "" {
		targetID = bookmark.UncategorizedID
	}
	target, err := s.getCategory(ctx, targetID)
	if err != nil {
		return err
	}
	targetPath, err := s.categoryPathString(ctx, target.ID)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	placeholders := make([]string, 0, len(deleteIDs))
	args := []any{}
	for _, categoryID := range deleteIDs {
		placeholders = append(placeholders, "?")
		args = append(args, categoryID)
	}
	if input.DeleteBookmarks {
		// 勾选删除书签时先清理书签，预览图记录由外键级联删除。
		if _, err = tx.ExecContext(ctx, `DELETE FROM bookmarks WHERE category_id IN (`+strings.Join(placeholders, ",")+`)`, args...); err != nil {
			return err
		}
	} else {
		updateArgs := append([]any{target.ID, targetPath}, args...)
		if _, err = tx.ExecContext(ctx, `UPDATE bookmarks SET category_id = ?, folder = ? WHERE category_id IN (`+strings.Join(placeholders, ",")+`)`, updateArgs...); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM categories WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...); err != nil {
		return err
	}
	return tx.Commit()
}

// ApplyAnalysis 合并分析建议而不是覆盖字段，保护用户手工维护的标签、关键词和别名。
func (s *Store) ApplyAnalysis(ctx context.Context, id string, result bookmark.AnalysisResult) (bookmark.Bookmark, error) {
	item, err := s.GetBookmark(ctx, id)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	now := time.Now().UTC()
	tags := bookmark.MergeLists(item.Tags, result.Tags)
	keywords := bookmark.MergeLists(item.Keywords, result.Keywords)
	aliases := bookmark.MergeLists(item.Aliases, result.Aliases)
	if _, err := s.db.ExecContext(ctx, `UPDATE bookmarks SET tags = ?, keywords = ?, aliases = ?, last_analyzed_at = ?, updated_at = ? WHERE id = ?`,
		mustJSON(tags), mustJSON(keywords), mustJSON(aliases), formatTime(now), formatTime(now), id); err != nil {
		return bookmark.Bookmark{}, err
	}
	_, _ = s.db.ExecContext(ctx, `INSERT INTO analysis_jobs(bookmark_id, status, engine, error, updated_at)
		VALUES(?, 'done', ?, '', ?)
		ON CONFLICT(bookmark_id) DO UPDATE SET status = excluded.status, engine = excluded.engine, error = '', updated_at = excluded.updated_at`,
		id, result.Engine, formatTime(now))
	return s.GetBookmark(ctx, id)
}

func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO settings(key, value) VALUES(?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

func (s *Store) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

func (s *Store) GetPreferences(ctx context.Context) (bookmark.AppPreferences, error) {
	value, err := s.GetSetting(ctx, "open_browser")
	if err != nil {
		return bookmark.AppPreferences{}, err
	}
	return bookmark.AppPreferences{OpenBrowser: bookmark.NormalizeBrowserSetting(value)}, nil
}

func (s *Store) SavePreferences(ctx context.Context, prefs bookmark.AppPreferences) (bookmark.AppPreferences, error) {
	normalized := bookmark.AppPreferences{OpenBrowser: bookmark.NormalizeBrowserSetting(prefs.OpenBrowser)}
	if err := s.SetSetting(ctx, "open_browser", normalized.OpenBrowser); err != nil {
		return bookmark.AppPreferences{}, err
	}
	return normalized, nil
}

// PreviewFileServer 只暴露 previews 目录，前端用相对 URL 展示数据库中保存的本地预览图。
func (s *Store) PreviewFileServer() http.Handler {
	return http.StripPrefix("/previews/", http.FileServer(http.Dir(filepath.Join(s.dataDir, "previews"))))
}

func (s *Store) SaveBookmarkPreview(ctx context.Context, bookmarkID string, input bookmark.PreviewInput) (bookmark.Preview, error) {
	if _, err := s.GetBookmark(ctx, bookmarkID); err != nil {
		return bookmark.Preview{}, err
	}
	input.Source = strings.TrimSpace(input.Source)
	if input.Source == "" {
		input.Source = "manual"
	}
	sourcePath := strings.TrimSpace(input.Path)
	if sourcePath == "" {
		return bookmark.Preview{}, errors.New("preview path is required")
	}
	stat, err := os.Stat(sourcePath)
	if err != nil {
		return bookmark.Preview{}, err
	}
	if stat.IsDir() {
		return bookmark.Preview{}, errors.New("preview path must be a file")
	}
	file, err := os.Open(sourcePath)
	if err != nil {
		return bookmark.Preview{}, err
	}
	defer file.Close()
	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return bookmark.Preview{}, err
	}
	hashName := fmt.Sprintf("%x", hash.Sum(nil))
	ext := strings.ToLower(filepath.Ext(sourcePath))
	if ext == "" {
		ext = ".img"
	}
	relativePath := filepath.ToSlash(filepath.Join("previews", "original", hashName+ext))
	targetPath := filepath.Join(s.dataDir, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return bookmark.Preview{}, err
	}
	if err := copyLocalFile(sourcePath, targetPath); err != nil {
		return bookmark.Preview{}, err
	}
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	preview := bookmark.Preview{
		ID:          newID("preview"),
		BookmarkID:  bookmarkID,
		Source:      input.Source,
		FilePath:    relativePath,
		ThumbPath:   relativePath,
		OriginalURL: strings.TrimSpace(input.OriginalURL),
		Mime:        mimeType,
		Size:        stat.Size(),
		CreatedAt:   time.Now().UTC(),
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO bookmark_previews
		(id, bookmark_id, source, file_path, thumb_path, original_url, mime, width, height, size, created_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(bookmark_id) DO UPDATE SET
			id = excluded.id,
			source = excluded.source,
			file_path = excluded.file_path,
			thumb_path = excluded.thumb_path,
			original_url = excluded.original_url,
			mime = excluded.mime,
			width = excluded.width,
			height = excluded.height,
			size = excluded.size,
			created_at = excluded.created_at`,
		preview.ID, preview.BookmarkID, preview.Source, preview.FilePath, preview.ThumbPath, preview.OriginalURL,
		preview.Mime, preview.Width, preview.Height, preview.Size, formatTime(preview.CreatedAt))
	if err != nil {
		return bookmark.Preview{}, err
	}
	preview.FilePath = previewAssetURL(preview.FilePath)
	preview.ThumbPath = previewAssetURL(preview.ThumbPath)
	return preview, nil
}

func (s *Store) FetchBookmarkPreview(ctx context.Context, bookmarkID string) (bookmark.Preview, error) {
	item, err := s.GetBookmark(ctx, bookmarkID)
	if err != nil {
		return bookmark.Preview{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.URL, nil)
	if err != nil {
		return bookmark.Preview{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return bookmark.Preview{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return bookmark.Preview{}, fmt.Errorf("preview page status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return bookmark.Preview{}, err
	}
	previewURL := extractOpenGraphImage(string(body))
	if previewURL == "" {
		return bookmark.Preview{}, errors.New("preview image not found")
	}
	resolvedURL, err := resolveURL(item.URL, previewURL)
	if err != nil {
		return bookmark.Preview{}, err
	}
	sourcePath, cleanup, err := s.downloadPreviewImage(ctx, resolvedURL)
	if err != nil {
		return bookmark.Preview{}, err
	}
	defer cleanup()
	return s.SaveBookmarkPreview(ctx, bookmarkID, bookmark.PreviewInput{Source: "auto", Path: sourcePath, OriginalURL: resolvedURL})
}

type bookmarkScanner interface {
	Scan(dest ...any) error
}

func scanBookmark(scanner bookmarkScanner) (bookmark.Bookmark, error) {
	var item bookmark.Bookmark
	var tagsJSON, keywordsJSON, aliasesJSON string
	var createdAt, updatedAt string
	var lastAnalyzed sql.NullString
	var preview bookmark.Preview
	var previewID, previewBookmarkID, previewSource, previewFilePath, previewThumbPath, previewOriginalURL, previewMime, previewCreatedAt sql.NullString
	var previewWidth, previewHeight, previewSize sql.NullInt64
	if err := scanner.Scan(&item.ID, &item.Title, &item.URL, &item.Description, &item.Folder, &item.CategoryID, &item.CategoryName,
		&tagsJSON, &keywordsJSON, &aliasesJSON, &item.Domain, &createdAt, &updatedAt, &lastAnalyzed,
		&previewID, &previewBookmarkID, &previewSource, &previewFilePath, &previewThumbPath, &previewOriginalURL, &previewMime,
		&previewWidth, &previewHeight, &previewSize, &previewCreatedAt); err != nil {
		return bookmark.Bookmark{}, err
	}
	if item.CategoryID == "" {
		item.CategoryID = bookmark.UncategorizedID
	}
	if item.CategoryName == "" {
		item.CategoryName = item.Folder
	}
	item.CategoryPath = splitCategoryPath(item.Folder)
	item.Tags = parseJSONList(tagsJSON)
	item.Keywords = parseJSONList(keywordsJSON)
	item.Aliases = parseJSONList(aliasesJSON)
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	if lastAnalyzed.Valid && lastAnalyzed.String != "" {
		parsed := parseTime(lastAnalyzed.String)
		item.LastAnalyzedAt = &parsed
	}
	if previewID.Valid {
		preview.ID = previewID.String
		preview.BookmarkID = previewBookmarkID.String
		preview.Source = previewSource.String
		preview.FilePath = previewFilePath.String
		preview.ThumbPath = previewThumbPath.String
		preview.OriginalURL = previewOriginalURL.String
		preview.Mime = previewMime.String
		preview.Width = int(previewWidth.Int64)
		preview.Height = int(previewHeight.Int64)
		preview.Size = previewSize.Int64
		preview.CreatedAt = parseTime(previewCreatedAt.String)
		item.Preview = &preview
	}
	return item, nil
}

func (s *Store) decorateBookmark(ctx context.Context, item *bookmark.Bookmark) error {
	if item == nil {
		return nil
	}
	path, err := s.categoryPath(ctx, item.CategoryID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		path = []string{bookmark.UncategorizedName}
		item.CategoryID = bookmark.UncategorizedID
		item.CategoryName = bookmark.UncategorizedName
	}
	if len(path) == 0 {
		path = []string{bookmark.UncategorizedName}
	}
	item.CategoryPath = path
	item.Folder = strings.Join(path, " / ")
	item.CategoryName = path[len(path)-1]
	if item.Preview != nil {
		item.Preview.FilePath = previewAssetURL(item.Preview.FilePath)
		item.Preview.ThumbPath = previewAssetURL(item.Preview.ThumbPath)
	}
	return nil
}

func bookmarkSelectSQL() string {
	return `SELECT b.id, b.title, b.url, b.description, b.folder, b.category_id, COALESCE(c.name, b.folder),
		b.tags, b.keywords, b.aliases, b.domain, b.created_at, b.updated_at, b.last_analyzed_at,
		p.id, p.bookmark_id, p.source, p.file_path, p.thumb_path, p.original_url, p.mime, p.width, p.height, p.size, p.created_at
		FROM bookmarks b
		LEFT JOIN categories c ON c.id = b.category_id
		LEFT JOIN bookmark_previews p ON p.bookmark_id = b.id`
}

// buildWhere 仅返回参数化 SQL 片段；排序字段由调用方白名单选择，避免拼接用户输入。
func (s *Store) buildWhere(ctx context.Context, req bookmark.SearchRequest) (string, []any, error) {
	clauses := []string{"1=1"}
	args := []any{}
	if req.Query != "" {
		like := "%" + escapeLike(strings.ToLower(req.Query)) + "%"
		clauses = append(clauses, `(lower(b.title) LIKE ? ESCAPE '\' OR lower(b.url) LIKE ? ESCAPE '\' OR lower(b.description) LIKE ? ESCAPE '\' OR lower(b.domain) LIKE ? ESCAPE '\' OR lower(b.tags) LIKE ? ESCAPE '\' OR lower(b.keywords) LIKE ? ESCAPE '\' OR lower(b.aliases) LIKE ? ESCAPE '\')`)
		for i := 0; i < 7; i++ {
			args = append(args, like)
		}
	}
	if req.Folder != "" {
		clauses = append(clauses, `b.folder = ?`)
		args = append(args, req.Folder)
	}
	if req.CategoryID != "" {
		categoryIDs := []string{req.CategoryID}
		if req.IncludeDescendants {
			descendants, err := s.descendantCategoryIDs(ctx, req.CategoryID)
			if err != nil {
				return "", nil, err
			}
			categoryIDs = append(categoryIDs, descendants...)
		}
		placeholders := make([]string, 0, len(categoryIDs))
		for _, id := range categoryIDs {
			placeholders = append(placeholders, "?")
			args = append(args, id)
		}
		clauses = append(clauses, `b.category_id IN (`+strings.Join(placeholders, ",")+`)`)
	}
	for _, tag := range req.Tags {
		clauses = append(clauses, `lower(b.tags) LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(strings.ToLower(tag))+"%")
	}
	return "WHERE " + strings.Join(clauses, " AND "), args, nil
}

func (s *Store) resolveInputCategory(ctx context.Context, input bookmark.BookmarkInput) (string, string, error) {
	return s.resolveInputCategoryWithExec(ctx, s.db, input)
}

func (s *Store) resolveInputCategoryTx(ctx context.Context, tx *sql.Tx, input bookmark.BookmarkInput) (string, string, error) {
	return s.resolveInputCategoryWithExec(ctx, tx, input)
}

type sqlExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *Store) resolveInputCategoryWithExec(ctx context.Context, exec sqlExecutor, input bookmark.BookmarkInput) (string, string, error) {
	if input.CategoryID != "" && input.CategoryID != bookmark.RootCategoryID {
		node, err := getCategoryWithExec(ctx, exec, input.CategoryID)
		if err != nil {
			return "", "", err
		}
		path, err := categoryPathWithExec(ctx, exec, node.ID)
		if err != nil {
			return "", "", err
		}
		return node.ID, strings.Join(path, " / "), nil
	}
	categoryID, err := s.ensureCategoryPathWithExec(ctx, exec, splitCategoryPath(input.Folder))
	if err != nil {
		return "", "", err
	}
	path, err := categoryPathWithExec(ctx, exec, categoryID)
	if err != nil {
		return "", "", err
	}
	return categoryID, strings.Join(path, " / "), nil
}

func (s *Store) ensureCategoryPath(ctx context.Context, path []string) (string, error) {
	return s.ensureCategoryPathWithExec(ctx, s.db, path)
}

func (s *Store) ensureCategoryPathWithExec(ctx context.Context, exec sqlExecutor, path []string) (string, error) {
	if len(path) == 0 {
		return bookmark.UncategorizedID, nil
	}
	parentID := ""
	categoryID := bookmark.UncategorizedID
	now := formatTime(time.Now().UTC())
	for _, name := range path {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if strings.EqualFold(name, bookmark.UncategorizedName) || strings.EqualFold(name, "Unsorted") {
			categoryID = bookmark.UncategorizedID
			parentID = categoryID
			continue
		}
		var existingID string
		err := exec.QueryRowContext(ctx, `SELECT id FROM categories WHERE name = ? AND COALESCE(parent_id, '') = ?`, name, parentID).Scan(&existingID)
		if errors.Is(err, sql.ErrNoRows) {
			existingID = newID("cat")
			if _, err := exec.ExecContext(ctx, `INSERT INTO categories(id, name, parent_id, sort_order, is_system, created_at, updated_at)
				VALUES(?, ?, ?, 0, 0, ?, ?)`, existingID, name, nullableParent(parentID), now, now); err != nil {
				return "", err
			}
		} else if err != nil {
			return "", err
		}
		categoryID = existingID
		parentID = existingID
	}
	return categoryID, nil
}

func (s *Store) descendantCategoryIDs(ctx context.Context, id string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `WITH RECURSIVE tree(id) AS (
		SELECT id FROM categories WHERE parent_id = ?
		UNION ALL
		SELECT c.id FROM categories c JOIN tree t ON c.parent_id = t.id
	) SELECT id FROM tree`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var categoryID string
		if err := rows.Scan(&categoryID); err != nil {
			return nil, err
		}
		ids = append(ids, categoryID)
	}
	return ids, rows.Err()
}

func (s *Store) getCategory(ctx context.Context, id string) (bookmark.CategoryNode, error) {
	return getCategoryWithExec(ctx, s.db, id)
}

func (s *Store) categoryNameExists(ctx context.Context, parentID, name, excludeID string) (bool, error) {
	var existingID string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM categories
		WHERE lower(name) = lower(?) AND COALESCE(parent_id, '') = ? AND id != ?
		LIMIT 1`, name, parentID, excludeID).Scan(&existingID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) categoryPath(ctx context.Context, id string) ([]string, error) {
	return categoryPathWithExec(ctx, s.db, id)
}

func (s *Store) categoryPathString(ctx context.Context, id string) (string, error) {
	path, err := s.categoryPath(ctx, id)
	if err != nil {
		return "", err
	}
	return strings.Join(path, " / "), nil
}

func (s *Store) syncBookmarkFoldersForCategory(ctx context.Context, id string) error {
	categoryIDs := []string{id}
	descendants, err := s.descendantCategoryIDs(ctx, id)
	if err != nil {
		return err
	}
	categoryIDs = append(categoryIDs, descendants...)
	for _, categoryID := range categoryIDs {
		path, err := s.categoryPathString(ctx, categoryID)
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE bookmarks SET folder = ? WHERE category_id = ?`, path, categoryID); err != nil {
			return err
		}
	}
	return nil
}

func rebalanceCategorySortOrderTx(ctx context.Context, tx *sql.Tx, movingID, parentID string, targetIndex int, movingIsSystem bool) error {
	rows, err := tx.QueryContext(ctx, `SELECT id, is_system FROM categories
		WHERE COALESCE(parent_id, '') = ? AND id != ?
		ORDER BY sort_order ASC, lower(name) ASC`, parentID, movingID)
	if err != nil {
		return err
	}
	type sibling struct {
		id       string
		isSystem bool
	}
	var systemSiblings []sibling
	var normalSiblings []sibling
	for rows.Next() {
		var item sibling
		var isSystem int
		if err := rows.Scan(&item.id, &isSystem); err != nil {
			_ = rows.Close()
			return err
		}
		item.isSystem = isSystem == 1
		if item.isSystem {
			systemSiblings = append(systemSiblings, item)
		} else {
			normalSiblings = append(normalSiblings, item)
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	// 普通分类的置顶只在普通同级内生效，避免把“未分类”等系统分类挤到后面。
	if !movingIsSystem {
		targetIndex -= len(systemSiblings)
	}
	if targetIndex < 0 {
		targetIndex = 0
	}
	if targetIndex > len(normalSiblings) {
		targetIndex = len(normalSiblings)
	}

	ordered := make([]sibling, 0, len(systemSiblings)+len(normalSiblings)+1)
	ordered = append(ordered, systemSiblings...)
	if movingIsSystem {
		ordered = append(ordered, sibling{id: movingID, isSystem: true})
		ordered = append(ordered, normalSiblings...)
	} else {
		ordered = append(ordered, normalSiblings[:targetIndex]...)
		ordered = append(ordered, sibling{id: movingID})
		ordered = append(ordered, normalSiblings[targetIndex:]...)
	}
	for index, item := range ordered {
		if _, err := tx.ExecContext(ctx, `UPDATE categories SET sort_order = ? WHERE id = ?`, index, item.id); err != nil {
			return err
		}
	}
	return nil
}

func getCategoryWithExec(ctx context.Context, exec sqlExecutor, id string) (bookmark.CategoryNode, error) {
	if id == bookmark.RootCategoryID {
		return bookmark.CategoryNode{ID: bookmark.RootCategoryID, Name: "全部分类", IsSystem: true}, nil
	}
	var node bookmark.CategoryNode
	var isSystem int
	var createdAt, updatedAt string
	err := exec.QueryRowContext(ctx, `SELECT id, name, COALESCE(parent_id, ''), sort_order, is_system, created_at, updated_at
		FROM categories WHERE id = ?`, id).Scan(&node.ID, &node.Name, &node.ParentID, &node.SortOrder, &isSystem, &createdAt, &updatedAt)
	if err != nil {
		return bookmark.CategoryNode{}, err
	}
	node.IsSystem = isSystem == 1
	node.CreatedAt = parseTime(createdAt)
	node.UpdatedAt = parseTime(updatedAt)
	node.Children = []bookmark.CategoryNode{}
	return node, nil
}

func categoryPathWithExec(ctx context.Context, exec sqlExecutor, id string) ([]string, error) {
	id = strings.TrimSpace(id)
	if id == "" || id == bookmark.RootCategoryID || id == bookmark.UncategorizedID {
		return []string{bookmark.UncategorizedName}, nil
	}
	parts := []string{}
	seen := map[string]struct{}{}
	currentID := id
	for currentID != "" && currentID != bookmark.RootCategoryID {
		if _, ok := seen[currentID]; ok {
			return nil, errors.New("category tree contains a cycle")
		}
		seen[currentID] = struct{}{}
		node, err := getCategoryWithExec(ctx, exec, currentID)
		if err != nil {
			return nil, err
		}
		parts = append([]string{node.Name}, parts...)
		currentID = normalizeParentID(node.ParentID)
	}
	if len(parts) == 0 {
		return []string{bookmark.UncategorizedName}, nil
	}
	return parts, nil
}

func (s *Store) nextCategorySortOrder(ctx context.Context, parentID string) int {
	var next sql.NullInt64
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order), -1) + 1 FROM categories WHERE COALESCE(parent_id, '') = ?`, parentID).Scan(&next)
	return int(next.Int64)
}

func splitCategoryPath(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "Unsorted") || strings.EqualFold(raw, bookmark.UncategorizedName) {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '/' || r == '\\' || r == '>' || r == '｜' || r == '|'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func normalizeParentID(parentID string) string {
	parentID = strings.TrimSpace(parentID)
	if parentID == "" || parentID == bookmark.RootCategoryID {
		return ""
	}
	return parentID
}

func nullableParent(parentID string) any {
	if parentID == "" || parentID == bookmark.RootCategoryID {
		return nil
	}
	return parentID
}

func previewAssetURL(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	if path == "" {
		return ""
	}
	path = strings.TrimPrefix(path, "/")
	if strings.HasPrefix(path, "previews/") {
		return "/" + path
	}
	return path
}

func (s *Store) insertFolder(ctx context.Context, folder string) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO folders(id, name, sort_order) VALUES(?, ?, 0)`, folderID(folder), folder)
	return err
}

func mustJSON(values []string) string {
	data, err := json.Marshal(bookmark.NormalizeList(values))
	if err != nil {
		return "[]"
	}
	return string(data)
}

func parseJSONList(raw string) []string {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}
	return bookmark.NormalizeList(values)
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UTC().UnixNano(), idSequence.Add(1))
}

func folderID(folder string) string {
	id := strings.ToLower(strings.TrimSpace(folder))
	id = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}, id)
	if id == "" {
		id = "unsorted"
	}
	return "folder_" + id
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(raw string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
}

func copyLocalFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func extractOpenGraphImage(html string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`),
		regexp.MustCompile(`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:image["']`),
		regexp.MustCompile(`(?is)<meta[^>]+name=["']twitter:image["'][^>]+content=["']([^"']+)["']`),
		regexp.MustCompile(`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+name=["']twitter:image["']`),
	}
	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(html); len(match) == 2 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func resolveURL(baseURL, rawURL string) (string, error) {
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	parsedRaw, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	return parsedBase.ResolveReference(parsedRaw).String(), nil
}

func (s *Store) downloadPreviewImage(ctx context.Context, imageURL string) (string, func(), error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", func() {}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", func() {}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", func() {}, fmt.Errorf("preview image status %d", resp.StatusCode)
	}
	ext := ".img"
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		if exts, _ := mime.ExtensionsByType(strings.Split(contentType, ";")[0]); len(exts) > 0 {
			ext = exts[0]
		}
	}
	tempFile, err := os.CreateTemp(s.dataDir, "preview-*"+ext)
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() {
		_ = os.Remove(tempFile.Name())
	}
	defer tempFile.Close()
	if _, err := io.Copy(tempFile, io.LimitReader(resp.Body, 8<<20)); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return tempFile.Name(), cleanup, nil
}
