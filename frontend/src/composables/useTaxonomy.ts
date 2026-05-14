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
