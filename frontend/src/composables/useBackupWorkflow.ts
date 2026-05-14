// useBackupWorkflow 管理备份页的确认口令、历史展示和忙碌状态，文件复制由后端 backup 包处理。
import { computed, ref } from 'vue'
import { canRestoreBackup, createBackupHistoryEntry, getRestoreValidationMessage } from '../helpers/backup'
import type { BackupHistoryEntry } from '../helpers/workflow'
import { useBookmarkStore } from '../stores/bookmarks'

export const useBackupWorkflow = () => {
  const store = useBookmarkStore()
  const backupPath = ref('')
  const restorePath = ref('')
  const restoreConfirm = ref('')
  const busy = ref(false)
  const backupHistory = ref<BackupHistoryEntry[]>([])
  const canRestore = computed(() => canRestoreBackup(restoreConfirm.value))

  const createBackup = async () => {
    // 备份历史是页面级提示信息，真实备份文件位置以后端返回路径为准。
    busy.value = true
    try {
      const result = await store.createBackup(backupPath.value)
      backupHistory.value.unshift(createBackupHistoryEntry(result.path))
    } finally {
      busy.value = false
    }
  }

  const removeBackup = (index: number) => {
    backupHistory.value.splice(index, 1)
  }

  const selectBackupForRestore = (filename: string) => {
    restorePath.value = filename
  }

  const restoreBackup = async () => {
    // 恢复是破坏性操作，前端先做确认口令校验，后端再创建覆盖前快照。
    const validationMessage = getRestoreValidationMessage(restoreConfirm.value)
    if (validationMessage) {
      store.setStatus(validationMessage)
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

  return {
    backupPath,
    restorePath,
    restoreConfirm,
    busy,
    backupHistory,
    canRestore,
    createBackup,
    removeBackup,
    restoreBackup,
    selectBackupForRestore
  }
}
