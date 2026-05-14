<!-- 文件说明：frontend/src/pages/BackupPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { ref } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NCheckbox,
  NEmpty,
  NIcon,
  NInput,
  NList,
  NListItem,
  NModal,
  NSpace,
  NTag
} from 'naive-ui'
import { useBackupWorkflow } from '../composables/useBackupWorkflow'
import { RESTORE_CONFIRMATION } from '../helpers/backup'
import { appIcons } from '../icons'

const workflow = useBackupWorkflow()
const restoreModalOpen = ref(false)

const openRestoreModal = (path?: string) => {
  if (path) workflow.selectBackupForRestore(path)
  restoreModalOpen.value = true
}

const confirmRestore = async () => {
  await workflow.restoreBackup()
  if (workflow.canRestore.value) {
    restoreModalOpen.value = false
  }
}
</script>

<template>
  <section class="page backup-page">
    <NCard class="workflow-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">BACKUP</span>
          <h2>备份与恢复</h2>
        </div>
      </template>

      <div class="backup-grid">
        <section class="backup-panel">
          <h3>创建备份</h3>
          <NInput v-model:value="workflow.backupPath.value" placeholder="备份路径；留空使用默认目录" />
          <NButton type="primary" :loading="workflow.busy.value" @click="workflow.createBackup">
            <template #icon>
              <NIcon :component="appIcons.backup" />
            </template>
            立即备份
          </NButton>
        </section>

        <section class="backup-panel danger-panel">
          <h3>恢复备份</h3>
          <NAlert type="error" :show-icon="false">恢复会替换当前数据库；后端会先创建恢复前快照。</NAlert>
          <NInput v-model:value="workflow.restorePath.value" placeholder="要恢复的 .db 备份路径" />
          <NButton type="error" secondary @click="openRestoreModal()">打开恢复确认</NButton>
        </section>

        <section class="backup-panel">
          <h3>自动备份</h3>
          <NSpace vertical>
            <NCheckbox disabled>启用自动备份</NCheckbox>
            <NCheckbox checked disabled>恢复前自动快照</NCheckbox>
          </NSpace>
          <p class="muted-copy">自动备份策略需要后端定时任务支持，本轮先隔离危险恢复入口。</p>
        </section>
      </div>
    </NCard>

    <NCard class="workflow-side-card" :bordered="false">
      <template #header>备份列表</template>
      <NList v-if="workflow.backupHistory.value.length">
        <NListItem v-for="(entry, index) in workflow.backupHistory.value" :key="entry.time">
          <NSpace vertical :size="2">
            <strong>{{ entry.filename }}</strong>
            <span class="muted-copy">{{ entry.size }} · {{ entry.time }}</span>
          </NSpace>
          <template #suffix>
            <NSpace>
              <NButton size="small" secondary @click="openRestoreModal(entry.filename)">恢复</NButton>
              <NButton size="small" type="error" secondary @click="workflow.removeBackup(index)">删除</NButton>
            </NSpace>
          </template>
        </NListItem>
      </NList>
      <NEmpty v-else description="暂无备份记录，创建备份后将会显示在此处" />
    </NCard>

    <NModal v-model:show="restoreModalOpen" preset="card" title="确认恢复备份" class="restore-modal">
      <NSpace vertical :size="14">
        <NAlert type="error">恢复会替换当前数据库。请输入 {{ RESTORE_CONFIRMATION }} 后继续。</NAlert>
        <NInput v-model:value="workflow.restorePath.value" placeholder="要恢复的 .db 备份路径" />
        <NInput v-model:value="workflow.restoreConfirm.value" placeholder="输入 RESTORE 后才能恢复" />
        <NSpace justify="space-between" align="center">
          <NTag type="error" round>危险操作</NTag>
          <NSpace>
            <NButton @click="restoreModalOpen = false">取消</NButton>
            <NButton type="error" :disabled="!workflow.canRestore.value" :loading="workflow.busy.value" @click="confirmRestore">
              确认恢复
            </NButton>
          </NSpace>
        </NSpace>
      </NSpace>
    </NModal>
  </section>
</template>
