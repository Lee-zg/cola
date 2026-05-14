<!-- 文件说明：frontend/src/pages/ExportPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed } from 'vue'
import { NAlert, NButton, NCard, NEmpty, NIcon, NInput, NRadioButton, NRadioGroup, NSpace, NTag } from 'naive-ui'
import { useExportWorkflow } from '../composables/useExportWorkflow'
import { appIcons, workflowIcons } from '../icons'

const workflow = useExportWorkflow()

const selectedTemplateName = computed(() => workflow.selectedTemplate.value?.name ?? '未选择模板')
</script>

<template>
  <section class="page export-page">
    <NCard class="workflow-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">THEMES</span>
          <h2>模板画廊</h2>
        </div>
      </template>

      <div class="template-grid">
        <button
          v-for="template in workflow.templates.value"
          :key="template.id"
          class="template-card"
          :class="{ active: workflow.templateId.value === template.id }"
          type="button"
          @click="workflow.selectTemplate(template.id)"
        >
          <span class="template-preview">
            <NIcon :component="workflowIcons[template.id] || workflowIcons.default" />
          </span>
          <strong>{{ template.name }}</strong>
          <small>{{ template.description || template.author || template.id }}</small>
          <NTag v-if="workflow.templateId.value === template.id" size="small" round type="success">选中</NTag>
        </button>
        <NEmpty v-if="!workflow.templates.value.length" description="暂无可用模板" />
      </div>
    </NCard>

    <NCard class="workflow-side-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">EXPORT</span>
          <h2>导出 HTML</h2>
        </div>
      </template>

      <NAlert type="info" :show-icon="false">
        当前模板：{{ selectedTemplateName }}。导出会调用现有后端生成静态 HTML。
      </NAlert>

      <NRadioGroup v-model:value="workflow.exportScope.value">
        <NSpace>
          <NRadioButton value="all">全部书签</NRadioButton>
          <NRadioButton value="filtered">当前筛选结果</NRadioButton>
        </NSpace>
      </NRadioGroup>

      <NInput v-model:value="workflow.exportPath.value" placeholder="导出 HTML 路径，例如 D:\\bookmarks.html">
        <template #prefix>
          <NIcon :component="appIcons.document" />
        </template>
      </NInput>

      <NButton
        type="primary"
        :disabled="workflow.exporting.value || !workflow.templates.value.length"
        :loading="workflow.exporting.value"
        @click="workflow.exportHtml"
      >
        <template #icon>
          <NIcon :component="appIcons.export" />
        </template>
        导出 HTML
      </NButton>
    </NCard>
  </section>
</template>
