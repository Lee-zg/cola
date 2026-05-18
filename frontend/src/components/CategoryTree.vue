<!-- 文件说明：frontend/src/components/CategoryTree.vue，基于 Naive UI Tree 二次封装的书签分类树。 -->
<script setup lang="ts">
import { computed, h, ref, watch, type Component } from 'vue'
import { NButton, NCheckbox, NDropdown, NIcon, NInput, NModal, NSpace, NTooltip, NTree } from 'naive-ui'
import type { DropdownOption, TreeOption } from 'naive-ui'
import {
  buildCategoryContextMap,
  canCreateChildCategory,
  canEditCategory,
  canReorderCategory,
  collectBatchActionCategoryIds,
  collectBatchDeleteCategoryIds,
  getBatchCategoryMoveState,
  hasPinnedCategory,
  hasUnpinnedCategory,
  isCheckableCategory,
  isRootCategory,
  isUncategorizedCategory,
  type CategoryReorderDirection
} from '../helpers/categoryTree'
import { appIcons } from '../icons'
import type { CategoryNode } from '../types'

type CategoryEditorMode = 'create' | 'rename'
type DeleteMode = 'single' | 'batch'
type ClearBookmarksTarget = 'category' | 'uncategorized'

interface CategoryTreeOption extends TreeOption {
  key: string
  label: string
  category: CategoryNode
  checkboxDisabled?: boolean
  children?: CategoryTreeOption[]
}

const props = defineProps<{
  categories: CategoryNode[]
  selectedId: string
}>()

const emit = defineEmits<{
  batchDelete: [ids: string[], deleteBookmarks: boolean]
  batchPin: [ids: string[], pinned: boolean]
  batchReorder: [ids: string[], direction: CategoryReorderDirection]
  clearBookmarks: [id: string]
  create: [name: string, parentId: string]
  delete: [id: string, deleteBookmarks: boolean]
  pin: [id: string, pinned: boolean]
  rename: [id: string, name: string]
  reorder: [id: string, direction: CategoryReorderDirection]
  select: [id: string]
}>()

const searchQuery = ref('')
const expandedKeys = ref<string[]>(['category_all'])
const checkedKeys = ref<string[]>([])
const editorOpen = ref(false)
const editorMode = ref<CategoryEditorMode>('create')
const editorDraft = ref('')
const editorTargetId = ref('')
const editorParentId = ref('category_all')
const deleteOpen = ref(false)
const deleteMode = ref<DeleteMode>('single')
const deleteTarget = ref<CategoryTreeOption | null>(null)
const deleteIds = ref<string[]>([])
const deleteBookmarks = ref(false)
const clearBookmarksOpen = ref(false)
const clearBookmarksTarget = ref<CategoryTreeOption | null>(null)

const selectedKeys = computed(() => [props.selectedId])

const normalizeText = (value: string) => value.trim().toLowerCase()

const toTreeOption = (node: CategoryNode): CategoryTreeOption => ({
  key: node.id,
  label: node.name,
  category: node,
  checkboxDisabled: !isCheckableCategory(node),
  children: node.children.map(toTreeOption)
})

const rootOption = computed(() => (props.categories[0] ? toTreeOption(props.categories[0]) : null))

const filterTreeOption = (node: CategoryTreeOption, pattern: string): CategoryTreeOption | null => {
  if (!pattern) return node
  const children = (node.children || []).map((child) => filterTreeOption(child, pattern)).filter(Boolean) as CategoryTreeOption[]
  const selfMatched = normalizeText(node.label).includes(pattern)
  if (node.key === 'category_all' || selfMatched || children.length) {
    return { ...node, children }
  }
  return null
}

const treeData = computed(() => {
  if (!rootOption.value) return []
  const filtered = filterTreeOption(rootOption.value, normalizeText(searchQuery.value))
  return filtered ? [filtered] : []
})

