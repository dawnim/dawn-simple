<script setup lang="ts">
import { ref } from 'vue'

const modal = ref<HTMLDialogElement>()
const expanded = ref(false)
</script>

<template>
  <div class="diagram">
    <button
      class="diagram-expand"
      type="button"
      title="全屏查看图表"
      aria-label="全屏查看图表"
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
    <dialog ref="modal" class="diagram-modal" aria-label="图表全屏预览" @close="expanded = false">
      <button type="button" class="diagram-close" aria-label="关闭图表" @click="modal?.close()">
        关闭 <kbd>Esc</kbd>
      </button>
    </dialog>
  </div>
</template>
