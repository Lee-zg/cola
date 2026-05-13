package importer

import (
	"html"
	"os"
	"regexp"
	"strings"

	"cola/internal/bookmark"
)

var (
	linkRe   = regexp.MustCompile(`(?i)<a\s+([^>]*href\s*=\s*["'][^"']+["'][^>]*)>(.*?)</a>`)
	hrefRe   = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)
	folderRe = regexp.MustCompile(`(?i)<h3[^>]*>(.*?)</h3>`)
	tagRe    = regexp.MustCompile(`(?s)<[^>]+>`)
)

func ParseNetscapeHTML(data []byte) []bookmark.BookmarkInput {
	lines := strings.Split(string(data), "\n")
	currentFolder := "Unsorted"
	var items []bookmark.BookmarkInput
	for _, line := range lines {
		if match := folderRe.FindStringSubmatch(line); len(match) == 2 {
			currentFolder = cleanHTMLText(match[1])
			if currentFolder == "" {
				currentFolder = "Unsorted"
			}
			continue
		}
		match := linkRe.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		href := hrefRe.FindStringSubmatch(match[1])
		if len(href) != 2 {
			continue
		}
		title := cleanHTMLText(match[2])
		if title == "" {
			title = href[1]
		}
		items = append(items, bookmark.BookmarkInput{
			Title:  title,
			URL:    html.UnescapeString(href[1]),
			Folder: currentFolder,
		})
	}
	return items
}

func ParseNetscapeHTMLFile(path string) ([]bookmark.BookmarkInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseNetscapeHTML(data), nil
}

func cleanHTMLText(raw string) string {
	raw = tagRe.ReplaceAllString(raw, "")
	raw = html.UnescapeString(raw)
	return strings.TrimSpace(raw)
}
