<template>
  <div class="admin-moment-editor-wrap">
    <div ref="host" class="admin-moment-editor"></div>
    <div v-if="bodyIds.length" class="admin-body-images">
      <span class="admin-body-images-label">{{ lang() === 'en' ? 'Images in body' : '正文图片' }} · {{ bodyIds.length }}</span>
      <button
        v-for="id in bodyIds"
        :key="id"
        type="button"
        class="admin-body-image"
        :title="`${lang() === 'en' ? 'Image' : '图片'} #${id}`"
        @click="openImage(id)"
      >
        <img v-if="thumbs[id]" :src="thumbs[id]" :alt="`图片 ${id}`" />
        <span v-else class="admin-body-image-fallback">#{{ id }}</span>
      </button>
    </div>
    <!-- 图片查看用浮层而不是 window.open：支持正文多图左右切换，且不会离开编辑页。 -->
    <Teleport to="body">
      <div v-if="viewerVisible" class="moment-editor-viewer">
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
  import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElImageViewer } from 'element-plus'
  import {
    createMarkdownEditor,
    enhanceMarkdown,
    renderMarkdown,
    request,
    type EditorConfig,
    type MarkdownEditor
  } from '@dawnim/admin'
  import { attachmentPath, attachmentURL, parseImages } from './attachments'

  const props = defineProps<{
    modelValue: string
    format?: string
    hidden?: string[]
    recordId?: number
    images?: string
  }>()
  const emit = defineEmits<{ 'update:modelValue': [string]; 'update:images': [string] }>()
  const { locale } = useI18n()
  const host = ref<HTMLElement | null>(null)
  let editor: MarkdownEditor | undefined
  const thumbs = reactive<Record<number, string>>({})
  const viewerVisible = ref(false)
  const viewerIndex = ref(0)
  const viewerUrls = ref<string[]>([])

  const lang = () => (locale.value.startsWith('en') ? 'en' : 'zh') as 'zh' | 'en'
  const config = (): EditorConfig => ({ hidden: props.hidden || [] })

  // 正文可能写 attachment://、upload://，也可能已经是官网/后台的绝对图片地址。
  function bodyImageIds(text?: string): number[] {
    const found = new Set<number>()
    const source = String(text || '')
    const pattern = /(?:attachment|upload):\/\/(\d+)|(?:\/api\/moments\/\d+\/images|\/admin\/attachments)\/(\d+)/g
    for (const match of source.matchAll(pattern)) {
      const id = Number(match[1] || match[2])
      if (Number.isInteger(id) && id > 0) found.add(id)
    }
    return [...found]
  }
  const bodyIds = computed(() => bodyImageIds(props.modelValue))

  // 业务方言：官网注入完整的 DawnMomentMarkdown，后台退回框架通用渲染。
  function render(text: string): string {
    const source = String(text || '').replace(/(?:attachment|upload):\/\/(\d+)/g, (_, id: string) =>
      attachmentPath(Number(id))
    )
    const engine = (window as any).DawnMomentMarkdown
    return engine?.render ? engine.render(source) : renderMarkdown(source)
  }
  function enhance(root: HTMLElement) {
    const engine = (window as any).DawnMomentMarkdown
    if (engine?.enhanceMarkdown) engine.enhanceMarkdown(root)
    else enhanceMarkdown(root)
    // 后台附件走鉴权接口，<img> 直接请求会 401，替换成带 token 拉取的 blob URL。
    root.querySelectorAll<HTMLImageElement>('img[src*="/attachments/"], img[src*="/moments/"][src*="/images/"]').forEach((image) => {
      const id = Number(
        image.src.match(/\/attachments\/(\d+)\/content/)?.[1] || image.src.match(/\/images\/(\d+)/)?.[1]
      )
      if (id) attachmentURL(id).then((url) => (image.src = url)).catch(() => undefined)
    })
  }
  async function loadThumbs(ids: number[]) {
    for (const id of ids) {
      if (thumbs[id]) continue
      try { thumbs[id] = await attachmentURL(id) } catch { /* 附件不存在或无权限时保留编号占位 */ }
    }
  }
  /** 预览区 <img> 可能已被 enhance 换成 blob URL，需要从 thumbs 反查附件 id。 */
  function idFromSrc(src: string): number | null {
    const match = src.match(/\/attachments\/(\d+)\/content/) || src.match(/\/images\/(\d+)/)
    if (match) return Number(match[1])
    const hit = Object.entries(thumbs).find(([, url]) => url === src)
    return hit ? Number(hit[0]) : null
  }
  /** 后台附件需要 Bearer，先把正文里引用的图都转成 blob URL 再交给查看器。 */
  async function openImage(target: number | string) {
    const id = typeof target === 'number' ? target : idFromSrc(target)
    const ids = [...bodyIds.value]
    if (id && !ids.includes(id)) ids.unshift(id)
    await loadThumbs(ids)
    const loaded = ids.filter((item) => thumbs[item])
    const urls = loaded.map((item) => thumbs[item])
    if (!urls.length && typeof target === 'number') return
    viewerUrls.value = urls.length ? urls : [String(target)]
    viewerIndex.value = id ? Math.max(0, loaded.indexOf(id)) : 0
    viewerVisible.value = true
  }
  function saveDraft() {
    if (!editor) return
    const key = 'admin.moment.drafts'
    const drafts = JSON.parse(localStorage.getItem(key) || '{}')
    drafts[props.recordId || 'new'] = editor.getValue()
    localStorage.setItem(key, JSON.stringify(drafts))
  }
  const imageMIMEs = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  const displayURL = async (id: number) => attachmentURL(id).catch(() => attachmentPath(id))
  // 选完立即上传：先按 SHA-256 秒传，命中复用附件 id，未命中才 POST /attachments。
  const imageUploader = {
    fingerprint: async (file: File) => ({
      sha256: [...new Uint8Array(await crypto.subtle.digest('SHA-256', await file.arrayBuffer()))]
        .map(value => value.toString(16).padStart(2, '0')).join('')
    }),
    check: async (files: File[], prints: { sha256?: string }[]) => {
      const hits = []
      for (let i = 0; i < files.length; i++) {
        const row = await request.post<{ exists: boolean; id: number }>({
          url: '/attachment-checks-sha',
          data: { name: files[i].name, size: files[i].size, sha256: prints[i].sha256 }
        })
        hits.push(row?.exists ? { id: row.id, name: files[i].name, url: await displayURL(row.id) } : null)
      }
      return hits
    },
    upload: async (files: File[]) => {
      const rows = []
      for (const file of files) {
        const form = new FormData()
        form.append('file', file)
        const row = await request.post<{ id: number }>({ url: '/attachments', data: form, timeout: 120000 })
        rows.push({ id: row.id, name: file.name, url: await displayURL(row.id) })
      }
      return rows
    }
  }
  const imageLibrary = {
    list: async (keyword?: string) => {
      const payload: any = await request.get({ url: '/attachments', params: { mine: 1, mime: 'image', name: keyword || undefined, page: 1, size: 60 } })
      const rows: any[] = Array.isArray(payload) ? payload : payload?.data || []
      return Promise.all(rows.filter(row => imageMIMEs.includes(row.mime)).map(async row => ({
        id: row.id, name: row.name, url: await displayURL(Number(row.id))
      })))
    }
  }
  const emitImages = (ids: (string | number)[]) => {
    const all = [...new Set([...bodyImageIds(props.modelValue), ...ids.map(Number)])]
    emit('update:images', JSON.stringify(all))
  }
  function applyOptions() {
    editor?.setOptions({
      locale: lang(),
      format: props.format === 'post' ? 'post' : 'moment',
      maxLength: props.format === 'post' ? 10000 : 1024,
      imageTarget: props.format === 'post' ? 'choose' : 'gallery',
      config
    })
  }

  // 正文引用过的附件必须同时登记到 images：官网 /api/moments/{id}/images/{image_id} 按这个数组取图，
  // 否则后台预览能看到、发布后官网却是坏图。
  watch(bodyIds, (ids) => {
    if (!ids.length) return
    const current = parseImages(props.images)
    const missing = ids.filter((id) => !current.includes(id))
    if (missing.length) emit('update:images', JSON.stringify([...current, ...missing]))
    void loadThumbs(ids)
  }, { immediate: true })

  onMounted(() => {
    if (!host.value) return
    editor = createMarkdownEditor(host.value, {
      locale: lang(),
      format: props.format === 'post' ? 'post' : 'moment',
      maxLength: props.format === 'post' ? 10000 : 1024,
      value: props.modelValue || '',
      alwaysFull: true,
      imageTarget: props.format === 'post' ? 'choose' : 'gallery',
      maxImages: 9,
      imageLibrary,
      imageUploader,
      onImagesChange: emitImages,
      customTools: [{
        key: 'my-images',
        title: ['我的图片', 'My images'],
        icon: 'image',
        where: 'attachments',
        onClick: ({ editor: instance, anchor }) => instance.openImageLibrary(anchor)
      }],
      config,
      render,
      enhance,
      onChange: (text) => emit('update:modelValue', text),
      onSaveDraft: saveDraft,
      mentions: async (query: string) => {
        const response = await fetch(`/api/mentions/search?q=${encodeURIComponent(query)}`)
        const payload = await response.json()
        const rows = payload?.data || payload || []
        return rows.map((row: any) => ({
          label: `${row.nickname || row.unique_id} · ${row.unique_id}`,
          value: String(row.unique_id)
        }))
      },
      previewImage: (src: string) => {
        void openImage(src)
      }
    })
  })
  watch(() => props.modelValue, (value) => {
    if (editor && value !== editor.getValue()) editor.setValue(value || '')
  })
  watch(() => [props.format, props.hidden, locale.value], applyOptions)
  onBeforeUnmount(() => editor?.destroy())
</script>
<style scoped>
.admin-moment-editor-wrap { display: flex; flex-direction: column; gap: 8px; width: 100%; }
.admin-moment-editor { display: flex; flex: 1; min-height: 360px; width: 100%; }
.admin-body-images { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.admin-body-images-label { color: var(--el-text-color-secondary, #909399); font-size: 12px; }
.admin-body-image {
  display: grid; place-items: center; width: 56px; height: 56px; padding: 0;
  border: 1px solid var(--el-border-color, #dcdfe6); border-radius: 6px;
  background: var(--el-fill-color-light, #f5f7fa); cursor: zoom-in; overflow: hidden;
}
.admin-body-image img { width: 100%; height: 100%; object-fit: cover; }
.admin-body-image-fallback { color: var(--el-text-color-secondary, #909399); font-size: 11px; }
</style>
<style>
/* 内置浏览器里 Element Plus 的 viewer-fade 过渡会停在 enter-from(opacity:0)，
   遮罩与大图其实已渲染却看不见；这里只对本模块的查看器去掉淡入依赖。 */
.moment-editor-viewer .el-image-viewer__wrapper { opacity: 1 !important; }
</style>
