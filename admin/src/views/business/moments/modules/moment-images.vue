<template>
  <div class="moment-images">
    <div v-for="id in ids" :key="id" class="moment-image">
      <button v-if="images[id]" type="button" class="moment-image-preview" @click="preview = id">
        <img :src="images[id].url" :alt="images[id].name || ''" />
      </button>
      <span v-else>#{{ id }}</span>
      <button
        type="button"
        class="moment-image-remove"
        :aria-label="t('移除图片')"
        @click="write(ids.filter((item) => item !== id))"
        >×</button
      >
    </div>
    <ElButton plain @click="open">{{ t('添加图片') }}</ElButton>
  </div>
  <ElDialog v-model="visible" :title="t('图库图片')" width="min(480px, 95vw)" append-to-body>
    <div class="moment-image-tools">
      <ElInput v-model="keyword" :placeholder="t('搜索图片')" @input="loadLibrary" />
      <ElButton :loading="uploading" @click="fileInput?.click()">{{ t('上传图片') }}</ElButton>
      <input ref="fileInput" type="file" :accept="imageTypes" multiple hidden @change="upload" />
    </div>
    <div v-loading="loading" class="moment-images">
      <button
        v-for="image in library"
        :key="image.id"
        type="button"
        class="moment-image-preview"
        :title="image.name"
        @click="add([Number(image.id)])"
      >
        <img :src="image.url" :alt="image.name || ''" />
      </button>
      <span v-if="!loading && !library.length">{{ t('暂无图片') }}</span>
    </div>
  </ElDialog>
  <ElImageViewer
    v-if="preview"
    :url-list="viewerIDs.map((id) => images[id].url)"
    :initial-index="viewerIDs.indexOf(preview)"
    teleported
    @close="preview = 0"
  />
</template>

<script setup lang="ts">
  import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElButton, ElDialog, ElImageViewer, ElInput, ElMessage, vLoading } from 'element-plus'
  import type { EditorImage } from '@dawnim/admin'
  import { imageLibrary, imageTypes, imageUploader, loadImage, parseImages } from './attachments'

  const props = defineProps<{ modelValue?: string }>()
  const emit = defineEmits<{ 'update:modelValue': [string] }>()
  const { t } = useI18n()
  const images = reactive<Record<number, EditorImage>>({})
  const ids = computed(() => parseImages(props.modelValue))
  const viewerIDs = computed(() => ids.value.filter((id) => images[id]))
  const preview = ref(0)
  const visible = ref(false)
  const keyword = ref('')
  const loading = ref(false)
  const uploading = ref(false)
  const fileInput = ref<HTMLInputElement>()
  const library = ref<EditorImage[]>([])
  let searchVersion = 0
  let mounted = true

  const write = (next: number[]) => emit('update:modelValue', JSON.stringify(next))
  watch(
    ids,
    (next) => {
      for (const id of next)
        void loadImage(id)
          .then((image) => {
            images[id] = image
          })
          .catch(() => {})
    },
    { immediate: true }
  )

  function add(selected: number[]) {
    const next = [...new Set([...ids.value, ...selected])]
    if (next.length > 9) {
      ElMessage.warning(t('最多选择9张图片'))
      return
    }
    write(next)
    visible.value = false
  }
  async function loadLibrary() {
    const version = ++searchVersion
    loading.value = true
    try {
      const rows = await imageLibrary.list(keyword.value)
      if (version === searchVersion) library.value = rows
    } catch {
      if (version === searchVersion) library.value = []
    } finally {
      if (version === searchVersion) loading.value = false
    }
  }
  function open() {
    visible.value = true
    keyword.value = ''
    void loadLibrary()
  }
  async function upload(event: Event) {
    const input = event.target as HTMLInputElement
    const files = Array.from(input.files || [])
    input.value = ''
    if (!files.length || uploading.value) return
    if (files.length + ids.value.length > 9) {
      ElMessage.warning(t('最多选择9张图片'))
      return
    }
    uploading.value = true
    try {
      const prints = await Promise.all(files.map(imageUploader.fingerprint!))
      const hits = await imageUploader.check!(files, prints)
      const missing = files.filter((_, index) => !hits[index])
      const uploaded = await imageUploader.upload(
        missing,
        prints.filter((_, index) => !hits[index])
      )
      if (mounted && visible.value) add(hits.map((hit) => Number((hit || uploaded.shift()!).id)))
    } catch {
      if (mounted) ElMessage.error(t('上传失败'))
    } finally {
      uploading.value = false
    }
  }
  onBeforeUnmount(() => {
    mounted = false
  })
</script>

<style scoped>
  .moment-images {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    width: 100%;
  }
  .moment-image {
    position: relative;
    display: grid;
    place-items: center;
    width: 60px;
    height: 60px;
    border: 1px solid var(--el-border-color);
    border-radius: 8px;
  }
  .moment-image-preview {
    display: block;
    padding: 0;
    border-radius: 8px;
    overflow: hidden;
    cursor: zoom-in;
  }
  .moment-image-preview img {
    display: block;
    width: 58px;
    height: 58px;
    object-fit: cover;
  }
  .moment-image-remove {
    position: absolute;
    top: 2px;
    right: 2px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #000a;
    color: white;
    cursor: pointer;
  }
  .moment-image-tools {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }
</style>
