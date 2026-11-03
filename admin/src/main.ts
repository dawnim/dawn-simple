import { createAdmin } from '@dawnim/admin'
import '@dawnim/admin/style.css'
import HeaderTools from './components/HeaderTools.vue'

createAdmin({
  apiBase: import.meta.env.VITE_ADMIN_PREFIX || '/admin',
  header: { before: HeaderTools },
  messages: {
    en: {
      '实时示例': 'Realtime example',
      '通过共享WebSocket读取业务缓存，可在缓存管理中查看或清理': 'Read a business cache over the shared WebSocket; inspect or clear it in cache management',
      '读取问候': 'Read greeting'
    }
  },
  pages: {
    '/examples/realtime': () => import('./views/examples/realtime.vue'),
    '/dashboard/statistics': () => import('./views/dashboard/index.vue'),
    '/dashboard/work': () => import('./views/dashboard/work.vue')
  }
}).mount('#app')
