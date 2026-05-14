// 文件说明：internal/exporter/html.go，负责应用后端或核心业务实现。
package exporter

import (
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"sort"

	"cola/internal/bookmark"
)

type CatalogData struct {
	Title     string              `json:"title"`
	Bookmarks []bookmark.Bookmark `json:"bookmarks"`
	Folders   []string            `json:"folders"`
	Tags      []string            `json:"tags"`
	Generated string              `json:"generated"`
}

func BuildCatalog(title string, items []bookmark.Bookmark) CatalogData {
	folderSeen := map[string]struct{}{}
	tagSeen := map[string]struct{}{}
	for _, item := range items {
		folderSeen[item.Folder] = struct{}{}
		for _, tag := range item.Tags {
			tagSeen[tag] = struct{}{}
		}
	}
	folders := keys(folderSeen)
	tags := keys(tagSeen)
	return CatalogData{
		Title:     title,
		Bookmarks: items,
		Folders:   folders,
		Tags:      tags,
	}
}

func RenderCatalogHTML(data CatalogData, templateID string) (string, error) {
	if data.Title == "" {
		data.Title = "Bookmark Catalog"
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	tpl := classicTemplate
	if templateID == "compact" {
		tpl = compactTemplate
	}
	parsed, err := template.New("catalog").Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = parsed.Execute(&buf, struct {
		Title   string
		Payload template.JS
	}{
		Title:   data.Title,
		Payload: template.JS(payload),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func WriteCatalogHTML(path string, data CatalogData, templateID string) error {
	rendered, err := RenderCatalogHTML(data, templateID)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(rendered), 0o644)
}

func keys(set map[string]struct{}) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		if value != "" {
			values = append(values, value)
		}
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	return values
}

const classicTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
:root{color-scheme:light;--bg:#f5f7fa;--panel:#ffffff;--text:#17202a;--muted:#5f6b7a;--line:#d9e0ea;--accent:#2563eb;--tag:#eef4ff}
*{box-sizing:border-box}body{margin:0;font-family:Inter,Segoe UI,Arial,sans-serif;background:var(--bg);color:var(--text)}
header{position:sticky;top:0;z-index:2;padding:18px 24px;background:var(--panel);border-bottom:1px solid var(--line)}
h1{margin:0 0 12px;font-size:24px}.toolbar{display:grid;grid-template-columns:1fr 220px 220px;gap:12px}
input,select{width:100%;height:38px;border:1px solid var(--line);border-radius:6px;padding:0 12px;background:#fff;color:var(--text)}
main{display:grid;grid-template-columns:260px 1fr;gap:18px;padding:18px 24px}.sidebar{align-self:start;position:sticky;top:96px}
.panel{background:var(--panel);border:1px solid var(--line);border-radius:8px;padding:14px}.panel h2{font-size:14px;margin:0 0 10px;color:var(--muted)}
.chip{display:block;width:100%;text-align:left;border:0;background:transparent;padding:8px;border-radius:6px;color:var(--text);cursor:pointer}.chip.active,.chip:hover{background:var(--tag);color:var(--accent)}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:14px}.card{background:var(--panel);border:1px solid var(--line);border-radius:8px;padding:14px;min-height:136px}
.card a{font-weight:700;color:var(--accent);text-decoration:none;overflow-wrap:anywhere}.url{margin-top:6px;color:var(--muted);font-size:12px;overflow-wrap:anywhere}.desc{margin-top:10px;color:var(--text);line-height:1.45}
.tags{display:flex;flex-wrap:wrap;gap:6px;margin-top:12px}.tag{font-size:12px;background:var(--tag);color:#1d4ed8;padding:4px 7px;border-radius:999px}
.empty{padding:48px;text-align:center;color:var(--muted)}@media(max-width:820px){.toolbar,main{grid-template-columns:1fr}.sidebar{position:static}}
</style>
</head>
<body>
<header><h1>{{.Title}}</h1><div class="toolbar"><input id="q" placeholder="搜索标题、网址、标签、关键词"><select id="folder"></select><select id="tag"></select></div></header>
<main><aside class="sidebar"><section class="panel"><h2>分类</h2><div id="folders"></div></section><section class="panel" style="margin-top:12px"><h2>标签</h2><div id="tags"></div></section></aside><section><div id="count" class="url"></div><div id="list" class="grid"></div></section></main>
<script>const catalog={{.Payload}};let state={q:"",folder:"",tag:""};const $=id=>document.getElementById(id);function esc(s){return String(s||"").replace(/[&<>"']/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[c]))}function options(el,all,items){el.innerHTML="<option value=''>"+all+"</option>"+items.map(x=>"<option>"+esc(x)+"</option>").join("")}function chips(el,items,key){el.innerHTML=[""].concat(items).map(x=>"<button class='chip "+(state[key]===x?"active":"")+"' data-value='"+esc(x)+"'>"+esc(x||"全部")+"</button>").join("");el.querySelectorAll("button").forEach(b=>b.onclick=()=>{state[key]=b.dataset.value;render()})}function match(b){const q=state.q.toLowerCase();const blob=[b.title,b.url,b.description,b.domain,(b.tags||[]).join(" "),(b.keywords||[]).join(" "),(b.aliases||[]).join(" ")].join(" ").toLowerCase();return(!q||blob.includes(q))&&(!state.folder||b.folder===state.folder)&&(!state.tag||(b.tags||[]).includes(state.tag))}function render(){const items=catalog.bookmarks.filter(match);$("count").textContent=items.length+" / "+catalog.bookmarks.length+" 个书签";$("list").innerHTML=items.length?items.map(b=>"<article class='card'><a href='"+esc(b.url)+"' target='_blank' rel='noopener noreferrer'>"+esc(b.title)+"</a><div class='url'>"+esc(b.url)+"</div>"+(b.description?"<div class='desc'>"+esc(b.description)+"</div>":"")+"<div class='tags'>"+(b.tags||[]).map(t=>"<span class='tag'>"+esc(t)+"</span>").join("")+"</div></article>").join(""):"<div class='empty'>没有匹配的书签</div>";options($("folder"),"全部分类",catalog.folders);options($("tag"),"全部标签",catalog.tags);$("folder").value=state.folder;$("tag").value=state.tag;chips($("folders"),catalog.folders,"folder");chips($("tags"),catalog.tags,"tag")} $("q").oninput=e=>{state.q=e.target.value;render()};$("folder").onchange=e=>{state.folder=e.target.value;render()};$("tag").onchange=e=>{state.tag=e.target.value;render()};render();</script>
</body>
</html>`

const compactTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
body{margin:0;font:14px/1.45 Segoe UI,Arial,sans-serif;color:#111827;background:#ffffff}header{padding:14px 18px;border-bottom:1px solid #d1d5db;position:sticky;top:0;background:#fff}h1{font-size:20px;margin:0 0 10px}.bar{display:flex;gap:8px}input,select{height:34px;border:1px solid #cbd5e1;border-radius:4px;padding:0 8px}input{flex:1}main{padding:14px 18px}.row{display:grid;grid-template-columns:1.4fr 2fr 160px;gap:10px;padding:9px 0;border-bottom:1px solid #e5e7eb}.row a{color:#0f62fe;text-decoration:none;font-weight:600;overflow-wrap:anywhere}.muted{color:#667085;overflow-wrap:anywhere}.tags{display:flex;gap:4px;flex-wrap:wrap}.tag{font-size:12px;background:#f1f5f9;padding:2px 5px;border-radius:4px}@media(max-width:760px){.bar,.row{display:block}.row>*{margin:5px 0}}
</style>
</head>
<body>
<header><h1>{{.Title}}</h1><div class="bar"><input id="q" placeholder="搜索"><select id="folder"></select><select id="tag"></select></div></header><main><div id="count" class="muted"></div><div id="list"></div></main>
<script>const catalog={{.Payload}};let state={q:"",folder:"",tag:""};const $=id=>document.getElementById(id);function esc(s){return String(s||"").replace(/[&<>"']/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[c]))}function options(el,all,items){el.innerHTML="<option value=''>"+all+"</option>"+items.map(x=>"<option>"+esc(x)+"</option>").join("")}function match(b){const q=state.q.toLowerCase();const blob=[b.title,b.url,b.description,b.domain,(b.tags||[]).join(" "),(b.keywords||[]).join(" "),(b.aliases||[]).join(" ")].join(" ").toLowerCase();return(!q||blob.includes(q))&&(!state.folder||b.folder===state.folder)&&(!state.tag||(b.tags||[]).includes(state.tag))}function render(){const items=catalog.bookmarks.filter(match);$("count").textContent=items.length+" / "+catalog.bookmarks.length+" 个书签";$("list").innerHTML=items.map(b=>"<div class='row'><a href='"+esc(b.url)+"' target='_blank' rel='noopener noreferrer'>"+esc(b.title)+"</a><div class='muted'>"+esc(b.url)+"</div><div class='tags'>"+(b.tags||[]).map(t=>"<span class='tag'>"+esc(t)+"</span>").join("")+"</div></div>").join("");options($("folder"),"全部分类",catalog.folders);options($("tag"),"全部标签",catalog.tags);$("folder").value=state.folder;$("tag").value=state.tag}["q","folder","tag"].forEach(id=>$(id).oninput=e=>{state[id==="q"?"q":id]=e.target.value;render()});render();</script>
</body>
</html>`
