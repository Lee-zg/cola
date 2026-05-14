// 文件说明：frontend/src/composables/useExportWorkflow.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { computed, ref, watch } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'

export type ExportScope = 'all' | 'filtered'

export const useExportWorkflow = () => {
  const store = useBookmarkStore()
  const exportPath = ref('')
  const templateId = ref('classic')
  const exporting = ref(false)
  const exportScope = ref<ExportScope>('all')
  const templates = computed(() => store.templates)
  const selectedTemplate = computed(() => store.templates.find((template) => template.id === templateId.value))

  watch(
    templates,
    (availableTemplates) => {
      if (availableTemplates.length && !availableTemplates.some((template) => template.id === templateId.value)) {
        templateId.value = availableTemplates[0].id
      }
    },
    { immediate: true }
  )

  const selectTemplate = (id: string) => {
    templateId.value = id
  }

  const exportHtml = async () => {
    exporting.value = true
    try {
      await store.exportTo(exportPath.value, templateId.value)
    } finally {
      exporting.value = false
    }
  }

  return {
    exportPath,
    templateId,
    exporting,
    exportScope,
    templates,
    selectedTemplate,
    selectTemplate,
    exportHtml
  }
}
