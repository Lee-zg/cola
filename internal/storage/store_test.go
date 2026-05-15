// 文件说明：internal/storage/store_test.go，负责应用后端或核心业务实现。
package storage

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"cola/internal/bookmark"
)

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

func TestFetchBookmarkPreviewFromOpenGraph(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/image.png" {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47})
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