const contextMap = computed(() => buildCategoryContextMap(props.categories))
const checkedActionableIds = computed(() => collectBatchActionCategoryIds(checkedKeys.value, contextMap.value))
const checkedBatchDeleteIds = computed(() => collectBatchDeleteCategoryIds(checkedKeys.value, contextMap.value))
const batchDeleteAllSelected = computed(() => checkedKeys.value.includes('category_all'))
const showBulkBar = computed(() => checkedActionableIds.value.length > 0 || checkedBatchDeleteIds.value.length > 0)
const batchMoveState = computed(() => getBatchCategoryMoveState(checkedActionableIds.value, contextMap.value))
const batchHasPinned = computed(() => hasPinnedCategory(checkedActionableIds.value, contextMap.value))
const batchHasUnpinned = computed(() => hasUnpinnedCategory(checkedActionableIds.value, contextMap.value))
const deleteTitle = computed(() =>
  deleteMode.value === 'batch'
    ? batchDeleteAllSelected.value
      ? '删除全部普通分类'
      : `删除 ${checkedBatchDeleteIds.value.length} 个分类`
    : '删除分类'
)
const deleteDescription = computed(() =>
  batchDeleteAllSelected.value
    ? '将删除全部普通分类及其子分类，不包含未分类。'
    : '默认删除后，分类及子分类下书签会移动到对应父级。'
)
const clearBookmarksType = computed<ClearBookmarksTarget>(() =>
  clearBookmarksTarget.value && isUncategorizedCategory(clearBookmarksTarget.value.key) ? 'uncategorized' : 'category'
)
const clearBookmarksTitle = computed(() => (clearBookmarksType.value === 'uncategorized' ? '清空未分类书签' : '清空分类书签'))
const clearBookmarksDescription = computed(() =>
  clearBookmarksType.value === 'uncategorized'
    ? '将删除未分类下的所有书签，但不会删除“未分类”系统分类。'
    : '将删除该分类及子分类下的所有书签，但不会删除分类节点。'
)

const collectExpandableKeys = (nodes: CategoryTreeOption[]): string[] =>
  nodes.flatMap((node) => {
    const childKeys = collectExpandableKeys(node.children || [])
    return node.children?.length ? [node.key].concat(childKeys) : childKeys
  })

const collectExistingKeys = (nodes: CategoryTreeOption[]): Set<string> => {
  const keys = new Set<string>()
  const visit = (items: CategoryTreeOption[]) => {
    for (const item of items) {
      keys.add(item.key)
      visit(item.children || [])
    }
  }
  visit(nodes)
  return keys
}

const isRoot = (option: CategoryTreeOption) => isRootCategory(option.key)

const stopNodeAction = (event: MouseEvent) => {
  event.preventDefault()
  event.stopPropagation()
}

const openCreateEditor = (parentId: string, event?: MouseEvent) => {
  if (event) stopNodeAction(event)
  editorMode.value = 'create'
  editorParentId.value = parentId
  editorTargetId.value = ''
  editorDraft.value = ''
  editorOpen.value = true
}

const openRenameEditor = (option: CategoryTreeOption, event?: MouseEvent) => {
  if (event) stopNodeAction(event)
  if (!canEditCategory(option.category)) return
  editorMode.value = 'rename'
  editorTargetId.value = option.key
  editorParentId.value = option.category.parentId || 'category_all'
  editorDraft.value = option.label
  editorOpen.value = true
}

const submitEditor = () => {
  const name = editorDraft.value.trim()
  if (!name) return
  if (editorMode.value === 'create') {
    emit('create', name, editorParentId.value || 'category_all')
  } else {
    emit('rename', editorTargetId.value, name)
  }
  editorOpen.value = false
}

const openSingleDeleteConfirm = (option: CategoryTreeOption) => {
  if (!canEditCategory(option.category)) return
  deleteMode.value = 'single'
  deleteTarget.value = option
  deleteIds.value = []
  deleteBookmarks.value = false
  deleteOpen.value = true
}

const openBatchDeleteConfirm = () => {
  if (!checkedBatchDeleteIds.value.length) return
  deleteMode.value = 'batch'
  deleteTarget.value = null
  deleteIds.value = checkedBatchDeleteIds.value
  deleteBookmarks.value = false
  deleteOpen.value = true
}

const openClearBookmarksConfirm = (option: CategoryTreeOption) => {
  if (isRoot(option)) return
  clearBookmarksTarget.value = option
  clearBookmarksOpen.value = true
}

