// workflow helper 保存跨页面复用的展示型选项和历史记录结构，不负责持久化。
export interface ImportSourceOption {
  id: string
  name: string
  description: string
}

export interface ImportHistoryRecord {
  source: string
  message: string
  time: string
}

export interface BackupHistoryEntry {
  filename: string
  size: string
  time: string
}

// id 必须匹配后端 BrowserImporter 支持的 sourceType。
export const importSourceOptions: ImportSourceOption[] = [
  { id: 'html', name: 'HTML 文件', description: '从浏览器导出的 bookmarks.html 导入' },
  { id: 'chrome', name: 'Chrome', description: '自动扫描 Chrome 默认书签数据' },
  { id: 'edge', name: 'Edge', description: '自动扫描 Edge 默认书签数据' },
  { id: 'firefox', name: 'Firefox', description: '自动扫描 Firefox 默认书签数据' }
]
