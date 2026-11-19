# Markdown editor

`@dawnim/admin` ships a complete Markdown editor kernel: skin, toolbar, floating panels, live preview, fullscreen, emoji, and table interactions. Applications only provide rendering rules, image endpoints, and a few options — the editor UI itself is never reimplemented.

The website (`app/web/assets/site.js`, via `window.DawnMomentEditor`) and the admin panel (`createMarkdownEditor`) use the same kernel.

## Features

| Area | Capabilities |
|---|---|
| Writing | Headings, bold / italic / underline / strikethrough / highlight, inline and block code, quotes, lists, alignment, tables, links, spoilers, video, `@` mentions |
| Views | Write, live preview, and preview modes; side-by-side on wide screens, stacked on narrow screens; fullscreen and line numbers |
| Emoji | Full built-in dataset with grouped scrolling, sticky anchors, following tabs, and remembered recent emojis |
| Images | History library, dedupe checks, immediate upload, file picker / drag-and-drop / clipboard paste, body or attachment-gallery targets |
| Extension | Hide any tool, add custom tools, multiple isolated instances, and safe editing of existing content without re-uploading |

On small screens the toolbar stays on one line and scrolls horizontally; it never wraps and pushes the textarea down.

## Quick start

```ts
import { createMarkdownEditor, editorTemplate, renderMarkdown } from '@dawnim/admin'
import '@dawnim/admin/style.css'

const root = document.querySelector<HTMLElement>('#editor')!
root.innerHTML = editorTemplate('zh')

const editor = createMarkdownEditor(root, {
  render: (text) => renderMarkdown(text),
  maxLength: 1024,
  onChange: (text) => console.log(text)
})
```

If the page already contains the same editor DOM, pass that container directly and the kernel adopts it instead of generating a second copy.

For plain script tags, use the published `dist/dawn-editor.js` and `dist/dawn-editor.css`. The script is an IIFE exposing `window.DawnEditor` with `createMarkdownEditor`, `editorTemplate`, `renderEditorIcons`, `renderMarkdown`, and `enhanceMarkdown`.

## Key options

| Option | Purpose |
|---|---|
| `render(text)` | Required. Application Markdown renderer used by previews and summaries |
| `enhance(root)` | Post-render DOM enhancement such as lightboxes or syntax highlighting |
| `value` / `maxLength` | Initial content (default limit 300) and maximum length |
| `format` | `moment` for compact posts or `post` for the full editor |
| `alwaysFull` | Keep the full toolbar and preview even for `moment` (admin use) |
| `locale` | `zh` or `en` for built-in copy |
| `mentions(query)` | Resolve `{ label, value }[]` for `@` suggestions |
| `config()` | Return `{ hidden: string[] }` to hide tools by key |
| `previewImage(src)` | Handle image clicks in previews |
| `onChange(text)` / `onSaveDraft()` | Content changes and `Cmd/Ctrl+S` |
| `onImagesChange(ids)` | Referenced image ids changed |

Instance methods: `getValue()`, `setValue(text)`, `insertText(text)`, `focus()`, `getImages()`, `setImages(images)`, `openImageLibrary(anchor?)`, `setOptions(...)`, and `destroy()`. Use `setImages` to restore attachments when editing existing content; it never re-uploads.

## Images

All image providers are optional. Without them the image button falls back to a plain file picker.

```ts
const editor = createMarkdownEditor(root, {
  render: renderMarkdown,
  imageTarget: 'choose',
  imageLibrary: {
    list: (keyword) => api.get('/api/moment-images/mine', { params: { keyword } }),
    refresh() { grid.reload() }
  },
  imageUploader: {
    fingerprint: async (file) => ({ md5: await md5(file), sha256: await sha256(file) }),
    check: (files, prints) => api.post('/api/moment-image-checks', { files: prints }),
    upload: (files, prints) => api.upload('/api/moment-images', files, prints),
    toMarkdown: (image, name) => `![${name || image.name}](${image.url})`
  },
  onImagesChange: (ids) => console.log(ids)
})
```

`check` must return one result per input file: an existing image for a dedupe hit, `null` otherwise. Without `check`, every file is uploaded. `toMarkdown` controls the syntax inserted into the body.

`imageTarget` picks the default destination:

| Value | Behaviour |
|---|---|
| `body` | Insert into the document |
| `gallery` | Only add to the attachment gallery |
| `choose` | Let the user pick in the library panel |

Picker, drag-and-drop, and clipboard paste share the same pipeline: fingerprint, dedupe check, then upload. Blob URLs are local previews only and are never persisted again.

## Custom tools

```ts
createMarkdownEditor(root, {
  render: renderMarkdown,
  config: () => ({ hidden: ['media', 'alignment'] }),
  customTools: [{
    key: 'my-images',
    title: ['我的图片', 'My images'],
    icon: 'image',
    where: 'both',
    onClick: ({ editor, anchor }) => editor.openImageLibrary(anchor)
  }]
})
```

`where` accepts `toolbar`, `attachments`, or `both`. On pages with multiple editors, pass a distinct `attachmentHost` to each instance.

## Static site integration

The website build concatenates the framework `dist/dawn-editor.js/css` with the application render bundle into `moment-markdown.js/css`, which Go embeds. The application keeps only its Markdown rules:

```js
const { enhanceMarkdown, createMarkdownEditor } = window.DawnEditor
window.DawnMomentMarkdown = { render: renderMarkdown, enhanceMarkdown }
window.DawnMomentEditor = { create: createMarkdownEditor }
```

Upgrading the framework therefore never requires rebuilding framework sources into the website.

## Keyboard and interactions

`Cmd/Ctrl+S` calls `onSaveDraft()`; `Cmd/Ctrl+B/I/U` apply bold, italic, and underline. Link, table, media, and spoiler panels are anchored to their toolbar buttons, close on outside click or `Esc`, and never submit the host form. Panels and emoji popovers are isolated per editor instance.

See [Frontend](frontend.md) and [Community example](moments.md) for complete integrations.
