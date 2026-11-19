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
