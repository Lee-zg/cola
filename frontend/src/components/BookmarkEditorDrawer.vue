<script setup lang="ts">
import { useBookmarkStore } from '../stores/bookmarks'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const store = useBookmarkStore()

const splitList = (raw: string): string[] =>
  raw
    .split(',')
    .map((value) => value.trim())
    .filter(Boolean)

const joinList = (values: string[]): string => values.join(', ')

const save = async () => {
  await store.save()
}

const remove = async () => {
  await store.removeSelected()
  emit('close')
}
</script>

<template>
  <div v-if="props.open" class="drawer-scrim" @click.self="emit('close')">
    <form class="editor drawer" @submit.prevent="save">
      <div class="editor-head">
        <div>
          <span class="eyebrow">{{ store.selected ? '编辑' : '新增' }}</span>
          <h2>{{ store.selected ? store.draft.title || '编辑书签' : '新增书签' }}</h2>
        </div>
        <div class="actions">
          <button type="button" @click="emit('close')">关闭</button>
          <button type="button" :disabled="!store.selected" @click="store.analyzeSelected()">AI 分析</button>
          <button type="button" class="danger" :disabled="!store.selected" @click="remove">删除</button>
          <button type="submit" class="primary-action">保存</button>
        </div>
      </div>

      <label class="field">
        <span>标题</span>
        <input v-model="store.draft.title" required />
      </label>
      <label class="field">
        <span>网址</span>
        <input v-model="store.draft.url" required />
      </label>
      <label class="field">
        <span>描述</span>
        <textarea v-model="store.draft.description" rows="5"></textarea>
      </label>
      <div class="two-col">
        <label class="field">
          <span>分类</span>
          <input v-model="store.draft.folder" />
        </label>
        <label class="field">
          <span>标签</span>
          <input
            :value="joinList(store.draft.tags)"
            @input="store.draft.tags = splitList(($event.target as HTMLInputElement).value)"
          />
        </label>
      </div>
      <div class="two-col">
        <label class="field">
          <span>关键词</span>
          <input
            :value="joinList(store.draft.keywords)"
            @input="store.draft.keywords = splitList(($event.target as HTMLInputElement).value)"
          />
        </label>
        <label class="field">
          <span>别名</span>
          <input
            :value="joinList(store.draft.aliases)"
            @input="store.draft.aliases = splitList(($event.target as HTMLInputElement).value)"
          />
        </label>
      </div>
      <p class="status">{{ store.status }}</p>
    </form>
  </div>
</template>
