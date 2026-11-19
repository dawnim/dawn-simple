const prefix = (import.meta.env.VITE_ADMIN_PREFIX as string) || '/admin'
const cache = new Map<number, string>()

export function parseImages(value?: string): number[] {
  try {
    return (JSON.parse(value || '[]') as unknown[])
      .map(Number)
      .filter((id) => Number.isInteger(id) && id > 0)
  } catch {
    return []
  }
}

export function attachmentPath(id: number): string {
  return `${prefix}/attachments/${id}/content`
}

/** 后台附件需要 Bearer token，转成 blob URL 后 <img> 才能直接显示。 */
export function attachmentURL(id: number): Promise<string> {
  const hit = cache.get(id)
  if (hit) return Promise.resolve(hit)
  const token = JSON.parse(localStorage.getItem('admin:user') || '{}').accessToken
  return fetch(attachmentPath(id), {
    headers: token ? { Authorization: `Bearer ${token}` } : {}
  })
    .then((response) => {
      if (!response.ok) throw new Error(String(response.status))
      return response.blob()
    })
    .then((blob) => {
      const url = URL.createObjectURL(blob)
      cache.set(id, url)
      return url
    })
}
