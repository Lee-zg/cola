<!-- 文件说明：frontend/src/pages/WebServerPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed } from 'vue'
import { NAlert, NButton, NCard, NDescriptions, NDescriptionsItem, NIcon, NInput, NSpace, NStatistic, NTag, useMessage } from 'naive-ui'
import { useDesktopRuntime } from '../composables/useDesktopRuntime'
import { useServerWorkflow } from '../composables/useServerWorkflow'
import { appIcons } from '../icons'

const workflow = useServerWorkflow()
const runtime = useDesktopRuntime()
const message = useMessage()
const serverStateLabel = computed(() => (workflow.server.value.running ? '运行中' : '未启动'))

const copyAccessUrl = async () => {
  const copied = await runtime.copyText(workflow.server.value.url || workflow.accessLabel.value)
  if (copied) message.success('访问地址已复制')
}

const openAccessUrl = () => {
  if (workflow.server.value.url) runtime.openExternal(workflow.server.value.url)
}
</script>

<template>
  <section class="page webserver-page split-page">
    <NCard class="server-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">WEB SERVER</span>
          <h2>{{ workflow.server.value.running ? '服务运行中' : '服务未启动' }}</h2>
        </div>
      </template>

      <NSpace vertical :size="18">
        <NStatistic label="当前状态" :value="serverStateLabel" />
        <p>{{ workflow.serverDescription.value }}</p>
        <NButton :type="workflow.server.value.running ? 'error' : 'primary'" @click="workflow.toggleServer">
          <template #icon>
            <NIcon :component="workflow.server.value.running ? appIcons.stop : appIcons.server" />
          </template>
          {{ workflow.server.value.running ? '停止服务' : '启动服务' }}
        </NButton>
      </NSpace>
    </NCard>

    <NCard class="server-card" :bordered="false">
      <template #header>访问入口</template>
      <NSpace vertical :size="16">
        <NInput :value="workflow.accessLabel.value" readonly>
          <template #prefix>
            <NIcon :component="appIcons.link" />
          </template>
        </NInput>
        <div class="qr-panel">
          <NIcon :component="appIcons.web" />
          <span>QR</span>
        </div>
        <NSpace>
          <NButton secondary :disabled="!workflow.server.value.running" @click="copyAccessUrl">
            <template #icon>
              <NIcon :component="appIcons.copy" />
            </template>
            复制地址
          </NButton>
          <NButton secondary :disabled="!workflow.server.value.running" @click="openAccessUrl">
            <template #icon>
              <NIcon :component="appIcons.open" />
            </template>
            打开
          </NButton>
        </NSpace>
        <NAlert type="info" :show-icon="false">二维码区域已预留；后续可接入前端二维码生成。</NAlert>
      </NSpace>
    </NCard>

    <NCard class="server-card" :bordered="false">
      <template #header>服务配置</template>
      <NDescriptions label-placement="left" :column="1" bordered>
        <NDescriptionsItem label="端口">{{ workflow.serverPort.value }}</NDescriptionsItem>
        <NDescriptionsItem label="地址">{{ workflow.server.value.addr || '未启动' }}</NDescriptionsItem>
        <NDescriptionsItem label="状态">
          <NTag :type="workflow.server.value.running ? 'success' : 'default'" round>{{ serverStateLabel }}</NTag>
        </NDescriptionsItem>
      </NDescriptions>
    </NCard>
  </section>
</template>
