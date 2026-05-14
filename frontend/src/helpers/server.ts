// 文件说明：frontend/src/helpers/server.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import type { ServerStatus } from '../types'

export const getServerPort = (server: ServerStatus): string => {
  if (!server.addr) return '默认'
  const parts = server.addr.split(':')
  return parts[parts.length - 1] || '默认'
}

export const getServerAccessLabel = (server: ServerStatus): string =>
  server.running ? server.url : '启动服务后显示链接'
