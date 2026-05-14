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
