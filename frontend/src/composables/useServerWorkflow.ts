// useServerWorkflow 将本地 Web 服务状态转换成页面可展示的端口、链接和操作。
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
    // 开关动作通过 store 执行，保证其他页面依赖的 server 状态同步刷新。
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
