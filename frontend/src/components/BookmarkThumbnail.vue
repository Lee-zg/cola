<!-- 文件说明：frontend/src/components/BookmarkThumbnail.vue，统一书签缩略图展示和图片预览交互。 -->
<script setup lang="ts">
import { computed } from 'vue'
import { NImage } from 'naive-ui'

const props = withDefaults(
  defineProps<{
    src?: string
    fallbackText?: string
    compact?: boolean
    actionLabel?: string
  }>(),
  {
    src: '',
    fallbackText: '?',
    compact: false,
    actionLabel: ''
  }
)

const emit = defineEmits<{
  action: []
}>()

const fallbackLetter = computed(() => props.fallbackText.slice(0, 1).toUpperCase() || '?')
</script>

<template>
  <div class="bookmark-thumbnail" :class="{ small: props.compact, 'has-image': Boolean(props.src) }" @click.stop>
    <!-- NImage 保留 Naive UI 原生预览层，列表和详情都能点击查看原图。 -->
    <NImage
      v-if="props.src"
      class="bookmark-thumbnail__image"
      :src="props.src"
      :preview-src="props.src"
      object-fit="cover"
    />
    <span v-else>{{ fallbackLetter }}</span>
    <button v-if="props.actionLabel" class="preview-action" type="button" @click.stop="emit('action')">{{ props.actionLabel }}</button>
  </div>
</template>
