// useTaxonomy 复用全局筛选状态，让分类页点击后能影响书签库列表。
import { computed } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'

export const useTaxonomy = () => {
  const store = useBookmarkStore()

  const selectFolder = (folder: string) => {
    store.setFolder(folder)
  }

  const selectTag = (tag: string) => {
    store.setTag(tag)
  }

  return {
    folders: computed(() => store.folders),
    tags: computed(() => store.tags),
    selectFolder,
    selectTag
  }
}
