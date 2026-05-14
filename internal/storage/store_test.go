// 文件说明：internal/storage/store_test.go，负责应用后端或核心业务实现。
package storage

import (
	"context"
	"path/filepath"
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
