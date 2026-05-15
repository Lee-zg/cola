// 文件说明：internal/importer/html_test.go，负责应用后端或核心业务实现。
package importer

import "testing"

func TestParseNetscapeHTML(t *testing.T) {
	data := []byte(`
		<DL><p>
		<DT><H3>Dev</H3>
		<DT><A HREF="https://go.dev/doc/">Go Docs</A>
		<DT><A HREF="javascript:alert(1)">Bad</A>
		</DL><p>`)
	items := ParseNetscapeHTML(data)
	if len(items) != 2 {
		t.Fatalf("expected 2 parsed items, got %d", len(items))
	}
	if items[0].Title != "Go Docs" || items[0].Folder != "Dev" {
		t.Fatalf("unexpected first item: %#v", items[0])
	}
}

func TestParseNetscapeHTMLNestedAndMinified(t *testing.T) {
	data := []byte(`<DL><p><DT><H3>Dev</H3><DL><p><DT><H3>Go</H3><DL><p><DT><A HREF="https://pkg.go.dev/">Pkg</A></DL><p><DT><A HREF="https://go.dev/">Go</A></DL><p></DL><p>`)
	items := ParseNetscapeHTML(data)
	if len(items) != 2 {
		t.Fatalf("expected 2 parsed items, got %d", len(items))
	}
	if items[0].Folder != "Dev / Go" {
		t.Fatalf("expected nested folder path, got %#v", items[0])
	}
	if items[1].Folder != "Dev" {
		t.Fatalf("expected parent folder path after nested close, got %#v", items[1])
	}
}
