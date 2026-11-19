<script setup lang="ts">
import { ref } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const english: Record<string, string> = {
  全屏查看图表: 'View diagram fullscreen', 图表全屏预览: 'Fullscreen diagram', 关闭图表: 'Close diagram', 关闭: 'Close'
}
const t = (key: string) => lang.value.startsWith('zh') ? key : english[key] || key
const modal = ref<HTMLDialogElement>()
const expanded = ref(false)
</script>

<template>
  <div class="diagram">
    <button
      class="diagram-expand"
      type="button"
      :title="t('全屏查看图表')"
      :aria-label="t('全屏查看图表')"
      @click="(expanded = true, modal?.showModal())"
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.6"
        aria-hidden="true"
      >
        <path d="M8 3H3v5m13-5h5v5M3 16v5h5m13-5v5h-5" />
      </svg>
    </button>
    <Teleport :to="modal" :disabled="!expanded">
      <slot />
    </Teleport>
    <dialog ref="modal" class="diagram-modal" :aria-label="t('图表全屏预览')" @close="expanded = false">
      <button type="button" class="diagram-close" :aria-label="t('关闭图表')" @click="modal?.close()">
        {{ t('关闭') }} <kbd>Esc</kbd>
      </button>
    </dialog>
  </div>
</template>
