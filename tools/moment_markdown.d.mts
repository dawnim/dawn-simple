export function renderMarkdown(
  source: string,
  emojis?: Record<string, string>,
  resolveAttachment?: (id: string) => { url?: string; name?: string } | undefined
): string
