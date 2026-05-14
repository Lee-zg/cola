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
