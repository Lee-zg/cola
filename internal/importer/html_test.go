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
