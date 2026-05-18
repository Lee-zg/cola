// bookmarkLists 放置编辑表单和列表字段的纯转换逻辑，避免组件重复处理 draft 结构。
import type { Bookmark, BookmarkInput, CategoryNode } from '../types'

const rootCategoryId = 'category_all'
const uncategorizedId = 'category_uncategorized'
const uncategorizedName = '未分类'

export const blankBookmarkInput = () => ({
  title: '',
  url: '',
  description: '',
  folder: '未分类',
  categoryId: 'category_uncategorized',
  tags: [],
  keywords: [],
  aliases: []
})

// 列表字段在前端用逗号编辑，最终的去重、排序和大小写规则由后端再次规范化。
export const splitList = (raw: string): string[] =>
  raw
    .split(',')
    .map((value) => value.trim())
    .filter(Boolean)

export const joinList = (values: string[]): string => values.join(', ')

export const toBookmarkInput = (bookmark: Bookmark) => ({
  title: bookmark.title,
  url: bookmark.url,
  description: bookmark.description,
  folder: bookmark.folder,
  categoryId: bookmark.categoryId || 'category_uncategorized',
  tags: [...bookmark.tags],
  keywords: [...bookmark.keywords],
  aliases: [...bookmark.aliases]
})

export const findCategoryPath = (categories: CategoryNode[], categoryId: string): string[] => {
  let matchedPath: string[] = []

  const visit = (nodes: CategoryNode[], path: string[]) => {
    for (const node of nodes) {
      const nextPath = node.id === rootCategoryId ? path : path.concat(node.name)
      if (node.id === categoryId) {
        matchedPath = nextPath
        return true
      }
      if (visit(node.children, nextPath)) return true
    }
    return false
  }

  visit(categories, [])
  return matchedPath
}

export const normalizeBookmarkCategoryInput = (draft: BookmarkInput, categories: CategoryNode[]): BookmarkInput => {
  const categoryId = (draft.categoryId || '').trim()

  if (!categoryId || categoryId === rootCategoryId || categoryId === uncategorizedId) {
    // 系统“全部分类”不能写入书签，清空或选未分类时统一保存到未分类。
    return { ...draft, categoryId: uncategorizedId, folder: uncategorizedName }
  }

  const categoryPath = findCategoryPath(categories, categoryId)
  if (!categoryPath.length) {
    return { ...draft, categoryId, folder: draft.folder?.trim() || uncategorizedName }
  }

  // 后端以 categoryId 为准，这里同步 folder 便于旧数据和列表展示保持一致。
  return { ...draft, categoryId, folder: categoryPath.join(' / ') }
}
