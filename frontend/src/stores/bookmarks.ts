// bookmarks store 是前端业务状态聚合层，负责协调书签、元数据、导入导出、备份和本地服务状态。
import { defineStore } from 'pinia'
import { api } from '../api'
import { blankBookmarkInput, toBookmarkInput } from '../helpers/bookmarkLists'
import type { AppPreferences, Bookmark, BookmarkInput, CategoryNode, ServerStatus, ThemeManifest } from '../types'

export type LibraryViewMode = 'table' | 'cards' | 'list'
export type CategoryReorderDirection = 'top' | 'up' | 'down'

const viewLimits: Record<LibraryViewMode, number> = {
  table: 50,
  cards: 24,
  list: 100
}

export const useBookmarkStore = defineStore('bookmarks', {
  state: () => ({
    items: [] as Bookmark[],
    folders: [] as string[],
    categories: [] as CategoryNode[],
    tags: [] as string[],
    templates: [] as ThemeManifest[],
    preferences: { openBrowser: 'default' } as AppPreferences,
    selected: null as Bookmark | null,
    draft: blankBookmarkInput() as BookmarkInput,
    query: '',
    folder: '',
    categoryId: '',
    tag: '',
    viewMode: 'cards' as LibraryViewMode,
    pageSize: viewLimits.cards,
    offset: 0,
    loadedCount: 0,
    total: 0,
    loading: false,
    loadingMode: '' as '' | 'replace' | 'append',
    status: '',
    server: { running: false, url: '', addr: '' } as ServerStatus
  }),
  actions: {
    // refresh 同步列表、筛选元数据、Web 服务状态和导出模板，作为页面进入和操作后的统一刷新入口。
    async refresh(options?: { append?: boolean; offset?: number }) {
      this.loading = true
      this.loadingMode = options?.append ? 'append' : 'replace'
      try {
        this.pageSize = viewLimits[this.viewMode]
        const offset = options?.offset ?? (options?.append ? this.loadedCount : this.offset)
        const [result, categories, folders, tags, server, templates, preferences] = await Promise.all([
          api.listBookmarks({
            query: this.query,
            folder: this.folder,
            categoryId: this.categoryId,
            includeDescendants: true,
            viewMode: this.viewMode,
            tags: this.tag ? [this.tag] : [],
            limit: this.pageSize,
            offset,
            sort: 'updated'
          }),
          api.listCategories(),
          api.listFolders(),
          api.listTags(),
          api.getLocalServerStatus(),
          api.listExportTemplates(),
          api.getPreferences()
        ])
        this.items = options?.append ? this.items.concat(result.items) : result.items
        this.total = result.total
        this.offset = offset
        this.loadedCount = this.items.length
        this.categories = categories
        this.folders = folders
        this.tags = tags
        this.server = server
        this.templates = templates
        this.preferences = preferences
      } finally {
        this.loading = false
        this.loadingMode = ''
      }
    },
    setStatus(message: string) {
      this.status = message
    },
    setQuery(query: string) {
      if (this.query !== query) {
        this.loading = true
        this.loadingMode = 'replace'
      }
      this.query = query
      this.offset = 0
      this.loadedCount = 0
    },
    setFolder(folder: string) {
      if (this.folder !== folder || this.categoryId !== '') {
        this.loading = true
        this.loadingMode = 'replace'
      }
      this.folder = folder
      this.categoryId = ''
      this.offset = 0
      this.loadedCount = 0
    },
    setCategory(categoryId: string) {
      const nextCategoryId = categoryId === 'category_all' ? '' : categoryId
      if (this.categoryId !== nextCategoryId || this.folder !== '') {
        this.loading = true
        this.loadingMode = 'replace'
      }
      this.categoryId = nextCategoryId
      this.folder = ''
      this.offset = 0
      this.loadedCount = 0
    },
    setTag(tag: string) {
      if (this.tag !== tag) {
        this.loading = true
        this.loadingMode = 'replace'
      }
      this.tag = tag
      this.offset = 0
      this.loadedCount = 0
    },
    setViewMode(mode: LibraryViewMode) {
      if (this.viewMode !== mode) {
        this.loading = true
        this.loadingMode = 'replace'
      }
      this.viewMode = mode
      this.offset = 0
      this.loadedCount = 0
      this.pageSize = viewLimits[mode]
    },
    clearFilters() {
      if (this.folder || this.categoryId || this.tag) {
        this.loading = true
        this.loadingMode = 'replace'
      }
      this.folder = ''
      this.categoryId = ''
      this.tag = ''
      this.offset = 0
      this.loadedCount = 0
    },
    updateDraft(patch: Partial<BookmarkInput>) {
      this.draft = {
        ...this.draft,
        ...patch
      }
    },
    select(item: Bookmark | null) {
      // 编辑抽屉始终使用 draft 副本，避免用户输入直接污染当前选中书签。
      this.selected = item
      this.draft = item ? toBookmarkInput(item) : blankBookmarkInput()
    },
    async save() {
      const saved = this.selected
        ? await api.updateBookmark(this.selected.id, this.draft)
        : await api.createBookmark(this.draft)
      this.selected = saved
      this.status = '已保存'
      await this.refresh()
    },
    async loadMore() {
      if (this.loading || this.loadedCount >= this.total) return
      await this.refresh({ append: true })
    },
    async setPage(page: number) {
      const nextPage = Math.max(1, page)
      this.offset = (nextPage - 1) * viewLimits[this.viewMode]
      this.loadedCount = 0
      await this.refresh({ offset: this.offset })
    },
    async removeSelected() {
      if (!this.selected) return
      await api.deleteBookmark(this.selected.id)
      this.select(null)
      this.status = '已删除'
      await this.refresh()
    },
    async analyzeSelected() {
      if (!this.selected) return
      this.selected = await api.analyzeBookmark(this.selected.id)
      this.status = '分析完成'
      await this.refresh()
    },
    async analyzeAll() {
      const count = await api.analyzeAllBookmarks()
      this.status = `已分析 ${count} 个书签`
      await this.refresh()
    },
    async importFrom(sourceType: string, path: string) {
      // 导入去重规则在后端存储层执行，前端只负责提交来源和展示汇总结果。
      const result = await api.importBookmarks(sourceType, path)
      this.status = `导入 ${result.imported} 个，跳过 ${result.skipped} 个`
      await this.refresh()
    },
    async exportTo(path: string, templateId: string) {
      const output = await api.exportBookmarks(path, templateId)
      this.status = `已导出到 ${output}`
    },
    async createCategory(name: string, parentId: string) {
      await api.createCategory({ name, parentId })
      this.status = '分类已创建'
      await this.refresh()
    },
    async renameCategory(id: string, name: string) {
      const current = this.findCategory(id)
      await api.updateCategory(id, { name, parentId: current?.parentId || '' })
      this.status = '分类已重命名'
      await this.refresh()
    },
    async moveCategory(id: string, parentId: string) {
      await api.moveCategory(id, { parentId, sortOrder: -1 })
      this.status = '分类已移动'
      await this.refresh()
    },
    async deleteCategory(id: string, deleteBookmarks = false) {
      await api.deleteCategoryWithOptions(id, { deleteBookmarks })
      if (this.categoryId === id) this.categoryId = ''
      this.status = deleteBookmarks ? '分类及书签已删除' : '分类已删除，书签已移动到父级'
      await this.refresh()
    },
    findCategory(id: string) {
      const stack = [...this.categories]
      while (stack.length) {
        const node = stack.shift()
        if (!node) continue
        if (node.id === id) return node
        stack.push(...node.children)
      }
      return null
    },
    findCategoryContext(id: string) {
      const visit = (nodes: CategoryNode[], parent: CategoryNode | null): { node: CategoryNode; parent: CategoryNode | null; siblings: CategoryNode[]; index: number } | null => {
        for (let index = 0; index < nodes.length; index += 1) {
          const node = nodes[index]
          if (node.id === id) {
            return { node, parent, siblings: nodes, index }
          }
          const found = visit(node.children, node)
          if (found) return found
        }
        return null
      }
      return visit(this.categories, null)
    },
    async reorderCategory(id: string, direction: CategoryReorderDirection) {
      const context = this.findCategoryContext(id)
      if (!context || context.node.isSystem) return

      const firstMovableIndex = context.siblings.findIndex((category) => !category.isSystem)
      let targetIndex = context.index
      if (direction === 'top') {
        targetIndex = firstMovableIndex >= 0 ? firstMovableIndex : 0
      } else if (direction === 'up') {
        targetIndex = Math.max(context.index - 1, firstMovableIndex >= 0 ? firstMovableIndex : 0)
      } else {
        targetIndex = Math.min(context.index + 1, context.siblings.length - 1)
      }
      if (targetIndex === context.index) return

      await api.moveCategory(id, {
        parentId: context.node.parentId || 'category_all',
        sortOrder: targetIndex
      })
      this.status = direction === 'top' ? '分类已置顶' : '分类顺序已调整'
      await this.refresh()
    },
    async savePreview(path: string) {
      if (!this.selected) return
      await api.saveBookmarkPreview(this.selected.id, { source: 'manual', path, originalUrl: '' })
      this.status = '预览图已更新'
      await this.refresh()
    },
    async fetchPreview(id?: string) {
      const bookmarkId = id || this.selected?.id
      if (!bookmarkId) return
      await api.fetchBookmarkPreview(bookmarkId)
      this.status = '预览图已获取'
      await this.refresh()
    },
    async openBookmark(id: string) {
      await api.openBookmark(id)
    },
    async savePreferences(preferences: AppPreferences) {
      this.preferences = await api.savePreferences(preferences)
      this.status = '设置已保存'
    },
    async createBackup(path: string) {
      const result = await api.createBackup(path)
      this.status = `备份已创建：${result.path}`
      return result
    },
    async restoreBackup(path: string) {
      const result = await api.restoreBackup(path)
      this.status = `已恢复，恢复前快照：${result.path}`
      await this.refresh()
    },
    async toggleServer() {
      // 服务开关后立即 refresh，确保状态栏、仪表盘和 Web 服务页面看到同一份运行状态。
      if (this.server.running) {
        await api.stopLocalServer()
      } else {
        this.server = await api.startLocalServer()
      }
      await this.refresh()
    }
  }
})
