# Project Workflow

## Version Control

The project uses a simple trunk-based model:

- `main` is protected and releasable.
- Work happens on short-lived branches.
- Pull requests are required for changes to `main`.
- CI must pass before merge.

Recommended branch names:

- `feature/<topic>`
- `fix/<topic>`
- `docs/<topic>`
- `test/<topic>`
- `ci/<topic>`
- `chore/<topic>`
- `release/vX.Y.Z`

## Commit Messages

Use Conventional Commits:

```text
feat(scope): add bookmark export template
fix(storage): preserve aliases during analysis
docs: document release process
```

Scopes should be short package or subsystem names, such as `storage`, `importer`, `exporter`, `frontend`, `ci`, or `docs`.

## Code Quality Gates

Required local checks:

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

CI also runs:

- Go race-enabled tests with coverage.
- Frontend lint, format, typecheck, tests, and build.
- Windows Wails build smoke test.
- CodeQL analysis.
- Go vulnerability scan.
- npm production dependency audit.

## Code Review

Pull requests should be small and focused. Reviewers should check:

- Correct behavior and adequate tests.
- Privacy-preserving local-only data flow.
- Safe HTML parsing and rendering.
- No accidental generated or local files.
- Clear documentation for user-facing changes.

## Release Management

Version numbers follow SemVer:

- `MAJOR`: incompatible public behavior or data format changes.
- `MINOR`: backward-compatible features.
- `PATCH`: backward-compatible fixes.

Release checklist:

1. Create `release/vX.Y.Z`.
2. Update `VERSION`.
3. Update `CHANGELOG.md`.
4. Run all local checks.
5. Merge release PR to `main`.
6. Tag `vX.Y.Z`.
7. Push the tag.
8. Confirm GitHub Release and artifacts were created.

## Dependency Management

Dependabot runs weekly for:

- Go modules.
- npm packages under `frontend`.
- GitHub Actions.

Dependency PRs should not include unrelated code changes. Runtime dependency additions need a short rationale in the PR description.

## Security Scanning

Security workflows run on PRs, pushes to `main`, and weekly schedules. Any high-severity result should be triaged before release.
