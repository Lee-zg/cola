// 文件说明：frontend/src/composables/useSettingsSummary.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { computed } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'

export const useSettingsSummary = () => {
  const store = useBookmarkStore()

  return {
    total: computed(() => store.total),
    folderCount: computed(() => store.folders.length),
    tagCount: computed(() => store.tags.length)
  }
}
