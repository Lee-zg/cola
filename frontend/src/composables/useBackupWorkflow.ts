// 文件说明：frontend/src/composables/useBackupWorkflow.ts，对应当前模块的数据结构、状态逻辑或工具函数。
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
