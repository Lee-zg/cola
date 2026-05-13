# Release Guide

## Prepare a Release

```powershell
git checkout main
git pull --ff-only
git checkout -b release/vX.Y.Z
```

Update:

- `VERSION`
- `CHANGELOG.md`
- Any user-facing documentation affected by the release

Run:

```powershell
go test . ./internal/...
cd frontend
npm ci
npm run lint
npm run format:check
npm run typecheck
npm run test
npm run build
cd ..
wails build
```

Open a pull request titled:

```text
release: vX.Y.Z
```

## Tag and Publish

After the release PR is merged:

```powershell
git checkout main
git pull --ff-only
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

The release workflow builds the Windows artifact and creates a GitHub Release.

## Hotfixes

For urgent fixes:

1. Branch from `main` with `fix/<issue>`.
2. Apply the smallest safe change.
3. Add regression tests.
4. Release as a patch version.
