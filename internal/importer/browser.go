// browser importer 负责从本机浏览器默认数据文件读取书签，不上传或访问网络。
package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"cola/internal/bookmark"

	_ "modernc.org/sqlite"
)

type BrowserImporter struct{}

// Import 支持显式路径优先；未传路径时只在当前系统约定目录中探测浏览器数据文件。
func (BrowserImporter) Import(ctx context.Context, sourceType, explicitPath string) ([]bookmark.BookmarkInput, error) {
	sourceType = strings.ToLower(strings.TrimSpace(sourceType))
	switch sourceType {
	case "html", "netscape":
		if explicitPath == "" {
			return nil, errors.New("html import requires a file path")
		}
		return ParseNetscapeHTMLFile(explicitPath)
	case "chrome", "edge":
		path := explicitPath
		if path == "" {
			candidates := browserBookmarkCandidates(sourceType)
			for _, candidate := range candidates {
				if fileExists(candidate) {
					path = candidate
					break
				}
			}
		}
		if path == "" {
			return nil, fmt.Errorf("%s bookmarks file not found", sourceType)
		}
		return parseChromiumBookmarks(path)
	case "firefox":
		path := explicitPath
		if path == "" {
			candidates := firefoxPlacesCandidates()
			for _, candidate := range candidates {
				if fileExists(candidate) {
					path = candidate
					break
				}
			}
		}
		if path == "" {
			return nil, errors.New("firefox places.sqlite not found")
		}
		return parseFirefoxPlaces(ctx, path)
	default:
		return nil, fmt.Errorf("unsupported import source %q", sourceType)
	}
}

type chromiumBookmarkFile struct {
	Roots map[string]chromiumNode `json:"roots"`
}

type chromiumNode struct {
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	URL      string         `json:"url"`
	Children []chromiumNode `json:"children"`
}

func parseChromiumBookmarks(path string) ([]bookmark.BookmarkInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var parsed chromiumBookmarkFile
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	var items []bookmark.BookmarkInput
	for name, root := range parsed.Roots {
		if name == "sync_transaction_version" {
			continue
		}
		walkChromium(root, humanFolderName(name), &items)
	}
	return items, nil
}

// walkChromium 将 Chromium 的树状目录压平成书签列表，保留叶子节点所在的最近文件夹名。
func walkChromium(node chromiumNode, folder string, items *[]bookmark.BookmarkInput) {
	walkChromiumPath(node, []string{folder}, items)
}

func walkChromiumPath(node chromiumNode, path []string, items *[]bookmark.BookmarkInput) {
	switch node.Type {
	case "url":
		*items = append(*items, bookmark.BookmarkInput{
			Title:  node.Name,
			URL:    node.URL,
			Folder: strings.Join(cleanPath(path), " / "),
		})
	case "folder":
		nextPath := slices.Clone(path)
		if node.Name != "" {
			nextPath = append(nextPath, node.Name)
		}
		for _, child := range node.Children {
			walkChromiumPath(child, nextPath, items)
		}
	default:
		for _, child := range node.Children {
			walkChromiumPath(child, path, items)
		}
	}
}

// parseFirefoxPlaces 以只读 immutable 模式打开 Firefox places.sqlite，避免干扰浏览器正在使用的数据库。
func parseFirefoxPlaces(ctx context.Context, path string) ([]bookmark.BookmarkInput, error) {
	dsn := "file:" + filepath.ToSlash(path) + "?mode=ro&immutable=1"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.QueryContext(ctx, `SELECT b.title, p.url
		FROM moz_bookmarks b
		JOIN moz_places p ON p.id = b.fk
		WHERE b.type = 1 AND p.url IS NOT NULL
		ORDER BY b.dateAdded DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []bookmark.BookmarkInput
	for rows.Next() {
		var title, rawURL string
		if err := rows.Scan(&title, &rawURL); err != nil {
			return nil, err
		}
		items = append(items, bookmark.BookmarkInput{
			Title:  title,
			URL:    rawURL,
			Folder: "Firefox",
		})
	}
	return items, rows.Err()
}

func browserBookmarkCandidates(browser string) []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	local := os.Getenv("LOCALAPPDATA")
	switch browser {
	case "chrome":
		return []string{
			filepath.Join(local, "Google", "Chrome", "User Data", "Default", "Bookmarks"),
			filepath.Join(local, "Google", "Chrome", "User Data", "Profile 1", "Bookmarks"),
		}
	case "edge":
		return []string{
			filepath.Join(local, "Microsoft", "Edge", "User Data", "Default", "Bookmarks"),
			filepath.Join(local, "Microsoft", "Edge", "User Data", "Profile 1", "Bookmarks"),
		}
	default:
		return nil
	}
}

func firefoxPlacesCandidates() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	roaming := os.Getenv("APPDATA")
	profilesRoot := filepath.Join(roaming, "Mozilla", "Firefox", "Profiles")
	entries, err := os.ReadDir(profilesRoot)
	if err != nil {
		return nil
	}
	var candidates []string
	for _, entry := range entries {
		if entry.IsDir() {
			candidates = append(candidates, filepath.Join(profilesRoot, entry.Name(), "places.sqlite"))
		}
	}
	return candidates
}

func humanFolderName(root string) string {
	switch root {
	case "bookmark_bar":
		return "Bookmarks Bar"
	case "other":
		return "Other Bookmarks"
	case "synced":
		return "Mobile Bookmarks"
	default:
		if root == "" {
			return "Imported"
		}
		return root
	}
}

func cleanPath(path []string) []string {
	out := make([]string, 0, len(path))
	for _, part := range path {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return []string{bookmark.UncategorizedName}
	}
	return out
}

func fileExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}
