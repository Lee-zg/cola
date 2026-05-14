<script setup lang="ts">
import { useDashboard } from '../composables/useDashboard'

const emit = defineEmits<{
  navigate: [path: '/library' | '/import' | '/webserver' | '/ai']
}>()

const dashboard = useDashboard()
</script>

<template>
  <section class="page dashboard-page">
    <div class="metric-grid">
      <article class="metric-card">
        <span>总书签</span>
        <strong>{{ dashboard.stats.value.total }}</strong>
      </article>
      <article class="metric-card">
        <span>分类数</span>
        <strong>{{ dashboard.stats.value.folderCount }}</strong>
      </article>
      <article class="metric-card">
        <span>标签数</span>
        <strong>{{ dashboard.stats.value.tagCount }}</strong>
      </article>
      <article class="metric-card">
        <span>待 AI 分析</span>
        <strong>{{ dashboard.stats.value.pendingAiCount }}</strong>
      </article>
    </div>

    <div class="dashboard-grid">
      <section class="surface">
        <div class="section-head">
          <div>
            <span class="eyebrow">Recent</span>
            <h2>最近更新</h2>
          </div>
          <button type="button" @click="emit('navigate', '/library')">打开书签库</button>
        </div>
        <div class="compact-list">
          <button
            v-for="item in dashboard.recentItems.value"
            :key="item.id"
            class="compact-row"
            type="button"
            @click="emit('navigate', '/library')"
          >
            <span>{{ item.title }}</span>
            <small>{{ item.folder }} · {{ item.domain || item.url }}</small>
          </button>
          <div v-if="!dashboard.recentItems.value.length" class="empty">暂无书签数据</div>
        </div>
      </section>

      <section class="surface">
        <div class="section-head">
          <div>
            <span class="eyebrow">Tags</span>
            <h2>Top 标签</h2>
          </div>
        </div>
        <div class="tag-cloud">
          <button
            v-for="tag in dashboard.topTags.value"
            :key="tag"
            class="tag large"
            type="button"
            @click="dashboard.selectTag(tag)"
          >
            {{ tag }}
          </button>
          <div v-if="!dashboard.topTags.value.length" class="empty inline-empty">暂无标签</div>
        </div>
      </section>

      <section class="surface action-card">
        <div>
          <span class="eyebrow">Local Web</span>
          <h2>本地 Web 服务</h2>
          <p>{{ dashboard.webServerSummary.value }}</p>
        </div>
        <button type="button" class="primary-action" @click="emit('navigate', '/webserver')">管理服务</button>
      </section>

      <section class="surface action-card">
        <div>
          <span class="eyebrow">AI Queue</span>
          <h2>AI 分析队列</h2>
          <p>当前列表中还有 {{ dashboard.stats.value.pendingAiCount }} 条书签未分析。</p>
        </div>
        <button type="button" @click="emit('navigate', '/ai')">查看队列</button>
      </section>

      <section class="surface onboarding-card">
        <div>
          <span class="eyebrow">Start</span>
          <h2>首次使用</h2>
          <p>从浏览器或 HTML 文件导入书签，再回到书签库整理分类与标签。</p>
        </div>
        <button type="button" @click="emit('navigate', '/import')">去导入</button>
      </section>
    </div>
  </section>
</template>
