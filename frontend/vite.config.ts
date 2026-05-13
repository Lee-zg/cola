import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    strictPort: false
  },
  test: {
    environment: 'jsdom',
    coverage: {
      exclude: ['dist/**', 'node_modules/**', 'wailsjs/**', 'coverage/**', '*.config.*', 'vite.config.ts']
    }
  }
})
