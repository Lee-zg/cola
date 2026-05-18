// 文件说明：frontend/src/icons.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import type { Component } from 'vue'
import {
  AddCircleOutline,
  AddOutline,
  AnalyticsOutline,
  ArchiveOutline,
  BookmarksOutline,
  BrowsersOutline,
  ChevronDownOutline,
  ChevronForwardOutline,
  CloseOutline,
  CloudDownloadOutline,
  CloudUploadOutline,
  CopyOutline,
  DocumentTextOutline,
  EllipsisHorizontalOutline,
  FolderOpenOutline,
  GlobeOutline,
  HomeOutline,
  LibraryOutline,
  LinkOutline,
  MoonOutline,
  OpenOutline,
  PencilOutline,
  PinOutline,
  Pin,
  PlayForwardOutline,
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
  TrashOutline,
  TrashBinOutline,
  ArrowDownOutline,
  ArrowUpOutline
} from '@vicons/ionicons5'

export type AppIconKey =
  | 'add'
  | 'addPlain'
  | 'ai'
  | 'backup'
  | 'bookmarks'
  | 'chevronDown'
  | 'chevronRight'
  | 'close'
  | 'copy'
  | 'dashboard'
  | 'document'
  | 'dots'
  | 'export'
  | 'folder'
  | 'import'
  | 'library'
  | 'link'
  | 'maximise'
  | 'minimise'
  | 'moon'
  | 'open'
  | 'pencil'
  | 'pin'
  | 'pinFilled'
  | 'overflow'
  | 'refresh'
  | 'save'
  | 'search'
  | 'server'
  | 'settings'
  | 'stats'
  | 'stop'
  | 'sun'
  | 'trash'
  | 'trashBin'
  | 'up'
  | 'down'
  | 'web'

export const appIcons: Record<AppIconKey, Component> = {
  add: AddCircleOutline,
  addPlain: AddOutline,
  ai: SparklesOutline,
  backup: ArchiveOutline,
  bookmarks: BookmarksOutline,
  chevronDown: ChevronDownOutline,
  chevronRight: ChevronForwardOutline,
  close: CloseOutline,
  copy: CopyOutline,
  dashboard: HomeOutline,
  document: DocumentTextOutline,
  dots: EllipsisHorizontalOutline,
  down: ArrowDownOutline,
  export: CloudUploadOutline,
  folder: FolderOpenOutline,
  import: CloudDownloadOutline,
  library: LibraryOutline,
  link: LinkOutline,
  maximise: SquareOutline,
  minimise: RemoveOutline,
  moon: MoonOutline,
  open: OpenOutline,
  pencil: PencilOutline,
  pin: PinOutline,
  pinFilled: Pin,
  overflow: PlayForwardOutline,
  refresh: RefreshOutline,
  save: SaveOutline,
  search: SearchOutline,
  server: ServerOutline,
  settings: SettingsOutline,
  stats: StatsChartOutline,
  stop: StopOutline,
  sun: SunnyOutline,
  trash: TrashOutline,
  trashBin: TrashBinOutline,
  up: ArrowUpOutline,
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
