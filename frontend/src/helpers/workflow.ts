// 文件说明：frontend/src/helpers/workflow.ts，对应当前模块的数据结构、状态逻辑或工具函数。
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

export const importSourceOptions: ImportSourceOption[] = [
  { id: 'html', name: 'HTML 文件', description: '从浏览器导出的 bookmarks.html 导入' },
  { id: 'chrome', name: 'Chrome', description: '自动扫描 Chrome 默认书签数据' },
  { id: 'edge', name: 'Edge', description: '自动扫描 Edge 默认书签数据' },
  { id: 'firefox', name: 'Firefox', description: '自动扫描 Firefox 默认书签数据' }
]
