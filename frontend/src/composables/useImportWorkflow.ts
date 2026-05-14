// 文件说明：frontend/src/composables/useImportWorkflow.ts，对应当前模块的数据结构、状态逻辑或工具函数。
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
    importing.value = true
    try {
      await store.importFrom(importSource.value, importPath.value)
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
