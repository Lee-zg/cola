// 文件说明：frontend/src/composables/workflows.test.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { useExportWorkflow } from './useExportWorkflow'
import { useImportWorkflow } from './useImportWorkflow'
import { useBackupWorkflow } from './useBackupWorkflow'
import { useBookmarkStore } from '../stores/bookmarks'

describe('workflow composables', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initializes import workflow state and records successful import', async () => {
    const store = useBookmarkStore()
    store.importFrom = vi.fn(async () => {
      store.setStatus('导入 2 个，跳过 1 个')
    })

    const workflow = useImportWorkflow()
    expect(workflow.importSource.value).toBe('html')
    expect(workflow.skipDuplicates.value).toBe(true)

    await workflow.startImport()

    expect(store.importFrom).toHaveBeenCalledWith('html', '')
    expect(workflow.importing.value).toBe(false)
    expect(workflow.importHistory.value[0].message).toBe('导入 2 个，跳过 1 个')
  })

  it('selects first available export template and exports with chosen path', async () => {
    const store = useBookmarkStore()
    store.templates = [
      {
        id: 'modern',
        name: 'Modern',
        version: '1.0.0',
        templateApiVersion: '1',
        entry: 'index.html',
        css: [],
        assets: [],
        author: '',
        description: ''
      }
    ]
    store.exportTo = vi.fn(async () => undefined)

    const workflow = useExportWorkflow()
    await nextTick()
    workflow.exportPath.value = 'D:\\bookmarks.html'
    await workflow.exportHtml()

    expect(workflow.templateId.value).toBe('modern')
    expect(store.exportTo).toHaveBeenCalledWith('D:\\bookmarks.html', 'modern')
    expect(workflow.exporting.value).toBe(false)
  })

  it('blocks restore until confirmation is valid', async () => {
    const store = useBookmarkStore()
    store.restoreBackup = vi.fn(async () => undefined)

    const workflow = useBackupWorkflow()
    await workflow.restoreBackup()

    expect(store.restoreBackup).not.toHaveBeenCalled()
    expect(store.status).toBe('请输入 RESTORE 确认恢复')

    workflow.restoreConfirm.value = 'RESTORE'
    workflow.restorePath.value = 'cola.db'
    await workflow.restoreBackup()

    expect(store.restoreBackup).toHaveBeenCalledWith('cola.db')
    expect(workflow.restoreConfirm.value).toBe('')
  })
})
