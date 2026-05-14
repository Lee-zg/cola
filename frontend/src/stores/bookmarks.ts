// bookmarks store 是前端业务状态聚合层，负责协调书签、元数据、导入导出、备份和本地服务状态。
import { defineStore } from 'pinia'
import { api } from '../api'
import { blankBookmarkInput, toBookmarkInput } from '../helpers/bookmarkLists'
import type { Bookmark, BookmarkInput, ServerStatus, ThemeManifest } from '../types'

export const useBookmarkStore = defineStore('bookmarks', {
  state: () => ({
    items: [] as Bookmark[],
    folders: [] as string[],
    tags: [] as string[],
    templates: [] as ThemeManifest[],
    selected: null as Bookmark | null,
    draft: blankBookmarkInput() as BookmarkInput,
    query: '',
    folder: '',
    tag: '',
    total: 0,
    loading: false,
    status: '',
    server: { running: false, url: '', addr: '' } as ServerStatus
  }),
  actions: {
    // refresh 同步列表、筛选元数据、Web 服务状态和导出模板，作为页面进入和操作后的统一刷新入口。
    async refresh() {
      this.loading = true
      try {
        const [result, folders, tags, server, templates] = await Promise.all([
          api.listBookmarks({
            query: this.query,
            folder: this.folder,
            tags: this.tag ? [this.tag] : [],
            limit: 200,
            offset: 0,
            sort: 'updated'
          }),
          api.listFolders(),
          api.listTags(),
          api.getLocalServerStatus(),
          api.listExportTemplates()
        ])
        this.items = result.items
        this.total = result.total
        this.folders = folders
        this.tags = tags
        this.server = server
        this.templates = templates
      } finally {
        this.loading = false
      }
    },
    setStatus(message: string) {
      this.status = message
    },
    setQuery(query: string) {
      this.query = query
    },
    setFolder(folder: string) {
      this.folder = folder
    },
    setTag(tag: string) {
      this.tag = tag
    },
    clearFilters() {
      this.folder = ''
      this.tag = ''
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
