<!-- 文件说明：frontend/src/components/BookmarkEditorDrawer.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { NButton, NDrawer, NDrawerContent, NDynamicTags, NForm, NFormItem, NIcon, NInput, NPopconfirm, NSpace, NTag } from 'naive-ui'
import { appIcons } from '../icons'
import type { Bookmark, BookmarkInput } from '../types'

const props = defineProps<{
  open: boolean
  selected: Bookmark | null
  draft: BookmarkInput
  status: string
}>()

const emit = defineEmits<{
  analyze: []
  close: []
  remove: []
  save: []
  'update:draft': [patch: Partial<BookmarkInput>]
}>()

const updateTextField = (field: keyof Pick<BookmarkInput, 'title' | 'url' | 'description' | 'folder'>, value: string) => {
  emit('update:draft', { [field]: value })
}

const updateListField = (field: keyof Pick<BookmarkInput, 'tags' | 'keywords' | 'aliases'>, value: string[]) => {
  emit('update:draft', { [field]: value })
}

const handleDrawerUpdate = (value: boolean) => {
  if (!value) emit('close')
}
</script>

<template>
  <NDrawer
    :show="props.open"
    :width="520"
    placement="right"
    :trap-focus="false"
    class="bookmark-editor-drawer"
    @update:show="handleDrawerUpdate"
  >
    <NDrawerContent closable>
      <template #header>
        <div>
          <span class="eyebrow">{{ props.selected ? 'EDIT' : 'CREATE' }}</span>
          <h2 class="drawer-title">{{ props.selected ? props.draft.title || '编辑书签' : '新增书签' }}</h2>
        </div>
      </template>

      <NForm class="editor-form" label-placement="top" @submit.prevent="emit('save')">
        <NFormItem label="标题" required>
          <NInput
            :value="props.draft.title"
            placeholder="书签标题"
            @update:value="updateTextField('title', $event)"
          />
        </NFormItem>

        <NFormItem label="网址" required>
          <NInput :value="props.draft.url" placeholder="https://example.com" @update:value="updateTextField('url', $event)">
            <template #prefix>
              <NIcon :component="appIcons.link" />
            </template>
          </NInput>
        </NFormItem>

        <NFormItem label="描述">
          <NInput
            :value="props.draft.description"
            type="textarea"
            :autosize="{ minRows: 4, maxRows: 8 }"
            placeholder="用于搜索和 AI 复审的简短描述"
            @update:value="updateTextField('description', $event)"
          />
        </NFormItem>

        <NFormItem label="分类">
          <NInput :value="props.draft.folder" placeholder="例如 Work" @update:value="updateTextField('folder', $event)">
            <template #prefix>
              <NIcon :component="appIcons.folder" />
            </template>
          </NInput>
        </NFormItem>

        <NFormItem label="标签">
          <NDynamicTags :value="props.draft.tags" @update:value="updateListField('tags', $event)" />
        </NFormItem>

        <NFormItem label="关键词">
          <NDynamicTags :value="props.draft.keywords" @update:value="updateListField('keywords', $event)" />
        </NFormItem>

        <NFormItem label="别名">
          <NDynamicTags :value="props.draft.aliases" @update:value="updateListField('aliases', $event)" />
        </NFormItem>

        <div v-if="props.selected" class="editor-meta">
          <NTag size="small" round>{{ props.selected.domain || '本地书签' }}</NTag>
          <span>更新于 {{ props.selected.updatedAt || '-' }}</span>
        </div>
      </NForm>

      <template #footer>
        <NSpace justify="space-between" align="center" class="drawer-footer">
          <span class="drawer-status">{{ props.status || '等待编辑' }}</span>
          <NSpace :size="8">
            <NButton :disabled="!props.selected" secondary @click="emit('analyze')">
              <template #icon>
                <NIcon :component="appIcons.ai" />
              </template>
              AI
            </NButton>
            <NPopconfirm :disabled="!props.selected" positive-text="删除" negative-text="取消" @positive-click="emit('remove')">
              <template #trigger>
                <NButton :disabled="!props.selected" type="error" secondary>
                  <template #icon>
                    <NIcon :component="appIcons.trash" />
                  </template>
                </NButton>
              </template>
              删除当前书签？
            </NPopconfirm>
            <NButton type="primary" @click="emit('save')">
              <template #icon>
                <NIcon :component="appIcons.save" />
              </template>
              保存
            </NButton>
          </NSpace>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