const confirmDelete = () => {
  if (deleteMode.value === 'batch') {
    emit('batchDelete', deleteIds.value, deleteBookmarks.value)
    checkedKeys.value = []
  } else if (deleteTarget.value) {
    emit('delete', deleteTarget.value.key, deleteBookmarks.value)
  }
  deleteOpen.value = false
  deleteTarget.value = null
  deleteIds.value = []
  deleteBookmarks.value = false
}

const confirmClearBookmarks = () => {
  if (!clearBookmarksTarget.value) return
  emit('clearBookmarks', clearBookmarksTarget.value.key)
  clearBookmarksOpen.value = false
  clearBookmarksTarget.value = null
}

const handleBatchPin = (pinned: boolean) => {
  if (!checkedActionableIds.value.length) return
  emit('batchPin', checkedActionableIds.value, pinned)
  checkedKeys.value = []
}

const handleBatchReorder = (direction: CategoryReorderDirection) => {
  if (!checkedActionableIds.value.length) return
  emit('batchReorder', checkedActionableIds.value, direction)
  checkedKeys.value = []
}

const clearChecked = () => {
  checkedKeys.value = []
}

const handleMoreSelect = (key: string | number, option: CategoryTreeOption) => {
  const action = String(key)
  if (action === 'create') {
    openCreateEditor(option.key)
    return
  }
  if (action === 'rename') {
    openRenameEditor(option)
    return
  }
  if (action === 'delete') {
    openSingleDeleteConfirm(option)
    return
  }
  if (action === 'clear-bookmarks') {
    openClearBookmarksConfirm(option)
    return
  }
  if (action === 'pin') {
    emit('pin', option.key, true)
    return
  }
  if (action === 'unpin') {
    emit('pin', option.key, false)
    return
  }
  if (action === 'up' || action === 'down') {
    emit('reorder', option.key, action)
  }
}

const toggleExpanded = (option: CategoryTreeOption, event: MouseEvent) => {
  stopNodeAction(event)
  if (isRoot(option) || !option.children?.length) return
  const next = new Set(expandedKeys.value)
  if (next.has(option.key)) {
    next.delete(option.key)
  } else {
    next.add(option.key)
  }
  next.add('category_all')
  expandedKeys.value = Array.from(next)
}

const renderActionIcon = (component: Component) => () => h(NIcon, { component })

const moreOptions = (option: CategoryTreeOption): DropdownOption[] => {
  const context = contextMap.value.get(option.key)
  const options: DropdownOption[] = [
    {
      key: 'create',
      label: isRoot(option) ? '新增一级分类' : '新增子分类',
      disabled: !canCreateChildCategory(option.category),
      icon: renderActionIcon(appIcons.addPlain)
    },
    {
      key: 'rename',
      label: '重命名',
      disabled: !canEditCategory(option.category),
      icon: renderActionIcon(appIcons.pencil)
    },
    {
      key: option.category.isPinned ? 'unpin' : 'pin',
      label: option.category.isPinned ? '取消置顶' : '置顶',
      disabled: !canEditCategory(option.category),
      icon: renderActionIcon(option.category.isPinned ? appIcons.pinFilled : appIcons.pin)
    },
    {
      key: 'divider-order',
      type: 'divider'
    },
    {
      key: 'up',
      label: '上移',
      disabled: !canReorderCategory(option.category, context, 'up'),
      icon: renderActionIcon(appIcons.up)
    },
    {
      key: 'down',
      label: '下移',
      disabled: !canReorderCategory(option.category, context, 'down'),
      icon: renderActionIcon(appIcons.down)
    },
    {
      key: 'divider-danger',
      type: 'divider'
    },
  ]
  if (isUncategorizedCategory(option.key)) {
    options.push({
      key: 'clear-bookmarks',
      label: '清空书签',
      disabled: option.category.count <= 0,
      icon: renderActionIcon(appIcons.trashBin)
    })
  } else {
    options.push({
      key: 'delete',
      label: '删除',
      disabled: !canEditCategory(option.category),
      icon: renderActionIcon(appIcons.trashBin)
    })
  }
  return options
}

