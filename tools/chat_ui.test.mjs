import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { JSDOM } from 'jsdom'

const root = new URL('../app/web/assets/', import.meta.url)
const read = (name) => readFileSync(new URL(name, root), 'utf8')
const dom = new JSDOM(read('index.html'), { url: 'http://127.0.0.1/', runScripts: 'outside-only', pretendToBeVisual: true })
const { window } = dom
globalThis.window = window
globalThis.document = window.document
globalThis.Node = window.Node

const content = [
    '# 一级标题',
    '**加粗** ~~删除~~ ==高亮== <u>下划线</u>',
    '',
    '```js',
    'let a = 1',
    '```',
    '',
    '> 引用内容',
    '',
    '!!剧透内容!!',
    '',
    '- [ ] 没完成哦',
    '- [x] 已经完成啦',
].join('\n')
const routes = {
    '/api/site': { support_key: 'demo', site_name: 'Dawn' },
    '/api/moment-config': { toolbar_hidden: [], emojis: {} },
    '/api/support-identities': { user: { type: 'member', unique_id: 1001, nickname: 'Doreen' }, support: {} },
    '/api/community-chats/7/messages': [{ id: 11, type: 'message', sender_subject: '2001', nickname: 'anonymous', content }],
    '/api/community-chats/7/read': {},
    '/api/community-chats': [{ id: 7, type: 'room', content: '# 一级标题\n!!秘密内容!!', unread: 0 }],
    '/api/moments': [],
}
window.fetch = async (path) => {
    const key = Object.keys(routes).find((route) => String(path).startsWith(route))
    return { ok: true, status: 200, json: async () => ({ data: key ? routes[key] : {} }) }
}
window.matchMedia = () => ({ matches: false, addEventListener() {}, removeEventListener() {} })
window.WebSocket = class { close() {} }
window.SparkMD5 = { ArrayBuffer: { hash: () => '' } }
window.crypto = { subtle: { digest: async () => new ArrayBuffer(32) } }
window.DawnSupport = {
    init(config) { void config.resolveIdentity({ visitor_ticket: 'x' }).catch(() => {}) },
    close() {}, setTheme() {}, setLocale() {}, syncIdentity: async () => ({}), reset() {},
    createRealtimeTicket: () => new Promise(() => {}),
}
window.HTMLDialogElement.prototype.showModal = function () { this.open = true }
window.HTMLDialogElement.prototype.close = function () { this.open = false }
window.URL.createObjectURL = () => 'blob:http://127.0.0.1/chat'
window.URL.revokeObjectURL = () => {}
window.scrollTo = () => {}
for (const name of ['spark-md5.min.js', 'moment-markdown.js', 'site.js']) window.eval(read(name))
const $ = (id) => window.document.getElementById(id)
const tick = () => new Promise((resolve) => setTimeout(resolve, 80))
const pointer = (target) => target.dispatchEvent(new window.Event('pointerdown', { bubbles: true }))

test('聊天消息按 Markdown 渲染并带通用增强', async () => {
    await tick()
    $('nav-chat').click()
    await tick()
    const body = $('chat-messages').querySelector('.chat-message .markdown-body')
    assert.ok(body, '消息正文没有走 markdown-body')
    assert.ok(body.querySelector('h1'), '标题未渲染')
    assert.ok(body.querySelector('strong'), '加粗未渲染')
    assert.ok(body.querySelector('mark'), '高亮未渲染')
    assert.ok(body.querySelector('pre.md-code-block'), '代码块未渲染')
    assert.ok(body.querySelector('.md-copy'), '代码复制按钮缺失')
    assert.ok(body.querySelector('blockquote'), '引用未渲染')
    assert.ok(body.querySelector('.task-list-item input'), '待办未渲染')
    const mask = body.querySelector('.md-mask')
    assert.ok(mask, '遮罩未渲染')
    mask.click()
    assert.ok(mask.classList.contains('revealed'), '遮罩点击不能展开')
    assert.equal($('chat-messages').querySelector('.chat-message > strong').textContent, 'anonymous', '昵称被破坏')
    const preview = $('chat-list').querySelector('.chat-preview')
    assert.ok(!preview.textContent.includes('#'), '会话摘要仍显示 Markdown 标记')
    assert.ok(!preview.textContent.includes('秘密内容') && preview.textContent.includes('•••'), '剧透在摘要里未遮盖')
})

test('点击聊天浮层外部快捷关闭，浮层内部与移动端不关闭', () => {
    const chat = $('community-chat')
    chat.hidden = false
    pointer(window.document.body)
    assert.equal(chat.hidden, true, '桌面点击外部未关闭')
    chat.hidden = false
    pointer($('chat-title'))
    assert.equal(chat.hidden, false, '点击浮层内部被误关')
    Object.defineProperty(window, 'innerWidth', { value: 390, configurable: true })
    pointer(window.document.body)
    assert.equal(chat.hidden, false, '移动端不应点外部关闭')
    Object.defineProperty(window, 'innerWidth', { value: 1280, configurable: true })
})

test('聊天内昵称与代码块按钮不受消息内元素影响', () => {
    const css = read('site.css')
    assert.match(css, /\.chat-message > strong \{/)
    assert.match(css, /\.chat-message > button \{/)
    assert.ok(!/\.chat-message strong \{/.test(css))
    assert.ok(!/\.chat-message \{[^}]*white-space: pre-wrap/.test(css))
})
