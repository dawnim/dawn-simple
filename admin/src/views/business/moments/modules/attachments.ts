import {
  request,
  type EditorImage,
  type EditorImageLibrary,
  type EditorImageUploader
} from '@dawnim/admin'

const prefix = (import.meta.env.VITE_ADMIN_PREFIX as string) || '/admin'
const images = new Map<number, Promise<EditorImage>>()
let accessToken: string
const imageMIMEs = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
export const imageTypes = imageMIMEs.join(',')

export function parseImages(value?: string): number[] {
  return JSON.parse(value || '[]') || []
}

export function loadImage(id: number): Promise<EditorImage> {
  const token = JSON.parse(localStorage.getItem('admin:user') || '{}').accessToken
  if (token !== accessToken) {
    for (const image of images.values())
      void image.then(
        (row) => URL.revokeObjectURL(row.url),
        () => {}
      )
    images.clear()
    accessToken = token
  }
  const cached = images.get(id)
  if (cached) return cached
  const pending = fetch(`${prefix}/attachments/${id}/content`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {}
  })
    .then(async (response) => {
      if (!response.ok) throw new Error(String(response.status))
      const disposition = response.headers.get('Content-Disposition') || ''
      const encoded = disposition.match(/filename\*=utf-8''([^;]+)/i)?.[1]
      const quoted = disposition.match(/filename="((?:\\.|[^"\\])*)"/)?.[1]
      const name = encoded
        ? decodeURIComponent(encoded)
        : (quoted?.replace(/\\(.)/g, '$1') ?? disposition.match(/filename=([^;]+)/)?.[1])
      return { id, name, url: URL.createObjectURL(await response.blob()) }
    })
    .catch((error) => {
      if (images.get(id) === pending) images.delete(id)
      throw error
    })
  images.set(id, pending)
  return pending
}

export const imageUploader: EditorImageUploader = {
  fingerprint: async (file) => ({
    sha256: [...new Uint8Array(await crypto.subtle.digest('SHA-256', await file.arrayBuffer()))]
      .map((value) => value.toString(16).padStart(2, '0'))
      .join('')
  }),
  check: (files, prints) =>
    Promise.all(
      files.map(async (file, index) => {
        const row = await request.post<{ exists: boolean; id: number }>({
          url: '/attachment-checks-sha',
          data: { name: file.name, size: file.size, sha256: prints[index].sha256 }
        })
        return row?.exists ? loadImage(row.id) : null
      })
    ),
  upload: async (files) => {
    const rows = []
    for (const file of files) {
      const form = new FormData()
      form.append('file', file)
      const row = await request.post<{ id: number }>({
        url: '/attachments',
        data: form,
        timeout: 120000
      })
      rows.push(await loadImage(row.id))
    }
    return rows
  },
  toMarkdown: (image) => `[attachment://${image.id}]`
}

export const imageLibrary: EditorImageLibrary = {
  list: async (keyword) => {
    const { data: rows } = await request.get<{ data: { id: number; mime: string }[] }>({
      url: '/attachments',
      params: { mine: 1, mime: 'image/', name: keyword || undefined, page_size: 60 }
    })
    return Promise.all(
      rows.filter((row) => imageMIMEs.includes(row.mime)).map((row) => loadImage(row.id))
    )
  }
}
