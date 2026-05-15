<!-- 文件说明：frontend/src/components/CategoryTree.vue，基于 Naive UI Tree 二次封装的书签分类树。 -->
<script setup lang="ts">
import { computed, h, ref, watch, type Component } from 'vue'
import { NButton, NCheckbox, NDropdown, NIcon, NInput, NModal, NSpace, NTooltip, NTree } from 'naive-ui'
import type { DropdownOption, TreeOption } from 'naive-ui'
import { appIcons } from '../icons'
import type { CategoryNode } from '../types'

type CategoryReorderDirection = 'top' | 'up' | 'down'
type CategoryEditorMode = 'create' | 'rename'

interface CategoryTreeOption extends TreeOption {
  key: string
  label: string
  category: CategoryNode
  children?: CategoryTreeOption[]
}

const props = defineProps<{
  categories: CategoryNode[]
  selectedId: string
}>()

const emit = defineEmits<{
  create: [name: string, parentId: string]
  delete: [id: string, deleteBookmarks: boolean]
  rename: [id: string, name: string]
  reorder: [id: string, direction: CategoryReorderDirection]
  select: [id: string]
}>()

const searchQuery = ref('')
const expandedKeys = ref<string[]>(['category_all'])
const editorOpen = ref(false)
const editorMode = ref<CategoryEditorMode>('create')
const editorDraft = ref('')
const editorTargetId = ref('')
const editorParentId = ref('category_all')
const deleteOpen = ref(false)
const deleteTarget = ref<CategoryTreeOption | null>(null)
const deleteBookmarks = ref(false)

const selectedKeys = computed(() => [props.selectedId])

const normalizeText = (value: string) => value.trim().toLowerCase()

