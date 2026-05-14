// api 模块是前端访问 Wails 后端的唯一边界，页面和 store 不直接读取 window.go。
import type {
  Bookmark,
  BookmarkInput,
  ImportResult,
  SearchRequest,
  SearchResult,
  ServerStatus,
  ThemeManifest
} from './types'

type Backend = {
  CreateBookmark(input: BookmarkInput): Promise<Bookmark>
  UpdateBookmark(id: string, input: BookmarkInput): Promise<Bookmark>
  DeleteBookmark(id: string): Promise<void>
  ListBookmarks(req: SearchRequest): Promise<SearchResult>
  ListFolders(): Promise<string[]>
  ListTags(): Promise<string[]>
  ImportBookmarks(req: { sourceType: string; path: string }): Promise<ImportResult>
  ExportBookmarks(req: { path: string; templateId: string }): Promise<string>
  AnalyzeBookmark(id: string): Promise<Bookmark>
  AnalyzeAllBookmarks(): Promise<number>
  CreateBackup(path: string): Promise<{ path: string }>
  RestoreBackup(path: string): Promise<{ path: string }>
  StartLocalServer(): Promise<ServerStatus>
  StopLocalServer(): Promise<void>
  GetLocalServerStatus(): Promise<ServerStatus>
  ListExportTemplates(): Promise<ThemeManifest[]>
}

declare global {
  interface Window {
    go?: { main?: { App?: Backend } }
  }
}

// backend 在运行时校验 Wails 注入对象，便于开发环境未通过 wails 启动时给出明确错误。
function backend(): Backend {
  const api = window.go?.main?.App
  if (!api) {
    throw new Error('Wails backend is not available. Run with wails dev or wails build.')
  }
  return api
}

// api 将 Go 方法名收敛成前端语义化方法，避免组件层依赖 Wails 生成的命名约定。
export const api = {
  createBookmark: (input: BookmarkInput) => backend().CreateBookmark(input),
  updateBookmark: (id: string, input: BookmarkInput) => backend().UpdateBookmark(id, input),
  deleteBookmark: (id: string) => backend().DeleteBookmark(id),
  listBookmarks: (req: SearchRequest) => backend().ListBookmarks(req),
  listFolders: () => backend().ListFolders(),
  listTags: () => backend().ListTags(),
  importBookmarks: (sourceType: string, path: string) => backend().ImportBookmarks({ sourceType, path }),
  exportBookmarks: (path: string, templateId: string) => backend().ExportBookmarks({ path, templateId }),
  analyzeBookmark: (id: string) => backend().AnalyzeBookmark(id),
  analyzeAllBookmarks: () => backend().AnalyzeAllBookmarks(),
  createBackup: (path: string) => backend().CreateBackup(path),
  restoreBackup: (path: string) => backend().RestoreBackup(path),
  startLocalServer: () => backend().StartLocalServer(),
  stopLocalServer: () => backend().StopLocalServer(),
  getLocalServerStatus: () => backend().GetLocalServerStatus(),
  listExportTemplates: () => backend().ListExportTemplates()
}
