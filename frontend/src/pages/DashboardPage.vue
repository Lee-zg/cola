<!-- 文件说明：frontend/src/pages/DashboardPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NCard, NEmpty, NIcon, NList, NListItem, NProgress, NSpace, NStatistic, NTag, NThing } from 'naive-ui'
import ColaScrollbar from '../components/ColaScrollbar.vue'
import { useDashboard } from '../composables/useDashboard'
import { appIcons } from '../icons'

const emit = defineEmits<{
  navigate: [path: '/library' | '/import' | '/webserver' | '/ai']
}>()

const dashboard = useDashboard()

const pendingPercentage = computed(() => {
  const total = dashboard.stats.value.total
  if (!total) return 0
  return Math.round((dashboard.stats.value.pendingAiCount / total) * 100)
})
</script>

<template>
  <section class="page dashboard-page">
    <div class="dashboard-hero panel-surface">
      <div>
        <span class="eyebrow">OVERVIEW</span>
        <h2>书签工作台</h2>
        <p>整理、导入、发布和复审都从这里进入；高频浏览与编辑留给书签库。</p>
      </div>
      <NSpace :size="10">
        <NButton type="primary" @click="emit('navigate', '/library')">
          <template #icon>
            <NIcon :component="appIcons.library" />
          </template>
          打开书签库
        </NButton>
        <NButton secondary @click="emit('navigate', '/import')">
          <template #icon>
            <NIcon :component="appIcons.import" />
          </template>
          导入
        </NButton>
      </NSpace>
    </div>

    <ColaScrollbar class="dashboard-scroll" aria-label="仪表盘卡片滚动区">
      <div class="dashboard-grid">
        <NCard class="metric-card" :bordered="false">
          <NStatistic label="总书签" :value="dashboard.stats.value.total" />
        </NCard>
        <NCard class="metric-card" :bordered="false">
          <NStatistic label="分类数" :value="dashboard.stats.value.folderCount" />
        </NCard>
        <NCard class="metric-card" :bordered="false">
          <NStatistic label="标签数" :value="dashboard.stats.value.tagCount" />
        </NCard>
        <NCard class="metric-card" :bordered="false">
          <NStatistic label="待 AI 分析" :value="dashboard.stats.value.pendingAiCount" />
        </NCard>

        <NCard class="dashboard-panel recent-panel" :bordered="false">
          <template #header>
            <span class="panel-title">最近更新</span>
          </template>
          <NList v-if="dashboard.recentItems.value.length" hoverable clickable>
            <NListItem v-for="item in dashboard.recentItems.value" :key="item.id" @click="emit('navigate', '/library')">
              <NThing :title="item.title" :description="`${item.folder || 'Unsorted'} · ${item.domain || item.url}`" />
            </NListItem>
          </NList>
          <NEmpty v-else description="暂无书签数据" />
        </NCard>

        <NCard class="dashboard-panel" :bordered="false">
          <template #header>
            <span class="panel-title">Top 标签</span>
          </template>
          <NSpace v-if="dashboard.topTags.value.length" wrap>
            <NTag v-for="tag in dashboard.topTags.value" :key="tag" round type="info" checkable @click="dashboard.selectTag(tag)">
              {{ tag }}
            </NTag>
          </NSpace>
          <NEmpty v-else description="暂无标签" />
        </NCard>

        <NCard class="dashboard-panel action-panel" :bordered="false">
          <template #header>
            <span class="panel-title">本地 Web 服务</span>
          </template>
          <p>{{ dashboard.webServerSummary.value }}</p>
          <NButton secondary @click="emit('navigate', '/webserver')">管理服务</NButton>
        </NCard>

        <NCard class="dashboard-panel action-panel" :bordered="false">
          <template #header>
            <span class="panel-title">AI 分析队列</span>
          </template>
          <NProgress type="line" :percentage="pendingPercentage" :indicator-placement="'inside'" />
          <p>当前列表中还有 {{ dashboard.stats.value.pendingAiCount }} 条书签未分析。</p>
          <NButton secondary @click="emit('navigate', '/ai')">查看队列</NButton>
        </NCard>
      </div>
    </ColaScrollbar>
  </section>
</template>
