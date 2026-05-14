// 文件说明：frontend/src/icons.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import type { Component } from 'vue'
import {
  AddCircleOutline,
  AnalyticsOutline,
  ArchiveOutline,
  BookmarksOutline,
  BrowsersOutline,
  CloseOutline,
  CloudDownloadOutline,
  CloudUploadOutline,
  CopyOutline,
  DocumentTextOutline,
  FolderOpenOutline,
  GlobeOutline,
  HomeOutline,
  LibraryOutline,
  LinkOutline,
  MoonOutline,
  OpenOutline,
  PlayForwardOutline,
  PricetagsOutline,
  RefreshOutline,
  RemoveOutline,
  RocketOutline,
  SaveOutline,
  SearchOutline,
  ServerOutline,
  SettingsOutline,
  SparklesOutline,
  SquareOutline,
  StatsChartOutline,
  StopOutline,
  SunnyOutline,
  TrashOutline
} from '@vicons/ionicons5'

export type AppIconKey =
  | 'add'
  | 'ai'
  | 'backup'
  | 'bookmarks'
  | 'close'
  | 'copy'
  | 'dashboard'
  | 'document'
  | 'export'
  | 'folder'
  | 'import'
  | 'library'
  | 'link'
  | 'maximise'
  | 'minimise'
  | 'moon'
  | 'open'
  | 'overflow'
  | 'refresh'
  | 'save'
  | 'search'
  | 'server'
  | 'settings'
  | 'stats'
  | 'stop'
  | 'sun'
  | 'tags'
  | 'trash'
  | 'web'

export const appIcons: Record<AppIconKey, Component> = {
  add: AddCircleOutline,
  ai: SparklesOutline,
  backup: ArchiveOutline,
  bookmarks: BookmarksOutline,
  close: CloseOutline,
  copy: CopyOutline,
  dashboard: HomeOutline,
  document: DocumentTextOutline,
  export: CloudUploadOutline,
  folder: FolderOpenOutline,
  import: CloudDownloadOutline,
  library: LibraryOutline,
  link: LinkOutline,
  maximise: SquareOutline,
  minimise: RemoveOutline,
  moon: MoonOutline,
  open: OpenOutline,
  overflow: PlayForwardOutline,
  refresh: RefreshOutline,
  save: SaveOutline,
  search: SearchOutline,
  server: ServerOutline,
  settings: SettingsOutline,
  stats: StatsChartOutline,
  stop: StopOutline,
  sun: SunnyOutline,
  tags: PricetagsOutline,
  trash: TrashOutline,
  web: GlobeOutline
}

export const sourceIcons: Record<string, Component> = {
  chrome: BrowsersOutline,
  edge: BrowsersOutline,
  firefox: BrowsersOutline,
  html: BookmarksOutline
}

export const workflowIcons: Record<string, Component> = {
  classic: BookmarksOutline,
  minimal: DocumentTextOutline,
  modern: AnalyticsOutline,
  default: RocketOutline
}
