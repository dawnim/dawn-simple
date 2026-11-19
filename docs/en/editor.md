# Markdown editor

The editor provides a toolbar, three editing views, fullscreen mode, line numbers, emoji, keyboard shortcuts, and isolated instances. On narrow screens, the toolbar scrolls horizontally and live preview stacks vertically. Any frontend with a DOM container can mount it and supply its own rendering and upload handlers.

## Quick start

```ts
import { createMarkdownEditor, renderMarkdown } from '@dawnim/admin/editor'
import '@dawnim/admin/editor.css'

const editor = createMarkdownEditor(document.querySelector<HTMLElement>('#editor')!, {
  render: renderMarkdown,
  maxLength: 10000,
  placeholder: 'Write something'
})

// Read the Markdown source when the form is submitted
const markdown = editor.getValue()
```

The standalone entry point uses the browser DOM. Admin apps can import the same API from `@dawnim/admin`. For a server-rendered template or plain HTML, load `dist/dawn-editor.js` and `dist/dawn-editor.css` from the package and mount an empty container with `window.DawnEditor.createMarkdownEditor`.

## Configuration and instance methods

| Option | Purpose |
|---|---|
| `render(text)` | Return sanitized HTML from Markdown or an application-defined dialect |
| `enhance(root)` | Add interactions after preview rendering |
| `value` / `placeholder` / `maxLength` | Initial content, input hint, and character limit |
| `compact` / `locale` | Compact view and `zh` or `en` UI text |
| `config()` | Return `{ hidden: string[] }` to hide tools by key, such as `image` or `media` |
| `mentions(query)` / `previewImage(src)` | Mention suggestions and preview image clicks |
| `onChange(text)` / `onSaveDraft()` | Document changes and `Cmd/Ctrl+S` |
| `imageLibrary` / `imageUploader` | Query existing images and upload new files |
| `pasteImages` | Off by default. With an uploader, paste uploads and inserts images into the document independently of button visibility |
| `mediaMarkup(url, type)` | Customize media insertion text. The default is a Markdown link |
| `customTools` | Register a toolbar button and its action |

Use `getValue()`, `setValue(text)`, `insertText(textOrTransform)`, `setOptions(options)`, `focus()`, or `destroy()` on an instance. A function passed to `insertText` receives the current selection.

## Images and custom syntax

Image providers return `{ id, url, name? }`. The default inserts a standard Markdown image with a persistable URL. The image button can query `imageLibrary.list(keyword)` or open a file picker. Supply `imageUploader.fingerprint` and `check` for deduplication. `check` returns one match or `null` per file, and `upload` returns images in input order. Use `imageUploader.toMarkdown(image, name)` to insert an application-specific reference.

An application can extend both the toolbar and renderer. This button wraps the selection in `[login]…[/login]`.

```ts
customTools: [{
  key: 'login',
  title: ['登录可见', 'Sign-in required'],
  onClick: ({ editor }) => editor.insertText(text => `[login]${text}[/login]`)
}]
```

Connect your parser through `render` for previews. For published content, the application backend decides which sections to return based on the viewer's permissions. See the [community example](moments.md) for Simple's attachment syntax.
