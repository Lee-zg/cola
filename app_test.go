package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cola/internal/bookmark"
)

func TestAppIntegration(t *testing.T) {
	app := NewAppForTest(t.TempDir())
	app.startup(context.Background())
	defer app.shutdown(context.Background())

	item, err := app.CreateBookmark(bookmark.BookmarkInput{
		Title: "Vue Docs",
		URL:   "https://vuejs.org/guide/",
		Tags:  []string{"Development"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.AnalyzeBookmark(item.ID); err != nil {
		t.Fatal(err)
	}
	search, err := app.ListBookmarks(bookmark.SearchRequest{Query: "vue"})
	if err != nil {
		t.Fatal(err)
	}
	if search.Total != 1 {
		t.Fatalf("expected 1 search result, got %d", search.Total)
	}
	exportPath := filepath.Join(t.TempDir(), "bookmarks.html")
	if _, err := app.ExportBookmarks(bookmark.ExportRequest{Path: exportPath, TemplateID: "classic"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Vue Docs") {
		t.Fatal("exported html missing bookmark")
	}
	status, err := app.StartLocalServer()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(status.URL + "/api/bookmarks?q=vue")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected web status: %s", resp.Status)
	}
}
