// 文件说明：internal/storage/store_test.go，负责应用后端或核心业务实现。
package storage

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"cola/internal/bookmark"
)

const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII="

func TestStoreCRUDSearchAndAnalysis(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{
		Title:  "Go Documentation",
		URL:    "https://go.dev/doc/",
		Folder: "Development",
		Tags:   []string{"Go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := store.ListBookmarks(ctx, bookmark.SearchRequest{Query: "go", Tags: []string{"Go"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || result.Items[0].ID != item.ID {
		t.Fatalf("unexpected search result: %#v", result)
	}
	updated, err := store.UpdateBookmark(ctx, item.ID, bookmark.BookmarkInput{
		Title:  "Go Packages",
		URL:    "https://pkg.go.dev/",
		Folder: "Development",
		Tags:   []string{"Docs"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.URL != "https://pkg.go.dev/" {
		t.Fatalf("unexpected updated item: %#v", updated)
	}
	analyzed, err := store.ApplyAnalysis(ctx, item.ID, bookmark.AnalysisResult{Tags: []string{"Development"}, Keywords: []string{"golang"}, Engine: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(analyzed.Keywords) == 0 {
		t.Fatal("expected keywords after analysis")
	}
	if err := store.DeleteBookmark(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
}

func TestStoreCanDeleteUncategorizedBookmark(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{
		Title: "No Category",
		URL:   "https://uncategorized-delete.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.CategoryID != bookmark.UncategorizedID {
		t.Fatalf("expected uncategorized bookmark, got %#v", item)
	}
	page, err := store.ListBookmarks(ctx, bookmark.SearchRequest{CategoryID: bookmark.UncategorizedID, IncludeDescendants: true, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0].ID != item.ID {
		t.Fatalf("expected bookmark under uncategorized filter, got %#v", page)
	}
	if err := store.DeleteBookmark(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	page, err = store.ListBookmarks(ctx, bookmark.SearchRequest{CategoryID: bookmark.UncategorizedID, IncludeDescendants: true, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 {
		t.Fatalf("expected uncategorized bookmark deleted, got %#v", page)
	}
}

func TestStoreUpsertSkipsDuplicates(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	result, err := store.UpsertBookmarks(ctx, []bookmark.BookmarkInput{
		{Title: "One", URL: "https://example.com"},
		{Title: "Two", URL: "https://example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 || result.Skipped != 1 {
		t.Fatalf("unexpected import result: %#v", result)
	}
}

func TestStoreUpsertUpdatesDuplicatesWhenAllowed(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	initial, err := store.UpsertBookmarksWithOptions(ctx, []bookmark.BookmarkInput{
		{Title: "Old", URL: "https://example.com/update", Folder: "Old"},
	}, bookmark.ImportOptions{SkipDuplicates: true, KeepFolders: true})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpsertBookmarksWithOptions(ctx, []bookmark.BookmarkInput{
		{Title: "New", URL: "https://example.com/update", Folder: "New / Folder"},
	}, bookmark.ImportOptions{SkipDuplicates: false, KeepFolders: true})
	if err != nil {
		t.Fatal(err)
	}
	if initial.Imported != 1 || updated.Updated != 1 || updated.Skipped != 0 {
		t.Fatalf("unexpected upsert results: initial=%#v updated=%#v", initial, updated)
	}
	page, err := store.ListBookmarks(ctx, bookmark.SearchRequest{Query: "New", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || strings.Join(page.Items[0].CategoryPath, " / ") != "New / Folder" {
		t.Fatalf("expected updated bookmark category, got %#v", page)
	}
}

func TestStoreUpsertCanDropImportedFolders(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	result, err := store.UpsertBookmarksWithOptions(ctx, []bookmark.BookmarkInput{
		{Title: "Flat", URL: "https://flat.example", Folder: "Nested / Folder"},
	}, bookmark.ImportOptions{SkipDuplicates: true, KeepFolders: false})
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected one imported bookmark, got %#v", result)
	}
	page, err := store.ListBookmarks(ctx, bookmark.SearchRequest{Query: "Flat", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0].CategoryID != bookmark.UncategorizedID {
		t.Fatalf("expected uncategorized import, got %#v", page)
	}
}

func TestStoreUpsertCreatesCategoryPath(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	result, err := store.UpsertBookmarks(ctx, []bookmark.BookmarkInput{
		{Title: "Imported", URL: "https://import.example", Folder: "Imported / Dev"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected one imported bookmark, got %#v", result)
	}
	page, err := store.ListBookmarks(ctx, bookmark.SearchRequest{Query: "Imported", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 {
		t.Fatalf("expected imported bookmark in list, got %#v", page)
	}
	if got := strings.Join(page.Items[0].CategoryPath, " / "); got != "Imported / Dev" {
		t.Fatalf("expected imported category path, got %q", got)
	}
}

func TestStoreUpsertSkipsInvalidRowsAndKeepsValidImports(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	result, err := store.UpsertBookmarks(ctx, []bookmark.BookmarkInput{
		{Title: "Good", URL: "https://good.example", Folder: "HTML / Good"},
		{Title: "Bad", URL: "javascript:alert(1)", Folder: "HTML / Bad"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 || result.Skipped != 1 {
		t.Fatalf("expected valid import and invalid skip, got %#v", result)
	}
}

func TestOpenMigratesLegacyBookmarksWithoutLosingData(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE bookmarks (
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
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO bookmarks
		(id, title, url, description, folder, tags, keywords, aliases, domain, created_at, updated_at)
		VALUES('old_1', 'Old Go', 'https://go.dev', '', 'Dev / Go', '[]', '[]', '[]', 'go.dev', '2026-05-14T00:00:00Z', '2026-05-14T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	store, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	result, err := store.ListBookmarks(ctx, bookmark.SearchRequest{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || result.Items[0].Title != "Old Go" {
		t.Fatalf("expected legacy bookmark after migration, got %#v", result)
	}
	if got := strings.Join(result.Items[0].CategoryPath, " / "); got != "Dev / Go" {
		t.Fatalf("expected migrated category path, got %q", got)
	}
}

func TestOpenRepairsOrphanCategoryIDFromFolder(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Repair", URL: "https://repair.example", Folder: "Ops / Runbook"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.ExecContext(ctx, `UPDATE bookmarks SET category_id = 'missing_category', folder = 'Ops / Runbook' WHERE id = ?`, item.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(ctx, store.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	repaired, err := reopened.GetBookmark(ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if repaired.CategoryID == "missing_category" {
		t.Fatalf("expected orphan category id to be repaired, got %#v", repaired)
	}
	if got := strings.Join(repaired.CategoryPath, " / "); got != "Ops / Runbook" {
		t.Fatalf("expected repaired folder path, got %q", got)
	}
}

func TestCategoryTreeDeleteMovesBookmarksToParent(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	parent, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "开发"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "Go", ParentID: parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Go", URL: "https://go.dev", CategoryID: child.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteCategory(ctx, child.ID); err != nil {
		t.Fatal(err)
	}
	updated, err := store.GetBookmark(ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.CategoryID != parent.ID {
		t.Fatalf("expected bookmark moved to parent category, got %#v", updated)
	}
}

func TestCategoryTreeDeleteCanRemoveBookmarks(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	parent, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "资料"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "临时", ParentID: parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Temp", URL: "https://temp.example", CategoryID: child.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteCategoryWithOptions(ctx, child.ID, bookmark.DeleteCategoryInput{DeleteBookmarks: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetBookmark(ctx, item.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected bookmark to be deleted with category, got %v", err)
	}
}

func TestCategoryTreeCountsPathsAndDescendantSearch(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	parent, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "开发"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "Go", ParentID: parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Go", URL: "https://go.dev", CategoryID: child.ID}); err != nil {
		t.Fatal(err)
	}

	tree, err := store.ListCategories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if tree[0].Count != 1 || tree[0].Children[1].Count != 1 || tree[0].Children[1].Children[0].Count != 1 {
		t.Fatalf("expected recursive category counts, got %#v", tree)
	}

	result, err := store.ListBookmarks(ctx, bookmark.SearchRequest{CategoryID: parent.ID, IncludeDescendants: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Fatalf("expected descendant search to include child bookmark, got %#v", result)
	}
	if got := strings.Join(result.Items[0].CategoryPath, " / "); got != "开发 / Go" {
		t.Fatalf("expected full category path, got %q", got)
	}
}

func TestMoveCategoryReordersOnlySiblings(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	first, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "B"})
	if err != nil {
		t.Fatal(err)
	}
	third, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "C"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.MoveCategory(ctx, third.ID, bookmark.MoveCategoryInput{ParentID: bookmark.RootCategoryID, SortOrder: 1}); err != nil {
		t.Fatal(err)
	}
	tree, err := store.ListCategories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, child := range tree[0].Children {
		if !child.IsSystem {
			got = append(got, child.Name)
		}
	}
	want := []string{third.Name, first.Name, second.Name}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected reordered siblings %v, got %v", want, got)
	}
}

func TestCategoryPinOrdersNewestFirstAndUnpinKeepsNormalOrder(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	first, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "B"})
	if err != nil {
		t.Fatal(err)
	}
	third, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "C"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetCategoryPinned(ctx, first.ID, bookmark.CategoryPinnedInput{Pinned: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetCategoryPinned(ctx, third.ID, bookmark.CategoryPinnedInput{Pinned: true}); err != nil {
		t.Fatal(err)
	}
	tree, err := store.ListCategories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	gotPinned := []string{}
	for _, child := range tree[0].Children {
		if child.IsPinned {
			gotPinned = append(gotPinned, child.Name)
		}
	}
	if strings.Join(gotPinned, ",") != "C,A" {
		t.Fatalf("expected newest pinned first, got %v", gotPinned)
	}
	if _, err := store.SetCategoryPinned(ctx, third.ID, bookmark.CategoryPinnedInput{Pinned: false}); err != nil {
		t.Fatal(err)
	}
	tree, err = store.ListCategories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, child := range tree[0].Children {
		if !child.IsSystem {
			got = append(got, child.Name)
		}
	}
	if strings.Join(got, ",") != strings.Join([]string{first.Name, second.Name, third.Name}, ",") {
		t.Fatalf("expected unpinned category to return to normal order, got %v", got)
	}
}

func TestBatchDeleteCategoriesMovesBookmarksToEachParent(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	dev, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "开发"})
	if err != nil {
		t.Fatal(err)
	}
	goCat, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "Go", ParentID: dev.ID})
	if err != nil {
		t.Fatal(err)
	}
	ops, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "运维"})
	if err != nil {
		t.Fatal(err)
	}
	runbook, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "手册", ParentID: ops.ID})
	if err != nil {
		t.Fatal(err)
	}
	goItem, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Go", URL: "https://go.dev/batch", CategoryID: goCat.ID})
	if err != nil {
		t.Fatal(err)
	}
	opsItem, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Ops", URL: "https://ops.example/batch", CategoryID: runbook.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.BatchDeleteCategories(ctx, bookmark.BatchCategoryDeleteInput{IDs: []string{goCat.ID, runbook.ID}}); err != nil {
		t.Fatal(err)
	}
	updatedGo, err := store.GetBookmark(ctx, goItem.ID)
	if err != nil {
		t.Fatal(err)
	}
	updatedOps, err := store.GetBookmark(ctx, opsItem.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedGo.CategoryID != dev.ID || updatedOps.CategoryID != ops.ID {
		t.Fatalf("expected bookmarks moved to each parent, got %#v and %#v", updatedGo, updatedOps)
	}
}

func TestBatchReorderRequiresSameParentAndUnpinned(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	first, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "B"})
	if err != nil {
		t.Fatal(err)
	}
	third, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "C"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "Child", ParentID: first.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.BatchReorderCategories(ctx, bookmark.BatchCategoryReorderInput{IDs: []string{second.ID, third.ID}, Direction: "up"}); err != nil {
		t.Fatal(err)
	}
	tree, err := store.ListCategories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, child := range tree[0].Children {
		if !child.IsSystem {
			got = append(got, child.Name)
		}
	}
	if strings.Join(got, ",") != "B,C,A" {
		t.Fatalf("expected selected block moved up, got %v", got)
	}
	if err := store.BatchReorderCategories(ctx, bookmark.BatchCategoryReorderInput{IDs: []string{first.ID, child.ID}, Direction: "down"}); err == nil {
		t.Fatal("expected cross-parent batch reorder to fail")
	}
	if _, err := store.SetCategoryPinned(ctx, second.ID, bookmark.CategoryPinnedInput{Pinned: true}); err != nil {
		t.Fatal(err)
	}
	if err := store.BatchReorderCategories(ctx, bookmark.BatchCategoryReorderInput{IDs: []string{second.ID}, Direction: "down"}); err == nil {
		t.Fatal("expected pinned category reorder to fail")
	}
}

func TestBatchReorderKeepsPinnedSortSlotForUnpin(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	first, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "B"})
	if err != nil {
		t.Fatal(err)
	}
	third, err := store.CreateCategory(ctx, bookmark.CategoryInput{Name: "C"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetCategoryPinned(ctx, pinned.ID, bookmark.CategoryPinnedInput{Pinned: true}); err != nil {
		t.Fatal(err)
	}
	if err := store.BatchReorderCategories(ctx, bookmark.BatchCategoryReorderInput{IDs: []string{third.ID}, Direction: "up"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetCategoryPinned(ctx, pinned.ID, bookmark.CategoryPinnedInput{Pinned: false}); err != nil {
		t.Fatal(err)
	}
	tree, err := store.ListCategories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, child := range tree[0].Children {
		if !child.IsSystem {
			got = append(got, child.Name)
		}
	}
	want := []string{third.Name, pinned.Name, first.Name}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected unpinned category to keep reserved sort slot, got %v", got)
	}
}

func TestSystemCategoryCannotBePinnedOrBatchDeleted(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if _, err := store.SetCategoryPinned(ctx, bookmark.UncategorizedID, bookmark.CategoryPinnedInput{Pinned: true}); err == nil {
		t.Fatal("expected system category pin to fail")
	}
	if err := store.BatchDeleteCategories(ctx, bookmark.BatchCategoryDeleteInput{IDs: []string{bookmark.UncategorizedID}}); err == nil {
		t.Fatal("expected system category batch delete to fail")
	}
}

func TestFetchBookmarkPreviewFromOpenGraph(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/image.png" {
			w.Header().Set("Content-Type", "image/png")
			data, _ := base64.StdEncoding.DecodeString(tinyPNGBase64)
			_, _ = w.Write(data)
			return
		}
		_, _ = w.Write([]byte(`<html><head><meta property="og:image" content="/image.png"></head></html>`))
	}))
	defer server.Close()

	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Preview", URL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := store.FetchBookmarkPreview(ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Source != "auto" || !strings.HasPrefix(preview.FilePath, "/previews/original/") {
		t.Fatalf("unexpected preview result: %#v", preview)
	}
}

func TestThumbnailCustomUploadAndAutoModeSwitch(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Thumb", URL: "https://thumb.example"})
	if err != nil {
		t.Fatal(err)
	}
	thumbnail, err := store.SaveCustomThumbnail(ctx, item.ID, bookmark.ThumbnailUploadInput{
		FileName: "thumb.png",
		Mime:     "image/png",
		Data:     tinyPNGBase64,
	})
	if err != nil {
		t.Fatal(err)
	}
	if thumbnail.UseAuto || thumbnail.DisplaySource != "upload" || !strings.HasPrefix(thumbnail.DisplayPath, "/previews/") {
		t.Fatalf("expected custom thumbnail display, got %#v", thumbnail)
	}
	thumbnail, err = store.SetThumbnailAutoMode(ctx, item.ID, bookmark.ThumbnailModeInput{UseAuto: true})
	if err != nil {
		t.Fatal(err)
	}
	if !thumbnail.UseAuto || thumbnail.DisplayPath != "" {
		t.Fatalf("expected auto mode without cached image, got %#v", thumbnail)
	}
	thumbnail, err = store.SetThumbnailAutoMode(ctx, item.ID, bookmark.ThumbnailModeInput{UseAuto: false})
	if err != nil {
		t.Fatal(err)
	}
	if thumbnail.DisplaySource != "upload" {
		t.Fatalf("expected custom image after switching back, got %#v", thumbnail)
	}
}

func TestCustomThumbnailURLRejectsNonImage(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("not an image"))
	}))
	defer server.Close()

	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Bad", URL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveCustomThumbnailURL(ctx, item.ID, bookmark.ThumbnailURLInput{URL: server.URL}); err == nil {
		t.Fatal("expected non-image thumbnail url to be rejected")
	}
}

