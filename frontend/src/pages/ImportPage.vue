<script setup lang="ts">
import { useImportWorkflow } from '../composables/useImportWorkflow'

const workflow = useImportWorkflow()
</script>

<template>
  <section class="page import-page">
    <section class="surface flow-panel">
      <div class="section-head">
        <div>
          <span class="eyebrow">Import</span>
          <h2>选择来源</h2>
        </div>
      </div>

      <div class="option-grid">
        <button
          v-for="source in workflow.sourceOptions"
          :key="source.id"
          class="option-card"
          :class="{ active: workflow.importSource.value === source.id }"
          type="button"
          @click="workflow.importSource.value = source.id"
        >
          <strong>{{ source.name }}</strong>
          <span>{{ source.description }}</span>
        </button>
      </div>

      <label class="field">
        <span>文件路径</span>
        <input v-model="workflow.importPath.value" placeholder="HTML/浏览器数据文件路径；自动扫描可留空" />
      </label>

      <div class="check-row">
        <label><input v-model="workflow.skipDuplicates.value" type="checkbox" /> 跳过重复</label>
        <label><input v-model="workflow.autoAnalyze.value" type="checkbox" /> 自动 AI 分析</label>
        <label><input v-model="workflow.keepFolders.value" type="checkbox" /> 保留原分类</label>
      </div>

      <div class="progress-block" :class="{ active: workflow.importing.value }">
        <div class="progress-track"><span></span></div>
        <button class="primary-action" type="button" :disabled="workflow.importing.value" @click="workflow.startImport">
          {{ workflow.importing.value ? '导入中' : '开始导入' }}
        </button>
      </div>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">History</span>
          <h2>导入记录</h2>
        </div>
      </div>
      <div class="compact-list">
        <div v-for="record in workflow.importHistory.value" :key="`${record.source}-${record.time}`" class="compact-row static">
          <span>{{ record.source }} · {{ record.message }}</span>
          <small>{{ record.time }}</small>
        </div>
        <div v-if="!workflow.importHistory.value.length" class="empty">本次会话暂无导入记录</div>
      </div>
    </section>
  </section>
</template>
