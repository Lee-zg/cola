// 文件说明：internal/bookmark/models.go，负责应用后端或核心业务实现。
package bookmark

import (
	"errors"
	"net"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	DefaultLimit = 100
	MaxLimit     = 500
)

var allowedSchemes = map[string]struct{}{
	"http":  {},
	"https": {},
	"ftp":   {},
}

type Bookmark struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	URL            string     `json:"url"`
	Description    string     `json:"description"`
	Folder         string     `json:"folder"`
	Tags           []string   `json:"tags"`
	Keywords       []string   `json:"keywords"`
	Aliases        []string   `json:"aliases"`
	Domain         string     `json:"domain"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	LastAnalyzedAt *time.Time `json:"lastAnalyzedAt,omitempty"`
}

type BookmarkInput struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Folder      string   `json:"folder"`
	Tags        []string `json:"tags"`
	Keywords    []string `json:"keywords"`
	Aliases     []string `json:"aliases"`
}

type SearchRequest struct {
	Query  string   `json:"query"`
	Folder string   `json:"folder"`
	Tags   []string `json:"tags"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
	Sort   string   `json:"sort"`
}

type SearchResult struct {
	Items  []Bookmark `json:"items"`
	Total  int        `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

type ImportRequest struct {
	SourceType string `json:"sourceType"`
	Path       string `json:"path"`
}

type ImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}

type ExportRequest struct {
	Path       string `json:"path"`
	TemplateID string `json:"templateId"`
}

type BackupResult struct {
	Path string `json:"path"`
}

type ServerStatus struct {
	Running bool   `json:"running"`
	URL     string `json:"url"`
	Addr    string `json:"addr"`
}

type AnalysisResult struct {
	Tags     []string `json:"tags"`
	Keywords []string `json:"keywords"`
	Aliases  []string `json:"aliases"`
	Engine   string   `json:"engine"`
	Version  string   `json:"version"`
}

type ThemeManifest struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Version            string   `json:"version"`
	TemplateAPIVersion string   `json:"templateApiVersion"`
	Entry              string   `json:"entry"`
	CSS                []string `json:"css"`
	Assets             []string `json:"assets"`
	Author             string   `json:"author"`
	Description        string   `json:"description"`
}

func NormalizeInput(input BookmarkInput) (BookmarkInput, string, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.URL = strings.TrimSpace(input.URL)
	input.Description = strings.TrimSpace(input.Description)
	input.Folder = strings.TrimSpace(input.Folder)
	input.Tags = NormalizeList(input.Tags)
	input.Keywords = NormalizeList(input.Keywords)
	input.Aliases = NormalizeList(input.Aliases)

	normalizedURL, domain, err := NormalizeURL(input.URL)
	if err != nil {
		return BookmarkInput{}, "", err
	}
	input.URL = normalizedURL
	if input.Title == "" {
		input.Title = domain
	}
	if input.Folder == "" {
		input.Folder = "Unsorted"
	}
	return input, domain, nil
}

func NormalizeURL(raw string) (string, string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", "", errors.New("url is required")
	}
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", "", err
	}
	if parsed.Scheme == "" {
		parsed, err = url.Parse("https://" + strings.TrimSpace(raw))
		if err != nil {
			return "", "", err
		}
	}
	scheme := strings.ToLower(parsed.Scheme)
	if _, ok := allowedSchemes[scheme]; !ok {
		return "", "", errors.New("unsupported url scheme")
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return "", "", errors.New("url host is required")
	}
	if ip := net.ParseIP(host); ip == nil {
		host = strings.TrimPrefix(host, "www.")
	}
	parsed.Scheme = scheme
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	return parsed.String(), host, nil
}

func NormalizeList(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	slices.SortFunc(out, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return out
}

func MergeLists(base []string, additions ...[]string) []string {
	merged := append([]string{}, base...)
	for _, list := range additions {
		merged = append(merged, list...)
	}
	return NormalizeList(merged)
}

func NormalizeSearch(req SearchRequest) SearchRequest {
	req.Query = strings.TrimSpace(req.Query)
	req.Folder = strings.TrimSpace(req.Folder)
	req.Tags = NormalizeList(req.Tags)
	if req.Limit <= 0 {
		req.Limit = DefaultLimit
	}
	if req.Limit > MaxLimit {
		req.Limit = MaxLimit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	switch req.Sort {
	case "title", "created", "updated":
	default:
		req.Sort = "updated"
	}
	return req
}
