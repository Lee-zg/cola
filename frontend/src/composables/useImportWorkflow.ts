// useImportWorkflow 管理导入页临时状态，导入解析、去重和可选分析由后端完成。
import { computed, ref } from 'vue'
import { getImportSourceOption, importSourceOptions, type ImportHistoryRecord } from '../helpers/workflow'
import { useBookmarkStore } from '../stores/bookmarks'
import type { ImportResult } from '../types'

const emptyImportResult = (): ImportResult => ({
  imported: 0,
  updated: 0,
  skipped: 0,
  analyzed: 0,
  errors: []
})

const formatImportSummary = (result: ImportResult) => {
  const parts = [`导入 ${result.imported} 个`, `更新 ${result.updated} 个`, `跳过 ${result.skipped} 个`]
  if (result.analyzed) parts.push(`分析 ${result.analyzed} 个`)
  return parts.join('，')
}

export const useImportWorkflow = () => {
  const store = useBookmarkStore()
  const importPath = ref('')
  const importSource = ref('html')
  const importing = ref(false)
  const importHistory = ref<ImportHistoryRecord[]>([])
  const skipDuplicates = ref(true)
  const autoAnalyze = ref(false)
  const keepFolders = ref(true)
  const lastResult = ref<ImportResult | null>(null)
  const lastError = ref('')

  const selectedSource = computed(() => getImportSourceOption(importSource.value))
  const importPathRequired = computed(() => !selectedSource.value.autoPathSupported)
  const canStartImport = computed(() => !importing.value && (!importPathRequired.value || Boolean(importPath.value.trim())))
  const pathValidationMessage = computed(() =>
    importPathRequired.value && !importPath.value.trim() ? 'HTML 文件导入需要填写文件路径。' : ''
  )
  const resultSummary = computed(() => (lastResult.value ? formatImportSummary(lastResult.value) : '尚未执行导入'))

  const setImportSource = (source: string) => {
    importSource.value = source
    lastError.value = ''
  }

  const setImportPath = (path: string) => {
    importPath.value = path.trim()
    lastError.value = ''
  }

  const applyDroppedFiles = (paths: string[]) => {
    const firstPath = paths.find(Boolean)
    if (firstPath) setImportPath(firstPath)
  }

  const startImport = async () => {
    if (!canStartImport.value) {
      lastError.value = pathValidationMessage.value || '请先补全导入信息。'
      store.setStatus(lastError.value)
      return
    }
    // 历史记录只用于当前前端会话展示，不写入数据库或配置。
    importing.value = true
    lastError.value = ''
    try {
      const result = await store.importFrom({
        sourceType: importSource.value,
        path: importPath.value.trim(),
        skipDuplicates: skipDuplicates.value,
        autoAnalyze: autoAnalyze.value,
        keepFolders: keepFolders.value
      })
      lastResult.value = result
      importHistory.value.unshift({
        source: selectedSource.value.name,
        message: formatImportSummary(result),
        time: new Date().toLocaleString(),
        type: 'success',
        imported: result.imported,
        updated: result.updated,
        skipped: result.skipped,
        analyzed: result.analyzed,
        errors: result.errors
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      const result = emptyImportResult()
      lastResult.value = result
      lastError.value = message
      store.setStatus(`导入失败：${message}`)
      importHistory.value.unshift({
        source: selectedSource.value.name,
        message: store.status,
        time: new Date().toLocaleString(),
        type: 'error',
        imported: 0,
        updated: 0,
        skipped: 0,
        analyzed: 0,
        errors: [message]
      })
    } finally {
      importing.value = false
    }
  }

  return {
    importPath,
    importSource,
    importing,
    importHistory,
    skipDuplicates,
    autoAnalyze,
    keepFolders,
    selectedSource,
    importPathRequired,
    canStartImport,
    pathValidationMessage,
    lastResult,
    lastError,
    resultSummary,
    sourceOptions: importSourceOptions,
    setImportSource,
    setImportPath,
    applyDroppedFiles,
    startImport
  }
}
