// 文件说明：frontend/src/composables/useSettingsSummary.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { computed } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'
import type { CategoryNode } from '../types'

const countEditableCategories = (nodes: CategoryNode[]): number =>
  nodes.reduce((count, node) => {
    // 设置页展示真实用户分类数量，系统分类不计入数据规模。
    const selfCount = node.isSystem ? 0 : 1
    return count + selfCount + countEditableCategories(node.children)
  }, 0)

export const useSettingsSummary = () => {
  const store = useBookmarkStore()

  return {
    total: computed(() => store.total),
    categoryCount: computed(() => countEditableCategories(store.categories)),
    tagCount: computed(() => store.tags.length)
  }
}
