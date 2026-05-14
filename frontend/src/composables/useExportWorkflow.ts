// useExportWorkflow 负责导出页的路径、模板选择和 loading 状态；实际导出始终由后端读取数据库。
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

  // 模板清单来自后端，若当前选择不可用则回落到第一项，避免提交不存在的 templateId。
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
    // exportScope 目前只影响界面意图展示，后端导出完整书签目录。
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
