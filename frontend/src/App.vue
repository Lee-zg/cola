<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, watch, type Component } from 'vue'
import NavRail from './components/NavRail.vue'
import StatusBar from './components/StatusBar.vue'
import TopBar from './components/TopBar.vue'
import { routes, type RoutePath } from './navigation'
import DashboardPage from './pages/DashboardPage.vue'
import AiPage from './pages/AiPage.vue'
import BackupPage from './pages/BackupPage.vue'
import ExportPage from './pages/ExportPage.vue'
import ImportPage from './pages/ImportPage.vue'
import LibraryPage from './pages/LibraryPage.vue'
import SettingsPage from './pages/SettingsPage.vue'
import TaxonomyPage from './pages/TaxonomyPage.vue'
import WebServerPage from './pages/WebServerPage.vue'
import { useHashRoute } from './router'
import { useBookmarkStore } from './stores/bookmarks'

const store = useBookmarkStore()
const { currentPath, navigate } = useHashRoute()

const routeComponents: Record<RoutePath, Component> = {
  '/': DashboardPage,
  '/library': LibraryPage,
  '/taxonomy': TaxonomyPage,
  '/import': ImportPage,
  '/export': ExportPage,
  '/webserver': WebServerPage,
  '/ai': AiPage,
  '/backup': BackupPage,
  '/settings': SettingsPage
}

const currentComponent = computed(() => routeComponents[currentPath.value])
const currentTitle = computed(() => routes.find((route) => route.path === currentPath.value)?.title ?? '仪表盘')

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
    store.status = err instanceof Error ? err.message : String(err)
  }
}

const createBookmark = () => {
  navigate('/library')
  store.select(null, true)
}

const analyzeAll = async () => {
  await store.analyzeAll()
}

const handleShortcut = (event: KeyboardEvent) => {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    document.querySelector<HTMLInputElement>('#global-search-input')?.focus()
  }
}

onMounted(() => {
  refresh()
  window.addEventListener('keydown', handleShortcut)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleShortcut)
})
</script>

<template>
  <div class="app-shell">
    <NavRail :active-path="currentPath" @navigate="navigate" />

    <main class="app-main">
      <TopBar
        :active-path="currentPath"
        :loading="store.loading"
        :query="store.query"
        :title="currentTitle"
        @analyze="analyzeAll"
        @create="createBookmark"
        @refresh="refresh"
        @update:query="store.query = $event"
      />

      <div class="workspace">
        <component :is="currentComponent" @navigate="navigate" />
      </div>

      <StatusBar :loading="store.loading" :server="store.server" :status="store.status" :total="store.total" />
    </main>
  </div>
</template>
