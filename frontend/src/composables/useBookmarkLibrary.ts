// useBookmarkLibrary 封装书签库页面的筛选、视图模式和编辑抽屉状态。
import { computed, ref } from 'vue'
import { type CategoryReorderDirection, type LibraryViewMode, useBookmarkStore } from '../stores/bookmarks'
import type { Bookmark, BookmarkInput } from '../types'

export type BookmarkViewMode = LibraryViewMode

// 该 composable 只编排页面交互，真实持久化和刷新仍交给 Pinia store。
export const useBookmarkLibrary = () => {
  const store = useBookmarkStore()
  const editorOpen = ref(false)

  const items = computed(() => store.items)
  const folders = computed(() => store.folders)
  const categories = computed(() => store.categories)
  const tags = computed(() => store.tags)
  const total = computed(() => store.total)
  const loadedCount = computed(() => store.loadedCount)
  const offset = computed(() => store.offset)
  const pageSize = computed(() => store.pageSize)
  const loading = computed(() => store.loading)
  const loadingMode = computed(() => store.loadingMode)
  const selected = computed(() => store.selected)
  const draft = computed(() => store.draft)
  const status = computed(() => store.status)
  const selectedId = computed(() => store.selected?.id ?? '')
  const activeFilters = computed(() => [store.categoryId, store.folder].filter(Boolean))
  const viewMode = computed({
    get: () => store.viewMode,
    set: (value: BookmarkViewMode) => store.setViewMode(value)
  })
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
  const categoryId = computed({
    get: () => store.categoryId || 'category_all',
    set: (value: string) => store.setCategory(value)
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

  const openBookmark = async (item: Bookmark, event: MouseEvent) => {
    if (!event.ctrlKey) {
      selectBookmark(item)
      return
    }
    event.preventDefault()
    await store.openBookmark(item.id)
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

  const loadMore = async () => {
    await store.loadMore()
  }

  const setPage = async (page: number) => {
    await store.setPage(page)
  }

  const createCategory = async (name: string, parentId: string) => {
    await store.createCategory(name, parentId)
  }

  const renameCategory = async (id: string, name: string) => {
    await store.renameCategory(id, name)
  }

  const moveCategory = async (id: string, parentId: string) => {
    await store.moveCategory(id, parentId)
  }

  const deleteCategory = async (id: string, deleteBookmarks = false) => {
    await store.deleteCategory(id, deleteBookmarks)
  }

  const reorderCategory = async (id: string, direction: CategoryReorderDirection) => {
    await store.reorderCategory(id, direction)
  }

  const savePreview = async (path: string) => {
    await store.savePreview(path)
  }

  const fetchPreview = async (id?: string) => {
    await store.fetchPreview(id)
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
    categories,
    tags,
    total,
    loadedCount,
    offset,
    pageSize,
    loading,
    loadingMode,
    selected,
    draft,
    status,
    selectedId,
    activeFilters,
    query,
    folder,
    tag,
    categoryId,
    editorOpen,
    viewMode,
    clearFilters,
    closeEditor,
    createBookmark,
    selectBookmark,
    openBookmark,
    updateDraft,
    loadMore,
    setPage,
    createCategory,
    renameCategory,
    moveCategory,
    deleteCategory,
    reorderCategory,
    savePreview,
    fetchPreview,
    save,
    removeSelected,
    analyzeSelected
  }
}
