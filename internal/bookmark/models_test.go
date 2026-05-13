package bookmark

import "testing"

func TestNormalizeURLRejectsDangerousSchemes(t *testing.T) {
	if _, _, err := NormalizeURL("javascript:alert(1)"); err == nil {
		t.Fatal("expected javascript url to be rejected")
	}
}

func TestNormalizeInputDefaults(t *testing.T) {
	input, domain, err := NormalizeInput(BookmarkInput{URL: "example.com", Tags: []string{"Dev", "dev", " AI "}})
	if err != nil {
		t.Fatal(err)
	}
	if domain != "example.com" {
		t.Fatalf("domain = %q", domain)
	}
	if input.Title != "example.com" || input.Folder != "Unsorted" {
		t.Fatalf("unexpected defaults: %#v", input)
	}
	if len(input.Tags) != 2 {
		t.Fatalf("expected normalized unique tags, got %#v", input.Tags)
	}
}
