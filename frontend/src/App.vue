<script setup lang="ts">
import { computed, type Component } from 'vue'
import NavRail from './components/NavRail.vue'
import StatusBar from './components/StatusBar.vue'
import TopBar from './components/TopBar.vue'
import { useAppShell } from './composables/useAppShell'
import { type RoutePath } from './navigation'
import DashboardPage from './pages/DashboardPage.vue'
import AiPage from './pages/AiPage.vue'
import BackupPage from './pages/BackupPage.vue'
import ExportPage from './pages/ExportPage.vue'
import ImportPage from './pages/ImportPage.vue'
import LibraryPage from './pages/LibraryPage.vue'
import SettingsPage from './pages/SettingsPage.vue'
import TaxonomyPage from './pages/TaxonomyPage.vue'
import WebServerPage from './pages/WebServerPage.vue'

const shell = useAppShell()

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

const currentComponent = computed(() => routeComponents[shell.currentPath.value])
</script>

<template>
  <div class="app-shell">
    <NavRail :active-path="shell.currentPath.value" @navigate="shell.navigate" />

    <main class="app-main">
      <TopBar
        :active-path="shell.currentPath.value"
        :loading="shell.loading.value"
        :query="shell.query.value"
        :title="shell.currentTitle.value"
        @analyze="shell.analyzeAll"
        @create="shell.createBookmark"
        @refresh="shell.refresh"
        @update:query="shell.query.value = $event"
      />

      <div class="workspace">
        <component :is="currentComponent" @navigate="shell.navigate" />
      </div>

      <StatusBar
        :loading="shell.loading.value"
        :server="shell.server.value"
        :status="shell.status.value"
        :total="shell.total.value"
      />
    </main>
  </div>
</template>
