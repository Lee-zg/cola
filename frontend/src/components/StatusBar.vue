<!-- 文件说明：frontend/src/components/StatusBar.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { NIcon, NProgress } from 'naive-ui'
import { appIcons } from '../icons'
import type { ServerStatus } from '../types'

const props = defineProps<{
  total: number
  server: ServerStatus
  status: string
  loading: boolean
}>()
</script>

<template>
  <footer class="status-bar">
    <span class="status-item">
      <NIcon :component="appIcons.library" />
      {{ props.total }} 个书签
    </span>
    <span class="status-item">
      <span class="status-dot" :class="{ online: props.server.running }"></span>
      {{ props.server.running ? `Web 服务运行中 ${props.server.url}` : 'Web 服务未启动' }}
    </span>
    <span class="status-message">{{ props.loading ? '正在同步数据' : props.status || '就绪' }}</span>
    <NProgress v-if="props.loading" class="status-progress" type="line" :percentage="66" :show-indicator="false" processing />
  </footer>
</template>
