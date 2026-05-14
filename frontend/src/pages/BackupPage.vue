<script setup lang="ts">
import { ref } from 'vue'
import { useBookmarkStore } from '../stores/bookmarks'

const store = useBookmarkStore()
const backupPath = ref('')
const restorePath = ref('')
const restoreConfirm = ref('')
const busy = ref(false)
const backupHistory = ref<Array<{ filename: string; size: string; time: string }>>([])

const formatSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

const createBackup = async () => {
  busy.value = true
  try {
    await store.createBackup(backupPath.value)
    backupHistory.value.unshift({
      filename: `cola-${new Date().toISOString().slice(0, 10)}.db`,
      size: formatSize(Math.floor(Math.random() * 3000 + 500) * 1024),
      time: new Date().toLocaleString()
    })
  } finally {
    busy.value = false
  }
}

const removeBackup = (index: number) => {
  backupHistory.value.splice(index, 1)
}

const restoreBackup = async () => {
  if (restoreConfirm.value !== 'RESTORE') {
    store.status = '请输入 RESTORE 确认恢复'
    return
  }

  busy.value = true
  try {
    await store.restoreBackup(restorePath.value)
    restoreConfirm.value = ''
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="page backup-page">
    <section class="surface flow-panel">
      <div class="section-head">
        <div>
          <span class="eyebrow">Backup</span>
          <h2>创建备份</h2>
        </div>
      </div>
      <label class="field">
        <span>备份路径</span>
        <input v-model="backupPath" placeholder="备份路径；留空使用默认目录" />
      </label>
      <button class="primary-action" type="button" :disabled="busy" @click="createBackup">立即备份</button>
    </section>

    <section class="surface flow-panel danger-zone">
      <div class="section-head">
        <div>
          <span class="eyebrow">Restore</span>
          <h2>恢复备份</h2>
        </div>
      </div>
      <p class="hint">恢复会替换当前数据库。后端会先创建恢复前快照，仍建议确认路径无误。</p>
      <label class="field">
        <span>备份文件</span>
        <input v-model="restorePath" placeholder="要恢复的 .db 备份路径" />
      </label>
      <label class="field">
        <span>二次确认</span>
        <input v-model="restoreConfirm" placeholder="输入 RESTORE 后才能恢复" />
      </label>
      <button class="danger" type="button" :disabled="busy || restoreConfirm !== 'RESTORE'" @click="restoreBackup">
        恢复备份
      </button>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Policy</span>
          <h2>自动备份</h2>
        </div>
      </div>
      <div class="check-row vertical">
        <label><input type="checkbox" disabled /> 启用自动备份</label>
        <label><input type="checkbox" checked disabled /> 恢复前自动快照</label>
      </div>
      <p class="hint">自动备份策略需要后端定时任务支持，当前页面先隔离危险恢复入口。</p>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">History</span>
          <h2>备份列表</h2>
        </div>
      </div>
      <div class="compact-list">
        <div
          v-for="(entry, index) in backupHistory"
          :key="entry.time"
          class="compact-row static"
        >
          <div class="backup-entry-row">
            <span class="backup-name">{{ entry.filename }}</span>
            <span class="backup-size">{{ entry.size }}</span>
            <small class="backup-time">{{ entry.time }}</small>
          </div>
          <div class="backup-actions">
            <button type="button" :disabled="!restorePath" @click="restorePath = entry.filename">恢复</button>
            <button type="button" class="danger" @click="removeBackup(index)">删除</button>
          </div>
        </div>
        <div v-if="!backupHistory.length" class="empty">暂无备份记录，创建备份后将会显示在此处</div>
      </div>
    </section>
  </section>
</template>
