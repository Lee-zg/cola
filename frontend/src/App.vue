<!-- 文件说明：frontend/src/App.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed, type Component } from 'vue'
import {
  darkTheme as naiveDarkTheme,
  NConfigProvider,
  NDialogProvider,
  NLoadingBarProvider,
  NMessageProvider
} from 'naive-ui'
import NavRail from './components/NavRail.vue'
import StatusBar from './components/StatusBar.vue'
import TopBar from './components/TopBar.vue'
import WindowTitleBar from './components/WindowTitleBar.vue'
import { useDesktopRuntime } from './composables/useDesktopRuntime'
import { useAppShell } from './composables/useAppShell'
import { useThemePreference } from './composables/useThemePreference'
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
import { darkThemeOverrides, lightThemeOverrides } from './theme/tokens'

const shell = useAppShell()
const theme = useThemePreference()
const runtime = useDesktopRuntime()

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
const activeNaiveTheme = computed(() => (theme.isDark.value ? naiveDarkTheme : null))
const activeThemeOverrides = computed(() => (theme.isDark.value ? darkThemeOverrides : lightThemeOverrides))
</script>

<template>
  <NConfigProvider :theme="activeNaiveTheme" :theme-overrides="activeThemeOverrides">
    <NMessageProvider placement="bottom-right">
      <NDialogProvider>
        <NLoadingBarProvider>
          <div class="desktop-frame">
            <WindowTitleBar
              :runtime-available="runtime.isAvailable"
              @close="runtime.closeWindow"
              @minimise="runtime.minimise"
              @toggle-maximise="runtime.toggleMaximise"
            />

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
                  <Transition name="route-fade" mode="out-in">
                    <component :is="currentComponent" :key="shell.currentPath.value" @navigate="shell.navigate" />
                  </Transition>
                </div>

                <StatusBar
                  :loading="shell.loading.value"
                  :server="shell.server.value"
                  :status="shell.status.value"
                  :total="shell.total.value"
                />
              </main>
            </div>
          </div>
        </NLoadingBarProvider>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>
