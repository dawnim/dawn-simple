import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { JSDOM } from 'jsdom'

const dom = new JSDOM('<!doctype html><html><body></body></html>', { runScripts: 'outside-only' })
dom.window.eval(readFileSync(new URL('../app/web/assets/moment-markdown.js', import.meta.url), 'utf8'))
const engine = dom.window.DawnMomentMarkdown

test('官网产物导出渲染与增强能力', () => {
    assert.equal(typeof engine?.render, 'function')
    assert.equal(typeof engine?.enhanceMarkdown, 'function')
    assert.match(engine.render('**粗**'), /<strong>粗<\/strong>/)
})

test('站内图片使用附件原名作为 alt，忽略代码和转义文本', () => {
    const ids = []
    const markup = engine.render(
        '[attachment://3]\n\n`[attachment://4]`\n\n```md\n[attachment://5]\n```\n\n\\[attachment://6]\n\n[attachment://7](https://example.com)',
        {},
        id => { ids.push(id); return { url: '/img/3', name: '原始"文件名".png' } }
    )
    const fragment = dom.window.document.createElement('template')
    fragment.innerHTML = markup
    const image = fragment.content.querySelector('img')
    assert.deepEqual(ids, ['3'])
    assert.equal(image.alt, '原始"文件名".png')
    assert.equal(image.getAttribute('src'), '/img/3')
    assert.equal(fragment.content.querySelector('code').textContent, '[attachment://4]')
    assert.equal(fragment.content.querySelector('a').textContent, 'attachment://7')
})

test('官网和通用渲染器保留标准 Markdown 的相对路径和 alt', () => {
    for (const render of [engine.render, dom.window.DawnEditor.renderMarkdown]) {
        const html = render('![说明][figure]\n\n[figure]: ./a.png "标题"')
        assert.match(html, /src="\.\/a\.png"/)
        assert.match(html, /alt="说明"/)
        assert.match(html, /title="标题"/)
    }
})
