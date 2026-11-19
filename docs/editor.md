# Markdown 编辑器

`@dawnim/admin` 自带一套完整的 Markdown 编辑器内核：皮肤、工具按钮、浮层、实时预览、全屏、表情与表格交互都由框架提供。业务只负责渲染规则、图片接口与少量配置，不需要自己实现编辑器 UI。

官网和后台使用同一个内核：官网 `app/web/assets/site.js` 通过 `window.DawnMomentEditor` 调用，后台页面通过 `createMarkdownEditor` 直接调用。

## 能力一览

| 分类 | 能力 |
|---|---|
| 编写 | 标题、粗体/斜体/下划线/删除线/高亮、行内代码与代码块、引用、有序/无序列表、分割线、对齐、表格、链接、剧透、视频、@ 提及 |
| 视图 | 正文、实时预览、预览三态；实时预览在宽屏左右分栏，窄屏自动上下堆叠；支持全屏与行号 |
| 表情 | 内置完整 emoji 数据，按分组滚动、锚点吸顶、底部标签跟随；最近使用会记住 |
| 图片 | 历史图库、秒传检查、立即上传、点选/拖拽/剪贴板粘贴；正文与下方附件区两种落点 |
| 扩展 | 隐藏任意工具按钮、添加自定义按钮、多实例隔离、编辑已有内容时不重复上传 |

移动端工具栏保持单行，可左右滑动；按钮不会换行挤占正文高度。

## 快速开始

```ts
import { createMarkdownEditor, editorTemplate, renderMarkdown } from '@dawnim/admin'
import '@dawnim/admin/style.css'

const root = document.querySelector<HTMLElement>('#editor')!
root.innerHTML = editorTemplate('zh')

const editor = createMarkdownEditor(root, {
  render: (text) => renderMarkdown(text),
  value: '',
  maxLength: 1024,
  onChange: (text) => console.log(text)
})
```

`editorTemplate()` 生成完整 DOM；如果页面已有同款结构（例如官网的 `#moment-*` 静态 DOM），直接把容器交给 `createMarkdownEditor` 即可接管，不会重复生成。

独立脚本场景可直接使用发布产物 `dist/dawn-editor.js` 与 `dist/dawn-editor.css`。JS 是 IIFE，加载后暴露全局 `window.DawnEditor`，其中包含 `createMarkdownEditor`、`editorTemplate`、`renderEditorIcons`、`renderMarkdown` 和 `enhanceMarkdown`。

## 常用选项

| 选项 | 作用 |
|---|---|
| `render(text)` | 必填，业务自己的 Markdown 渲染规则；预览与列表摘要都用它 |
| `enhance(root)` | 渲染后增强 DOM，如图片灯箱、代码高亮 |
| `value` / `maxLength` | 初始值（默认 300）与字数上限 |
| `format` | `moment` 精简动态或 `post` 完整帖子 |
| `alwaysFull` | 后台场景：即使是动态格式也保留完整工具栏与预览 |
| `locale` | `zh` 或 `en`，影响内置文案 |
| `mentions(query)` | 返回 `@` 候选 `{ label, value }[]` |
| `config()` | 返回 `{ hidden: string[] }`，按工具 key 隐藏按钮 |
| `previewImage(src)` | 预览里点击图片的回调；不传则新窗口打开 |
| `onChange(text)` / `onSaveDraft()` | 正文变化、`Cmd/Ctrl+S` 保存草稿 |
| `onImagesChange(ids)` | 附件区或正文引用的图片 id 变化 |

实例方法：`getValue()`、`setValue(text)`、`insertText(text)`、`focus()`、`getImages()`、`setImages(images)`、`openImageLibrary(anchor?)`、`setOptions(...)`、`destroy()`。`setImages` 用于编辑已有内容时回填附件区，不会重新上传。

## 图片能力接入

图片接口是可选注入；一个都不接时，图片按钮退化为「直接选择本地文件」。

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

- `fingerprint` 计算摘要；内核不假设算法，业务自行决定 md5、sha256 与大小。
- `check` 按 `files` 顺序返回等长数组，命中秒传的位置返回已有图片，其余为 `null`。不提供 `check` 时全部走 `upload`。
- `upload` 返回与入参等序的图片数组，每项为 `{ id, url, name? }`。
- `toMarkdown` 决定正文里的图片语法；不传时用内核默认标记。

`imageTarget` 控制图片按钮的默认落点：

| 值 | 行为 |
|---|---|
| `body` | 选完直接插入正文 |
| `gallery` | 只加入下方附件区（如动态格式） |
| `choose` | 图库浮层里让用户选「插正文」或「仅加入九宫格」 |

粘贴、拖拽和点选都走同一条上传链路：先摘要、再秒传检查、未命中才上传。附件区图片用 `getImages()` 读取，保存时与正文一起提交；blob 只用于本地预览，不会重新入库。

## 自定义按钮与隐藏按钮

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

`where` 可取 `toolbar`、`attachments` 或 `both`。同一页面有多个编辑器时，各自传入 `attachmentHost`，避免抢同一个附件容器。

## 独立静态站接入

官网静态资源生成脚本先把框架 `dist/dawn-editor.js/css` 与业务自己的渲染包拼成 `moment-markdown.js/css`，再由 Go `embed` 下发。业务侧只保留 Markdown 渲染规则：

```js
const { enhanceMarkdown, createMarkdownEditor } = window.DawnEditor
window.DawnMomentMarkdown = { render: renderMarkdown, enhanceMarkdown }
window.DawnMomentEditor = { create: createMarkdownEditor }
```

这样框架升级不会要求官网重新打包框架源码，两边的编辑器能力始终来自同一个发布包。

## 键盘与交互

- `Cmd/Ctrl+S` 触发 `onSaveDraft()`；`Cmd/Ctrl+B/I/U` 对应粗体、斜体、下划线。
- 链接、表格、媒体、剧透等是锚定工具栏按钮的浮层，点外部或 `Esc` 关闭，不遮挡宿主表单；`Enter` 在浮层里执行插入，不会提交外层表单。
- 帮助面板、浮层与 emoji 面板都按编辑器实例隔离；同页多个编辑器不会串台。
- 移动端工具栏单行横滑，实时预览自动变成上下结构。

后台与官网的完整接入示例见[前端开发](frontend.md)和[网站动态开发实例](moments.md)。
