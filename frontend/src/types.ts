// 文件说明：frontend/src/types.ts，对应当前模块的数据结构、状态逻辑或工具函数。
export interface Bookmark {
  id: string
  title: string
  url: string
  description: string
  folder: string
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
  tags: string[]
  keywords: string[]
  aliases: string[]
}

export interface SearchRequest {
  query: string
  folder: string
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
