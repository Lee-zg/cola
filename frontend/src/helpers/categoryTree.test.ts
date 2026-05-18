// categoryTree.test 验证分类树批量选择和移动边界。
import { describe, expect, it } from 'vitest'
import {
  buildCategoryContextMap,
  collectBatchActionCategoryIds,
  collectBatchDeleteCategoryIds,
  getBatchCategoryMoveState,
  hasPinnedCategory,
  hasUnpinnedCategory,
  isCheckableCategory,
  pruneCoveredCategoryIds
} from './categoryTree'
import type { CategoryNode } from '../types'

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

describe('category tree helpers', () => {
  it('allows checking root and normal categories but keeps uncategorized disabled', () => {
    const tree = [
      category({
        id: 'category_all',
        name: '全部',
        parentId: '',
        isSystem: true,
        children: [
          category({ id: 'category_uncategorized', name: '未分类', isSystem: true }),
          category({ id: 'cat_a', name: 'A' })
        ]
      })
    ]

    expect(isCheckableCategory(tree[0])).toBe(true)
    expect(isCheckableCategory(tree[0].children[0])).toBe(false)
    expect(isCheckableCategory(tree[0].children[1])).toBe(true)
  })

  it('maps checked root to normal top-level delete targets without uncategorized', () => {
    const tree = [
      category({
        id: 'category_all',
        name: '全部',
        parentId: '',
        isSystem: true,
        children: [
          category({ id: 'category_uncategorized', name: '未分类', isSystem: true }),
          category({ id: 'cat_a', name: 'A', children: [category({ id: 'cat_a_child', name: 'A 子级', parentId: 'cat_a' })] }),
          category({ id: 'cat_b', name: 'B' })
        ]
      })
    ]
    const contextMap = buildCategoryContextMap(tree)

    expect(collectBatchDeleteCategoryIds(['category_all', 'category_uncategorized', 'cat_a_child'], contextMap)).toEqual([
      'cat_a',
      'cat_b'
    ])
  })

  it('prunes child delete targets when their parent is already selected', () => {
    const tree = [
      category({
        id: 'category_all',
        name: '全部',
        parentId: '',
        isSystem: true,
        children: [
          category({
            id: 'cat_a',
            name: 'A',
            children: [
              category({ id: 'cat_a_child', name: 'A 子级', parentId: 'cat_a' }),
              category({ id: 'cat_a_child_2', name: 'A 子级 2', parentId: 'cat_a' })
            ]
          }),
          category({ id: 'cat_b', name: 'B' })
        ]
      })
    ]
    const contextMap = buildCategoryContextMap(tree)

    expect(pruneCoveredCategoryIds(['cat_a_child', 'cat_a', 'cat_b', 'cat_a_child_2'], contextMap)).toEqual(['cat_a', 'cat_b'])
    expect(collectBatchDeleteCategoryIds(['category_uncategorized', 'cat_a_child'], contextMap)).toEqual(['cat_a_child'])
  })

  it('collects batch action ids from root without including system categories', () => {
    const tree = [
      category({
        id: 'category_all',
        name: '全部',
        parentId: '',
        isSystem: true,
        children: [
          category({ id: 'category_uncategorized', name: '未分类', isSystem: true }),
          category({ id: 'cat_a', name: 'A', children: [category({ id: 'cat_a_child', name: 'A 子级', parentId: 'cat_a' })] })
        ]
      })
    ]
    const contextMap = buildCategoryContextMap(tree)

    expect(collectBatchActionCategoryIds(['category_all', 'category_uncategorized'], contextMap)).toEqual(['cat_a', 'cat_a_child'])
  })

  it('allows batch move only inside one unpinned sibling group', () => {
    const tree = [
      category({
        id: 'category_all',
        name: '全部',
        parentId: '',
        isSystem: true,
        children: [
          category({ id: 'cat_a', name: 'A' }),
          category({ id: 'cat_b', name: 'B' }),
          category({ id: 'cat_c', name: 'C' })
        ]
      })
    ]
    const contextMap = buildCategoryContextMap(tree)

    expect(getBatchCategoryMoveState(['cat_b', 'cat_c'], contextMap)).toEqual({ canMoveUp: true, canMoveDown: false })
    expect(getBatchCategoryMoveState(['cat_a'], contextMap)).toEqual({ canMoveUp: false, canMoveDown: true })
  })

  it('disables batch move when pinned categories are selected', () => {
    const tree = [
      category({
        id: 'category_all',
        name: '全部',
        parentId: '',
        isSystem: true,
        children: [
          category({ id: 'cat_a', name: 'A', isPinned: true }),
          category({ id: 'cat_b', name: 'B' })
        ]
      })
    ]
    const contextMap = buildCategoryContextMap(tree)

    expect(getBatchCategoryMoveState(['cat_a'], contextMap)).toEqual({ canMoveUp: false, canMoveDown: false })
    expect(hasPinnedCategory(['cat_a', 'cat_b'], contextMap)).toBe(true)
    expect(hasUnpinnedCategory(['cat_a', 'cat_b'], contextMap)).toBe(true)
  })
})
