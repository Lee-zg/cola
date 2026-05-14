<script setup lang="ts">
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

const updateQuery = (event: Event) => {
  emit('update:query', (event.target as HTMLInputElement).value)
}
</script>

<template>
  <header class="topbar">
    <div class="topbar-title">
      <span class="eyebrow">Cola Bookmarks</span>
      <h1>{{ props.title }}</h1>
    </div>

    <label class="global-search" title="Ctrl+K 聚焦搜索">
      <span>搜索</span>
      <input
        id="global-search-input"
        :value="props.query"
        placeholder="标题、网址、标签、关键词"
        @input="updateQuery"
      />
      <kbd>Ctrl K</kbd>
    </label>

    <div class="topbar-actions">
      <button v-if="props.activePath === '/library'" class="primary-action" type="button" @click="emit('create')">
        新建
      </button>
      <button v-if="props.activePath === '/ai'" type="button" @click="emit('analyze')">全部分析</button>
      <button type="button" :disabled="props.loading" @click="emit('refresh')">
        {{ props.loading ? '刷新中' : '刷新' }}
      </button>
    </div>
  </header>
</template>
