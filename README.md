# Cola Bookmarks

Cola Bookmarks is a local-first desktop bookmark manager built with Go, Wails v2, Vue 3, and TypeScript.

![CI](https://github.com/Lee-zg/cola/actions/workflows/ci.yml/badge.svg)
![CodeQL](https://github.com/Lee-zg/cola/actions/workflows/codeql.yml/badge.svg)
![Security](https://github.com/Lee-zg/cola/actions/workflows/security.yml/badge.svg)

## Features

- Import Netscape bookmark HTML, Chrome, Edge, and Firefox bookmarks.
- Create, edit, delete, search, categorize, and tag bookmarks.
- Run offline rule-based analysis to generate tags, keywords, and aliases.
- Keep all bookmark data in a local SQLite database.
- Create and restore local database backups.
- Start a `127.0.0.1` read-only web catalog with search, folders, and tags.
- Export an offline static HTML catalog.
- Keep export/web theme package architecture ready with `theme.json + CSS + assets`.

## Project Status

Current version: `0.1.0`

The application is an early MVP. The storage, import/export, local Web catalog, and governance workflows are in place, while richer AI model downloads and a third-party theme ecosystem are future work.

## Development

Requirements:

- Go 1.25+
- Wails v2.12+
- Node.js 22.13+ with `npm` available on `PATH`
- Windows WebView2 on Windows

Commands:

```powershell
go mod download
go test . ./internal/...
cd frontend
npm ci
npm run lint
npm run format:check
npm run typecheck
npm run test
npm run build
cd ..
wails dev
```

## Quality and Security

- CI runs Go tests, frontend linting, format checks, type checks, tests, builds, and a Windows Wails build smoke test.
- CodeQL scans Go and TypeScript.
- `govulncheck`, `npm audit`, dependency review, Dependabot, and OSSF Scorecard are configured.
- Pull request titles follow Conventional Commits.
- Releases are created from `vX.Y.Z` tags.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) for branch naming, commit format, test expectations, review policy, release steps, and dependency management.

Supporting docs:

- [Workflow](docs/WORKFLOW.md)
- [Release guide](docs/RELEASE.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Security policy](SECURITY.md)
- [Code of conduct](CODE_OF_CONDUCT.md)
- [Changelog](CHANGELOG.md)

The current implementation does not upload telemetry or bookmark content. Larger offline AI models are intentionally left as a future local download integration.

## License

MIT. See [LICENSE](LICENSE).
