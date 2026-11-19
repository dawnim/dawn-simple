import DefaultTheme from 'vitepress/theme'
import { useData, useRouter, withBase } from 'vitepress'
import { onMounted, onUnmounted, watch } from 'vue'
import type { Theme } from 'vitepress'
import Diagram from './Diagram.vue'
import './style.css'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('Diagram', Diagram)
  },
  setup() {
    const { lang, page } = useData()
    const router = useRouter()
    const key = 'dawn.site.locale'
    const change = (locale: string) => {
      const path = page.value.relativePath.replace(/^en\//, '').replace(/\.md$/, '.html').replace(/index\.html$/, '')
      if (locale !== (lang.value.startsWith('zh') ? 'zh' : 'en'))
        void router.go(withBase(`${locale === 'en' ? '/en/' : '/'}${path}`))
    }
    const sync = (event: StorageEvent) => {
      if (event.key === key && event.newValue) change(event.newValue)
    }
    onMounted(() => {
      const locale = localStorage.getItem(key) || (lang.value === 'en' ? 'en' : navigator.language.startsWith('zh') ? 'zh' : 'en')
      localStorage.setItem(key, locale)
      change(locale)
      addEventListener('storage', sync)
    })
    watch(() => page.value.relativePath, () => {
      localStorage.setItem(key, lang.value.startsWith('zh') ? 'zh' : 'en')
    })
    onUnmounted(() => removeEventListener('storage', sync))
  }
} satisfies Theme
