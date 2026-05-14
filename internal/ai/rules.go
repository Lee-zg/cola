// ai 包当前只提供离线规则分析器，后续可在 Analyzer 接口下替换为本地模型实现。
package ai

import (
	"context"
	"net/url"
	"strings"

	"cola/internal/bookmark"
)

type Analyzer interface {
	Analyze(ctx context.Context, item bookmark.Bookmark) (bookmark.AnalysisResult, error)
}

// RuleAnalyzer 不访问网络，也不读取外部模型；结果只作为可合并的标签、关键词和别名建议。
type RuleAnalyzer struct{}

func (RuleAnalyzer) Analyze(_ context.Context, item bookmark.Bookmark) (bookmark.AnalysisResult, error) {
	text := strings.ToLower(strings.Join([]string{item.Title, item.URL, item.Description, item.Domain}, " "))
	tags := []string{}
	keywords := []string{item.Domain}
	aliases := []string{}

	categories := map[string][]string{
		"Development":  {"github", "gitlab", "stackoverflow", "developer", "docs", "api", "golang", "typescript", "vue", "react"},
		"Design":       {"figma", "dribbble", "behance", "design", "font", "icon"},
		"News":         {"news", "medium", "substack", "36kr", "hacker news", "bbc", "nytimes"},
		"AI":           {"openai", "anthropic", "huggingface", "llm", "machine learning", "ai"},
		"Productivity": {"notion", "trello", "calendar", "todo", "task", "workspace"},
		"Learning":     {"course", "tutorial", "learn", "docs", "documentation", "book"},
	}
	// 分类规则保持显式枚举，方便维护者审查自动打标来源，避免隐藏的模型行为。
	for tag, needles := range categories {
		for _, needle := range needles {
			if strings.Contains(text, needle) {
				tags = append(tags, tag)
				break
			}
		}
	}
	if parsed, err := url.Parse(item.URL); err == nil {
		host := strings.TrimPrefix(parsed.Hostname(), "www.")
		parts := strings.Split(host, ".")
		if len(parts) > 0 && parts[0] != "" {
			aliases = append(aliases, parts[0])
			keywords = append(keywords, parts[0])
		}
	}
	for _, token := range strings.FieldsFunc(strings.ToLower(item.Title), splitKeyword) {
		if len([]rune(token)) >= 3 {
			keywords = append(keywords, token)
		}
	}
	return bookmark.AnalysisResult{
		Tags:     bookmark.NormalizeList(tags),
		Keywords: bookmark.NormalizeList(keywords),
		Aliases:  bookmark.NormalizeList(aliases),
		Engine:   "rules",
		Version:  "1.0.0",
	}, nil
}

func splitKeyword(r rune) bool {
	return r == ' ' || r == '-' || r == '_' || r == '/' || r == '\\' || r == '.' || r == ':' || r == '|'
}
