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
