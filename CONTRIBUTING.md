# Contributing to Cola Bookmarks

Thanks for helping improve Cola Bookmarks. This project is local-first desktop software, so privacy, predictable behavior, and maintainable code matter more than adding features quickly.

## Development Setup

Requirements:

- Go 1.25 or newer
- Node.js 22.13 or newer
- npm 10 or newer
- Wails v2.12.0 or newer
- Windows WebView2 for Windows builds

Install and verify:

```powershell
go mod download
cd frontend
npm ci
npm run build
cd ..
go test . ./internal/...
cd frontend
npm run lint
npm run typecheck
npm run test
npm run build
cd ..
wails dev
```

## Branch Strategy

- `main` is always expected to be releasable.
- Feature branches use `feature/<short-description>`.
- Bug fixes use `fix/<short-description>`.
- Documentation-only work uses `docs/<short-description>`.
- CI/build maintenance uses `chore/<short-description>`.
- Release preparation uses `release/vX.Y.Z`.

Keep branches focused. Avoid mixing feature work, refactoring, formatting, and dependency updates in one pull request.

## Commit Message Format

Use Conventional Commits:

```text
<type>(optional-scope): <summary>
```

Accepted types:

- `feat`: user-facing feature
- `fix`: bug fix
- `docs`: documentation
- `test`: tests
- `refactor`: behavior-preserving code change
- `perf`: performance improvement
- `build`: build system or dependency changes
- `ci`: CI configuration
- `chore`: maintenance
- `security`: vulnerability fix or hardening

Examples:

```text
feat(importer): support Firefox bookmark import
fix(exporter): escape bookmark descriptions in static HTML
ci: add CodeQL workflow
```

Breaking changes must include `BREAKING CHANGE:` in the commit body.

## Pull Request Requirements

Every pull request should include:

- A clear summary of the change.
- Linked issue when applicable.
- Test evidence, including exact commands run.
- Screenshots or short recordings for UI changes.
- Notes about privacy, local storage, or security impact when relevant.

Before requesting review, run:

```powershell
cd frontend
npm ci
npm run build
cd ..
go test . ./internal/...
cd frontend
npm run lint
npm run format:check
npm run typecheck
npm run test
npm run build
```

## Code Review

At least one maintainer approval is required before merging to `main`.

Reviewers should prioritize:

- Data privacy and local-only guarantees.
- Security boundaries for imports, exports, local Web access, and theme packages.
- Tests for new behavior and regression risk.
- Clear module ownership and maintainable APIs.
- UI behavior across typical desktop window sizes.

Use squash merge for ordinary PRs unless preserving individual commits adds clear value.

## Testing Policy

- Go unit tests live next to packages as `*_test.go`.
- Integration tests should exercise the public `App` facade or service boundaries.
- Frontend tests use Vitest.
- Any importer, exporter, backup, or local Web server change needs a test.
- Security-sensitive HTML handling needs tests for escaping or rejection behavior.

## Release Process

1. Update `VERSION` using SemVer.
2. Update `CHANGELOG.md`.
3. Run the full local verification commands.
4. Create a release PR from `release/vX.Y.Z` to `main`.
5. After merge, create and push tag `vX.Y.Z`.
6. GitHub Actions builds the release artifact and creates a GitHub Release.

## Dependency Management

- Dependabot opens weekly PRs for Go, npm, and GitHub Actions dependencies.
- Keep dependency PRs small and separate from feature work.
- Security updates should be reviewed and merged promptly after CI passes.
- New runtime dependencies require a short justification in the PR.

## Security and Privacy Expectations

- Do not add telemetry or network upload of bookmark data.
- Local Web server changes must default to `127.0.0.1`.
- Imported and exported HTML must be sanitized or escaped.
- Theme packages must not execute third-party JavaScript.
