<!-- 文件说明：frontend/src/pages/SettingsPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed } from 'vue'
import { NAlert, NCard, NDescriptions, NDescriptionsItem, NIcon, NRadioButton, NRadioGroup, NSpace, NSwitch, NTag } from 'naive-ui'
import { useSettingsSummary } from '../composables/useSettingsSummary'
import { useThemePreference } from '../composables/useThemePreference'
import { appIcons } from '../icons'

const summary = useSettingsSummary()
const theme = useThemePreference()
const runtimeFeatureLabel = computed(() => '前端 Runtime 适配已启用')

const updateThemeMode = (value: string | number | boolean) => {
  theme.setThemeMode(value === 'dark' ? 'dark' : 'light')
}
</script>

<template>
  <section class="page settings-page split-page">
    <NCard class="settings-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">THEME</span>
          <h2>外观</h2>
        </div>
      </template>
      <NSpace vertical :size="18">
        <NRadioGroup :value="theme.mode.value" @update:value="updateThemeMode">
          <NSpace>
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
        <NSpace align="center">
          <span>快速切换</span>
          <NSwitch :value="theme.mode.value === 'dark'" @update:value="theme.setThemeMode($event ? 'dark' : 'light')" />
          <NTag round>{{ theme.mode.value === 'dark' ? '深色模式' : '浅色模式' }}</NTag>
        </NSpace>
      </NSpace>
    </NCard>

    <NCard class="settings-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">DATA</span>
          <h2>数据概览</h2>
        </div>
      </template>
      <NDescriptions :column="1" bordered>
        <NDescriptionsItem label="书签数量">{{ summary.total.value }}</NDescriptionsItem>
        <NDescriptionsItem label="分类数量">{{ summary.folderCount.value }}</NDescriptionsItem>
        <NDescriptionsItem label="标签数量">{{ summary.tagCount.value }}</NDescriptionsItem>
      </NDescriptions>
    </NCard>

    <NCard class="settings-card" :bordered="false">
      <template #header>快捷键与桌面能力</template>
      <NDescriptions :column="1" bordered>
        <NDescriptionsItem label="全局搜索">Ctrl + K</NDescriptionsItem>
        <NDescriptionsItem label="窗口控制">{{ runtimeFeatureLabel }}</NDescriptionsItem>
        <NDescriptionsItem label="托盘能力">后续 Go/Wails 扩展</NDescriptionsItem>
      </NDescriptions>
      <NAlert class="settings-alert" type="info" :show-icon="false">Cola Bookmarks 是本地优先的桌面书签管理器。</NAlert>
    </NCard>
  </section>
</template>
