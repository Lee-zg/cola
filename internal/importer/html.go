// importer 的 HTML 解析器面向浏览器导出的 Netscape bookmark 文件，只提取标题、URL 和当前文件夹。
package importer

import (
	"html"
	"os"
	"regexp"
	"slices"
	"strings"

	"cola/internal/bookmark"
)

var (
	tokenRe  = regexp.MustCompile(`(?is)</dl\s*>|<h3[^>]*>.*?</h3>|<a\s+[^>]*href\s*=\s*["'][^"']+["'][^>]*>.*?</a>`)
	linkRe   = regexp.MustCompile(`(?is)<a\s+([^>]*href\s*=\s*["'][^"']+["'][^>]*)>(.*?)</a>`)
	hrefRe   = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)
	folderRe = regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`)
	tagRe    = regexp.MustCompile(`(?s)<[^>]+>`)
)

// ParseNetscapeHTML 按标签顺序解析常见导出格式；它不是通用 HTML 清洗器，后续字段校验交给 bookmark.NormalizeInput。
func ParseNetscapeHTML(data []byte) []bookmark.BookmarkInput {
	folderStack := []string{bookmark.UncategorizedName}
	var items []bookmark.BookmarkInput
	for _, token := range tokenRe.FindAllString(string(data), -1) {
		lowerToken := strings.ToLower(token)
		if strings.HasPrefix(lowerToken, "</dl") {
			if len(folderStack) > 1 {
				folderStack = folderStack[:len(folderStack)-1]
			}
			continue
		}
		if match := folderRe.FindStringSubmatch(token); len(match) == 2 {
			folderName := cleanHTMLText(match[1])
			if folderName == "" {
				folderName = bookmark.UncategorizedName
			}
			folderStack = append(folderStack, folderName)
			continue
		}
		match := linkRe.FindStringSubmatch(token)
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
			Folder: strings.Join(cleanHTMLPath(folderStack), " / "),
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

func cleanHTMLPath(path []string) []string {
	copied := slices.Clone(path)
	out := make([]string, 0, len(copied))
	for _, part := range copied {
		part = strings.TrimSpace(part)
		if part != "" && part != bookmark.UncategorizedName {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return []string{bookmark.UncategorizedName}
	}
	return out
}
