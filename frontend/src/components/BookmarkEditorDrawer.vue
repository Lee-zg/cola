<script setup lang="ts">
import { joinList, splitList } from '../helpers/bookmarkLists'
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

const updateListField = (field: keyof Pick<BookmarkInput, 'tags' | 'keywords' | 'aliases'>, value: string) => {
  emit('update:draft', { [field]: splitList(value) })
}
</script>

<template>
  <div v-if="props.open" class="drawer-scrim" @click.self="emit('close')">
    <form class="editor drawer" @submit.prevent="emit('save')">
      <div class="editor-head">
        <div>
          <span class="eyebrow">{{ props.selected ? '编辑' : '新增' }}</span>
          <h2>{{ props.selected ? props.draft.title || '编辑书签' : '新增书签' }}</h2>
        </div>
        <div class="actions">
          <button type="button" @click="emit('close')">关闭</button>
          <button type="button" :disabled="!props.selected" @click="emit('analyze')">AI 分析</button>
          <button type="button" class="danger" :disabled="!props.selected" @click="emit('remove')">删除</button>
          <button type="submit" class="primary-action">保存</button>
        </div>
      </div>

      <label class="field">
        <span>标题</span>
        <input :value="props.draft.title" required @input="updateTextField('title', ($event.target as HTMLInputElement).value)" />
      </label>
      <label class="field">
        <span>网址</span>
        <input :value="props.draft.url" required @input="updateTextField('url', ($event.target as HTMLInputElement).value)" />
      </label>
      <label class="field">
        <span>描述</span>
        <textarea
          :value="props.draft.description"
          rows="5"
          @input="updateTextField('description', ($event.target as HTMLTextAreaElement).value)"
        ></textarea>
      </label>
      <div class="two-col">
        <label class="field">
          <span>分类</span>
          <input :value="props.draft.folder" @input="updateTextField('folder', ($event.target as HTMLInputElement).value)" />
        </label>
        <label class="field">
          <span>标签</span>
          <input :value="joinList(props.draft.tags)" @input="updateListField('tags', ($event.target as HTMLInputElement).value)" />
        </label>
      </div>
      <div class="two-col">
        <label class="field">
          <span>关键词</span>
          <input
            :value="joinList(props.draft.keywords)"
            @input="updateListField('keywords', ($event.target as HTMLInputElement).value)"
          />
        </label>
        <label class="field">
          <span>别名</span>
          <input
            :value="joinList(props.draft.aliases)"
            @input="updateListField('aliases', ($event.target as HTMLInputElement).value)"
          />
        </label>
      </div>
      <p class="status">{{ props.status }}</p>
    </form>
  </div>
</template>
