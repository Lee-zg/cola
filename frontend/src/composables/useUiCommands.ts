// 文件说明：frontend/src/composables/useUiCommands.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { ref } from 'vue'

const createBookmarkRequestId = ref(0)

export const useUiCommands = () => {
  const requestCreateBookmark = () => {
    createBookmarkRequestId.value += 1
  }

  return {
    createBookmarkRequestId,
    requestCreateBookmark
  }
}
