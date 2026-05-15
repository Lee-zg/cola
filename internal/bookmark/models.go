// bookmark 包定义前后端共享的业务数据契约，并集中处理输入归一化规则。
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
	// DefaultLimit 和 MaxLimit 保护列表查询，避免桌面端一次性渲染或导出过大的分页结果。
	DefaultLimit = 100
	MaxLimit     = 500

	RootCategoryID        = "category_all"
	UncategorizedID       = "category_uncategorized"
	UncategorizedName     = "未分类"
	DefaultBrowserSetting = "default"
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
	CategoryID     string     `json:"categoryId"`
	CategoryName   string     `json:"categoryName"`
	CategoryPath   []string   `json:"categoryPath"`
	Preview        *Preview   `json:"preview,omitempty"`
	Thumbnail      *Thumbnail `json:"thumbnail,omitempty"`
	Tags           []string   `json:"tags"`
	Keywords       []string   `json:"keywords"`
	Aliases        []string   `json:"aliases"`
	Domain         string     `json:"domain"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	LastAnalyzedAt *time.Time `json:"lastAnalyzedAt,omitempty"`
}

// BookmarkInput 是用户可编辑字段；ID、Domain 和时间戳由后端生成，避免前端伪造存储状态。
type BookmarkInput struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Folder      string   `json:"folder"`
	CategoryID  string   `json:"categoryId"`
	Tags        []string `json:"tags"`
	Keywords    []string `json:"keywords"`
	Aliases     []string `json:"aliases"`
}

type SearchRequest struct {
	Query              string   `json:"query"`
	Folder             string   `json:"folder"`
	CategoryID         string   `json:"categoryId"`
	IncludeDescendants bool     `json:"includeDescendants"`
	ViewMode           string   `json:"viewMode"`
	Tags               []string `json:"tags"`
	Limit              int      `json:"limit"`
	Offset             int      `json:"offset"`
	Sort               string   `json:"sort"`
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

type CategoryNode struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	ParentID  string         `json:"parentId"`
	SortOrder int            `json:"sortOrder"`
	IsSystem  bool           `json:"isSystem"`
	Count     int            `json:"count"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Children  []CategoryNode `json:"children"`
}

type CategoryInput struct {
	Name     string `json:"name"`
	ParentID string `json:"parentId"`
}

type MoveCategoryInput struct {
	ParentID  string `json:"parentId"`
	SortOrder int    `json:"sortOrder"`
}

type DeleteCategoryInput struct {
	DeleteBookmarks bool `json:"deleteBookmarks"`
}

