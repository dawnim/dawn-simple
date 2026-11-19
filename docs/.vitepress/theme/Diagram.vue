<script setup lang="ts">
import { ref, useId } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const english: Record<string, string> = {
  全屏查看图表: 'View diagram fullscreen', 图表全屏预览: 'Fullscreen diagram', 退出全屏: 'Exit fullscreen'
}
const t = (key: string) => lang.value.startsWith('zh') ? key : english[key] || key
const diagram = ref<HTMLDivElement>()
const modal = ref<HTMLDialogElement>()
const modalId = useId()
const expanded = ref(false)

function expand() {
  const box = diagram.value?.querySelector<SVGSVGElement>('.docs-mermaid svg[viewBox]')?.viewBox.baseVal
  if (!box || !modal.value) return

  modal.value.style.setProperty('--diagram-ratio', String(box.width / box.height))
  expanded.value = true
  modal.value.showModal()
}
</script>

<template>
  <div ref="diagram" class="diagram">
    <button
      class="diagram-expand"
      type="button"
      :title="t('全屏查看图表')"
      :aria-label="t('全屏查看图表')"
      :aria-controls="modalId"
      @click="expand"
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.65"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M9 3H5a2 2 0 0 0-2 2v4M15 3h4a2 2 0 0 1 2 2v4M9 21H5a2 2 0 0 1-2-2v-4M15 21h4a2 2 0 0 0 2-2v-4" />
      </svg>
    </button>
    <dialog :id="modalId" ref="modal" class="diagram-modal" :aria-label="t('图表全屏预览')" @close="expanded = false">
      <button type="button" class="diagram-close" :title="t('退出全屏')" :aria-label="t('退出全屏')" @click="modal?.close()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.65" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <rect x="9.5" y="9.5" width="5" height="5" rx="1" />
          <path d="M4 8V5a1 1 0 0 1 1-1h3M20 8V5a1 1 0 0 0-1-1h-3M4 16v3a1 1 0 0 0 1 1h3M20 16v3a1 1 0 0 1-1 1h-3" />
        </svg>
      </button>
    </dialog>
    <Teleport defer :to="`#${modalId}`" :disabled="!expanded">
      <slot />
    </Teleport>
  </div>
</template>
