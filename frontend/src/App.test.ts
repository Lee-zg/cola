import { describe, expect, it } from 'vitest'

const filterBookmarks = (items: Array<{ title: string; tags: string[] }>, query: string, tag: string) =>
  items.filter((item) => item.title.toLowerCase().includes(query.toLowerCase()) && (!tag || item.tags.includes(tag)))

describe('bookmark filtering helpers', () => {
  it('filters by query and tag', () => {
    const items = [
      { title: 'Vue Documentation', tags: ['Development'] },
      { title: 'Design System', tags: ['Design'] }
    ]
    expect(filterBookmarks(items, 'vue', 'Development')).toHaveLength(1)
    expect(filterBookmarks(items, 'vue', 'Design')).toHaveLength(0)
  })
})