const renderLabel = ({ option }: { option: TreeOption }) => {
  const categoryOption = option as CategoryTreeOption
  return h(
    NTooltip,
    { delay: 650, trigger: 'hover' },
    {
      trigger: () =>
        h('span', { class: ['category-tree-label', { pinned: categoryOption.category.isPinned }] }, [
          categoryOption.category.isPinned
            ? h(NIcon, { class: 'category-tree-pin', component: appIcons.pinFilled })
            : null,
          h('span', { class: 'category-tree-label-text' }, categoryOption.label)
        ]),
      default: () => categoryOption.label
    }
  )
}

const renderPrefix = ({ option }: { option: TreeOption }) => {
  const categoryOption = option as CategoryTreeOption
  return h(NIcon, {
    class: ['category-tree-prefix', { pinned: categoryOption.category.isPinned }],
    component: isRoot(categoryOption) ? appIcons.library : appIcons.folder
  })
}

const renderSuffix = ({ option }: { option: TreeOption }) => {
  const categoryOption = option as CategoryTreeOption
  const hasChildren = Boolean(categoryOption.children?.length)
  const expanded = expandedKeys.value.includes(categoryOption.key)
  return h(
    'div',
    {
      class: 'category-tree-suffix',
      onClick: stopNodeAction
    },
    [
      h('span', { class: 'category-tree-count' }, String(categoryOption.category.count)),
      h(
        NDropdown,
        {
          options: moreOptions(categoryOption),
          placement: 'bottom-end',
          trigger: 'click',
          onSelect: (key: string | number) => handleMoreSelect(key, categoryOption)
        },
        {
          default: () =>
            h(
              NButton,
              {
                'aria-label': '更多分类操作',
                circle: true,
                class: 'category-tree-action',
                quaternary: true,
                size: 'tiny',
                title: '更多'
              },
              { icon: renderActionIcon(appIcons.dots) }
            )
        }
      ),
      h(
        NButton,
        {
          'aria-label': expanded ? '折叠分类' : '展开分类',
          circle: true,
          class: ['category-tree-expand', { hidden: isRoot(categoryOption) || !hasChildren }],
          quaternary: true,
          size: 'tiny',
          tabindex: isRoot(categoryOption) || !hasChildren ? -1 : 0,
          title: expanded ? '折叠' : '展开',
          onClick: (event: MouseEvent) => toggleExpanded(categoryOption, event)
        },
        { icon: renderActionIcon(expanded ? appIcons.chevronDown : appIcons.chevronRight) }
      )
    ]
  )
}

const renderSwitcherIcon = () => null

const handleSelectedKeys = (keys: Array<string | number>) => {
  const selectedKey = String(keys[0] || 'category_all')
  emit('select', selectedKey)
}

const handleCheckedKeys = (keys: Array<string | number>) => {
  checkedKeys.value = keys.map(String)
}

const handleExpandedKeys = (keys: Array<string | number>) => {
  expandedKeys.value = Array.from(new Set(['category_all'].concat(keys.map(String))))
}

const nodeProps = ({ option }: { option: TreeOption }) => {
  const categoryOption = option as CategoryTreeOption
  return {
    class: {
      'is-pinned': categoryOption.category.isPinned,
      'is-system': categoryOption.category.isSystem
    }
  }
}

watch(
  treeData,
  (data) => {
    const existingKeys = collectExistingKeys(data)
    const nextKeys = expandedKeys.value.filter((key) => existingKeys.has(key))
    checkedKeys.value = checkedKeys.value.filter((key) => existingKeys.has(key))
    if (!nextKeys.includes('category_all')) nextKeys.unshift('category_all')
    expandedKeys.value = nextKeys
  },
  { immediate: true }
)

watch(searchQuery, (value) => {
  if (value.trim()) {
    expandedKeys.value = collectExpandableKeys(treeData.value)
    if (!expandedKeys.value.includes('category_all')) expandedKeys.value.unshift('category_all')
  }
})
</script>

