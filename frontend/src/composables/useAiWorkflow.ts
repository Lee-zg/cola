// 文件说明：frontend/src/composables/useAiWorkflow.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { computed, ref } from 'vue'
import { getPendingAiItems } from '../helpers/dashboard'
import { useBookmarkStore } from '../stores/bookmarks'
import type { Bookmark } from '../types'

export const useAiWorkflow = () => {
  const store = useBookmarkStore()
  const analyzing = ref(false)
  const pendingItems = computed(() => getPendingAiItems(store.items))

  const analyzeAll = async () => {
    analyzing.value = true
    try {
      await store.analyzeAll()
    } finally {
      analyzing.value = false
    }
  }

  const selectForReview = (item: Bookmark) => {
    store.select(item)
  }

  return {
    analyzing,
    pendingItems,
    analyzeAll,
    selectForReview
  }
}
