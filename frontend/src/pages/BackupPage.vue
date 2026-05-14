<script setup lang="ts">
import { useBackupWorkflow } from '../composables/useBackupWorkflow'

const workflow = useBackupWorkflow()
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
        <input v-model="workflow.backupPath.value" placeholder="备份路径；留空使用默认目录" />
      </label>
      <button class="primary-action" type="button" :disabled="workflow.busy.value" @click="workflow.createBackup">立即备份</button>
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
        <input v-model="workflow.restorePath.value" placeholder="要恢复的 .db 备份路径" />
      </label>
      <label class="field">
        <span>二次确认</span>
        <input v-model="workflow.restoreConfirm.value" placeholder="输入 RESTORE 后才能恢复" />
      </label>
      <button class="danger" type="button" :disabled="workflow.busy.value || !workflow.canRestore.value" @click="workflow.restoreBackup">
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
          v-for="(entry, index) in workflow.backupHistory.value"
          :key="entry.time"
          class="compact-row static"
        >
          <div class="backup-entry-row">
            <span class="backup-name">{{ entry.filename }}</span>
            <span class="backup-size">{{ entry.size }}</span>
            <small class="backup-time">{{ entry.time }}</small>
          </div>
          <div class="backup-actions">
            <button type="button" @click="workflow.selectBackupForRestore(entry.filename)">恢复</button>
            <button type="button" class="danger" @click="workflow.removeBackup(index)">删除</button>
          </div>
        </div>
        <div v-if="!workflow.backupHistory.value.length" class="empty">暂无备份记录，创建备份后将会显示在此处</div>
      </div>
    </section>
  </section>
</template>
