// 文件说明：frontend/src/composables/useDesktopRuntime.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { onBeforeUnmount } from 'vue'
import {
  BrowserOpenURL,
  ClipboardSetText,
  OnFileDrop,
  OnFileDropOff,
  Quit,
  WindowMinimise,
  WindowToggleMaximise
} from '../../wailsjs/runtime/runtime'

declare global {
  interface Window {
    runtime?: unknown
  }
}

export interface DroppedFilePayload {
  x: number
  y: number
  paths: string[]
}

const isWailsRuntimeAvailable = () => typeof window !== 'undefined' && Boolean(window.runtime)

export const useDesktopRuntime = () => {
  const isAvailable = isWailsRuntimeAvailable()

  const minimise = () => {
    if (isWailsRuntimeAvailable()) WindowMinimise()
  }

  const toggleMaximise = () => {
    if (isWailsRuntimeAvailable()) WindowToggleMaximise()
  }

  const closeWindow = () => {
    if (isWailsRuntimeAvailable()) {
      Quit()
      return
    }
    window.close()
  }

  const openExternal = (url: string) => {
    if (!url) return
    if (isWailsRuntimeAvailable()) {
      BrowserOpenURL(url)
      return
    }
    window.open(url, '_blank', 'noopener,noreferrer')
  }

  const copyText = async (text: string) => {
    if (!text) return false
    if (isWailsRuntimeAvailable()) {
      return ClipboardSetText(text)
    }
    await navigator.clipboard?.writeText(text)
    return true
  }

  const onFileDrop = (callback: (payload: DroppedFilePayload) => void) => {
    if (!isWailsRuntimeAvailable()) return () => undefined
    OnFileDrop((x: number, y: number, paths: string[]) => callback({ x, y, paths }), true)
    return () => OnFileDropOff()
  }

  onBeforeUnmount(() => {
    if (isWailsRuntimeAvailable()) OnFileDropOff()
  })

  return {
    isAvailable,
    minimise,
    toggleMaximise,
    closeWindow,
    openExternal,
    copyText,
    onFileDrop
  }
}
