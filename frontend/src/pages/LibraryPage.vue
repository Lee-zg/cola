<script setup lang="ts">
import { watch } from 'vue'
import BookmarkEditorDrawer from '../components/BookmarkEditorDrawer.vue'
import { useBookmarkLibrary } from '../composables/useBookmarkLibrary'
import { useUiCommands } from '../composables/useUiCommands'

const library = useBookmarkLibrary()
const uiCommands = useUiCommands()

watch(
  uiCommands.createBookmarkRequestId,
  (requestId) => {
    if (requestId > 0) {
      library.createBookmark()
    }
  },
  { flush: 'post' }
)
</script>

<template>
  <section class="page library-page">
    <aside class="filter-panel surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Filters</span>
          <h2>筛选</h2>
        </div>
        <button type="button" :disabled="!library.activeFilters.value.length" @click="library.clearFilters">清空</button>
      </div>

      <section class="filter-block">
        <h3>分类</h3>
        <button class="filter-chip" :class="{ active: !library.folder.value }" type="button" @click="library.folder.value = ''">
          全部
        </button>
        <button
          v-for="folder in library.folders.value"
          :key="folder"
          class="filter-chip"
          :class="{ active: library.folder.value === folder }"
          type="button"
          @click="library.folder.value = folder"
        >
          {{ folder }}
        </button>
      </section>

      <section class="filter-block">
        <h3>标签</h3>
        <button class="filter-chip" :class="{ active: !library.tag.value }" type="button" @click="library.tag.value = ''">
          全部
        </button>
        <button
          v-for="tag in library.tags.value"
          :key="tag"
          class="filter-chip"
          :class="{ active: library.tag.value === tag }"
          type="button"
          @click="library.tag.value = tag"
        >
          {{ tag }}
        </button>
      </section>

      <button class="primary-action wide" type="button" @click="library.createBookmark">新建书签</button>
    </aside>

    <main class="library-workspace surface">
      <div class="library-toolbar">
        <div>
          <span class="eyebrow">Library</span>
          <h2>{{ library.total.value }} 个书签</h2>
        </div>
        <div class="segmented" aria-label="视图切换">
          <button type="button" :class="{ active: library.viewMode.value === 'list' }" @click="library.viewMode.value = 'list'">列表</button>
          <button type="button" :class="{ active: library.viewMode.value === 'cards' }" @click="library.viewMode.value = 'cards'">卡片</button>
          <button type="button" :class="{ active: library.viewMode.value === 'compact' }" @click="library.viewMode.value = 'compact'">
            紧凑
          </button>
        </div>
      </div>

      <div class="bookmark-list" :class="`mode-${library.viewMode.value}`">
        <button
          v-for="item in library.items.value"
          :key="item.id"
          class="bookmark-row"
          :class="{ active: library.selectedId.value === item.id }"
          type="button"
          @click="library.selectBookmark(item)"
        >
          <span class="title">{{ item.title }}</span>
          <span class="url">{{ item.url }}</span>
          <span class="meta">
            <span>{{ item.folder }}</span>
            <span v-for="tag in item.tags" :key="tag" class="tag">{{ tag }}</span>
          </span>
        </button>
        <div v-if="!library.items.value.length" class="empty">没有匹配的书签</div>
      </div>
    </main>

    <BookmarkEditorDrawer
      :draft="library.draft.value"
      :open="library.editorOpen.value"
      :selected="library.selected.value"
      :status="library.status.value"
      @analyze="library.analyzeSelected"
      @close="library.closeEditor"
      @remove="library.removeSelected"
      @save="library.save"
      @update:draft="library.updateDraft"
    />
  </section>
</template>
