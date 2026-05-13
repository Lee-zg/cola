<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useBookmarkStore } from './stores/bookmarks'
import type { Bookmark } from './types'

const store = useBookmarkStore()
const importPath = ref('')
const importSource = ref('html')
const exportPath = ref('')
const templateId = ref('classic')
const backupPath = ref('')
const restorePath = ref('')

let searchTimer = 0
watch(
  () => [store.query, store.folder, store.tag],
  () => {
    window.clearTimeout(searchTimer)
    searchTimer = window.setTimeout(() => store.refresh(), 220)
  }
)

onMounted(() => {
  store.refresh().catch((err) => {
    store.status = err instanceof Error ? err.message : String(err)
  })
})

const selectedId = computed(() => store.selected?.id ?? '')

function choose(item: Bookmark) {
  store.select(item)
}

function splitList(raw: string): string[] {
  return raw
    .split(',')
    .map((value) => value.trim())
    .filter(Boolean)
}

function joinList(values: string[]): string {
  return values.join(', ')
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <div>
          <h1>Cola Bookmarks</h1>
          <p>本地优先的浏览器书签管理器</p>
        </div>
        <button class="icon-button" title="新建书签" @click="store.select(null)">+</button>
      </div>

      <label class="field">
        <span>搜索</span>
        <input v-model="store.query" placeholder="标题、网址、标签、关键词" />
      </label>

      <label class="field">
        <span>分类</span>
        <select v-model="store.folder">
          <option value="">全部分类</option>
          <option v-for="folder in store.folders" :key="folder" :value="folder">{{ folder }}</option>
        </select>
      </label>

      <label class="field">
        <span>标签</span>
        <select v-model="store.tag">
          <option value="">全部标签</option>
          <option v-for="tag in store.tags" :key="tag" :value="tag">{{ tag }}</option>
        </select>
      </label>

      <section class="panel">
        <h2>导入</h2>
        <div class="row">
          <select v-model="importSource">
            <option value="html">HTML 文件</option>
            <option value="chrome">Chrome</option>
            <option value="edge">Edge</option>
            <option value="firefox">Firefox</option>
          </select>
          <button @click="store.importFrom(importSource, importPath)">导入</button>
        </div>
        <input v-model="importPath" placeholder="HTML/浏览器数据文件路径；自动扫描可留空" />
      </section>

      <section class="panel">
        <h2>Web 与导出</h2>
        <button class="wide" @click="store.toggleServer()">
          {{ store.server.running ? '关闭本地 Web' : '启动本地 Web' }}
        </button>
        <a v-if="store.server.running" class="server-link" :href="store.server.url" target="_blank" rel="noreferrer">
          {{ store.server.url }}
        </a>
        <div class="row">
          <select v-model="templateId">
            <option v-for="template in store.templates" :key="template.id" :value="template.id">{{ template.name }}</option>
          </select>
          <button @click="store.exportTo(exportPath, templateId)">导出</button>
        </div>
        <input v-model="exportPath" placeholder="导出 HTML 路径，例如 D:\\bookmarks.html" />
      </section>

      <section class="panel">
        <h2>备份恢复</h2>
        <input v-model="backupPath" placeholder="备份路径；留空使用默认目录" />
        <button class="wide" @click="store.createBackup(backupPath)">创建备份</button>
        <input v-model="restorePath" placeholder="要恢复的 .db 备份路径" />
        <button class="wide danger" @click="store.restoreBackup(restorePath)">恢复备份</button>
      </section>
    </aside>

    <main class="workspace">
      <header class="topbar">
        <div>
          <strong>{{ store.total }}</strong>
          <span> 个书签</span>
          <span v-if="store.loading" class="muted">正在加载</span>
        </div>
        <div class="actions">
          <button @click="store.analyzeAll()">全部 AI 分析</button>
          <button @click="store.refresh()">刷新</button>
        </div>
      </header>

      <section class="content">
        <div class="list">
          <button
            v-for="item in store.items"
            :key="item.id"
            class="bookmark-row"
            :class="{ active: selectedId === item.id }"
            @click="choose(item)"
          >
            <span class="title">{{ item.title }}</span>
            <span class="url">{{ item.url }}</span>
            <span class="meta">
              <span>{{ item.folder }}</span>
              <span v-for="tag in item.tags" :key="tag" class="tag">{{ tag }}</span>
            </span>
          </button>
          <div v-if="!store.items.length" class="empty">没有匹配的书签</div>
        </div>

        <form class="editor" @submit.prevent="store.save()">
          <div class="editor-head">
            <h2>{{ store.selected ? '编辑书签' : '新增书签' }}</h2>
            <div class="actions">
              <button type="button" :disabled="!store.selected" @click="store.analyzeSelected()">AI 分析</button>
              <button type="button" class="danger" :disabled="!store.selected" @click="store.removeSelected()">删除</button>
              <button type="submit">保存</button>
            </div>
          </div>

          <label class="field">
            <span>标题</span>
            <input v-model="store.draft.title" required />
          </label>
          <label class="field">
            <span>网址</span>
            <input v-model="store.draft.url" required />
          </label>
          <label class="field">
            <span>描述</span>
            <textarea v-model="store.draft.description" rows="5"></textarea>
          </label>
          <div class="two-col">
            <label class="field">
              <span>分类</span>
              <input v-model="store.draft.folder" />
            </label>
            <label class="field">
              <span>标签</span>
              <input :value="joinList(store.draft.tags)" @input="store.draft.tags = splitList(($event.target as HTMLInputElement).value)" />
            </label>
          </div>
          <div class="two-col">
            <label class="field">
              <span>关键词</span>
              <input :value="joinList(store.draft.keywords)" @input="store.draft.keywords = splitList(($event.target as HTMLInputElement).value)" />
            </label>
            <label class="field">
              <span>别名</span>
              <input :value="joinList(store.draft.aliases)" @input="store.draft.aliases = splitList(($event.target as HTMLInputElement).value)" />
            </label>
          </div>
          <p class="status">{{ store.status }}</p>
        </form>
      </section>
    </main>
  </div>
</template>
