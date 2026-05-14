<script setup lang="ts">
import { useServerWorkflow } from '../composables/useServerWorkflow'

const workflow = useServerWorkflow()
</script>

<template>
  <section class="page webserver-page split-page">
    <section class="surface action-card tall">
      <div>
        <span class="eyebrow">Web Server</span>
        <h2>{{ workflow.server.value.running ? '服务运行中' : '服务未启动' }}</h2>
        <p>{{ workflow.serverDescription.value }}</p>
      </div>
      <button class="primary-action" type="button" @click="workflow.toggleServer">
        {{ workflow.server.value.running ? '停止服务' : '启动服务' }}
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
        <span>{{ workflow.accessLabel.value }}</span>
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
          <dd>{{ workflow.serverPort.value }}</dd>
        </div>
        <div>
          <dt>地址</dt>
          <dd>{{ workflow.server.value.addr || '未启动' }}</dd>
        </div>
      </dl>
    </section>
  </section>
</template>
