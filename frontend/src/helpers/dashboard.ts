import type { Bookmark, ServerStatus } from '../types'

export interface DashboardStats {
  total: number
  folderCount: number
  tagCount: number
  pendingAiCount: number
}

export const getPendingAiItems = (items: Bookmark[]): Bookmark[] => items.filter((item) => !item.lastAnalyzedAt)

export const getDashboardStats = (input: {
  total: number
  folders: string[]
  tags: string[]
  items: Bookmark[]
}): DashboardStats => ({
  total: input.total,
  folderCount: input.folders.length,
  tagCount: input.tags.length,
  pendingAiCount: getPendingAiItems(input.items).length
})

export const getRecentItems = (items: Bookmark[], limit = 5): Bookmark[] => items.slice(0, limit)

export const getTopTags = (tags: string[], limit = 10): string[] => tags.slice(0, limit)

export const getWebServerSummary = (server: ServerStatus): string =>
  server.running ? server.url : '服务未启动，启动后可在局域网访问书签页面。'
