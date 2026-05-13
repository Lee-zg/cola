# Cola Bookmarks

Cola Bookmarks is a local-first desktop bookmark manager built with Go, Wails v2, Vue 3, and TypeScript.

## Features

- Import Netscape bookmark HTML, Chrome, Edge, and Firefox bookmarks.
- Create, edit, delete, search, categorize, and tag bookmarks.
- Run offline rule-based analysis to generate tags, keywords, and aliases.
- Keep all bookmark data in a local SQLite database.
- Create and restore local database backups.
- Start a `127.0.0.1` read-only web catalog with search, folders, and tags.
- Export an offline static HTML catalog.
- Keep export/web theme package architecture ready with `theme.json + CSS + assets`.

## Development

Requirements:

- Go 1.25+
- Wails v2
- Node.js with `npm` available on `PATH`
- Windows WebView2 on Windows

Commands:

```powershell
go mod download
go test ./...
cd frontend
npm install
npm run build
cd ..
wails dev
```

The current implementation does not upload telemetry or bookmark content. Larger offline AI models are intentionally left as a future local download integration.
