// 文件说明：frontend/src/composables/useTaxonomy.ts，对应当前模块的数据结构、状态逻辑或工具函数。
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
