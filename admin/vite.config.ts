import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const httpTarget = process.env.DAWN_HTTP_TARGET || 'http://127.0.0.1:5123'
const wsTarget = process.env.DAWN_WS_TARGET || 'ws://127.0.0.1:8851'
const adminPrefix = process.env.DAWN_ADMIN_PREFIX || '/admin'
const wsPath = process.env.DAWN_WS_PATH || '/'

export default defineConfig({
  plugins: [vue()],
  define: { 'import.meta.env.VITE_ADMIN_PREFIX': JSON.stringify(adminPrefix) },
  publicDir: '../public',
  optimizeDeps: { exclude: ['@dawnim/admin'] },
  resolve: { dedupe: ['vue', 'vue-router', 'pinia', 'element-plus', 'vue-i18n'] },
  server: {
    host: '127.0.0.1',
    strictPort: true,
    proxy: {
      [adminPrefix]: httpTarget,
      '/api': httpTarget,
      '/ws': { target: wsTarget, ws: true, rewrite: () => wsPath }
    }
  }
})
