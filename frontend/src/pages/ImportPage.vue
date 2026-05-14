<script setup lang="ts">
import { ref } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'

const store = useBookmarkStore()
const importPath = ref('')
const importSource = ref('html')
const importing = ref(false)
const importHistory = ref<Array<{ source: string; message: string; time: string }>>([])

const skipDuplicates = ref(true)
const autoAnalyze = ref(false)
const keepFolders = ref(true)

const sourceOptions = [
  { id: 'html', name: 'HTML 文件', description: '从浏览器导出的 bookmarks.html 导入' },
  { id: 'chrome', name: 'Chrome', description: '自动扫描 Chrome 默认书签数据' },
  { id: 'edge', name: 'Edge', description: '自动扫描 Edge 默认书签数据' },
  { id: 'firefox', name: 'Firefox', description: '自动扫描 Firefox 默认书签数据' }
]

const startImport = async () => {
  importing.value = true
  try {
    await store.importFrom(importSource.value, importPath.value)
    importHistory.value.unshift({
      source: importSource.value,
      message: store.status,
      time: new Date().toLocaleString()
    })
  } finally {
    importing.value = false
  }
}
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
          v-for="source in sourceOptions"
          :key="source.id"
          class="option-card"
          :class="{ active: importSource === source.id }"
          type="button"
          @click="importSource = source.id"
        >
          <strong>{{ source.name }}</strong>
          <span>{{ source.description }}</span>
        </button>
      </div>

      <label class="field">
        <span>文件路径</span>
        <input v-model="importPath" placeholder="HTML/浏览器数据文件路径；自动扫描可留空" />
      </label>

      <div class="check-row">
        <label><input v-model="skipDuplicates" type="checkbox" /> 跳过重复</label>
        <label><input v-model="autoAnalyze" type="checkbox" /> 自动 AI 分析</label>
        <label><input v-model="keepFolders" type="checkbox" /> 保留原分类</label>
      </div>

      <div class="progress-block" :class="{ active: importing }">
        <div class="progress-track"><span></span></div>
        <button class="primary-action" type="button" :disabled="importing" @click="startImport">
          {{ importing ? '导入中' : '开始导入' }}
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
        <div v-for="record in importHistory" :key="`${record.source}-${record.time}`" class="compact-row static">
          <span>{{ record.source }} · {{ record.message }}</span>
          <small>{{ record.time }}</small>
        </div>
        <div v-if="!importHistory.length" class="empty">本次会话暂无导入记录</div>
      </div>
    </section>
  </section>
</template>
