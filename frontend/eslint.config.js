import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import vueTs from '@vue/eslint-config-typescript'
import globals from 'globals'

export default [
  {
    ignores: ['dist/**', 'node_modules/**', 'wailsjs/**', 'coverage/**']
  },
  js.configs.recommended,
  ...pluginVue.configs['flat/recommended'],
  ...vueTs(),
  {
    files: ['**/*.{ts,vue}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.es2022
      }
    },
    rules: {
      'vue/multi-word-component-names': 'off',
      'vue/no-mutating-props': 'error',
      'vue/html-self-closing': 'off',
      'vue/max-attributes-per-line': 'off',
      'vue/singleline-html-element-content-newline': 'off',
      '@typescript-eslint/no-explicit-any': 'warn'
    }
  },
  {
    files: ['**/*.test.ts'],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2022
      }
    }
  },
  {
    files: ['scripts/**/*.mjs'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.node,
        ...globals.es2022
      }
    }
  }
]
