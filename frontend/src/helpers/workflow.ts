// workflow helper 保存跨页面复用的展示型选项和历史记录结构，不负责持久化。
export interface ImportSourceOption {
  id: string
  name: string
  description: string
  pathHint: string
  autoPathSupported: boolean
}

export interface ImportHistoryRecord {
  source: string
  message: string
  time: string
  type: 'success' | 'error'
  imported: number
  updated: number
  skipped: number
  analyzed: number
  errors: string[]
}

export interface BackupHistoryEntry {
  filename: string
  size: string
  time: string
}

// id 必须匹配后端 BrowserImporter 支持的 sourceType。
export const importSourceOptions: ImportSourceOption[] = [
  {
    id: 'html',
    name: 'HTML 文件',
    description: '从浏览器导出的 bookmarks.html 导入',
    pathHint: '需要填写 bookmarks.html 文件路径，或把文件拖到导入面板。',
    autoPathSupported: false
  },
  {
    id: 'chrome',
    name: 'Chrome',
    description: '自动扫描 Chrome 默认书签数据',
    pathHint: '可留空自动扫描默认配置，也可以填写 Bookmarks 文件路径。',
    autoPathSupported: true
  },
  {
    id: 'edge',
    name: 'Edge',
    description: '自动扫描 Edge 默认书签数据',
    pathHint: '可留空自动扫描默认配置，也可以填写 Bookmarks 文件路径。',
    autoPathSupported: true
  },
  {
    id: 'firefox',
    name: 'Firefox',
    description: '自动扫描 Firefox places.sqlite',
    pathHint: '可留空自动扫描默认配置，也可以填写 places.sqlite 文件路径。',
    autoPathSupported: true
  }
]

export const getImportSourceOption = (id: string) =>
  importSourceOptions.find((option) => option.id === id) || importSourceOptions[0]
