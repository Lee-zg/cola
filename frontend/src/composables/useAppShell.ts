import { computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { routes } from '../navigation'
import { useHashRoute } from '../router'
import { useBookmarkStore } from '../stores/bookmarks'
import { useUiCommands } from './useUiCommands'

export const useAppShell = () => {
  const store = useBookmarkStore()
  const { currentPath, navigate } = useHashRoute()
  const uiCommands = useUiCommands()
  const currentTitle = computed(() => routes.find((route) => route.path === currentPath.value)?.title ?? '仪表盘')
  const query = computed({
    get: () => store.query,
    set: (value: string) => store.setQuery(value)
  })

  let searchTimer = 0

  watch(
    () => [store.query, store.folder, store.tag],
    () => {
      window.clearTimeout(searchTimer)
      searchTimer = window.setTimeout(() => store.refresh(), 220)
    }
  )

  const refresh = async () => {
    try {
      await store.refresh()
    } catch (err) {
      store.setStatus(err instanceof Error ? err.message : String(err))
    }
  }

  const createBookmark = () => {
    navigate('/library')
    uiCommands.requestCreateBookmark()
  }

  const analyzeAll = async () => {
    await store.analyzeAll()
  }

  const focusGlobalSearch = () => {
    document.querySelector<HTMLInputElement>('#global-search-input')?.focus()
  }

  const handleShortcut = (event: KeyboardEvent) => {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault()
      focusGlobalSearch()
    }
  }

  onMounted(() => {
    refresh()
    window.addEventListener('keydown', handleShortcut)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleShortcut)
  })

  return {
    currentPath,
    currentTitle,
    loading: computed(() => store.loading),
    query,
    server: computed(() => store.server),
    status: computed(() => store.status),
    total: computed(() => store.total),
    analyzeAll,
    createBookmark,
    navigate,
    refresh
  }
}
