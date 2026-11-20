import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { JSDOM } from 'jsdom'

const asset = name => readFileSync(new URL(`../app/web/assets/${name}`, import.meta.url), 'utf8')

test('编辑器加载缓慢时首页先显示，身份恢复不重复请求列表', async () => {
    const dom = new JSDOM(asset('index.html'), { url: 'http://localhost/', runScripts: 'outside-only' })
    const { window } = dom
    window.matchMedia = () => ({ matches: true, addEventListener() {} })
    window.scrollTo = () => {}
    window.localStorage.setItem('dawn.member', 'existing')
    let resolveMarkdown
    window.momentMarkdownReady = new Promise(resolve => { resolveMarkdown = resolve })
    const requests = []
    window.fetch = async path => {
        requests.push(path)
        let data = {}
        if (String(path).startsWith('/api/moments?')) data = [{ id: 1, title: '测试帖子', content: '正文', images: '[]', created_at: 1 }]
        if (path === '/api/moment-overview') data = { posts: 1, replies: 0, members: 1, hot: [] }
        if (path === '/api/site') data = { support_key: 'demo' }
        if (path === '/api/support-identities') data = { user: { type: 'member', unique_id: '1001' }, support: {} }
        return { ok: true, status: 200, json: async () => ({ data }) }
    }
    window.DawnSupport = { init(options) { void options.resolveIdentity({ visitor_ticket: 'test' }) }, setTheme() {}, setLocale() {} }
    window.eval(asset('site.js'))
    await new Promise(resolve => setTimeout(resolve, 50))
    assert.equal(window.document.querySelector('.moment-title')?.textContent, '测试帖子')
    assert.equal(window.document.querySelector('#site-startup'), null)
    assert.equal(requests.filter(path => String(path).startsWith('/api/moments?')).length, 1)
    window.DawnMomentEditor = { create: () => ({}) }
    resolveMarkdown()
    await new Promise(resolve => setTimeout(resolve, 0))
    dom.window.close()
})
