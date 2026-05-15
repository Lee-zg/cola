<!-- 文件说明：frontend/src/pages/LibraryPage.vue，书签库核心工作台，包含树形分类和多视图书签列表。 -->
<script setup lang="ts">
import { computed, watch } from 'vue'
import {
  NButton,
  NEmpty,
  NIcon,
  NInput,
  NPagination,
  NRadioButton,
  NRadioGroup,
  NSpace,
  NTag,
  NTooltip
} from 'naive-ui'
import BookmarkEditorDrawer from '../components/BookmarkEditorDrawer.vue'
import ColaLoader from '../components/ColaLoader.vue'
import CategoryTree from '../components/CategoryTree.vue'
import ColaScrollbar from '../components/ColaScrollbar.vue'
import { useBookmarkLibrary } from '../composables/useBookmarkLibrary'
import { useUiCommands } from '../composables/useUiCommands'
import { appIcons } from '../icons'
import type { Bookmark } from '../types'

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
  { label: '表格', value: 'table' },
  { label: '卡片', value: 'cards' },
  { label: '列表', value: 'list' }
]

const page = computed({
  get: () => Math.floor(library.offset.value / library.pageSize.value) + 1,
  set: (value: number) => {
    void library.setPage(value)
  }
})
const pageCount = computed(() => Math.max(1, Math.ceil(library.total.value / library.pageSize.value)))
const showPagination = computed(() => library.viewMode.value !== 'list' && library.total.value > library.pageSize.value)

const getPreviewStyle = (item: Bookmark) => ({
  backgroundImage: item.preview?.thumbPath || item.preview?.filePath ? `url("${item.preview.thumbPath || item.preview.filePath}")` : ''
})

const createCategory = async (name: string, parentId: string) => {
  await library.createCategory(name, parentId)
}

const renameCategory = async (id: string, name: string) => {
  await library.renameCategory(id, name)
}

const deleteCategory = async (id: string, deleteBookmarks: boolean) => {
  await library.deleteCategory(id, deleteBookmarks)
}

const handleBookmarkClick = async (item: Bookmark, event: MouseEvent) => {
  await library.openBookmark(item, event)
}

const handleListReachBottom = async () => {
  if (library.viewMode.value === 'list') {
    await library.loadMore()
  }
}

const selectCategory = (id: string) => {
  library.categoryId.value = id
}

const reorderCategory = async (id: string, direction: 'top' | 'up' | 'down') => {
  await library.reorderCategory(id, direction)
}
</script>

