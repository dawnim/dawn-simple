import { createAdmin } from '@dawnim/admin'
import '@dawnim/admin/style.css'
import HeaderTools from './components/HeaderTools.vue'

createAdmin({
  apiBase: import.meta.env.VITE_ADMIN_PREFIX || '/admin',
  header: { before: HeaderTools },
  pages: {
    '/dashboard/console': () => import('./views/dashboard/index.vue')
  }
}).mount('#app')
