import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    allowedHosts: true, // ✅ 允许所有 Host（仅建议开发时）
    proxy: {
      '/api': {
        // 要和后端的端口对应上
        target: 'http://localhost:8000',
        changeOrigin: true,
        secure: false
      }
    }
  }
})
