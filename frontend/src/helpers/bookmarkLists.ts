// bookmarkLists 放置编辑表单和列表字段的纯转换逻辑，避免组件重复处理 draft 结构。
import type { Bookmark } from '../types'

export const blankBookmarkInput = () => ({
  title: '',
  url: '',
  description: '',
  folder: 'Unsorted',
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
  tags: [...bookmark.tags],
  keywords: [...bookmark.keywords],
  aliases: [...bookmark.aliases]
})
