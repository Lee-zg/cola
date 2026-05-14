// 文件说明：app.go，负责应用后端或核心业务实现。
package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cola/internal/ai"
	"cola/internal/backup"
	"cola/internal/bookmark"
	"cola/internal/exporter"
	"cola/internal/importer"
	"cola/internal/storage"
	"cola/internal/theme"
	"cola/internal/webserver"
)

type App struct {
	ctx       context.Context
	mu        sync.Mutex
	store     *storage.Store
	analyzer  ai.Analyzer
	importer  importer.BrowserImporter
	server    *webserver.Server
	dataDir   string
	dbPath    string
	themesDir string
}

func NewApp() *App {
	return &App{analyzer: ai.RuleAnalyzer{}}
}

func NewAppForTest(dataDir string) *App {
	app := NewApp()
	app.dataDir = dataDir
	app.dbPath = filepath.Join(dataDir, "cola.db")
	app.themesDir = filepath.Join(dataDir, "themes")
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.ensureReady(ctx); err != nil {
		println("startup error:", err.Error())
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.server != nil {
		_ = a.server.Stop(ctx)
	}
	if a.store != nil {
		_ = a.store.Close()
	}
}

func (a *App) CreateBookmark(input bookmark.BookmarkInput) (bookmark.Bookmark, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.Bookmark{}, err
	}
	return a.store.CreateBookmark(a.context(), input)
}

func (a *App) UpdateBookmark(id string, input bookmark.BookmarkInput) (bookmark.Bookmark, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.Bookmark{}, err
	}
	return a.store.UpdateBookmark(a.context(), id, input)
}

func (a *App) DeleteBookmark(id string) error {
	if err := a.ensureReady(a.context()); err != nil {
		return err
	}
	return a.store.DeleteBookmark(a.context(), id)
}

func (a *App) GetBookmark(id string) (bookmark.Bookmark, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.Bookmark{}, err
	}
	return a.store.GetBookmark(a.context(), id)
}

func (a *App) ListBookmarks(req bookmark.SearchRequest) (bookmark.SearchResult, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.SearchResult{}, err
	}
	return a.store.ListBookmarks(a.context(), req)
}

func (a *App) ListFolders() ([]string, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return nil, err
	}
	return a.store.ListFolders(a.context())
}

func (a *App) ListTags() ([]string, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return nil, err
	}
	return a.store.ListTags(a.context())
}

func (a *App) ImportBookmarks(req bookmark.ImportRequest) (bookmark.ImportResult, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.ImportResult{}, err
	}
	items, err := a.importer.Import(a.context(), req.SourceType, req.Path)
	if err != nil {
		return bookmark.ImportResult{}, err
	}
	return a.store.UpsertBookmarks(a.context(), items)
}

func (a *App) ExportBookmarks(req bookmark.ExportRequest) (string, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return "", err
	}
	if strings.TrimSpace(req.Path) == "" {
		return "", errors.New("export path is required")
	}
	items, err := a.store.AllBookmarks(a.context())
	if err != nil {
		return "", err
	}
	templateID := req.TemplateID
	if templateID == "" {
		templateID = "classic"
	}
	data := exporter.BuildCatalog("Cola Bookmarks", items)
	if err := exporter.WriteCatalogHTML(req.Path, data, templateID); err != nil {
		return "", err
	}
	return req.Path, nil
}

func (a *App) AnalyzeBookmark(id string) (bookmark.Bookmark, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.Bookmark{}, err
	}
	item, err := a.store.GetBookmark(a.context(), id)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	result, err := a.analyzer.Analyze(a.context(), item)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	return a.store.ApplyAnalysis(a.context(), id, result)
}

func (a *App) AnalyzeAllBookmarks() (int, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return 0, err
	}
	items, err := a.store.AllBookmarks(a.context())
	if err != nil {
		return 0, err
	}
	count := 0
	for _, item := range items {
		result, err := a.analyzer.Analyze(a.context(), item)
		if err != nil {
			return count, err
		}
		if _, err := a.store.ApplyAnalysis(a.context(), item.ID, result); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (a *App) CreateBackup(path string) (bookmark.BackupResult, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.BackupResult{}, err
	}
	backupPath, err := backup.Create(a.store.DBPath(), path)
	if err != nil {
		return bookmark.BackupResult{}, err
	}
	return bookmark.BackupResult{Path: backupPath}, nil
}

func (a *App) RestoreBackup(path string) (bookmark.BackupResult, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.BackupResult{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.server != nil {
		_ = a.server.Stop(a.context())
	}
	if err := a.store.Close(); err != nil {
		return bookmark.BackupResult{}, err
	}
	snapshot, err := backup.Restore(a.dbPath, path)
	if err != nil {
		reopened, openErr := storage.Open(a.context(), a.dbPath)
		if openErr == nil {
			a.store = reopened
			a.server = webserver.New(a.store)
		}
		return bookmark.BackupResult{Path: snapshot}, err
	}
	reopened, err := storage.Open(a.context(), a.dbPath)
	if err != nil {
		return bookmark.BackupResult{Path: snapshot}, err
	}
	a.store = reopened
	a.server = webserver.New(a.store)
	return bookmark.BackupResult{Path: snapshot}, nil
}

func (a *App) StartLocalServer() (bookmark.ServerStatus, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.ServerStatus{}, err
	}
	return a.server.Start(a.context())
}

func (a *App) StopLocalServer() error {
	if err := a.ensureReady(a.context()); err != nil {
		return err
	}
	return a.server.Stop(a.context())
}

func (a *App) GetLocalServerStatus() (bookmark.ServerStatus, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.ServerStatus{}, err
	}
	return a.server.Status(), nil
}

func (a *App) ListExportTemplates() []bookmark.ThemeManifest {
	return theme.BuiltinTemplates()
}

func (a *App) ValidateThemePackage(path string) (bookmark.ThemeManifest, error) {
	return theme.ValidatePackage(path)
}

func (a *App) InstallThemePackage(path string) (bookmark.ThemeManifest, error) {
	if err := a.ensureReady(a.context()); err != nil {
		return bookmark.ThemeManifest{}, err
	}
	return theme.InstallPackage(path, a.themesDir)
}

func (a *App) ensureReady(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store != nil {
		return nil
	}
	if a.dataDir == "" {
		config, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		a.dataDir = filepath.Join(config, "ColaBookmarks")
	}
	if a.dbPath == "" {
		a.dbPath = filepath.Join(a.dataDir, "cola.db")
	}
	if a.themesDir == "" {
		a.themesDir = filepath.Join(a.dataDir, "themes")
	}
	if err := os.MkdirAll(a.themesDir, 0o755); err != nil {
		return err
	}
	store, err := storage.Open(ctx, a.dbPath)
	if err != nil {
		return err
	}
	a.store = store
	a.server = webserver.New(store)
	return nil
}

func (a *App) context() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
