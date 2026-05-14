import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { routePaths, type RoutePath } from './navigation'

const isRoutePath = (path: string): path is RoutePath => routePaths.includes(path as RoutePath)

export const normalizeHashPath = (hash: string): RoutePath => {
  const rawPath = hash.replace(/^#/, '').split('?')[0] || '/'
  const normalized = rawPath.startsWith('/') ? rawPath : `/${rawPath}`

  if (normalized.startsWith('/library/')) {
    return '/library'
  }

  return isRoutePath(normalized) ? normalized : '/'
}

export const useHashRoute = () => {
  const currentPath = ref<RoutePath>(normalizeHashPath(window.location.hash))

  const syncFromHash = () => {
    currentPath.value = normalizeHashPath(window.location.hash)
  }

  const navigate = (path: RoutePath) => {
    if (currentPath.value === path) return
    window.location.hash = path
  }

  onMounted(() => {
    if (!window.location.hash) {
      window.location.hash = '/'
    }
    syncFromHash()
    window.addEventListener('hashchange', syncFromHash)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('hashchange', syncFromHash)
  })

  return {
    currentPath: computed(() => currentPath.value),
    navigate
  }
}