const toTreeOption = (node: CategoryNode): CategoryTreeOption => ({
  key: node.id,
  label: node.name,
  category: node,
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

const siblingMeta = computed(() => {
  const meta = new Map<string, { index: number; siblings: CategoryNode[] }>()
  const visit = (nodes: CategoryNode[]) => {
    nodes.forEach((node, index) => {
      meta.set(node.id, { index, siblings: nodes })
      visit(node.children)
    })
  }
  for (const category of props.categories) {
    meta.set(category.id, { index: 0, siblings: props.categories })
    visit(category.children)
  }
  return meta
})

const isRoot = (option: CategoryTreeOption) => option.key === 'category_all'
const isUncategorized = (option: CategoryTreeOption) => option.key === 'category_uncategorized'
const canCreateChild = (option: CategoryTreeOption) => !isUncategorized(option)
const canEdit = (option: CategoryTreeOption) => !option.category.isSystem

const firstMovableIndex = (siblings: CategoryNode[]) => {
  const index = siblings.findIndex((category) => !category.isSystem)
  return index >= 0 ? index : 0
}

const canReorder = (option: CategoryTreeOption, direction: CategoryReorderDirection) => {
  if (!canEdit(option)) return false
  const meta = siblingMeta.value.get(option.key)
  if (!meta) return false
  const firstIndex = firstMovableIndex(meta.siblings)
  if (direction === 'top') return meta.index > firstIndex
  if (direction === 'up') return meta.index > firstIndex
  return meta.index < meta.siblings.length - 1
}

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
  if (!canEdit(option)) return
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

const openDeleteConfirm = (option: CategoryTreeOption) => {
  if (!canEdit(option)) return
  deleteTarget.value = option
  deleteBookmarks.value = false
  deleteOpen.value = true
}

const confirmDelete = () => {
  if (deleteTarget.value) {
    emit('delete', deleteTarget.value.key, deleteBookmarks.value)
  }
  deleteOpen.value = false
  deleteTarget.value = null
  deleteBookmarks.value = false
}

const handleMoreSelect = (key: string | number, option: CategoryTreeOption) => {
  const action = String(key)
  if (action === 'delete') {
    openDeleteConfirm(option)
    return
  }
  if (action === 'top' || action === 'up' || action === 'down') {
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

const renderIconButton = (label: string, component: Component, disabled: boolean, onClick: (event: MouseEvent) => void) =>
  h(
    NTooltip,
    { delay: 550, trigger: 'hover' },
    {
      trigger: () =>
        h(
          NButton,
          {
            'aria-label': label,
            circle: true,
            class: 'category-tree-action',
            disabled,
            quaternary: true,
            size: 'tiny',
            title: label,
            onClick
          },
          { icon: renderActionIcon(component) }
        ),
      default: () => label
    }
  )

const moreOptions = (option: CategoryTreeOption): DropdownOption[] => [
  {
    key: 'up',
    label: '上移',
    disabled: !canReorder(option, 'up'),
    icon: renderActionIcon(appIcons.up)
  },
  {
    key: 'down',
    label: '下移',
    disabled: !canReorder(option, 'down'),
    icon: renderActionIcon(appIcons.down)
  },
  {
    key: 'top',
    label: '置顶',
    disabled: !canReorder(option, 'top'),
    icon: renderActionIcon(appIcons.pin)
  },
  {
    key: 'divider',
    type: 'divider'
  },
  {
    key: 'delete',
    label: '删除',
    disabled: !canEdit(option),
    icon: renderActionIcon(appIcons.trash)
  }
]

const renderLabel = ({ option }: { option: TreeOption }) => {
  const categoryOption = option as CategoryTreeOption
  return h(
    NTooltip,
    { delay: 650, trigger: 'hover' },
    {
      trigger: () => h('span', { class: 'category-tree-label' }, categoryOption.label),
      default: () => categoryOption.label
    }
  )
}

const renderPrefix = ({ option }: { option: TreeOption }) => {
  const categoryOption = option as CategoryTreeOption
  return h(NIcon, {
    class: 'category-tree-prefix',
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
      h('div', { class: 'category-tree-actions' }, [
        renderIconButton('新增子分类', appIcons.add, !canCreateChild(categoryOption), (event) => openCreateEditor(categoryOption.key, event)),
        renderIconButton('修改分类', appIcons.pencil, !canEdit(categoryOption), (event) => openRenameEditor(categoryOption, event)),
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
        )
      ]),
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

const handleExpandedKeys = (keys: Array<string | number>) => {
  expandedKeys.value = Array.from(new Set(['category_all'].concat(keys.map(String))))
}

watch(
  treeData,
  (data) => {
    const existingKeys = collectExistingKeys(data)
    const nextKeys = expandedKeys.value.filter((key) => existingKeys.has(key))
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
      <NTooltip trigger="hover">
        <template #trigger>
          <NButton circle size="small" type="primary" aria-label="新增一级分类" @click="openCreateEditor('category_all')">
            <template #icon>
              <NIcon :component="appIcons.add" />
            </template>
          </NButton>
        </template>
        新增一级分类
      </NTooltip>
    </div>

    <div class="category-tree-scroll">
      <NTree
        block-line
        block-node
        class="category-naive-tree"
        :data="treeData"
        :expand-on-click="false"
        :expanded-keys="expandedKeys"
        :indent="14"
        :render-label="renderLabel"
        :render-prefix="renderPrefix"
        :render-suffix="renderSuffix"
        :render-switcher-icon="renderSwitcherIcon"
        :selected-keys="selectedKeys"
        @update:expanded-keys="handleExpandedKeys"
        @update:selected-keys="handleSelectedKeys"
      />
    </div>

    <NModal
      v-model:show="editorOpen"
      preset="dialog"
      :title="editorMode === 'create' ? '新增分类' : '修改分类'"
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
      title="删除分类"
      type="warning"
      positive-text="删除"
      negative-text="取消"
      @positive-click="confirmDelete"
    >
      <NSpace vertical :size="12">
        <span>默认删除后，该分类及子分类下书签会移动到父级。</span>
        <NCheckbox v-model:checked="deleteBookmarks" class="category-delete-option">
          同时删除该分类及子分类下所有书签
        </NCheckbox>
      </NSpace>
    </NModal>
  </div>
</template>
