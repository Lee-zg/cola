<!-- 文件说明：frontend/src/pages/ImportPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { NAlert, NButton, NCard, NCheckbox, NEmpty, NIcon, NInput, NList, NListItem, NProgress, NSpace, NSteps, NStep } from 'naive-ui'
import { useImportWorkflow } from '../composables/useImportWorkflow'
import { sourceIcons } from '../icons'

const workflow = useImportWorkflow()
</script>

<template>
  <section class="page import-page">
    <NCard class="workflow-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">IMPORT</span>
          <h2>导入中心</h2>
        </div>
      </template>

      <NSteps :current="workflow.importing.value ? 2 : 1" size="small" class="workflow-steps">
        <NStep title="选择来源" />
        <NStep title="执行导入" />
        <NStep title="检查记录" />
      </NSteps>

      <div class="option-grid">
        <button
          v-for="source in workflow.sourceOptions"
          :key="source.id"
          class="option-card"
          :class="{ active: workflow.importSource.value === source.id }"
          type="button"
          @click="workflow.importSource.value = source.id"
        >
          <NIcon :component="sourceIcons[source.id]" />
          <strong>{{ source.name }}</strong>
          <span>{{ source.description }}</span>
        </button>
      </div>

      <NInput v-model:value="workflow.importPath.value" placeholder="HTML/浏览器数据文件路径；自动扫描可留空" />

      <NSpace wrap>
        <NCheckbox v-model:checked="workflow.skipDuplicates.value">跳过重复</NCheckbox>
        <NCheckbox v-model:checked="workflow.autoAnalyze.value">自动 AI 分析</NCheckbox>
        <NCheckbox v-model:checked="workflow.keepFolders.value">保留原分类</NCheckbox>
      </NSpace>

      <NAlert type="info" :show-icon="false">
        浏览器来源可留空路径，后端会尝试读取默认书签位置；HTML 导入建议填写完整文件路径。
      </NAlert>

      <div class="workflow-actions">
        <NProgress v-if="workflow.importing.value" type="line" :percentage="78" processing />
        <NButton type="primary" :loading="workflow.importing.value" @click="workflow.startImport">开始导入</NButton>
      </div>
    </NCard>

    <NCard class="workflow-side-card" :bordered="false">
      <template #header>导入记录</template>
      <NList v-if="workflow.importHistory.value.length">
        <NListItem v-for="record in workflow.importHistory.value" :key="`${record.source}-${record.time}`">
          <span>{{ record.source }} · {{ record.message }}</span>
          <template #suffix>{{ record.time }}</template>
        </NListItem>
      </NList>
      <NEmpty v-else description="本次会话暂无导入记录" />
    </NCard>
  </section>
</template>
