// 文件说明：frontend/src/composables/useServerWorkflow.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { computed } from 'vue'
import { getServerAccessLabel, getServerPort } from '../helpers/server'
import { useBookmarkStore } from '../stores/bookmarks'

export const useServerWorkflow = () => {
  const store = useBookmarkStore()
  const server = computed(() => store.server)
  const serverPort = computed(() => getServerPort(store.server))
  const accessLabel = computed(() => getServerAccessLabel(store.server))
  const serverDescription = computed(() =>
    store.server.running ? store.server.url : '启动后可在浏览器或局域网设备访问导出的书签页面。'
  )

  const toggleServer = async () => {
    await store.toggleServer()
  }

  return {
    server,
    serverPort,
    accessLabel,
    serverDescription,
    toggleServer
  }
}
