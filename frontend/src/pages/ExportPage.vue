<script setup lang="ts">
import { useExportWorkflow } from '../composables/useExportWorkflow'

const workflow = useExportWorkflow()
</script>

<template>
  <section class="page export-page">
    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Themes</span>
          <h2>模板画廊</h2>
        </div>
      </div>

      <div class="template-grid">
        <button
          v-for="template in workflow.templates.value"
          :key="template.id"
          class="template-card"
          :class="{ active: workflow.templateId.value === template.id }"
          type="button"
          @click="workflow.selectTemplate(template.id)"
        >
          <span class="template-preview">{{ template.name.slice(0, 2).toUpperCase() }}</span>
          <strong>{{ template.name }}</strong>
          <small>{{ template.description || template.author || template.id }}</small>
        </button>
        <div v-if="!workflow.templates.value.length" class="empty">暂无可用模板</div>
      </div>
    </section>

    <section class="surface flow-panel">
      <div class="section-head">
        <div>
          <span class="eyebrow">Export</span>
          <h2>导出 HTML</h2>
        </div>
      </div>

      <div v-if="workflow.selectedTemplate.value" class="selected-template">
        <strong>{{ workflow.selectedTemplate.value.name }}</strong>
        <span>{{ workflow.selectedTemplate.value.description || '当前选择的导出主题模板' }}</span>
      </div>

      <div class="radio-row">
        <label><input v-model="workflow.exportScope.value" type="radio" value="all" /> 全部书签</label>
        <label><input v-model="workflow.exportScope.value" type="radio" value="filtered" /> 当前筛选结果</label>
      </div>

      <label class="field">
        <span>输出路径</span>
        <input v-model="workflow.exportPath.value" placeholder="导出 HTML 路径，例如 D:\\bookmarks.html" />
      </label>

      <button
        class="primary-action"
        type="button"
        :disabled="workflow.exporting.value || !workflow.templates.value.length"
        @click="workflow.exportHtml"
      >
        {{ workflow.exporting.value ? '导出中' : '导出 HTML' }}
      </button>
    </section>
  </section>
</template>
