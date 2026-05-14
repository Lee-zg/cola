<script setup lang="ts">
import { computed } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'

const store = useBookmarkStore()

const serverPort = computed(() => {
  if (!store.server.addr) return '默认'
  const parts = store.server.addr.split(':')
  return parts[parts.length - 1] || '默认'
})
</script>

<template>
  <section class="page webserver-page split-page">
    <section class="surface action-card tall">
      <div>
        <span class="eyebrow">Web Server</span>
        <h2>{{ store.server.running ? '服务运行中' : '服务未启动' }}</h2>
        <p>{{ store.server.running ? store.server.url : '启动后可在浏览器或局域网设备访问导出的书签页面。' }}</p>
      </div>
      <button class="primary-action" type="button" @click="store.toggleServer()">
        {{ store.server.running ? '停止服务' : '启动服务' }}
      </button>
    </section>

    <section class="surface flow-panel">
      <div class="section-head">
        <div>
          <span class="eyebrow">Access</span>
          <h2>访问入口</h2>
        </div>
      </div>
      <div class="server-url-box">
        <span>{{ store.server.running ? store.server.url : '启动服务后显示链接' }}</span>
      </div>
      <div class="qr-placeholder">QR</div>
      <p class="hint">二维码占位已预留；当前后端未提供二维码生成接口，可后续接入前端生成。</p>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Config</span>
          <h2>服务配置</h2>
        </div>
      </div>
      <dl class="detail-list">
        <div>
          <dt>端口</dt>
          <dd>{{ serverPort }}</dd>
        </div>
        <div>
          <dt>地址</dt>
          <dd>{{ store.server.addr || '未启动' }}</dd>
        </div>
      </dl>
    </section>
  </section>
</template>
