<!-- 文件说明：frontend/src/pages/LibraryPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed, watch } from 'vue'
import {
  NButton,
  NCard,
  NEmpty,
  NIcon,
  NInput,
  NRadioButton,
  NRadioGroup,
  NScrollbar,
  NSpace,
  NTag,
  NThing
} from 'naive-ui'
import BookmarkEditorDrawer from '../components/BookmarkEditorDrawer.vue'
import { useBookmarkLibrary } from '../composables/useBookmarkLibrary'
import { useUiCommands } from '../composables/useUiCommands'
import { appIcons } from '../icons'

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

const viewOptions = [
  { label: '列表', value: 'list' },
  { label: '卡片', value: 'cards' },
  { label: '紧凑', value: 'compact' }
]

const folderOptions = computed(() => [''].concat(library.folders.value))
const tagOptions = computed(() => [''].concat(library.tags.value))
</script>

<template>
  <section class="page library-page">
    <aside class="library-sidebar panel-surface">
      <div class="panel-heading">
        <div>
          <span class="eyebrow">FILTERS</span>
          <h2>筛选</h2>
        </div>
        <NButton size="small" secondary :disabled="!library.activeFilters.value.length" @click="library.clearFilters">
          清空
        </NButton>
      </div>

      <section class="filter-block">
        <h3>分类</h3>
        <NScrollbar class="filter-scroll" trigger="none">
          <button
            v-for="folder in folderOptions"
            :key="folder || 'all-folders'"
            class="filter-chip"
            :class="{ active: library.folder.value === folder }"
            type="button"
            @click="library.folder.value = folder"
          >
            <NIcon :component="appIcons.folder" />
            <span>{{ folder || '全部分类' }}</span>
          </button>
        </NScrollbar>
      </section>

      <section class="filter-block">
        <h3>标签</h3>
        <NScrollbar class="filter-scroll" trigger="none">
          <button
            v-for="tag in tagOptions"
            :key="tag || 'all-tags'"
            class="filter-chip"
            :class="{ active: library.tag.value === tag }"
            type="button"
            @click="library.tag.value = tag"
          >
            <NIcon :component="appIcons.tags" />
            <span>{{ tag || '全部标签' }}</span>
          </button>
        </NScrollbar>
      </section>

      <NButton class="wide-action" type="primary" block @click="library.createBookmark">
        <template #icon>
          <NIcon :component="appIcons.add" />
        </template>
        新建书签
      </NButton>
    </aside>

    <main class="library-workspace panel-surface">
      <div class="library-toolbar">
        <div>
          <span class="eyebrow">LIBRARY</span>
          <h2>{{ library.total.value }} 个书签</h2>
        </div>
        <NSpace align="center" :size="10">
          <NInput
            class="library-inline-search"
            :value="library.query.value"
            clearable
            placeholder="在书签库中搜索"
            @update:value="library.query.value = $event"
          >
            <template #prefix>
              <NIcon :component="appIcons.search" />
            </template>
          </NInput>
          <NRadioGroup v-model:value="library.viewMode.value" size="small">
            <NRadioButton v-for="option in viewOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </NRadioButton>
          </NRadioGroup>
        </NSpace>
      </div>

      <NScrollbar class="library-scroll">
        <div v-if="library.items.value.length" class="bookmark-list" :class="`mode-${library.viewMode.value}`">
          <NCard
            v-for="item in library.items.value"
            :key="item.id"
            class="bookmark-card"
            :class="{ active: library.selectedId.value === item.id }"
            :bordered="false"
            size="small"
            hoverable
            @click="library.selectBookmark(item)"
          >
            <NThing>
              <template #avatar>
                <div class="bookmark-favicon">{{ (item.domain || item.title || '?').slice(0, 1).toUpperCase() }}</div>
              </template>
              <template #header>{{ item.title || item.url }}</template>
              <template #description>
                <span class="bookmark-url">{{ item.url }}</span>
              </template>
              <template #footer>
                <NSpace :size="6" align="center" wrap>
                  <NTag size="small" round>{{ item.folder || 'Unsorted' }}</NTag>
                  <NTag v-for="tag in item.tags" :key="tag" size="small" round type="info">{{ tag }}</NTag>
                </NSpace>
              </template>
            </NThing>
          </NCard>
        </div>
        <NEmpty v-else class="page-empty" description="没有匹配的书签">
          <template #extra>
            <NButton type="primary" @click="library.createBookmark">新增第一个书签</NButton>
          </template>
        </NEmpty>
      </NScrollbar>
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
