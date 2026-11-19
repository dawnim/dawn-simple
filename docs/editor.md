# Markdown 编辑器

编辑器提供工具栏、三种编辑视图、全屏、行号、表情、快捷键和多实例隔离。工具栏在窄屏单行横滑，实时预览上下排列。任何能提供 DOM 容器的前端都可以使用，渲染与上传接入由调用方提供。

## 快速开始

```ts
import { createMarkdownEditor, renderMarkdown } from '@dawnim/admin/editor'
import '@dawnim/admin/editor.css'

const editor = createMarkdownEditor(document.querySelector<HTMLElement>('#editor')!, {
  render: renderMarkdown,
  maxLength: 10000,
  placeholder: '输入正文'
})

// 提交表单时读取 Markdown 原文
const markdown = editor.getValue()
```

独立入口使用浏览器 DOM。后台可直接从 `@dawnim/admin` 导入同一套接口。服务端模板或纯 HTML 页面加载发布包的 `dist/dawn-editor.js` 与 `dist/dawn-editor.css`，通过 `window.DawnEditor.createMarkdownEditor` 挂载空容器。

## 配置与实例方法

| 参数 | 用途 |
|---|---|
| `render(text)` | 把源文本渲染为可直接插入页面的安全 HTML，也可接入自定义 Markdown 方言 |
| `enhance(root)` | 在预览 HTML 生成后补充图片查看、代码复制等交互 |
| `value` / `placeholder` / `maxLength` | 初始内容、输入提示和字数上限 |
| `compact` / `locale` | 精简视图与 `zh`、`en` 文案 |
| `config()` | 返回 `{ hidden: string[] }`，按 key 隐藏工具，例如 `image` 或 `media` |
| `mentions(query)` / `previewImage(src)` | 提及候选和点击预览图片 |
| `onChange(text)` / `onSaveDraft()` | 正文变化和 `Cmd/Ctrl+S` 草稿回调 |
| `imageLibrary` / `imageUploader` | 历史图片查询和上传 |
| `pasteImages` | 默认关闭。启用并接入上传后，粘贴图片会上传并插入正文，独立于图片按钮显隐 |
| `mediaMarkup(url, type)` | 自定义媒体插入文本，默认插入 Markdown 链接 |
| `customTools` | 注册自定义工具栏按钮和点击行为 |

实例可通过 `getValue()`、`setValue(text)`、`insertText(textOrTransform)`、`setOptions(options)`、`focus()` 和 `destroy()` 操作。`insertText` 的函数参数接收当前选区文本。

## 图片与自定义语法

图片 provider 返回 `{ id, url, name? }`。默认插入标准 Markdown 图片，`url` 使用可保存的地址。图片按钮可以查询 `imageLibrary.list(keyword)` 或选择本地文件。可用 `imageUploader.fingerprint` 与 `check` 做秒传检查，`check` 按文件顺序返回命中记录或 `null`，`upload` 返回同顺序的上传结果。业务可以用 `imageUploader.toMarkdown(image, name)` 生成自己的引用。

调用方可同时扩展工具和渲染器。例如按钮用 `[login]…[/login]` 包裹选区：

```ts
customTools: [{
  key: 'login',
  title: ['登录可见', 'Sign-in required'],
  onClick: ({ editor }) => editor.insertText(text => `[login]${text}[/login]`)
}]
```

在 `render` 中接入业务解析器来预览标记。正式展示时，由业务后端根据访问者权限决定返回哪些正文。Simple 的站内附件接法见[网站动态开发实例](moments.md)。
