<script setup lang="ts">
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
    <span>{{ props.total }} 个书签</span>
    <span class="status-dot" :class="{ online: props.server.running }"></span>
    <span>{{ props.server.running ? `Web 服务运行中 ${props.server.url}` : 'Web 服务未启动' }}</span>
    <span class="status-message">{{ props.loading ? '正在同步数据' : props.status || '就绪' }}</span>
  </footer>
</template>
