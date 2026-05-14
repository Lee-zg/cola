// server helper 把后端返回的本地服务状态转换成稳定的展示文案。
import type { ServerStatus } from '../types'

export const getServerPort = (server: ServerStatus): string => {
  // addr 由 Go net.Listener 生成，当前格式为 host:port；为空表示服务未启动。
  if (!server.addr) return '默认'
  const parts = server.addr.split(':')
  return parts[parts.length - 1] || '默认'
}

export const getServerAccessLabel = (server: ServerStatus): string =>
  server.running ? server.url : '启动服务后显示链接'
