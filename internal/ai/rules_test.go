package ai

import (
	"context"
	"testing"

	"cola/internal/bookmark"
)

func TestRuleAnalyzerTagsDevelopment(t *testing.T) {
	result, err := (RuleAnalyzer{}).Analyze(context.Background(), bookmark.Bookmark{
		Title:  "GitHub Developer Docs",
		URL:    "https://github.com/docs",
		Domain: "github.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tags) == 0 || result.Engine != "rules" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
