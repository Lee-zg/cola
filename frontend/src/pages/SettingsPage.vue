<!-- 文件说明：frontend/src/pages/SettingsPage.vue，承载多级设置导航、分组配置和偏好保存。 -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NDescriptions,
  NDescriptionsItem,
  NIcon,
  NRadioButton,
  NRadioGroup,
  NSpace,
  NSwitch,
  NTag
} from 'naive-ui'
import { useBookmarkStore } from '../stores/bookmarks'
import { useSettingsSummary } from '../composables/useSettingsSummary'
import { useThemePreference } from '../composables/useThemePreference'
import { appIcons, type AppIconKey } from '../icons'
import type { RoutePath } from '../navigation'
import type { AppPreferences } from '../types'
import type { ThemeMode } from '../theme/tokens'

type SettingsSectionKey = 'general' | 'appearance' | 'data' | 'ai' | 'network'

interface SettingsSubSection {
  key: string
  label: string
  description: string
  disabled?: boolean
}

interface SettingsSection {
  key: SettingsSectionKey
  label: string
  eyebrow: string
  icon: AppIconKey
  description: string
  children: SettingsSubSection[]
}

const defaultPreferences = (): AppPreferences => ({
  openBrowser: 'default',
  lazyFetchThumbnails: true
})

const browserOptions = [
  { label: '系统默认', value: 'default' },
  { label: 'Chrome', value: 'chrome' },
  { label: 'Edge', value: 'edge' },
  { label: 'Firefox', value: 'firefox' }
]

const settingsSections: SettingsSection[] = [
  {
    key: 'general',
    label: '常规设置',
    eyebrow: 'GENERAL',
    icon: 'settings',
    description: '配置书签打开方式、缩略图加载和桌面交互。',
    children: [
      { key: 'browser', label: '打开方式', description: '书签链接调用的浏览器。' },
      { key: 'thumbnails', label: '缩略图', description: '列表缩略图的后台补齐策略。' },
      { key: 'shortcuts', label: '快捷键', description: '桌面端高频操作提示。' }
    ]
  },
  {
    key: 'appearance',
    label: '外观设置',
    eyebrow: 'APPEARANCE',
    icon: 'sun',
    description: '切换界面主题和查看当前视觉状态。',
    children: [
      { key: 'theme', label: '主题模式', description: '浅色和深色主题立即生效。' }
    ]
  },
  {
    key: 'data',
    label: '数据管理',
    eyebrow: 'DATA',
    icon: 'stats',
    description: '查看当前数据规模，并定位到导入、导出和备份能力。',
    children: [
      { key: 'summary', label: '数据概览', description: '书签、分类和标签的当前规模。' },
      { key: 'maintenance', label: '维护入口', description: '导入、导出与备份功能入口。' }
    ]
  },
  {
    key: 'ai',
    label: 'AI 设置',
    eyebrow: 'AI',
    icon: 'ai',
    description: '预留 AI 分析、摘要和自动标签策略。',
    children: [
      { key: 'ai-placeholder', label: '能力预留', description: '后续接入 AI 配置项。', disabled: true }
    ]
  },
  {
    key: 'network',
    label: '网络设置',
    eyebrow: 'NETWORK',
    icon: 'web',
    description: '预留本地 Web 服务、代理和远程访问策略。',
    children: [
      { key: 'network-placeholder', label: '能力预留', description: '后续接入网络配置项。', disabled: true }
    ]
  }
]

const summary = useSettingsSummary()
const store = useBookmarkStore()
const theme = useThemePreference()
const emit = defineEmits<{
  navigate: [path: RoutePath]
}>()
const activeSectionKey = ref<SettingsSectionKey>('general')
const draftPreferences = ref<AppPreferences>(defaultPreferences())
const saving = ref(false)

const activeSection = computed(
  () => settingsSections.find((section) => section.key === activeSectionKey.value) || settingsSections[0]
)

const normalizedPreferences = computed(() => ({
  ...defaultPreferences(),
  ...store.preferences
}))

const defaultPreferenceState = computed(() => defaultPreferences())

const hasPreferenceChanges = computed(
  () =>
    draftPreferences.value.openBrowser !== normalizedPreferences.value.openBrowser ||
    draftPreferences.value.lazyFetchThumbnails !== normalizedPreferences.value.lazyFetchThumbnails
)

const canResetPreferences = computed(
  () =>
    draftPreferences.value.openBrowser !== defaultPreferenceState.value.openBrowser ||
    draftPreferences.value.lazyFetchThumbnails !== defaultPreferenceState.value.lazyFetchThumbnails
)

const activeBrowserLabel = computed(
  () => browserOptions.find((option) => option.value === draftPreferences.value.openBrowser)?.label || '系统默认'
)

const selectSection = (key: SettingsSectionKey) => {
  activeSectionKey.value = key
}