func TestCustomThumbnailURLCachesImage(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		data, _ := base64.StdEncoding.DecodeString(tinyPNGBase64)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Remote", URL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	thumbnail, err := store.SaveCustomThumbnailURL(ctx, item.ID, bookmark.ThumbnailURLInput{URL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if thumbnail.UseAuto || thumbnail.DisplaySource != "remote" || !strings.HasPrefix(thumbnail.DisplayPath, "/previews/") {
		t.Fatalf("expected remote thumbnail display, got %#v", thumbnail)
	}
}

func TestCustomThumbnailUploadRejectsInvalidImage(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.CreateBookmark(ctx, bookmark.BookmarkInput{Title: "Invalid", URL: "https://invalid-thumb.example"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveCustomThumbnail(ctx, item.ID, bookmark.ThumbnailUploadInput{
		FileName: "not-image.png",
		Mime:     "image/png",
		Data:     base64.StdEncoding.EncodeToString([]byte("not an image")),
	}); err == nil {
		t.Fatal("expected invalid uploaded thumbnail to be rejected")
	}
}

func TestOpenMigratesLegacyPreviewToAutoThumbnail(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "legacy-preview.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE bookmarks (
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
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE bookmark_previews (
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
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO bookmarks
		(id, title, url, description, folder, category_id, tags, keywords, aliases, domain, created_at, updated_at)
		VALUES('bm_legacy', 'Legacy', 'https://legacy.example', '', 'Unsorted', '', '[]', '[]', '[]', 'legacy.example', '2026-05-14T00:00:00Z', '2026-05-14T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO bookmark_previews
		(id, bookmark_id, source, file_path, thumb_path, original_url, mime, width, height, size, created_at)
		VALUES('preview_legacy', 'bm_legacy', 'auto', 'previews/original/old.png', 'previews/original/old.png', 'https://legacy.example/old.png', 'image/png', 12, 8, 4, '2026-05-14T00:00:00Z')`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	store, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	item, err := store.GetBookmark(ctx, "bm_legacy")
	if err != nil {
		t.Fatal(err)
	}
	if item.Thumbnail == nil || !item.Thumbnail.UseAuto || item.Thumbnail.AutoThumbPath != "/previews/original/old.png" {
		t.Fatalf("expected legacy preview migrated to auto thumbnail, got %#v", item.Thumbnail)
	}
}
