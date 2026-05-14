// useDashboard 从全局书签状态派生首页统计，不直接请求后端。
import { computed } from 'vue'
import { getDashboardStats, getRecentItems, getTopTags, getWebServerSummary } from '../helpers/dashboard'
import { useBookmarkStore } from '../stores/bookmarks'

export const useDashboard = () => {
  const store = useBookmarkStore()
  const stats = computed(() =>
    getDashboardStats({
      total: store.total,
      folders: store.folders,
      tags: store.tags,
      items: store.items
    })
  )
  const recentItems = computed(() => getRecentItems(store.items))
  const topTags = computed(() => getTopTags(store.tags))
  const webServerSummary = computed(() => getWebServerSummary(store.server))

  const selectTag = (tag: string) => {
    store.setTag(tag)
  }

  return {
    stats,
    recentItems,
    topTags,
    webServerSummary,
    selectTag
  }
}
