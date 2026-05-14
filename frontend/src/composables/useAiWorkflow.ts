// useAiWorkflow 管理离线分析页的批量分析状态和待分析书签列表。
import { computed, ref } from 'vue'
import { getPendingAiItems } from '../helpers/dashboard'
import { useBookmarkStore } from '../stores/bookmarks'
import type { Bookmark } from '../types'

export const useAiWorkflow = () => {
  const store = useBookmarkStore()
  const analyzing = ref(false)
  const pendingItems = computed(() => getPendingAiItems(store.items))

  const analyzeAll = async () => {
    // 批量分析由后端串行执行，前端只维护 loading 状态和触发刷新。
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
