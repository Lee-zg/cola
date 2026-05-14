<!-- 文件说明：frontend/src/components/WindowTitleBar.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { NButton, NIcon, NTooltip } from 'naive-ui'
import { appIcons } from '../icons'

const props = defineProps<{
  runtimeAvailable: boolean
}>()

const emit = defineEmits<{
  close: []
  minimise: []
  toggleMaximise: []
}>()
</script>

<template>
  <header class="window-titlebar">
    <div class="window-drag-region">
      <div class="brand-mark" aria-hidden="true">CB</div>
      <div class="window-titlecopy">
        <strong>Cola Bookmarks</strong>
        <span>{{ props.runtimeAvailable ? 'Wails Desktop' : 'Browser Preview' }}</span>
      </div>
    </div>

    <div class="window-controls">
      <NTooltip trigger="hover">
        <template #trigger>
          <NButton quaternary circle size="small" aria-label="最小化" @click="emit('minimise')">
            <template #icon>
              <NIcon :component="appIcons.minimise" />
            </template>
          </NButton>
        </template>
        最小化
      </NTooltip>
      <NTooltip trigger="hover">
        <template #trigger>
          <NButton quaternary circle size="small" aria-label="最大化或还原" @click="emit('toggleMaximise')">
            <template #icon>
              <NIcon :component="appIcons.maximise" />
            </template>
          </NButton>
        </template>
        最大化/还原
      </NTooltip>
      <NTooltip trigger="hover">
        <template #trigger>
          <NButton quaternary circle size="small" type="error" aria-label="关闭" @click="emit('close')">
            <template #icon>
              <NIcon :component="appIcons.close" />
            </template>
          </NButton>
        </template>
        关闭
      </NTooltip>
    </div>
  </header>
</template>
