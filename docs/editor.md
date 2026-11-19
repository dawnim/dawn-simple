# Markdown 编辑器

编辑器提供工具栏、三种编辑视图、全屏、行号、表情、快捷键和多实例隔离。工具栏在窄屏单行横滑，实时预览上下排列。任何能提供 DOM 容器的前端都可以使用，渲染与上传接入由调用方提供。

## 快速开始

使用 `@dawnim/admin` 0.1.7。在页面放置一个空容器，等 DOM 就绪后挂载。

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

提交时用 `editor.getValue()` 读取 Markdown 原文，卸载页面时调用 `editor.destroy()`。只读页面可调用 `renderMarkdown(text)`，插入结果后用 `enhanceMarkdown(container)` 启用 DOM 交互。

后台应用也可从 `@dawnim/admin` 导入接口。服务端模板或静态页面可将发布包 `dist/` 下的同名资源放到站点静态目录，直接通过全局对象挂载。

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
| `mediaMarkup(url, type)` | 自定义媒体插入文本，默认插入 Markdown 自动链接 |
| `customTools` | 注册自定义工具栏按钮和点击行为 |

实例可通过 `getValue()`、`setValue(text)`、`insertText(textOrTransform)`、`setOptions(options)`、`focus()` 和 `destroy()` 操作。`insertText` 的函数参数接收当前选区文本。

## 图片接入

内置渲染器支持标准图片语法 `![替代文字](图片地址)` 和引用式图片。图片回调返回 `{ id, url, name? }`，编辑器默认生成标准图片语法，`url` 使用持久可访问的地址。图片按钮可以查询 `imageLibrary.list(keyword)` 或选择本地文件。可用 `imageUploader.fingerprint` 与 `check` 做秒传检查，`check` 按文件顺序返回命中记录或 `null`，`upload` 返回同顺序的上传结果。业务可以用 `imageUploader.toMarkdown(image, name)` 生成自己的引用。

## 自定义语法

业务可以通过 `customTools` 插入自定义标记，通过 `render` 接入对应的解析器。下面的工具将当前选区包裹为 `[login]…[/login]`。

```ts
customTools: [{
  key: 'login',
  title: ['登录可见', 'Sign-in required'],
  onClick: ({ editor }) => editor.insertText(text => `[login]${text}[/login]`)
}]
```

`[login]`、`[reply]`、`[buy]` 的语义及解析器由使用方定义。发布后的内容由业务后端先按访问者权限过滤，再返回可见正文。Simple 的站内附件接法见[论坛动态开发实例](moments.md)。