type Preview struct {
	ID          string    `json:"id"`
	BookmarkID  string    `json:"bookmarkId"`
	Source      string    `json:"source"`
	FilePath    string    `json:"filePath"`
	ThumbPath   string    `json:"thumbPath"`
	OriginalURL string    `json:"originalUrl"`
	Mime        string    `json:"mime"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
}

type PreviewInput struct {
	Source      string `json:"source"`
	Path        string `json:"path"`
	OriginalURL string `json:"originalUrl"`
}

type Thumbnail struct {
	ID                string     `json:"id"`
	BookmarkID        string     `json:"bookmarkId"`
	UseAuto           bool       `json:"useAuto"`
	DisplayPath       string     `json:"displayPath"`
	DisplaySource     string     `json:"displaySource"`
	AutoSource        string     `json:"autoSource"`
	AutoFilePath      string     `json:"autoFilePath"`
	AutoThumbPath     string     `json:"autoThumbPath"`
	AutoOriginalURL   string     `json:"autoOriginalUrl"`
	AutoMime          string     `json:"autoMime"`
	AutoWidth         int        `json:"autoWidth"`
	AutoHeight        int        `json:"autoHeight"`
	AutoSize          int64      `json:"autoSize"`
	AutoStatus        string     `json:"autoStatus"`
	AutoError         string     `json:"autoError"`
	AutoFetchedAt     *time.Time `json:"autoFetchedAt,omitempty"`
	CustomSource      string     `json:"customSource"`
	CustomFilePath    string     `json:"customFilePath"`
	CustomThumbPath   string     `json:"customThumbPath"`
	CustomOriginalURL string     `json:"customOriginalUrl"`
	CustomMime        string     `json:"customMime"`
	CustomWidth       int        `json:"customWidth"`
	CustomHeight      int        `json:"customHeight"`
	CustomSize        int64      `json:"customSize"`
	CustomCreatedAt   *time.Time `json:"customCreatedAt,omitempty"`
}

type ThumbnailUploadInput struct {
	FileName string `json:"fileName"`
	Mime     string `json:"mime"`
	Data     string `json:"data"`
}

type ThumbnailURLInput struct {
	URL string `json:"url"`
}

type ThumbnailModeInput struct {
	UseAuto bool `json:"useAuto"`
}

type ThumbnailClearInput struct {
	ClearAuto bool `json:"clearAuto"`
}

type AppPreferences struct {
	OpenBrowser         string `json:"openBrowser"`
	LazyFetchThumbnails bool   `json:"lazyFetchThumbnails"`
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

// NormalizeInput 是创建、编辑和导入的统一入口，保证默认标题、默认分类和列表字段规则一致。
func NormalizeInput(input BookmarkInput) (BookmarkInput, string, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.URL = strings.TrimSpace(input.URL)
	input.Description = strings.TrimSpace(input.Description)
	input.Folder = strings.TrimSpace(input.Folder)
	input.CategoryID = strings.TrimSpace(input.CategoryID)
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
		input.Folder = UncategorizedName
	}
	return input, domain, nil
}

func NormalizeCategoryInput(input CategoryInput) (CategoryInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.ParentID = strings.TrimSpace(input.ParentID)
	if input.Name == "" {
		return CategoryInput{}, errors.New("category name is required")
	}
	if input.ParentID == RootCategoryID {
		input.ParentID = ""
	}
	if input.ParentID == "" {
		input.ParentID = RootCategoryID
	}
	return input, nil
}

func NormalizeBrowserSetting(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "chrome", "edge", "firefox":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return DefaultBrowserSetting
	}
}

// NormalizeURL 接受用户省略 scheme 的输入，但只允许桌面书签目录需要展示的常见协议。
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
	// 域名去掉 www. 便于搜索和统计；IP 地址保持原样，避免误改实际访问目标。
	if ip := net.ParseIP(host); ip == nil {
		host = strings.TrimPrefix(host, "www.")
	}
	parsed.Scheme = scheme
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	return parsed.String(), host, nil
}

// NormalizeList 对标签、关键词和别名统一去空、大小写不敏感去重并排序，保证持久化结果稳定。
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

// MergeLists 用于 AI 分析结果回写，新增建议只补充到现有人工数据中。
func MergeLists(base []string, additions ...[]string) []string {
	merged := append([]string{}, base...)
	for _, list := range additions {
		merged = append(merged, list...)
	}
	return NormalizeList(merged)
}

// NormalizeSearch 对外部查询做分页兜底和排序白名单处理，防止前端传入无界或未知查询。
func NormalizeSearch(req SearchRequest) SearchRequest {
	req.Query = strings.TrimSpace(req.Query)
	req.Folder = strings.TrimSpace(req.Folder)
	req.CategoryID = strings.TrimSpace(req.CategoryID)
	if req.CategoryID == RootCategoryID {
		req.CategoryID = ""
	}
	if req.ViewMode == "" {
		req.ViewMode = "cards"
	}
	req.Tags = NormalizeList(req.Tags)
	if req.Limit <= 0 {
		switch req.ViewMode {
		case "table":
			req.Limit = 50
		case "cards":
			req.Limit = 24
		default:
			req.Limit = DefaultLimit
		}
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
