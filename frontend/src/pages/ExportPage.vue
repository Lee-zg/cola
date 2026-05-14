<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'

const store = useBookmarkStore()
const exportPath = ref('')
const templateId = ref('classic')
const exporting = ref(false)
const exportScope = ref<'all' | 'filtered'>('all')

const selectedTemplate = computed(() => store.templates.find((template) => template.id === templateId.value))

watch(
  () => store.templates,
  (templates) => {
    if (templates.length && !templates.some((template) => template.id === templateId.value)) {
      templateId.value = templates[0].id
    }
  },
  { immediate: true }
)

const exportHtml = async () => {
  exporting.value = true
  try {
    await store.exportTo(exportPath.value, templateId.value)
  } finally {
    exporting.value = false
  }
}
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
          v-for="template in store.templates"
          :key="template.id"
          class="template-card"
          :class="{ active: templateId === template.id }"
          type="button"
          @click="templateId = template.id"
        >
          <span class="template-preview">{{ template.name.slice(0, 2).toUpperCase() }}</span>
          <strong>{{ template.name }}</strong>
          <small>{{ template.description || template.author || template.id }}</small>
        </button>
        <div v-if="!store.templates.length" class="empty">暂无可用模板</div>
      </div>
    </section>

    <section class="surface flow-panel">
      <div class="section-head">
        <div>
          <span class="eyebrow">Export</span>
          <h2>导出 HTML</h2>
        </div>
      </div>

      <div v-if="selectedTemplate" class="selected-template">
        <strong>{{ selectedTemplate.name }}</strong>
        <span>{{ selectedTemplate.description || '当前选择的导出主题模板' }}</span>
      </div>

      <div class="radio-row">
        <label><input v-model="exportScope" type="radio" value="all" /> 全部书签</label>
        <label><input v-model="exportScope" type="radio" value="filtered" /> 当前筛选结果</label>
      </div>

      <label class="field">
        <span>输出路径</span>
        <input v-model="exportPath" placeholder="导出 HTML 路径，例如 D:\\bookmarks.html" />
      </label>

      <button class="primary-action" type="button" :disabled="exporting || !store.templates.length" @click="exportHtml">
        {{ exporting ? '导出中' : '导出 HTML' }}
      </button>
    </section>
  </section>
</template>
