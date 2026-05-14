<!-- 文件说明：frontend/src/pages/AiPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed } from 'vue'
import { NAlert, NButton, NCard, NEmpty, NIcon, NList, NListItem, NProgress, NSpace, NThing } from 'naive-ui'
import { useAiWorkflow } from '../composables/useAiWorkflow'
import { useDashboard } from '../composables/useDashboard'
import { appIcons } from '../icons'

const workflow = useAiWorkflow()
const dashboard = useDashboard()
const aiProgress = computed(() => {
  const total = dashboard.stats.value.total
  if (!total) return 0
  return Math.max(0, Math.min(100, Math.round(((total - workflow.pendingItems.value.length) / total) * 100)))
})
</script>

<template>
  <section class="page ai-page">
    <NCard class="workflow-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">AI ASSISTANT</span>
          <h2>AI 助手</h2>
        </div>
      </template>
      <NSpace vertical :size="18">
        <p>为书签补全描述、关键词、标签等元数据。当前列表待分析 {{ workflow.pendingItems.value.length }} 条。</p>
        <NProgress type="line" :percentage="aiProgress" :processing="workflow.analyzing.value" />
        <NButton type="primary" :loading="workflow.analyzing.value" @click="workflow.analyzeAll">
          <template #icon>
            <NIcon :component="appIcons.ai" />
          </template>
          全部 AI 分析
        </NButton>
      </NSpace>
    </NCard>

    <NCard class="workflow-side-card" :bordered="false">
      <template #header>待分析队列</template>
      <NList v-if="workflow.pendingItems.value.length" hoverable clickable>
        <NListItem v-for="item in workflow.pendingItems.value" :key="item.id" @click="workflow.selectForReview(item)">
          <NThing :title="item.title" :description="item.url" />
        </NListItem>
      </NList>
      <NEmpty v-else description="当前列表没有待分析书签" />
    </NCard>

    <NCard class="workflow-side-card" :bordered="false">
      <template #header>规则配置</template>
      <NAlert type="info" :show-icon="false">当前后端已提供分析动作，规则编辑和失败队列需新增任务 API 后接入。</NAlert>
    </NCard>
  </section>
</template>
