// theme 包定义导出模板包的最低校验规则，确保第三方资源只能作为静态数据被安装。
package theme

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"cola/internal/bookmark"
)

const TemplateAPIVersion = "1"

// BuiltinTemplates 暴露当前内置模板清单，前端只用它来展示可选导出样式。
func BuiltinTemplates() []bookmark.ThemeManifest {
	return []bookmark.ThemeManifest{
		{
			ID:                 "classic",
			Name:               "Classic Catalog",
			Version:            "1.0.0",
			TemplateAPIVersion: TemplateAPIVersion,
			Entry:              "builtin",
			Author:             "Cola",
			Description:        "Readable light catalog optimized for exported bookmark pages.",
		},
		{
			ID:                 "compact",
			Name:               "Compact Index",
			Version:            "1.0.0",
			TemplateAPIVersion: TemplateAPIVersion,
			Entry:              "builtin",
			Author:             "Cola",
			Description:        "Dense layout for large bookmark collections.",
		},
	}
}

// ValidatePackage 校验 manifest、API 版本和资源路径，禁止绝对路径、目录逃逸和脚本类资源。
func ValidatePackage(packagePath string) (bookmark.ThemeManifest, error) {
	info, err := os.Stat(packagePath)
	if err != nil {
		return bookmark.ThemeManifest{}, err
	}
	if !info.IsDir() {
		return bookmark.ThemeManifest{}, errors.New("theme package must be a directory")
	}
	manifestPath := filepath.Join(packagePath, "theme.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return bookmark.ThemeManifest{}, err
	}
	var manifest bookmark.ThemeManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return bookmark.ThemeManifest{}, err
	}
	if strings.TrimSpace(manifest.ID) == "" || strings.TrimSpace(manifest.Name) == "" {
		return bookmark.ThemeManifest{}, errors.New("theme id and name are required")
	}
	if manifest.TemplateAPIVersion != TemplateAPIVersion {
		return bookmark.ThemeManifest{}, errors.New("unsupported template api version")
	}
	if manifest.Entry == "" {
		return bookmark.ThemeManifest{}, errors.New("theme entry is required")
	}
	for _, rel := range append(append([]string{}, manifest.CSS...), manifest.Assets...) {
		if err := validateThemeAssetPath(packagePath, rel); err != nil {
			return bookmark.ThemeManifest{}, err
		}
	}
	return manifest, nil
}

// InstallPackage 先复用完整校验，再复制到应用主题目录；复制过程中仍逐个校验资源路径。
func InstallPackage(packagePath, themesDir string) (bookmark.ThemeManifest, error) {
	manifest, err := ValidatePackage(packagePath)
	if err != nil {
		return bookmark.ThemeManifest{}, err
	}
	target := filepath.Join(themesDir, manifest.ID)
	if err := os.MkdirAll(target, 0o755); err != nil {
		return bookmark.ThemeManifest{}, err
	}
	if err := copyThemeDirectory(packagePath, target); err != nil {
		return bookmark.ThemeManifest{}, err
	}
	return manifest, nil
}

func validateThemeAssetPath(root, rel string) error {
	if rel == "" || filepath.IsAbs(rel) || strings.Contains(rel, "..") {
		return errors.New("theme asset path must be relative and stay inside package")
	}
	ext := strings.ToLower(filepath.Ext(rel))
	allowed := map[string]struct{}{
		".json": {}, ".css": {}, ".png": {}, ".jpg": {}, ".jpeg": {},
		".gif": {}, ".webp": {}, ".svg": {}, ".woff": {}, ".woff2": {},
	}
	if _, ok := allowed[ext]; !ok {
		return errors.New("theme assets may only contain json, css, fonts, and images")
	}
	full := filepath.Join(root, rel)
	info, err := os.Stat(full)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("theme asset path must point to a file")
	}
	return nil
}

func copyThemeDirectory(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := validateThemeAssetPath(src, rel); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
