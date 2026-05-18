// 文件说明：frontend/src/navigation.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import type { AppIconKey } from './icons'

export type RoutePath =
  | '/'
  | '/library'
  | '/import'
  | '/export'
  | '/webserver'
  | '/ai'
  | '/backup'
  | '/settings'

export interface NavItem {
  path: RoutePath
  label: string
  title: string
  shortLabel: string
  icon: AppIconKey
  group: 'primary' | 'workflow' | 'system'
}

export const routes: NavItem[] = [
  { path: '/', label: '仪表盘', title: '仪表盘', shortLabel: 'DB', icon: 'dashboard', group: 'primary' },
  { path: '/library', label: '书签库', title: '书签库', shortLabel: 'LB', icon: 'library', group: 'primary' },
  { path: '/import', label: '导入中心', title: '导入中心', shortLabel: 'IN', icon: 'import', group: 'workflow' },
  { path: '/export', label: '导出与主题', title: '导出与主题', shortLabel: 'EX', icon: 'export', group: 'workflow' },
  { path: '/webserver', label: '本地 Web', title: '本地 Web 服务', shortLabel: 'WB', icon: 'web', group: 'workflow' },
  { path: '/ai', label: 'AI 助手', title: 'AI 助手', shortLabel: 'AI', icon: 'ai', group: 'workflow' },
  { path: '/backup', label: '备份与恢复', title: '备份与恢复', shortLabel: 'BK', icon: 'backup', group: 'system' },
  { path: '/settings', label: '设置', title: '设置', shortLabel: 'ST', icon: 'settings', group: 'system' }
]

export const routePaths = routes.map((route) => route.path)
