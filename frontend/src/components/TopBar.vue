<!-- 文件说明：frontend/src/components/TopBar.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { NButton, NIcon, NInput, NSpace, NTooltip } from 'naive-ui'
import { appIcons } from '../icons'
import type { RoutePath } from '../navigation'

const props = defineProps<{
  title: string
  activePath: RoutePath
  query: string
  loading: boolean
}>()

const emit = defineEmits<{
  'update:query': [value: string]
  refresh: []
  create: []
  analyze: []
}>()

const updateQuery = (value: string) => emit('update:query', value)
</script>

<template>
  <header class="topbar">
    <div class="topbar-title">
      <span class="eyebrow">Cola Bookmarks</span>
      <h1>{{ props.title }}</h1>
    </div>

    <div class="global-search" title="Ctrl+K 聚焦搜索">
      <NInput
        id="global-search-input"
        :value="props.query"
        clearable
        placeholder="标题、网址、标签、关键词"
        @update:value="updateQuery"
      >
        <template #prefix>
          <NIcon :component="appIcons.search" />
        </template>
      </NInput>
      <kbd>Ctrl K</kbd>
    </div>

    <NSpace class="topbar-actions" :size="10" align="center">
      <NButton v-if="props.activePath === '/library'" type="primary" @click="emit('create')">
        <template #icon>
          <NIcon :component="appIcons.add" />
        </template>
        新建
      </NButton>
      <NButton v-if="props.activePath === '/ai'" tertiary @click="emit('analyze')">
        <template #icon>
          <NIcon :component="appIcons.ai" />
        </template>
        全部分析
      </NButton>
      <NTooltip trigger="hover">
        <template #trigger>
          <NButton :loading="props.loading" circle secondary aria-label="刷新" @click="emit('refresh')">
            <template #icon>
              <NIcon :component="appIcons.refresh" />
            </template>
          </NButton>
        </template>
        刷新数据
      </NTooltip>
    </NSpace>
  </header>
</template>
