<!-- 文件说明：frontend/src/pages/ImportPage.vue，书签导入中心，提供来源选择、导入选项和结果反馈。 -->
<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { NAlert, NButton, NCard, NCheckbox, NEmpty, NIcon, NInput, NProgress, NTag } from 'naive-ui'
import { useDesktopRuntime } from '../composables/useDesktopRuntime'
import { useImportWorkflow } from '../composables/useImportWorkflow'
import { appIcons, sourceIcons } from '../icons'

const workflow = useImportWorkflow()
const runtime = useDesktopRuntime()
let stopFileDrop: () => void = () => undefined

const handlePanelDrop = (event: DragEvent) => {
  event.preventDefault()
  const files = Array.from(event.dataTransfer?.files || [])
  const paths = files.map((file) => (file as File & { path?: string }).path || file.name)
  workflow.applyDroppedFiles(paths)
}

onMounted(() => {
  stopFileDrop = runtime.onFileDrop((payload) => workflow.applyDroppedFiles(payload.paths))
})

onBeforeUnmount(() => {
  stopFileDrop()
})
</script>

<template>
  <section class="page import-page import-workbench">
    <div class="import-main panel-surface" @drop="handlePanelDrop" @dragover.prevent>
      <header class="import-hero">
        <div>
          <span class="eyebrow">IMPORT</span>
          <h2>导入中心</h2>
          <p>选择来源，确认导入规则，然后把外部书签合并进当前书签库。</p>
        </div>
        <NTag :type="workflow.importing.value ? 'warning' : 'success'" round>
          {{ workflow.importing.value ? '导入中' : '准备就绪' }}
        </NTag>
      </header>

      <section class="import-section" aria-labelledby="import-source-title">
        <div class="import-section-title">
          <span>1</span>
          <div>
            <h3 id="import-source-title">选择来源</h3>
            <p>{{ workflow.selectedSource.value.description }}</p>
          </div>
        </div>

        <div class="import-source-grid">
          <button
            v-for="source in workflow.sourceOptions"
            :key="source.id"
            class="import-source-card"
            :class="{ active: workflow.importSource.value === source.id }"
            type="button"
            @click="workflow.setImportSource(source.id)"
          >
            <NIcon :component="sourceIcons[source.id]" />
            <span>
              <strong>{{ source.name }}</strong>
              <small>{{ source.autoPathSupported ? '支持自动扫描' : '需要文件路径' }}</small>
            </span>
          </button>
        </div>
      </section>

      <section class="import-section" aria-labelledby="import-path-title">
        <div class="import-section-title">
          <span>2</span>
          <div>
            <h3 id="import-path-title">路径与规则</h3>
            <p>{{ workflow.selectedSource.value.pathHint }}</p>
          </div>
        </div>

        <div class="import-path-panel" :class="{ required: workflow.importPathRequired.value && !workflow.importPath.value }">
          <NIcon :component="workflow.importPath.value ? appIcons.document : appIcons.import" />
          <div>
            <NInput
              :value="workflow.importPath.value"
              clearable
              :placeholder="workflow.selectedSource.value.pathHint"
              @update:value="workflow.setImportPath"
            />
            <span>支持拖放文件到当前页面。Chrome、Edge、Firefox 留空时会尝试读取默认书签数据。</span>
          </div>
        </div>

        <div class="import-option-grid">
          <label class="import-option-row">
            <NCheckbox v-model:checked="workflow.skipDuplicates.value" />
            <span>
              <strong>跳过重复 URL</strong>
              <small>关闭后会按 URL 更新已有书签的标题、分类和标签。</small>
            </span>
          </label>
          <label class="import-option-row">
            <NCheckbox v-model:checked="workflow.keepFolders.value" />
            <span>
              <strong>保留原分类</strong>
              <small>关闭后全部导入到未分类，避免生成大量来源目录。</small>
            </span>
          </label>
          <label class="import-option-row">
            <NCheckbox v-model:checked="workflow.autoAnalyze.value" />
            <span>
              <strong>自动 AI 分析</strong>
              <small>导入成功后立即为新增或更新的书签补充标签和关键词。</small>
            </span>
          </label>
        </div>
      </section>

      <section class="import-section import-run-section" aria-labelledby="import-run-title">
        <div class="import-section-title">
          <span>3</span>
          <div>
            <h3 id="import-run-title">执行导入</h3>
            <p>{{ workflow.resultSummary.value }}</p>
          </div>
        </div>

        <div class="import-run-panel">
          <NProgress
            :percentage="workflow.importing.value ? 72 : workflow.lastResult.value ? 100 : 0"
            :processing="workflow.importing.value"
            type="line"
          />
          <NButton type="primary" size="large" :loading="workflow.importing.value" :disabled="!workflow.canStartImport.value" @click="workflow.startImport">
            <template #icon>
              <NIcon :component="appIcons.import" />
            </template>
            开始导入
          </NButton>
        </div>

        <NAlert v-if="workflow.pathValidationMessage.value || workflow.lastError.value" type="error" :show-icon="false">
          {{ workflow.pathValidationMessage.value || workflow.lastError.value }}
        </NAlert>
      </section>
    </div>

    <aside class="import-side">
      <NCard class="workflow-side-card import-result-card" :bordered="false">
        <template #header>本次结果</template>
        <div class="import-result-grid">
          <div>
            <span>新增</span>
            <strong>{{ workflow.lastResult.value?.imported || 0 }}</strong>
          </div>
          <div>
            <span>更新</span>
            <strong>{{ workflow.lastResult.value?.updated || 0 }}</strong>
          </div>
          <div>
            <span>跳过</span>
            <strong>{{ workflow.lastResult.value?.skipped || 0 }}</strong>
          </div>
          <div>
            <span>分析</span>
            <strong>{{ workflow.lastResult.value?.analyzed || 0 }}</strong>
          </div>
        </div>
        <NAlert v-if="workflow.lastResult.value?.errors.length" type="warning" :show-icon="false">
          {{ workflow.lastResult.value.errors.slice(0, 3).join('；') }}
        </NAlert>
      </NCard>

      <NCard class="workflow-side-card import-history-card" :bordered="false">
        <template #header>导入记录</template>
        <div v-if="workflow.importHistory.value.length" class="import-history-list">
          <article v-for="record in workflow.importHistory.value" :key="`${record.source}-${record.time}`" class="import-history-item">
            <div>
              <NTag size="small" :type="record.type === 'success' ? 'success' : 'error'" round>{{ record.source }}</NTag>
              <strong>{{ record.message }}</strong>
              <span>{{ record.time }}</span>
            </div>
          </article>
        </div>
        <NEmpty v-else description="本次会话暂无导入记录" />
      </NCard>
    </aside>
  </section>
</template>
