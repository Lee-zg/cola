<script setup lang="ts">
import { useAiWorkflow } from '../composables/useAiWorkflow'

const workflow = useAiWorkflow()
</script>

<template>
  <section class="page ai-page">
    <section class="surface action-card">
      <div>
        <span class="eyebrow">AI Assistant</span>
        <h2>批量分析</h2>
        <p>为书签补全描述、关键词、标签等元数据。当前列表待分析 {{ workflow.pendingItems.value.length }} 条。</p>
      </div>
      <button class="primary-action" type="button" :disabled="workflow.analyzing.value" @click="workflow.analyzeAll">
        {{ workflow.analyzing.value ? '分析中' : '全部 AI 分析' }}
      </button>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Queue</span>
          <h2>待分析队列</h2>
        </div>
      </div>
      <div class="compact-list">
        <button
          v-for="item in workflow.pendingItems.value"
          :key="item.id"
          class="compact-row"
          type="button"
          @click="workflow.selectForReview(item)"
        >
          <span>{{ item.title }}</span>
          <small>{{ item.url }}</small>
        </button>
        <div v-if="!workflow.pendingItems.value.length" class="empty">当前列表没有待分析书签</div>
      </div>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Rules</span>
          <h2>规则配置</h2>
        </div>
      </div>
      <p class="hint">当前后端已提供分析动作，规则编辑和失败队列需要新增后端配置与任务 API 后接入。</p>
    </section>
  </section>
</template>
