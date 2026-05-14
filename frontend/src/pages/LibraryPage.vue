<script setup lang="ts">
import { computed, ref } from 'vue'
import BookmarkEditorDrawer from '../components/BookmarkEditorDrawer.vue'
import { useBookmarkStore } from '../stores/bookmarks'
import type { Bookmark } from '../types'

const store = useBookmarkStore()
const viewMode = ref<'list' | 'cards' | 'compact'>('list')

const selectedId = computed(() => store.selected?.id ?? '')
const activeFilters = computed(() => [store.folder, store.tag].filter(Boolean))

const choose = (item: Bookmark) => {
  store.select(item, true)
}

const createBookmark = () => {
  store.select(null, true)
}

const clearFilters = () => {
  store.folder = ''
  store.tag = ''
}

defineExpose({ createBookmark })
</script>

<template>
  <section class="page library-page">
    <aside class="filter-panel surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Filters</span>
          <h2>筛选</h2>
        </div>
        <button type="button" :disabled="!activeFilters.length" @click="clearFilters">清空</button>
      </div>

      <section class="filter-block">
        <h3>分类</h3>
        <button class="filter-chip" :class="{ active: !store.folder }" type="button" @click="store.folder = ''">
          全部
        </button>
        <button
          v-for="folder in store.folders"
          :key="folder"
          class="filter-chip"
          :class="{ active: store.folder === folder }"
          type="button"
          @click="store.folder = folder"
        >
          {{ folder }}
        </button>
      </section>

      <section class="filter-block">
        <h3>标签</h3>
        <button class="filter-chip" :class="{ active: !store.tag }" type="button" @click="store.tag = ''">
          全部
        </button>
        <button
          v-for="tag in store.tags"
          :key="tag"
          class="filter-chip"
          :class="{ active: store.tag === tag }"
          type="button"
          @click="store.tag = tag"
        >
          {{ tag }}
        </button>
      </section>

      <button class="primary-action wide" type="button" @click="createBookmark">新建书签</button>
    </aside>

    <main class="library-workspace surface">
      <div class="library-toolbar">
        <div>
          <span class="eyebrow">Library</span>
          <h2>{{ store.total }} 个书签</h2>
        </div>
        <div class="segmented" aria-label="视图切换">
          <button type="button" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">列表</button>
          <button type="button" :class="{ active: viewMode === 'cards' }" @click="viewMode = 'cards'">卡片</button>
          <button type="button" :class="{ active: viewMode === 'compact' }" @click="viewMode = 'compact'">
            紧凑
          </button>
        </div>
      </div>

      <div class="bookmark-list" :class="`mode-${viewMode}`">
        <button
          v-for="item in store.items"
          :key="item.id"
          class="bookmark-row"
          :class="{ active: selectedId === item.id }"
          type="button"
          @click="choose(item)"
        >
          <span class="title">{{ item.title }}</span>
          <span class="url">{{ item.url }}</span>
          <span class="meta">
            <span>{{ item.folder }}</span>
            <span v-for="tag in item.tags" :key="tag" class="tag">{{ tag }}</span>
          </span>
        </button>
        <div v-if="!store.items.length" class="empty">没有匹配的书签</div>
      </div>
    </main>

    <BookmarkEditorDrawer :open="store.editorOpen" @close="store.editorOpen = false" />
  </section>
</template>
