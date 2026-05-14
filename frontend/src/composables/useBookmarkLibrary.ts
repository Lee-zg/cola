// 文件说明：frontend/src/composables/useBookmarkLibrary.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { computed, ref } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'
import type { Bookmark, BookmarkInput } from '../types'

export type BookmarkViewMode = 'list' | 'cards' | 'compact'

export const useBookmarkLibrary = () => {
  const store = useBookmarkStore()
  const editorOpen = ref(false)
  const viewMode = ref<BookmarkViewMode>('list')

  const items = computed(() => store.items)
  const folders = computed(() => store.folders)
  const tags = computed(() => store.tags)
  const total = computed(() => store.total)
  const selected = computed(() => store.selected)
  const draft = computed(() => store.draft)
  const status = computed(() => store.status)
  const selectedId = computed(() => store.selected?.id ?? '')
  const activeFilters = computed(() => [store.folder, store.tag].filter(Boolean))
  const query = computed({
    get: () => store.query,
    set: (value: string) => store.setQuery(value)
  })
  const folder = computed({
    get: () => store.folder,
    set: (value: string) => store.setFolder(value)
  })
  const tag = computed({
    get: () => store.tag,
    set: (value: string) => store.setTag(value)
  })

  const openEditor = (item: Bookmark | null) => {
    store.select(item)
    editorOpen.value = true
  }

  const closeEditor = () => {
    editorOpen.value = false
  }

  const createBookmark = () => {
    openEditor(null)
  }

  const selectBookmark = (item: Bookmark) => {
    openEditor(item)
  }

  const clearFilters = () => {
    store.clearFilters()
  }

  const updateDraft = (patch: Partial<BookmarkInput>) => {
    store.updateDraft(patch)
  }

  const save = async () => {
    await store.save()
  }

  const removeSelected = async () => {
    await store.removeSelected()
    closeEditor()
  }

  const analyzeSelected = async () => {
    await store.analyzeSelected()
  }

  return {
    items,
    folders,
    tags,
    total,
    selected,
    draft,
    status,
    selectedId,
    activeFilters,
    query,
    folder,
    tag,
    editorOpen,
    viewMode,
    clearFilters,
    closeEditor,
    createBookmark,
    selectBookmark,
    updateDraft,
    save,
    removeSelected,
    analyzeSelected
  }
}