<template>
  <div class="category-tree-shell">
    <div class="category-tree-search">
      <NInput v-model:value="searchQuery" clearable size="small" placeholder="搜索分类">
        <template #prefix>
          <NIcon :component="appIcons.search" />
        </template>
      </NInput>
    </div>

    <div v-if="showBulkBar" class="category-bulk-bar">
      <span class="category-bulk-count">已选 {{ checkedActionableIds.length }}</span>
      <NSpace :size="6" align="center">
        <NTooltip trigger="hover">
          <template #trigger>
            <NButton circle quaternary size="tiny" aria-label="批量上移" :disabled="!batchMoveState.canMoveUp" @click="handleBatchReorder('up')">
              <template #icon>
                <NIcon :component="appIcons.up" />
              </template>
            </NButton>
          </template>
          批量上移
        </NTooltip>
        <NTooltip trigger="hover">
          <template #trigger>
            <NButton circle quaternary size="tiny" aria-label="批量下移" :disabled="!batchMoveState.canMoveDown" @click="handleBatchReorder('down')">
              <template #icon>
                <NIcon :component="appIcons.down" />
              </template>
            </NButton>
          </template>
          批量下移
        </NTooltip>
        <NTooltip trigger="hover">
          <template #trigger>
            <NButton circle quaternary size="tiny" aria-label="批量置顶" :disabled="!batchHasUnpinned" @click="handleBatchPin(true)">
              <template #icon>
                <NIcon :component="appIcons.pin" />
              </template>
            </NButton>
          </template>
          批量置顶
        </NTooltip>
        <NTooltip trigger="hover">
          <template #trigger>
            <NButton circle quaternary size="tiny" aria-label="批量取消置顶" :disabled="!batchHasPinned" @click="handleBatchPin(false)">
              <template #icon>
                <NIcon :component="appIcons.pinFilled" />
              </template>
            </NButton>
          </template>
          批量取消置顶
        </NTooltip>
        <NTooltip trigger="hover">
          <template #trigger>
            <NButton
              circle
              quaternary
              size="tiny"
              type="error"
              aria-label="批量删除"
              :disabled="!checkedBatchDeleteIds.length"
              @click="openBatchDeleteConfirm"
            >
              <template #icon>
                <NIcon :component="appIcons.trashBin" />
              </template>
            </NButton>
          </template>
          批量删除
        </NTooltip>
        <NTooltip trigger="hover">
          <template #trigger>
            <NButton circle quaternary size="tiny" aria-label="清空选择" @click="clearChecked">
              <template #icon>
                <NIcon :component="appIcons.close" />
              </template>
            </NButton>
          </template>
          清空选择
        </NTooltip>
      </NSpace>
    </div>

    <div class="category-tree-scroll">
      <NTree
        block-line
        block-node
        checkable
        cascade
        class="category-naive-tree"
        :checked-keys="checkedKeys"
        :data="treeData"
        :expand-on-click="false"
        :expanded-keys="expandedKeys"
        :indent="16"
        :node-props="nodeProps"
        :render-label="renderLabel"
        :render-prefix="renderPrefix"
        :render-suffix="renderSuffix"
        :render-switcher-icon="renderSwitcherIcon"
        :selected-keys="selectedKeys"
        @update:checked-keys="handleCheckedKeys"
        @update:expanded-keys="handleExpandedKeys"
        @update:selected-keys="handleSelectedKeys"
      />
    </div>

    <NModal
      v-model:show="editorOpen"
      preset="dialog"
      :title="editorMode === 'create' ? '新增分类' : '重命名分类'"
      positive-text="保存"
      negative-text="取消"
      @positive-click="submitEditor"
    >
      <NSpace vertical :size="10">
        <NInput v-model:value="editorDraft" autofocus placeholder="分类名称" @keyup.enter="submitEditor" />
      </NSpace>
    </NModal>

    <NModal
      v-model:show="deleteOpen"
      preset="dialog"
      :title="deleteTitle"
      type="warning"
      positive-text="删除"
      negative-text="取消"
      @positive-click="confirmDelete"
    >
      <NSpace vertical :size="12">
        <span>{{ deleteDescription }}</span>
        <NCheckbox v-model:checked="deleteBookmarks" class="category-delete-option">
          同时删除分类及子分类下所有书签
        </NCheckbox>
      </NSpace>
    </NModal>

    <NModal
      v-model:show="clearBookmarksOpen"
      preset="dialog"
      :title="clearBookmarksTitle"
      type="warning"
      positive-text="清空"
      negative-text="取消"
      @positive-click="confirmClearBookmarks"
    >
      <NSpace vertical :size="12">
        <span>{{ clearBookmarksDescription }}</span>
      </NSpace>
    </NModal>
  </div>
</template>
