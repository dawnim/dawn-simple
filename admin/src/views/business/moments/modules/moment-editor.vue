<template>
  <div ref="host" class="admin-moment-editor"></div>
  <ElImageViewer v-if="preview" :url-list="[preview]" teleported @close="preview = ''" />
</template>

<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElImageViewer } from 'element-plus'
  import { createMarkdownEditor, enhanceMarkdown, type MarkdownEditor } from '@dawnim/admin'
  import { renderMarkdown } from '../../../../../../tools/moment_markdown.mjs'
  import { imageLibrary, imageUploader, loadImage } from './attachments'

  const props = defineProps<{
    modelValue: string
    format?: string
    hidden?: string[]
    recordId?: number
  }>()
  const emit = defineEmits<{ 'update:modelValue': [string] }>()
  const { locale } = useI18n()
  const host = ref<HTMLElement>()
  const preview = ref('')
  let editor: MarkdownEditor | undefined
  const images = new Map<string, { url: string; name?: string }>()
  let emojis: Record<string, string> = {}

  const options = () => ({
    locale: locale.value.startsWith('en') ? ('en' as const) : ('zh' as const),
    maxLength: props.format === 'post' ? 10000 : 1024,
    pasteImages: props.format === 'post',
    config: () => ({
      hidden: [...(props.hidden || []), ...(props.format === 'post' ? [] : ['image', 'media'])]
    })
  })
  const render = (text: string) => renderMarkdown(text, emojis, (id) => images.get(id) || {})

  onMounted(() => {
    editor = createMarkdownEditor(host.value!, {
      ...options(),
      value: props.modelValue || '',
      imageLibrary,
      imageUploader,
      render,
      enhance(root) {
        enhanceMarkdown(root)
        root
          .querySelectorAll<HTMLImageElement>('img[data-moment-attachment]')
          .forEach(async (img) => {
            const id = img.dataset.momentAttachment!
            if (images.has(id)) return
            try {
              const image = await loadImage(Number(id))
              images.set(id, image)
              if (!img.isConnected) return
              img.alt = image.name || ''
              img.src = image.url
            } catch {
              img.hidden = true
            }
          })
      },
      mediaMarkup: (url, type) => `MEDIA::${type}::${url}`,
      previewImage: (src) => {
        preview.value = src
      },
      onChange: (text) => emit('update:modelValue', text),
      onSaveDraft() {
        const key = 'admin.moment.drafts'
        const drafts = JSON.parse(localStorage.getItem(key) || '{}')
        drafts[props.recordId || 'new'] = editor!.getValue()
        localStorage.setItem(key, JSON.stringify(drafts))
      },
      async mentions(query) {
        const response = await fetch(`/api/mentions/search?q=${encodeURIComponent(query)}`)
        const { data } = await response.json()
        return (data || []).map((row: { nickname: string; unique_id: string }) => ({
          label: `${row.nickname || row.unique_id} · ${row.unique_id}`,
          value: row.unique_id
        }))
      }
    })
    void fetch('/api/moment-config')
      .then((response) => response.json())
      .then(({ data }) => {
        emojis = data.emojis
        editor?.setOptions({ render })
      })
      .catch(() => undefined)
  })
  watch(
    () => props.modelValue,
    (value) => {
      if (editor && value !== editor.getValue()) editor.setValue(value || '')
    }
  )
  watch(
    () => [props.format, props.hidden, locale.value],
    () => editor?.setOptions(options())
  )
  onBeforeUnmount(() => {
    editor?.destroy()
    editor = undefined
  })
</script>

<style scoped>
  .admin-moment-editor {
    display: flex;
    width: 100%;
    min-width: 0;
    height: 380px;
  }
</style>
