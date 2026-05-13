# Architecture Overview

Cola Bookmarks is a local-first desktop application.

```mermaid
flowchart LR
  Desktop["Wails Desktop UI"] --> App["Go App Facade"]
  App --> Storage["SQLite Storage"]
  App --> Importer["Browser and HTML Importers"]
  App --> AI["Offline Rule Analyzer"]
  App --> Exporter["HTML Exporter"]
  App --> Web["127.0.0.1 Web Catalog"]
  App --> Backup["Backup and Restore"]
  Exporter --> Theme["Export Theme Templates"]
```

## Boundaries

- The Vue frontend only talks to the Wails `App` facade.
- Storage is local SQLite.
- The local Web server is read-only and binds to `127.0.0.1`.
- Theme packages are data and CSS only; JavaScript execution is out of scope.
- AI analysis is offline and rule-based in the current version.

## Privacy Model

Bookmark data, tags, aliases, analysis results, backups, and exports stay local. No telemetry or cloud upload is part of the current architecture.
