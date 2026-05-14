// 文件说明：internal/theme/theme_test.go，负责应用后端或核心业务实现。
package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePackage(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "style.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := `{"id":"sample","name":"Sample","version":"1.0.0","templateApiVersion":"1","entry":"index.html","css":["style.css"],"assets":[]}`
	if err := os.WriteFile(filepath.Join(dir, "theme.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed, err := ValidatePackage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ID != "sample" {
		t.Fatalf("unexpected manifest: %#v", parsed)
	}
}
