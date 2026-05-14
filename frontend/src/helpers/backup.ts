// 文件说明：frontend/src/helpers/backup.ts，对应当前模块的数据结构、状态逻辑或工具函数。
export const RESTORE_CONFIRMATION = 'RESTORE'

export const canRestoreBackup = (confirmation: string): boolean => confirmation === RESTORE_CONFIRMATION

export const getRestoreValidationMessage = (confirmation: string): string =>
  canRestoreBackup(confirmation) ? '' : `请输入 ${RESTORE_CONFIRMATION} 确认恢复`

export const formatSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export const createBackupHistoryEntry = (path: string, timestamp = new Date()) => {
  const fallbackName = `cola-${timestamp.toISOString().slice(0, 10)}.db`
  const normalized = path.trim().replace(/\\/g, '/')
  const filename = normalized ? normalized.split('/').filter(Boolean).pop() || fallbackName : fallbackName

  return {
    filename,
    size: '未知大小',
    time: timestamp.toLocaleString()
  }
}
