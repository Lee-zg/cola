import { defineStore } from 'pinia'
import { api } from '../api'
import type { Bookmark, BookmarkInput, ServerStatus, ThemeManifest } from '../types'

const blankInput = (): BookmarkInput => ({
  title: '',
  url: '',
  description: '',
  folder: 'Unsorted',
  tags: [],
  keywords: [],
  aliases: []
})

export const useBookmarkStore = defineStore('bookmarks', {
  state: () => ({
    items: [] as Bookmark[],
    folders: [] as string[],
    tags: [] as string[],
    templates: [] as ThemeManifest[],
    selected: null as Bookmark | null,
    draft: blankInput(),
    query: '',
    folder: '',
    tag: '',
    total: 0,
    loading: false,
    status: '',
    server: { running: false, url: '', addr: '' } as ServerStatus
  }),
  actions: {
    async refresh() {
      this.loading = true
      try {
        const [result, folders, tags, server, templates] = await Promise.all([
          api.listBookmarks({ query: this.query, folder: this.folder, tags: this.tag ? [this.tag] : [], limit: 200, offset: 0, sort: 'updated' }),
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
    select(item: Bookmark | null) {
      this.selected = item
      this.draft = item
        ? {
            title: item.title,
            url: item.url,
            description: item.description,
            folder: item.folder,
            tags: [...item.tags],
            keywords: [...item.keywords],
            aliases: [...item.aliases]
          }
        : blankInput()
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
    },
    async restoreBackup(path: string) {
      const result = await api.restoreBackup(path)
      this.status = `已恢复，恢复前快照：${result.path}`
      await this.refresh()
    },
    async toggleServer() {
      if (this.server.running) {
        await api.stopLocalServer()
      } else {
        this.server = await api.startLocalServer()
      }
      await this.refresh()
    }
  }
})
