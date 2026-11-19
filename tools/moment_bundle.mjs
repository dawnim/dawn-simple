import { renderMarkdown } from './moment_markdown.mjs'

// 渲染是官网业务定制；编辑器与增强来自框架发布的 dawn-editor.js（全局 DawnEditor）。
const { enhanceMarkdown, createMarkdownEditor } = window.DawnEditor
window.DawnMomentMarkdown = { render: renderMarkdown, renderMarkdown, enhanceMarkdown }
window.DawnMomentEditor = { create: createMarkdownEditor }
