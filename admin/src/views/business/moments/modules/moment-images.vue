<template>
  <div class="moment-images">
    <div v-for="id in ids" :key="id" class="moment-image">
      <img v-if="thumbs[id]" :src="thumbs[id]" :alt="`图片 ${id}`" @click="preview(thumbs[id])" />
      <span v-else class="moment-image-loading">加载中…</span>
      <span class="moment-image-id">#{{ id }}</span>
      <button type="button" class="moment-image-remove" title="移除" @click="remove(id)">×</button>
    </div>
    <label class="moment-image-add" title="上传图片">
      <span class="moment-image-plus">＋</span>
      <span>添加图片</span>
      <input type="file" accept="image/jpeg,image/png,image/gif,image/webp" multiple @change="add" />
    </label>
    <Teleport to="body">
      <div v-if="viewerVisible" class="moment-images-viewer">
        <!-- 不写死 z-index：ElImageViewer 会自己 nextZIndex()，始终盖在 ElDialog 之上。 -->
        <ElImageViewer
          :url-list="viewerUrls"
          :initial-index="viewerIndex"
          @close="viewerVisible = false"
        />
      </div>
    </Teleport>
  </div>
</template>
<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue'
  import { ElImageViewer } from 'element-plus'
  import { request } from '@dawnim/admin'
  import { attachmentURL, parseImages } from './attachments'

  const props = defineProps<{ modelValue?: string; recordId?: number }>()
  const emit = defineEmits<{ 'update:modelValue': [string] }>()
  const thumbs = reactive<Record<number, string>>({})
  const ids = computed(() => parseImages(props.modelValue))
  const viewerVisible = ref(false)
  const viewerIndex = ref(0)
  const viewerUrls = ref<string[]>([])

  function write(next: number[]) {
    emit('update:modelValue', JSON.stringify(next))
  }
  async function load() {
    for (const id of ids.value) {
      if (thumbs[id]) continue
      try {
        thumbs[id] = await attachmentURL(id)
      } catch {
        delete thumbs[id]
      }
    }
  }
  watch(ids, load, { immediate: true })

  async function add(event: Event) {
    const input = event.target as HTMLInputElement
    const next = ids.value.slice()
    for (const file of Array.from(input.files || [])) {
      const form = new FormData()
      form.append('file', file)
      const row = await request.post<{ id: number }>({ url: '/attachments', data: form, timeout: 120000 })
      next.push(row.id)
    }
    input.value = ''
    if (next.length !== ids.value.length) write(next)
  }
  function remove(id: number) {
    write(ids.value.filter((item) => item !== id))
  }
  function preview(url: string) {
    // 浮层查看，且把整组图片交给查看器，可左右切换，不再开新窗口。
    const list = ids.value.map((id) => thumbs[id]).filter((item): item is string => !!item)
    viewerUrls.value = list
    viewerIndex.value = Math.max(0, list.indexOf(url))
    viewerVisible.value = true
  }
</script>
<style scoped>
.moment-images { display: flex; flex-wrap: wrap; gap: 10px; width: 100%; }
.moment-image, .moment-image-add {
  position: relative; display: grid; place-items: center; width: 92px; height: 92px;
  border: 1px solid var(--el-border-color, #dcdfe6); border-radius: 8px; overflow: hidden;
}
.moment-image img { width: 100%; height: 100%; object-fit: cover; cursor: zoom-in; }
.moment-image-loading, .moment-image-id { font-size: 11px; color: var(--el-text-color-secondary, #909399); }
.moment-image-id { position: absolute; bottom: 2px; left: 4px; padding: 0 4px; border-radius: 4px; background: #0008; color: #fff; }
.moment-image-remove {
  position: absolute; top: 2px; right: 2px; width: 20px; height: 20px; border: 0; border-radius: 50%;
  background: #0009; color: #fff; font-size: 15px; line-height: 18px; cursor: pointer;
}
.moment-image-add { gap: 2px; border-style: dashed; color: var(--el-text-color-secondary, #909399); font-size: 11px; cursor: pointer; }
.moment-image-plus { font-size: 20px; line-height: 1; }
.moment-image-add input { display: none; }
</style>
<style>
/* 同 moment-editor：避免内置浏览器里 viewer-fade 卡在 enter-from 导致整层透明。 */
.moment-images-viewer .el-image-viewer__wrapper { opacity: 1 !important; }
</style>
