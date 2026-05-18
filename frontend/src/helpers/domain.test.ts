// 文件说明：frontend/src/helpers/domain.test.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { describe, expect, it } from 'vitest'
import { canRestoreBackup, createBackupHistoryEntry, formatSize, getRestoreValidationMessage } from './backup'
import { findCategoryPath, joinList, normalizeBookmarkCategoryInput, splitList } from './bookmarkLists'
import { getDashboardStats, getPendingAiItems, getRecentItems, getTopTags } from './dashboard'
import type { Bookmark, CategoryNode } from '../types'

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

const category = (patch: Partial<CategoryNode>): CategoryNode => ({
  id: '',
  name: '',
  parentId: 'category_all',
  sortOrder: 0,
  isSystem: false,
  isPinned: false,
  count: 0,
  createdAt: '',
  updatedAt: '',
  children: [],
  ...patch
})

describe('bookmark list helpers', () => {
  it('splits and joins comma separated list values', () => {
    expect(splitList(' vue, go, , rust ')).toEqual(['vue', 'go', 'rust'])
    expect(joinList(['vue', 'go'])).toBe('vue, go')
  })

  it('normalizes category id to folder path before saving', () => {
    const categories = [
      category({
        id: 'category_all',
        name: '全部分类',
        parentId: '',
        isSystem: true,
        children: [
          category({
            id: 'cat_work',
            name: '工作',
            children: [category({ id: 'cat_dev', name: '研发', parentId: 'cat_work' })]
          })
        ]
      })
    ]
    const draft = {
      title: 'Vue',
      url: 'https://vuejs.org',
      description: '',
      folder: '未分类',
      categoryId: 'cat_dev',
      tags: [],
      keywords: [],
      aliases: []
    }

    expect(findCategoryPath(categories, 'cat_dev')).toEqual(['工作', '研发'])
    expect(normalizeBookmarkCategoryInput(draft, categories)).toMatchObject({
      categoryId: 'cat_dev',
      folder: '工作 / 研发'
    })
  })

  it('keeps root and empty category drafts in uncategorized', () => {
    const draft = {
      title: 'Vue',
      url: 'https://vuejs.org',
      description: '',
      folder: '工作',
      categoryId: 'category_all',
      tags: [],
      keywords: [],
      aliases: []
    }

    expect(normalizeBookmarkCategoryInput(draft, [])).toMatchObject({
      categoryId: 'category_uncategorized',
      folder: '未分类'
    })
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
