// 文件说明：frontend/src/theme/tokens.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { darkTheme, type GlobalThemeOverrides } from 'naive-ui'

export type ThemeMode = 'light' | 'dark'

export const designTokens = {
  bg: '#F8F5F0',
  panel: '#FFFFFF',
  panelStrong: '#4A423C',
  text: '#4A423C',
  muted: '#8A817C',
  disabled: '#D3CCC6',
  line: '#E6DDD4',
  lineStrong: '#D3CCC6',
  accent: '#C28455',
  accentStrong: '#A66D43',
  accentSoft: '#F3E4D7',
  info: '#798C9F',
  infoSoft: '#E5EBF0',
  danger: '#A95C58',
  dangerSoft: '#F1DEDA',
  shadow: '0 18px 48px rgb(194 132 85 / 10%)'
} as const

export const darkDesignTokens = {
  bg: '#211D1A',
  panel: '#2B2622',
  panelStrong: '#F8F5F0',
  text: '#F4ECE5',
  muted: '#C8BBB1',
  disabled: '#70645C',
  line: '#463B34',
  lineStrong: '#5E5047',
  accent: '#C28455',
  accentStrong: '#D49A6E',
  accentSoft: '#3B2C22',
  info: '#9AABBB',
  infoSoft: '#28333B',
  danger: '#C77A74',
  dangerSoft: '#412A29',
  shadow: '0 18px 48px rgb(0 0 0 / 22%)'
} as const

export const lightThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: designTokens.accent,
    primaryColorHover: '#D49A6E',
    primaryColorPressed: designTokens.accentStrong,
    primaryColorSuppl: designTokens.accent,
    infoColor: designTokens.info,
    infoColorHover: '#8FA2B4',
    infoColorPressed: '#637789',
    errorColor: designTokens.danger,
    errorColorHover: '#BD706A',
    errorColorPressed: '#8D4946',
    warningColor: designTokens.accent,
    successColor: '#77936E',
    bodyColor: designTokens.bg,
    baseColor: designTokens.panel,
    cardColor: designTokens.panel,
    modalColor: designTokens.panel,
    popoverColor: designTokens.panel,
    textColorBase: designTokens.text,
    textColor1: designTokens.text,
    textColor2: designTokens.muted,
    textColor3: designTokens.muted,
    borderColor: designTokens.line,
    dividerColor: designTokens.line,
    borderRadius: '8px',
    borderRadiusSmall: '7px',
    fontFamily: "'Aptos', 'Segoe UI Variable', 'Microsoft YaHei UI', sans-serif"
  },
  Button: {
    borderRadiusMedium: '7px',
    heightMedium: '36px',
    textColor: designTokens.text
  },
  Card: {
    borderRadius: '8px',
    color: designTokens.panel,
    titleTextColor: designTokens.text,
    textColor: designTokens.text
  },
  Input: {
    borderRadius: '7px'
  },
  Select: {
    peers: {
      InternalSelection: {
        borderRadius: '7px'
      }
    }
  },
  Tag: {
    borderRadius: '6px'
  }
}

export const darkThemeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: darkDesignTokens.accent,
    primaryColorHover: darkDesignTokens.accentStrong,
    primaryColorPressed: '#B77649',
    primaryColorSuppl: darkDesignTokens.accent,
    infoColor: darkDesignTokens.info,
    infoColorHover: '#B0BFCC',
    infoColorPressed: '#7F91A2',
    errorColor: darkDesignTokens.danger,
    errorColorHover: '#D08D87',
    errorColorPressed: '#A95C58',
    warningColor: darkDesignTokens.accent,
    successColor: '#8DAA82',
    bodyColor: darkDesignTokens.bg,
    baseColor: darkDesignTokens.panel,
    cardColor: darkDesignTokens.panel,
    modalColor: darkDesignTokens.panel,
    popoverColor: darkDesignTokens.panel,
    textColorBase: darkDesignTokens.text,
    textColor1: darkDesignTokens.text,
    textColor2: darkDesignTokens.muted,
    textColor3: darkDesignTokens.muted,
    borderColor: darkDesignTokens.line,
    dividerColor: darkDesignTokens.line,
    borderRadius: '8px',
    borderRadiusSmall: '7px',
    fontFamily: "'Aptos', 'Segoe UI Variable', 'Microsoft YaHei UI', sans-serif"
  },
  Button: {
    borderRadiusMedium: '7px',
    heightMedium: '36px'
  },
  Card: {
    borderRadius: '8px',
    color: darkDesignTokens.panel,
    titleTextColor: darkDesignTokens.text,
    textColor: darkDesignTokens.text
  },
  Input: {
    borderRadius: '7px'
  },
  Tag: {
    borderRadius: '6px'
  }
}

export const naiveDarkTheme = darkTheme