const setThemeMode = (value: string | number | boolean) => {
  theme.setThemeMode(value === 'dark' ? 'dark' : 'light')
}

const patchDraftPreferences = (patch: Partial<AppPreferences>) => {
  draftPreferences.value = {
    ...draftPreferences.value,
    ...patch
  }
}

const savePreferences = async () => {
  if (!hasPreferenceChanges.value || saving.value) return
  saving.value = true
  try {
    await store.savePreferences(draftPreferences.value)
  } finally {
    saving.value = false
  }
}

const cancelPreferenceChanges = () => {
  draftPreferences.value = { ...normalizedPreferences.value }
}

const resetPreferences = () => {
  draftPreferences.value = defaultPreferences()
}

const navigateTo = (path: RoutePath) => {
  emit('navigate', path)
}

// store 刷新后同步草稿；用户未保存时保留当前草稿，避免后台刷新覆盖编辑。
watch(
  normalizedPreferences,
  (preferences) => {
    if (!hasPreferenceChanges.value) {
      draftPreferences.value = { ...preferences }
    }
  },
  { immediate: true }
)
</script>

<template>
  <section class="page settings-page">
    <aside class="settings-nav panel-surface" aria-label="设置分类">
      <div class="settings-nav-heading">
        <span class="eyebrow">SETTINGS</span>
        <h2>设置中心</h2>
        <p>按类别管理应用偏好，新增设置时可继续扩展当前层级。</p>
      </div>

      <nav class="settings-section-list">
        <button
          v-for="section in settingsSections"
          :key="section.key"
          class="settings-section-button"
          :class="{ active: section.key === activeSection.key }"
          type="button"
          @click="selectSection(section.key)"
        >
          <NIcon :component="appIcons[section.icon]" />
          <span>
            <strong>{{ section.label }}</strong>
            <small>{{ section.description }}</small>
          </span>
        </button>
      </nav>

      <div class="settings-subnav">
        <span class="settings-subnav-title">{{ activeSection.label }}</span>
        <button
          v-for="child in activeSection.children"
          :key="child.key"
          class="settings-subnav-item"
          :class="{ disabled: child.disabled }"
          type="button"
          :disabled="child.disabled"
        >
          <span>{{ child.label }}</span>
          <small>{{ child.description }}</small>
        </button>
      </div>
    </aside>

    <main class="settings-detail panel-surface">
      <header class="settings-detail-header">
        <div>
          <span class="eyebrow">{{ activeSection.eyebrow }}</span>
          <h2>{{ activeSection.label }}</h2>
          <p>{{ activeSection.description }}</p>
        </div>
        <NTag round :type="hasPreferenceChanges ? 'warning' : 'default'">
          {{ hasPreferenceChanges ? '有未保存更改' : '设置已同步' }}
        </NTag>
      </header>

      <div class="settings-detail-scroll">
        <template v-if="activeSection.key === 'general'">
          <section class="settings-group" aria-labelledby="settings-browser-title">
            <div class="settings-group-heading">
              <div>
                <span class="eyebrow">BROWSER</span>
                <h3 id="settings-browser-title">浏览器打开方式</h3>
              </div>
              <NTag round>{{ activeBrowserLabel }}</NTag>
            </div>
            <div class="settings-control-row">
              <div class="settings-control-copy">
                <strong>默认打开浏览器</strong>
                <span>找不到指定浏览器时会回退到系统默认浏览器。</span>
              </div>
              <NRadioGroup
                :value="draftPreferences.openBrowser"
                size="small"
                @update:value="patchDraftPreferences({ openBrowser: String($event) })"
              >
                <NSpace :size="8" wrap>
                  <NRadioButton v-for="option in browserOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </NRadioButton>
                </NSpace>
              </NRadioGroup>
            </div>
          </section>

          <section class="settings-group" aria-labelledby="settings-thumbnail-title">
            <div class="settings-group-heading">
              <div>
                <span class="eyebrow">THUMBNAILS</span>
                <h3 id="settings-thumbnail-title">缩略图</h3>
              </div>
              <NTag round>{{ draftPreferences.lazyFetchThumbnails === false ? '手动' : '自动' }}</NTag>
            </div>
            <div class="settings-control-row">
              <div class="settings-control-copy">
                <strong>列表中自动补齐缩略图</strong>
                <span>开启后会按 OpenGraph、Favicon、本地截图顺序缓存封面。</span>
              </div>
              <NSwitch
                :value="draftPreferences.lazyFetchThumbnails !== false"
                @update:value="patchDraftPreferences({ lazyFetchThumbnails: $event })"
              />
            </div>
          </section>

          <section class="settings-group" aria-labelledby="settings-shortcut-title">
            <div class="settings-group-heading">
              <div>
                <span class="eyebrow">SHORTCUTS</span>
                <h3 id="settings-shortcut-title">快捷键与桌面能力</h3>
              </div>
            </div>
            <NDescriptions label-placement="left" :column="1" bordered>
              <NDescriptionsItem label="全局搜索">Ctrl + K</NDescriptionsItem>
              <NDescriptionsItem label="打开书签">按住 Ctrl 点击书签标题或卡片</NDescriptionsItem>
              <NDescriptionsItem label="窗口控制">前端 Runtime 适配已启用</NDescriptionsItem>
            </NDescriptions>
          </section>
        </template>

        <template v-else-if="activeSection.key === 'appearance'">
          <section class="settings-group" aria-labelledby="settings-theme-title">
            <div class="settings-group-heading">
              <div>
                <span class="eyebrow">THEME</span>
                <h3 id="settings-theme-title">主题模式</h3>
              </div>
              <NTag round>{{ theme.mode.value === 'dark' ? '深色模式' : '浅色模式' }}</NTag>
            </div>
            <div class="settings-control-row">
              <div class="settings-control-copy">
                <strong>界面主题</strong>
                <span>主题设置保存在本地浏览器存储中，切换后立即生效。</span>
              </div>
              <NRadioGroup :value="theme.mode.value" size="small" @update:value="setThemeMode">
                <NSpace :size="8">
                  <NRadioButton value="light">
                    <NIcon :component="appIcons.sun" />
                    浅色
                  </NRadioButton>
                  <NRadioButton value="dark">
                    <NIcon :component="appIcons.moon" />
                    深色
                  </NRadioButton>
                </NSpace>
              </NRadioGroup>
            </div>
            <div class="settings-control-row">
              <div class="settings-control-copy">
                <strong>快速切换</strong>
                <span>用于验证深浅色下的卡片、按钮和边框层级。</span>
              </div>
              <NSwitch
                :value="theme.mode.value === 'dark'"
                @update:value="theme.setThemeMode($event ? ('dark' as ThemeMode) : ('light' as ThemeMode))"
              />
            </div>
          </section>
        </template>

        <template v-else-if="activeSection.key === 'data'">
          <section class="settings-group" aria-labelledby="settings-data-title">
            <div class="settings-group-heading">
              <div>
                <span class="eyebrow">OVERVIEW</span>
                <h3 id="settings-data-title">数据概览</h3>
              </div>
            </div>
            <div class="settings-metric-grid">
              <div class="settings-metric">
                <span>书签数量</span>
                <strong>{{ summary.total.value }}</strong>
              </div>
              <div class="settings-metric">
                <span>分类数量</span>
                <strong>{{ summary.categoryCount.value }}</strong>
              </div>
              <div class="settings-metric">
                <span>标签数量</span>
                <strong>{{ summary.tagCount.value }}</strong>
              </div>
            </div>
          </section>

          <section class="settings-group" aria-labelledby="settings-maintenance-title">
            <div class="settings-group-heading">
              <div>
                <span class="eyebrow">MAINTENANCE</span>
                <h3 id="settings-maintenance-title">维护入口</h3>
              </div>
            </div>
            <div class="settings-link-grid">
              <NButton secondary @click="navigateTo('/import')">
                <template #icon>
                  <NIcon :component="appIcons.import" />
                </template>
                导入中心
              </NButton>
              <NButton secondary @click="navigateTo('/export')">
                <template #icon>
                  <NIcon :component="appIcons.export" />
                </template>
                导出与主题
              </NButton>
              <NButton secondary @click="navigateTo('/backup')">
                <template #icon>
                  <NIcon :component="appIcons.backup" />
                </template>
                备份与恢复
              </NButton>
            </div>
          </section>
        </template>

        <template v-else>
          <section class="settings-group settings-placeholder" aria-label="预留设置">
            <NIcon :component="appIcons[activeSection.icon]" />
            <strong>{{ activeSection.label }}尚未开放</strong>
            <span>当前页面结构已支持多级设置。后续接入字段时，可在配置列表中新增分组和控件。</span>
          </section>
        </template>
      </div>

      <footer class="settings-actions">
        <NAlert v-if="activeSection.key === 'appearance'" type="info" :show-icon="false">
          外观设置立即生效；底部按钮只处理浏览器和缩略图等应用偏好。
        </NAlert>
        <span v-else class="settings-actions-note">保存后会调用现有偏好设置接口。</span>
        <NSpace :size="8">
          <NButton secondary :disabled="!hasPreferenceChanges || saving" @click="cancelPreferenceChanges">取消</NButton>
          <NButton secondary :disabled="!canResetPreferences || saving" @click="resetPreferences">重置</NButton>
          <NButton type="primary" :loading="saving" :disabled="!hasPreferenceChanges" @click="savePreferences">
            <template #icon>
              <NIcon :component="appIcons.save" />
            </template>
            保存
          </NButton>
        </NSpace>
      </footer>
    </main>
  </section>
</template>
