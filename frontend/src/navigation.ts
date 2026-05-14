import type { Component } from 'vue'

export type RoutePath =
  | '/'
  | '/library'
  | '/taxonomy'
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
  icon: string
  group: 'primary' | 'workflow' | 'system'
  component?: Component
}

export const routes: NavItem[] = [
  { path: '/', label: '仪表盘', title: '仪表盘', shortLabel: 'DB', icon: '🏠', group: 'primary' },
  { path: '/library', label: '书签库', title: '书签库', shortLabel: 'LB', icon: '📚', group: 'primary' },
  { path: '/taxonomy', label: '分类与标签', title: '分类与标签', shortLabel: 'TG', icon: '🏷️', group: 'primary' },
  { path: '/import', label: '导入中心', title: '导入中心', shortLabel: 'IN', icon: '⬇️', group: 'workflow' },
  { path: '/export', label: '导出与主题', title: '导出与主题', shortLabel: 'EX', icon: '⬆️', group: 'workflow' },
  { path: '/webserver', label: '本地 Web', title: '本地 Web 服务', shortLabel: 'WB', icon: '🌐', group: 'workflow' },
  { path: '/ai', label: 'AI 助手', title: 'AI 助手', shortLabel: 'AI', icon: '🤖', group: 'workflow' },
  { path: '/backup', label: '备份与恢复', title: '备份与恢复', shortLabel: 'BK', icon: '💾', group: 'system' },
  { path: '/settings', label: '设置', title: '设置', shortLabel: 'ST', icon: '⚙️', group: 'system' }
]

export const routePaths = routes.map((route) => route.path)
