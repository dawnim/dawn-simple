import { test } from 'node:test'
import assert from 'node:assert/strict'
import { JSDOM } from 'jsdom'

const dom = new JSDOM('<!doctype html><html><body></body></html>')
globalThis.window = dom.window
globalThis.document = dom.window.document
globalThis.Node = dom.window.Node
const { renderMarkdown } = await import('./moment_markdown.mjs')

const cases = [
    ['基础格式', '**粗** *斜* ~~删~~ ==高== ++插++ <u>下划线</u>',
        /<strong>粗<\/strong> <em>斜<\/em> <s>删<\/s> <mark>高<\/mark> <ins>插<\/ins> <u>下划线<\/u>/],
    ['对齐不产生非法嵌套', '<center>居中</center>\n\n<p align="right">右</p>',
        /<center>居中<\/center>\s*<p align="right">右<\/p>/],
    ['标题锚点与去重', '# 同一标题\n\n# 同一标题', /id="同一标题"[\s\S]*id="同一标题-1"/],
    ['任务列表', '- [ ] 待办\n- [x] 完成', /task-list-item[\s\S]*checked=""[\s\S]*完成/],
    ['提示块', ':::warning 注意\n风险\n:::', /md-callout-warning[\s\S]*注意[\s\S]*风险/],
    ['折叠剧透', ':::spoiler 剧透\n隐藏\n:::', /<details class="md-spoiler">[\s\S]*隐藏/],
    ['点击遮罩与提及', '!!遮罩!! @alice', /md-mask[\s\S]*<a class="md-mention" href="\/\?author=alice">/],
    ['自定义表情', ':party:', /<span class="md-emoji">🎉<\/span>/],
    ['行内公式', '$E=mc^2$', /class="katex"/],
    ['代码高亮与语言头', '```js\nconst a = 1\n```', /md-code-header[\s\S]*>js<[\s\S]*hljs-keyword/],
    ['Mermaid 容器', '```mermaid\ngraph TD\n  A-->B\n```', /<pre class="md-mermaid">graph TD/],
    ['iframe 域名白名单', 'MEDIA::iframe::https://player.bilibili.com/player.html?bvid=1', /<iframe[^>]*player\.bilibili\.com/],
    ['图片与外链安全', '![外链](https://example.com/a.png)\n\n[站点](https://example.com)', /class="md-image"[\s\S]*target="_blank"[\s\S]*noopener/],
    ['表格与脚注', '| a | b |\n| - | - |\n| 1 | 2 |\n\n文字[^1]\n\n[^1]: 注释', /<table>[\s\S]*footnotes/],
    ['上下标与缩写', 'x^2^ H~2~O', /<sup>2<\/sup>[\s\S]*<sub>2<\/sub>/],
    ['编辑器生成的表格', '| 列 1 | 列 2 |\n| --- | --- |\n| 内容 1-1 | 内容 1-2 |', /<table>[\s\S]*<th>列 1<\/th>[\s\S]*内容 1-1/],
    ['预览 blob 图片保留 src', '![shot.png](blob:http://127.0.0.1/preview)', /<img[^>]*src="blob:http:\/\/127\.0\.0\.1\/preview"/],
    ['编辑器媒体与链接', 'MEDIA::video::https://example.com/a.mp4\n\n[示例](https://example.com)', /<video[^>]*example\.com[\s\S]*<a[^>]*href="https:\/\/example\.com"/],
]

for (const [name, source, expected] of cases) {
    test(name, () => assert.match(renderMarkdown(source, { party: '🎉' }), expected))
}

const attacks = [
    ['script 标签', '<script>alert(1)</script>', /<script/i],
    ['javascript 链接', '[x](javascript:alert(1))', /<a[^>]*href="javascript/i],
    ['事件属性', '<img src=x onerror=alert(1)>', /onerror/i],
    ['内联样式注入', '<p style="position:fixed">x</p>', /style=/i],
    ['非白名单 iframe', 'MEDIA::iframe::https://evil.com/x', /<iframe/],
    ['任意 HTML iframe', '<iframe src="https://evil.com/x"></iframe>', /evil\.com/],
    ['form 注入', '<form action="https://evil.com"><input name="a"></form>', /<form/i],
    ['未知协议图片', '![x](javascript:alert(1))', /<img/i],
]

for (const [name, source, forbidden] of attacks) {
    test(`XSS: ${name}`, () => assert.doesNotMatch(renderMarkdown(source), forbidden))
}
