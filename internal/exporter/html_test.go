package exporter

import (
	"strings"
	"testing"

	"cola/internal/bookmark"
)

func TestRenderCatalogEscapesPayload(t *testing.T) {
	html, err := RenderCatalogHTML(BuildCatalog("Test", []bookmark.Bookmark{{
		Title:  `<script>alert(1)</script>`,
		URL:    "https://example.com",
		Folder: "Test",
	}}), "classic")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, `<script>alert(1)</script>`) {
		t.Fatal("expected title to be escaped inside JSON payload")
	}
	if !strings.Contains(html, "Bookmark") && !strings.Contains(html, "catalog") {
		t.Fatal("expected rendered catalog html")
	}
}
