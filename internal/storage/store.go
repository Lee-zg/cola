// storage 包负责本地 SQLite 持久化，是书签、分类、分析状态和设置的唯一写入点。
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cola/internal/bookmark"

	_ "modernc.org/sqlite"
)

type Store struct {
	db     *sql.DB
	dbPath string
}

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
	store := &Store{db: db, dbPath: dbPath}
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
			tags TEXT NOT NULL DEFAULT '[]',
			keywords TEXT NOT NULL DEFAULT '[]',
			aliases TEXT NOT NULL DEFAULT '[]',
			domain TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			last_analyzed_at TEXT
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
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO folders(id, name, sort_order) VALUES('folder_unsorted', 'Unsorted', 0)`)
	return err
}

// CreateBookmark 只创建单条用户输入的书签，URL 唯一冲突转成稳定的业务错误给前端展示。
func (s *Store) CreateBookmark(ctx context.Context, input bookmark.BookmarkInput) (bookmark.Bookmark, error) {
	normalized, domain, err := bookmark.NormalizeInput(input)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	now := time.Now().UTC()
	item := bookmark.Bookmark{
		ID:          newID("bm"),
		Title:       normalized.Title,
		URL:         normalized.URL,
		Description: normalized.Description,
		Folder:      normalized.Folder,
		Tags:        normalized.Tags,
		Keywords:    normalized.Keywords,
		Aliases:     normalized.Aliases,
		Domain:      domain,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.insertFolder(ctx, item.Folder); err != nil {
		return bookmark.Bookmark{}, err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO bookmarks
		(id, title, url, description, folder, tags, keywords, aliases, domain, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Title, item.URL, item.Description, item.Folder,
		mustJSON(item.Tags), mustJSON(item.Keywords), mustJSON(item.Aliases),
		item.Domain, formatTime(item.CreatedAt), formatTime(item.UpdatedAt)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return bookmark.Bookmark{}, errors.New("bookmark url already exists")
		}
		return bookmark.Bookmark{}, err
	}
	return item, nil
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
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO folders(id, name, sort_order) VALUES(?, ?, 0)`, folderID(normalized.Folder), normalized.Folder); err != nil {
			return result, err
		}
		res, execErr := tx.ExecContext(ctx, `INSERT OR IGNORE INTO bookmarks
			(id, title, url, description, folder, tags, keywords, aliases, domain, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			newID("bm"), normalized.Title, normalized.URL, normalized.Description, normalized.Folder,
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
	if err := s.insertFolder(ctx, normalized.Folder); err != nil {
		return bookmark.Bookmark{}, err
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `UPDATE bookmarks SET
		title = ?, url = ?, description = ?, folder = ?, tags = ?, keywords = ?, aliases = ?,
		domain = ?, updated_at = ?
		WHERE id = ?`,
		normalized.Title, normalized.URL, normalized.Description, normalized.Folder,
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
	row := s.db.QueryRowContext(ctx, `SELECT id, title, url, description, folder, tags, keywords, aliases,
		domain, created_at, updated_at, last_analyzed_at FROM bookmarks WHERE id = ?`, id)
	return scanBookmark(row)
}

func (s *Store) ListBookmarks(ctx context.Context, req bookmark.SearchRequest) (bookmark.SearchResult, error) {
	req = bookmark.NormalizeSearch(req)
	where, args := buildWhere(req)
	orderBy := "updated_at DESC"
	switch req.Sort {
	case "title":
		orderBy = "lower(title) ASC"
	case "created":
		orderBy = "created_at DESC"
	}

	countQuery := `SELECT count(*) FROM bookmarks ` + where
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return bookmark.SearchResult{}, err
	}

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, req.Limit, req.Offset)
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, url, description, folder, tags, keywords, aliases,
		domain, created_at, updated_at, last_analyzed_at FROM bookmarks `+where+` ORDER BY `+orderBy+` LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return bookmark.SearchResult{}, err
	}
	defer rows.Close()
	items := []bookmark.Bookmark{}
	for rows.Next() {
		item, err := scanBookmark(rows)
		if err != nil {
			return bookmark.SearchResult{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return bookmark.SearchResult{}, err
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
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT folder FROM bookmarks ORDER BY lower(folder)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var folders []string
	for rows.Next() {
		var folder string
		if err := rows.Scan(&folder); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}
	if len(folders) == 0 {
		folders = append(folders, "Unsorted")
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

type bookmarkScanner interface {
	Scan(dest ...any) error
}

func scanBookmark(scanner bookmarkScanner) (bookmark.Bookmark, error) {
	var item bookmark.Bookmark
	var tagsJSON, keywordsJSON, aliasesJSON string
	var createdAt, updatedAt string
	var lastAnalyzed sql.NullString
	if err := scanner.Scan(&item.ID, &item.Title, &item.URL, &item.Description, &item.Folder,
		&tagsJSON, &keywordsJSON, &aliasesJSON, &item.Domain, &createdAt, &updatedAt, &lastAnalyzed); err != nil {
		return bookmark.Bookmark{}, err
	}
	item.Tags = parseJSONList(tagsJSON)
	item.Keywords = parseJSONList(keywordsJSON)
	item.Aliases = parseJSONList(aliasesJSON)
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	if lastAnalyzed.Valid && lastAnalyzed.String != "" {
		parsed := parseTime(lastAnalyzed.String)
		item.LastAnalyzedAt = &parsed
	}
	return item, nil
}

// buildWhere 仅返回参数化 SQL 片段；排序字段由调用方白名单选择，避免拼接用户输入。
func buildWhere(req bookmark.SearchRequest) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}
	if req.Query != "" {
		like := "%" + escapeLike(strings.ToLower(req.Query)) + "%"
		clauses = append(clauses, `(lower(title) LIKE ? ESCAPE '\' OR lower(url) LIKE ? ESCAPE '\' OR lower(description) LIKE ? ESCAPE '\' OR lower(domain) LIKE ? ESCAPE '\' OR lower(tags) LIKE ? ESCAPE '\' OR lower(keywords) LIKE ? ESCAPE '\' OR lower(aliases) LIKE ? ESCAPE '\')`)
		for i := 0; i < 7; i++ {
			args = append(args, like)
		}
	}
	if req.Folder != "" {
		clauses = append(clauses, `folder = ?`)
		args = append(args, req.Folder)
	}
	for _, tag := range req.Tags {
		clauses = append(clauses, `lower(tags) LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(strings.ToLower(tag))+"%")
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
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
	return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
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
