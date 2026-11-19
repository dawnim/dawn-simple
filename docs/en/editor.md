# Markdown editor

The editor provides a toolbar, three editing views, fullscreen mode, line numbers, emoji, keyboard shortcuts, and isolated instances. On narrow screens, the toolbar scrolls horizontally and live preview stacks vertically. Any frontend with a DOM container can mount it and supply its own rendering and upload handlers.

## Quick start

Use `@dawnim/admin` 0.1.7. Add an empty container and mount the editor after the DOM is ready.

```html
<div id="editor"></div>
```

```ts
import { createMarkdownEditor, renderMarkdown, enhanceMarkdown } from '@dawnim/admin/editor'
import '@dawnim/admin/editor.css'

const editor = createMarkdownEditor(document.querySelector<HTMLElement>('#editor')!, {
  render: renderMarkdown,
  enhance: enhanceMarkdown,
  maxLength: 10000
})
```

Read the Markdown source with `editor.getValue()` on submit and call `editor.destroy()` when the view unmounts. On a read-only page, use `renderMarkdown(text)`, insert the result, then call `enhanceMarkdown(container)` for DOM interactions.

Admin apps can import the same API from `@dawnim/admin`. For server-rendered templates or static pages, serve the files from the package’s `dist/` directory and mount through the browser global.

```html
<link rel="stylesheet" href="/assets/dawn-editor.css">
<div id="editor"></div>
<script src="/assets/dawn-editor.js"></script>
<script>
  const { createMarkdownEditor, renderMarkdown, enhanceMarkdown } = window.DawnEditor
  const editor = createMarkdownEditor(document.getElementById('editor'), {
    render: renderMarkdown,
    enhance: enhanceMarkdown
  })
</script>
```

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
| `mediaMarkup(url, type)` | Customize media insertion text. The default is a Markdown autolink |
| `customTools` | Register a toolbar button and its action |

Use `getValue()`, `setValue(text)`, `insertText(textOrTransform)`, `setOptions(options)`, `focus()`, or `destroy()` on an instance. A function passed to `insertText` receives the current selection.

## Images

The built-in renderer supports standard `![alt text](image-url)` and reference-style images. Image providers return `{ id, url, name? }`, and the editor inserts standard image markup by default. Supply a URL that will stay accessible in saved documents. The image button can query `imageLibrary.list(keyword)` or open a file picker. Supply `imageUploader.fingerprint` and `check` for deduplication. `check` returns one match or `null` per file, and `upload` returns images in input order. Use `imageUploader.toMarkdown(image, name)` to insert an application-specific reference.

## Custom syntax

Use `customTools` to insert application markup and supply the parser through `render`. This button wraps the selection in `[login]…[/login]`.

```ts
customTools: [{
  key: 'login',
  title: ['登录可见', 'Sign-in required'],
  onClick: ({ editor }) => editor.insertText(text => `[login]${text}[/login]`)
}]
```

The application defines and parses tags such as `[login]`, `[reply]`, and `[buy]`. For published content, its backend filters the body by the viewer’s permissions before returning it. See the [forum development example](moments.md) for Simple's attachment syntax.