<template>
  <section class="page library-page">
    <aside class="library-sidebar category-sidebar panel-surface">
      <div class="panel-heading category-heading">
        <div>
          <span class="eyebrow">CATEGORIES</span>
          <h2>书签分类</h2>
        </div>
        <NTooltip trigger="hover">
          <template #trigger>
            <NButton circle size="small" secondary aria-label="刷新分类" @click="library.clearFilters">
              <template #icon>
                <NIcon :component="appIcons.refresh" />
              </template>
            </NButton>
          </template>
          查看全部分类
        </NTooltip>
      </div>

      <CategoryTree
        :categories="library.categories.value"
        :selected-id="library.categoryId.value"
        @create="createCategory"
        @delete="deleteCategory"
        @rename="renameCategory"
        @reorder="reorderCategory"
        @select="selectCategory"
      />
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
            placeholder="标题、网址、描述、标签、关键词"
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
          <NButton type="primary" @click="library.createBookmark">
            <template #icon>
              <NIcon :component="appIcons.add" />
            </template>
            新建
          </NButton>
        </NSpace>
      </div>

      <ColaScrollbar class="library-scroll" aria-label="书签列表滚动区" @reach-bottom="handleListReachBottom">
        <div class="library-list-stage" :class="{ 'is-loading': library.loading.value }">
          <ColaLoader v-if="library.loading.value && library.loadingMode.value !== 'append' && !library.items.value.length" label="可乐正在翻找书签" />
          <div v-else-if="library.items.value.length" class="bookmark-list" :class="`mode-${library.viewMode.value}`">
            <table v-if="library.viewMode.value === 'table'" class="bookmark-table">
              <thead>
                <tr>
                  <th>预览</th>
                  <th>标题</th>
                  <th>分类</th>
                  <th>更新时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in library.items.value" :key="item.id" :class="{ active: library.selectedId.value === item.id }">
                  <td>
                    <div class="bookmark-preview small" :style="getPreviewStyle(item)">
                      <span v-if="!item.preview">{{ (item.domain || item.title || '?').slice(0, 1).toUpperCase() }}</span>
                    </div>
                  </td>
                  <td>
                    <button
                      class="bookmark-title-link"
                      type="button"
                      title="按住 Ctrl 点击打开，普通点击编辑"
                      @click="handleBookmarkClick(item, $event)"
                    >
                      {{ item.title || item.url }}
                    </button>
                    <span class="bookmark-url">{{ item.domain || item.url }}</span>
                  </td>
                  <td>{{ item.categoryPath?.join(' / ') || item.folder }}</td>
                  <td>{{ item.updatedAt }}</td>
                  <td>
                    <NSpace :size="6">
                      <NButton size="small" secondary @click="library.selectBookmark(item)">编辑</NButton>
                      <NButton size="small" secondary @click="library.fetchPreview(item.id)">预览</NButton>
                    </NSpace>
                  </td>
                </tr>
              </tbody>
            </table>

            <article
              v-for="item in library.viewMode.value === 'cards' ? library.items.value : []"
              :key="item.id"
              class="bookmark-card-pro"
              :class="{ active: library.selectedId.value === item.id }"
              @click="handleBookmarkClick(item, $event)"
            >
              <div class="bookmark-preview" :style="getPreviewStyle(item)">
                <span v-if="!item.preview">{{ (item.domain || item.title || '?').slice(0, 1).toUpperCase() }}</span>
                <button class="preview-action" type="button" @click.stop="library.fetchPreview(item.id)">获取预览</button>
              </div>
              <div class="bookmark-card-body">
                <div class="bookmark-card-title">
                  <strong>{{ item.title || item.url }}</strong>
                  <span class="open-hint">
                    <NIcon :component="appIcons.open" />
                    Ctrl 打开
                  </span>
                </div>
                <span class="bookmark-url">{{ item.url }}</span>
                <p>{{ item.description || item.domain || '暂无描述' }}</p>
                <NSpace :size="6" align="center" wrap>
                  <NTag size="small" round>{{ item.categoryPath?.join(' / ') || item.folder }}</NTag>
                  <NTag v-for="tag in item.tags.slice(0, 3)" :key="tag" size="small" round type="info">{{ tag }}</NTag>
                </NSpace>
              </div>
            </article>

            <div v-if="library.viewMode.value === 'list'" class="bookmark-stream">
              <article
                v-for="item in library.items.value"
                :key="item.id"
                class="bookmark-row"
                :class="{ active: library.selectedId.value === item.id }"
                @click="handleBookmarkClick(item, $event)"
              >
                <div class="bookmark-preview small" :style="getPreviewStyle(item)">
                  <span v-if="!item.preview">{{ (item.domain || item.title || '?').slice(0, 1).toUpperCase() }}</span>
                </div>
                <div>
                  <strong>{{ item.title || item.url }}</strong>
                  <span class="bookmark-url">{{ item.url }}</span>
                  <p>{{ item.description || item.categoryPath?.join(' / ') || item.folder }}</p>
                </div>
              </article>
              <NButton v-if="library.loadedCount.value < library.total.value" secondary block :loading="library.loading.value" @click="library.loadMore">
                加载更多（{{ library.loadedCount.value }} / {{ library.total.value }}）
              </NButton>
            </div>
          </div>
          <NEmpty v-else class="page-empty" description="没有匹配的书签">
            <template #extra>
              <NButton type="primary" @click="library.createBookmark">新增第一个书签</NButton>
            </template>
          </NEmpty>
          <Transition name="cola-loader-pop">
            <div v-if="library.loading.value && library.loadingMode.value !== 'append' && library.items.value.length" class="library-loading-overlay">
              <ColaLoader label="正在刷新书签列表" />
            </div>
          </Transition>
        </div>
      </ColaScrollbar>

      <div v-if="showPagination" class="library-pagination">
        <NPagination v-model:page="page" :page-count="pageCount" :page-slot="7" />
      </div>
    </main>

    <BookmarkEditorDrawer
      :categories="library.categories.value"
      :draft="library.draft.value"
      :open="library.editorOpen.value"
      :selected="library.selected.value"
      :status="library.status.value"
      @analyze="library.analyzeSelected"
      @close="library.closeEditor"
      @fetch-preview="library.fetchPreview()"
      @remove="library.removeSelected"
      @save="library.save"
      @save-preview="library.savePreview"
      @update:draft="library.updateDraft"
    />
  </section>
</template>
