// types.ts 镜像 Go 后端暴露给 Wails 的 JSON 契约，字段名需要和 bookmark 包保持一致。
export interface Bookmark {
  id: string
  title: string
  url: string
  description: string
  folder: string
  categoryId: string
  categoryName: string
  categoryPath: string[]
  preview?: BookmarkPreview
  tags: string[]
  keywords: string[]
  aliases: string[]
  domain: string
  createdAt: string
  updatedAt: string
  lastAnalyzedAt?: string
}

export interface BookmarkInput {
  title: string
  url: string
  description: string
  folder: string
  categoryId: string
  tags: string[]
  keywords: string[]
  aliases: string[]
}

export interface SearchRequest {
  query: string
  folder: string
  categoryId: string
  includeDescendants: boolean
  viewMode: string
  tags: string[]
  limit: number
  offset: number
  sort: string
}

export interface SearchResult {
  items: Bookmark[]
  total: number
  limit: number
  offset: number
}

export interface ServerStatus {
  running: boolean
  url: string
  addr: string
}

export interface ImportResult {
  imported: number
  skipped: number
  errors: string[]
}

export interface CategoryNode {
  id: string
  name: string
  parentId: string
  sortOrder: number
  isSystem: boolean
  count: number
  createdAt: string
  updatedAt: string
  children: CategoryNode[]
}

export interface CategoryInput {
  name: string
  parentId: string
}

export interface MoveCategoryInput {
  parentId: string
  sortOrder: number
}

export interface DeleteCategoryInput {
  deleteBookmarks: boolean
}

export interface BookmarkPreview {
  id: string
  bookmarkId: string
  source: string
  filePath: string
  thumbPath: string
  originalUrl: string
  mime: string
  width: number
  height: number
  size: number
  createdAt: string
}

export interface PreviewInput {
  source: string
  path: string
  originalUrl: string
}

export interface AppPreferences {
  openBrowser: string
}

export interface ThemeManifest {
  id: string
  name: string
  version: string
  templateApiVersion: string
  entry: string
  css: string[]
  assets: string[]
  author: string
  description: string
}
