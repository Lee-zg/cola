// useImportWorkflow 管理导入页临时状态，导入解析和去重由后端 importer/storage 完成。
import { ref } from 'vue'
import { importSourceOptions, type ImportHistoryRecord } from '../helpers/workflow'
import { useBookmarkStore } from '../stores/bookmarks'

export const useImportWorkflow = () => {
  const store = useBookmarkStore()
  const importPath = ref('')
  const importSource = ref('html')
  const importing = ref(false)
  const importHistory = ref<ImportHistoryRecord[]>([])
  const skipDuplicates = ref(true)
  const autoAnalyze = ref(false)
  const keepFolders = ref(true)

  const startImport = async () => {
    // 历史记录只用于当前前端会话展示，不写入数据库或配置。
    importing.value = true
    try {
      await store.importFrom(importSource.value, importPath.value)
      importHistory.value.unshift({
        source: importSource.value,
        message: store.status,
        time: new Date().toLocaleString()
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      store.setStatus(`导入失败：${message}`)
      importHistory.value.unshift({
        source: importSource.value,
        message: store.status,
        time: new Date().toLocaleString()
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
    sourceOptions: importSourceOptions,
    startImport
  }
}
