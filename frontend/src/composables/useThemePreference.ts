// 文件说明：frontend/src/composables/useThemePreference.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { computed, ref, watchEffect } from 'vue'
import { darkDesignTokens, designTokens, type ThemeMode } from '../theme/tokens'

const STORAGE_KEY = 'cola-theme-mode'

const readStoredTheme = (): ThemeMode => {
  if (typeof window === 'undefined') return 'light'
  const stored = window.localStorage.getItem(STORAGE_KEY)
  return stored === 'dark' ? 'dark' : 'light'
}

const mode = ref<ThemeMode>(readStoredTheme())

export const useThemePreference = () => {
  const isDark = computed(() => mode.value === 'dark')
  const activeTokens = computed(() => (isDark.value ? darkDesignTokens : designTokens))

  const setThemeMode = (value: ThemeMode) => {
    mode.value = value
  }

  const toggleTheme = () => {
    mode.value = mode.value === 'dark' ? 'light' : 'dark'
  }

  watchEffect(() => {
    if (typeof document === 'undefined') return
    const tokens = activeTokens.value
    document.documentElement.dataset.theme = mode.value
    document.documentElement.style.colorScheme = mode.value

    Object.entries(tokens).forEach(([key, value]) => {
      const cssName = key.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)
      document.documentElement.style.setProperty(`--${cssName}`, value)
    })

    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, mode.value)
    }
  })

  return {
    mode,
    isDark,
    activeTokens,
    setThemeMode,
    toggleTheme
  }
}
