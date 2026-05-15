// 文件说明：frontend/src/helpers/domain.test.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { describe, expect, it } from 'vitest'
import { canRestoreBackup, createBackupHistoryEntry, formatSize, getRestoreValidationMessage } from './backup'
import { splitList, joinList } from './bookmarkLists'
import { getDashboardStats, getPendingAiItems, getRecentItems, getTopTags } from './dashboard'
import type { Bookmark } from '../types'

const makeBookmark = (id: string, lastAnalyzedAt?: string): Bookmark => ({
  id,
  title: `Bookmark ${id}`,
  url: `https://example.com/${id}`,
  description: '',
  folder: 'Work',
  categoryId: 'cat_work',
  categoryName: 'Work',
  categoryPath: ['Work'],
  tags: ['vue'],
  keywords: [],
  aliases: [],
  domain: 'example.com',
  createdAt: '2026-05-14T00:00:00Z',
  updatedAt: '2026-05-14T00:00:00Z',
  lastAnalyzedAt
})

describe('bookmark list helpers', () => {
  it('splits and joins comma separated list values', () => {
    expect(splitList(' vue, go, , rust ')).toEqual(['vue', 'go', 'rust'])
    expect(joinList(['vue', 'go'])).toBe('vue, go')
  })
})

describe('dashboard helpers', () => {
  it('calculates dashboard stats and derived lists', () => {
    const items = [makeBookmark('1'), makeBookmark('2', '2026-05-14T01:00:00Z'), makeBookmark('3')]

    expect(getDashboardStats({ total: 3, folders: ['Work'], tags: ['vue', 'go'], items })).toEqual({
      total: 3,
      folderCount: 1,
      tagCount: 2,
      pendingAiCount: 2
    })
    expect(getPendingAiItems(items).map((item) => item.id)).toEqual(['1', '3'])
    expect(getRecentItems(items, 2)).toHaveLength(2)
    expect(getTopTags(['a', 'b', 'c'], 2)).toEqual(['a', 'b'])
  })
})

describe('backup helpers', () => {
  it('validates restore confirmation and formats backup metadata', () => {
    expect(canRestoreBackup('RESTORE')).toBe(true)
    expect(canRestoreBackup('restore')).toBe(false)
    expect(getRestoreValidationMessage('no')).toBe('请输入 RESTORE 确认恢复')
    expect(formatSize(1536)).toBe('1.5 KB')
    expect(createBackupHistoryEntry('D:\\backup\\cola.db', new Date('2026-05-14T00:00:00Z')).filename).toBe(
      'cola.db'
    )
  })
})
