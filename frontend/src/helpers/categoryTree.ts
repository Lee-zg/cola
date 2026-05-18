// categoryTree 集中分类树纯逻辑，组件只负责渲染和事件接线。
import type { CategoryNode } from '../types'

export type CategoryReorderDirection = 'up' | 'down'

export interface CategoryContext {
  node: CategoryNode
  parent: CategoryNode | null
  siblings: CategoryNode[]
  index: number
}

export const isRootCategory = (id: string) => id === 'category_all'

export const isUncategorizedCategory = (id: string) => id === 'category_uncategorized'

export const canEditCategory = (node: CategoryNode) => !node.isSystem

export const isCheckableCategory = (node: CategoryNode) => isRootCategory(node.id) || (!node.isSystem && !isUncategorizedCategory(node.id))

export const canCreateChildCategory = (node: CategoryNode) => !isUncategorizedCategory(node.id)

export const buildCategoryContextMap = (categories: CategoryNode[]) => {
  const meta = new Map<string, CategoryContext>()
  const visit = (nodes: CategoryNode[], parent: CategoryNode | null) => {
    nodes.forEach((node, index) => {
      meta.set(node.id, { node, parent, siblings: nodes, index })
      visit(node.children, node)
    })
  }
  visit(categories, null)
  return meta
}

export const firstMovableIndex = (siblings: CategoryNode[]) => {
  const index = siblings.findIndex((category) => !category.isSystem && !category.isPinned)
  return index >= 0 ? index : 0
}

export const canReorderCategory = (node: CategoryNode, context: CategoryContext | undefined, direction: CategoryReorderDirection) => {
  if (!context || !canEditCategory(node) || node.isPinned) return false
  const movable = context.siblings.filter((category) => !category.isSystem && !category.isPinned)
  const movableIndex = movable.findIndex((category) => category.id === node.id)
  if (movableIndex < 0) return false
  return direction === 'up' ? movableIndex > 0 : movableIndex < movable.length - 1
}

const uniqueCategoryIds = (ids: string[]) => Array.from(new Set(ids))

const collectEditableDescendantIds = (node: CategoryNode): string[] =>
  node.children.flatMap((child) => {
    const descendantIds = collectEditableDescendantIds(child)
    return canEditCategory(child) ? [child.id].concat(descendantIds) : descendantIds
  })

export const collectBatchActionCategoryIds = (ids: string[], contextMap: Map<string, CategoryContext>) => {
  const uniqueIds = uniqueCategoryIds(ids)
  const rootNode = contextMap.get('category_all')?.node
  const actionIds = uniqueIds.some(isRootCategory) && rootNode ? collectEditableDescendantIds(rootNode) : uniqueIds
  return uniqueCategoryIds(actionIds).filter((id) => {
    const context = contextMap.get(id)
    return Boolean(context && canEditCategory(context.node))
  })
}

export const pruneCoveredCategoryIds = (ids: string[], contextMap: Map<string, CategoryContext>) => {
  const selected = new Set(uniqueCategoryIds(ids))
  return uniqueCategoryIds(ids).filter((id) => {
    const context = contextMap.get(id)
    if (!context || !canEditCategory(context.node)) return false
    let parent = context.parent
    // 删除父分类会覆盖其所有子分类，批量提交前先去掉被父节点覆盖的子节点。
    while (parent) {
      if (canEditCategory(parent) && selected.has(parent.id)) return false
      parent = contextMap.get(parent.id)?.parent || null
    }
    return true
  })
}

export const collectBatchDeleteCategoryIds = (ids: string[], contextMap: Map<string, CategoryContext>) => {
  const uniqueIds = uniqueCategoryIds(ids)
  const rootNode = contextMap.get('category_all')?.node
  if (uniqueIds.some(isRootCategory) && rootNode) {
    // 勾选“全部分类”时只提交普通一级分类，由后端级联删除其子分类，未分类保持保护态。
    return rootNode.children.filter((child) => canEditCategory(child)).map((child) => child.id)
  }
  return pruneCoveredCategoryIds(
    uniqueIds.filter((id) => {
      const context = contextMap.get(id)
      return Boolean(context && canEditCategory(context.node))
    }),
    contextMap
  )
}

export const getBatchCategoryMoveState = (ids: string[], contextMap: Map<string, CategoryContext>) => {
  const contexts = ids.map((id) => contextMap.get(id)).filter(Boolean) as CategoryContext[]
  if (!contexts.length) {
    return { canMoveUp: false, canMoveDown: false }
  }
  const parentId = contexts[0].node.parentId || 'category_all'
  const sameParent = contexts.every((context) => (context.node.parentId || 'category_all') === parentId)
  const allMovable = contexts.every((context) => canEditCategory(context.node) && !context.node.isPinned)
  if (!sameParent || !allMovable) {
    return { canMoveUp: false, canMoveDown: false }
  }
  const selected = new Set(contexts.map((context) => context.node.id))
  // 批量移动只在同级未置顶区内判断边界，避免跨过置顶区或系统分类。
  const movableSiblings = contexts[0].siblings.filter((category) => !category.isSystem && !category.isPinned)
  const selectedIndexes = movableSiblings
    .map((category, index) => (selected.has(category.id) ? index : -1))
    .filter((index) => index >= 0)
  if (!selectedIndexes.length) {
    return { canMoveUp: false, canMoveDown: false }
  }
  return {
    canMoveUp: selectedIndexes[0] > 0,
    canMoveDown: selectedIndexes[selectedIndexes.length - 1] < movableSiblings.length - 1
  }
}

export const hasPinnedCategory = (ids: string[], contextMap: Map<string, CategoryContext>) =>
  ids.some((id) => Boolean(contextMap.get(id)?.node.isPinned))

export const hasUnpinnedCategory = (ids: string[], contextMap: Map<string, CategoryContext>) =>
  ids.some((id) => {
    const node = contextMap.get(id)?.node
    return Boolean(node && canEditCategory(node) && !node.isPinned)
  })
