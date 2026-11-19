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
window.fetch = async () => ({ ok: true, json: async () => ({ data: {}, toolbar_hidden: [], emojis: {} }) })
window.matchMedia = () => ({ matches: false, addEventListener() {}, removeEventListener() {} })
window.WebSocket = class { close() {} }
window.SparkMD5 = { ArrayBuffer: { hash: () => '' } }
window.crypto = { subtle: { digest: async () => new ArrayBuffer(32) } }
window.DawnSupport = { init() {} }
window.HTMLDialogElement.prototype.showModal = function () { this.open = true }
window.HTMLDialogElement.prototype.close = function () { this.open = false }
window.URL.createObjectURL = () => 'blob:http://127.0.0.1/preview'
window.URL.revokeObjectURL = () => {}
window.scrollTo = () => {}
window.localStorage.setItem('dawn.site.locale', 'zh')
for (const name of ['spark-md5.min.js', 'moment-markdown.js', 'site.js']) window.eval(read(name))
const $ = (id) => window.document.getElementById(id)

const dialog = $('moment-dialog')
const ed = () => $('moment-editor-shell').querySelector('[data-editor-shell]')
const editorNode = (name) => ed().querySelector(`[data-editor-${name}]`)
const input = () => editorNode('input')
const preview = () => editorNode('preview')
const tool = (key) => ed().querySelector(`[data-editor-command="${key}"]`)
const panel = (name) => dialog.querySelector(`[data-editor-panel="${name}"]`)
const help = () => $('moment-editor-shell').querySelector('[data-editor-help]')

test('官网从通用模板挂载编辑器', () => {
    assert.ok($('moment-editor-shell').classList.contains('dawn-editor'))
    assert.equal(document.querySelectorAll('.md-editor-root, .md-editor-shell').length, 0)
    assert.equal(ed().querySelectorAll('.editor-toolbar .editor-tool').length, 17)
    assert.ok(ed().querySelectorAll('.editor-tool svg').length >= 10)
})

test('交流板块下拉正常填充（回归）', () => {
    assert.equal($('moment-board').options.length, 4)
})

test('四项命令打开锚定浮层而不是模态 dialog，再点一次收起', () => {
    $('format-post').click()
    for (const name of ['spoiler', 'link', 'table', 'media']) {
        tool(name).click()
        assert.equal(panel(name).hidden, false, `${name} 未打开`)
        assert.equal(panel(name).tagName, 'DIV', `${name} 不应是原生 dialog`)
        assert.ok(panel(name).classList.contains('editor-popover'))
        assert.equal(tool(name).classList.contains('active'), true, `${name} 按钮应高亮`)
        tool(name).click()
        assert.equal(panel(name).hidden, true, `${name} 未收起`)
        assert.equal(tool(name).classList.contains('active'), false)
    }
    assert.equal($('editor-emoji-dialog'), null, '旧的居中表情弹窗应已移除')
    assert.equal(help().tagName, 'DIALOG', 'Markdown 帮助仍是模态 dialog')
})

test('切回动态精简模式时自动收起浮层，避免锚点消失后跳到视口左上角', () => {
    const host = document.createElement('div')
    document.body.append(host)
    const editor = window.DawnMomentEditor.create(host, {
        locale: 'zh', render: (text) => text
    })
    const link = [...document.querySelectorAll('[data-editor-panel="link"]')].at(-1)
    const toolbar = host.querySelector('.editor-toolbar')
    host.querySelector('[data-editor-command="link"]').click()
    assert.equal(link.hidden, false, '帖子模式下浮层可打开')
    editor.setOptions({ compact: true })
    assert.equal(toolbar.hidden, true, '动态模式应隐藏完整工具栏')
    assert.equal(link.hidden, true, '工具栏消失后浮层必须一起收起')
    assert.equal(host.querySelector('[data-editor-command="link"]').classList.contains('active'), false)
    editor.setOptions({ compact: false })
    assert.equal(toolbar.hidden, false)
    host.querySelector('[data-editor-command="link"]').click()
    assert.equal(link.hidden, false, '切回帖子后浮层仍可正常打开')
    editor.destroy()
    host.remove()
})

test('表情按钮弹出内置分组浮层，插入 Unicode 表情且可连续插入', () => {
    $('format-post').click()
    const button = tool('emoji')
    const picker = document.querySelector('[data-emoji-panel]')
    assert.ok(picker, '内核没有生成表情浮层')
    assert.ok($('moment-dialog').contains(picker), '官网弹窗内的浮层必须挂在 dialog 内')
    button.click()
    assert.equal(picker.hidden, false)
    assert.equal(picker.querySelectorAll('[data-emoji-group]').length, 9, '最近使用 + 8 个分组')
    picker.querySelector('[data-emoji-group="smileys"]').click()
    assert.ok(picker.querySelector('[data-emoji-group="smileys"]').classList.contains('active'))
    assert.match(picker.querySelector('[data-emoji-anchor="smileys"]').textContent, /笑脸与人像|Smileys & People/)
    assert.equal(picker.querySelectorAll('[data-emoji-section]').length, 9, '9 个分组在同一个连续列表里')
    const items = picker.querySelectorAll('.editor-emoji-item')
    assert.ok(items.length > 1900, `全部 1914 个表情应一次性渲染，实际 ${items.length}`)

    input().value = ''
    items[0].click()
    assert.equal(picker.hidden, false, '插入后面板保持打开，方便连续选择')
    assert.equal(input().value, '😀', '应插入表情字符而不是 :短码:')
    assert.equal(JSON.parse(window.localStorage.getItem('dawn.editor.recentEmojis'))[0], '😀')
    items[1].click()
    assert.equal(input().value, '😀😃')
    picker.querySelector('[data-emoji-group="recent"]').click()
    assert.match(picker.querySelector('[data-emoji-anchor="recent"]').textContent, /最近使用|Recently used/)
    assert.equal(picker.querySelectorAll('[data-emoji-section="recent"] .editor-emoji-item').length, 2)

    const dialogOpen = $('moment-dialog').open
    document.dispatchEvent(new window.KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }))
    assert.equal(picker.hidden, true, 'ESC 只关闭表情浮层')
    assert.equal($('moment-dialog').open, dialogOpen, 'ESC 不应改变发帖弹窗状态')
})

test('表情分组与占位文案跟随编辑器语言包', () => {
    const host = document.createElement('div')
    document.body.append(host)
    const editor = window.DawnMomentEditor.create(host, {
        locale: 'en', render: (text) => text
    })
    const picker = [...document.querySelectorAll('[data-emoji-panel]')].at(-1)
    host.querySelector('[data-editor-command="emoji"]').click()
    assert.equal(picker.hidden, false)
    picker.querySelector('[data-emoji-group="smileys"]').click()
    assert.equal(picker.querySelector('[data-emoji-anchor="smileys"]').textContent, 'Smileys & People')
    assert.equal(picker.querySelector('[data-emoji-group="recent"]').title, 'Recently used')
    assert.equal(picker.querySelector('[data-emoji-close]').title, 'Close emoji picker')
    editor.setOptions({ locale: 'zh' })
    assert.equal(picker.querySelector('[data-emoji-anchor="smileys"]').textContent, '笑脸与人像')
    assert.equal(picker.querySelector('[data-emoji-group="recent"]').title, '最近使用')
    editor.destroy()
    assert.equal(document.body.contains(picker), false, '销毁时应移除注入的浮层')
    host.remove()
})

test('表情连续滚动：tab 跟随锚点高亮，点 tab 直接跳到锚点', () => {
    const host = document.createElement('div')
    document.body.append(host)
    const editor = window.DawnMomentEditor.create(host, {
        locale: 'zh', render: (text) => text
    })
    const picker = [...document.querySelectorAll('[data-emoji-panel]')].at(-1)
    host.querySelector('[data-editor-command="emoji"]').click()
    const scroll = picker.querySelector('[data-emoji-scroll]')
    // jsdom 不做布局，手动模拟各组锚点的纵向位置
    const sections = [...picker.querySelectorAll('[data-emoji-section]')]
    sections.forEach((section, index) => Object.defineProperty(section, 'offsetTop', { value: index * 100, configurable: true }))

    picker.querySelector('[data-emoji-group="food"]').click()
    assert.equal(scroll.scrollTop, 300, '点 tab 应直接跳到该组锚点')
    assert.ok(picker.querySelector('[data-emoji-group="food"]').classList.contains('active'))

    scroll.scrollTop = 250
    scroll.dispatchEvent(new window.Event('scroll'))
    assert.equal(picker.querySelector('[data-emoji-group].active').dataset.emojiGroup, 'animals', '滚到动物与自然时 tab 应跟随')
    scroll.scrollTop = 0
    scroll.dispatchEvent(new window.Event('scroll'))
    assert.equal(picker.querySelector('[data-emoji-group].active').dataset.emojiGroup, 'recent')
    editor.destroy()
    host.remove()
})

test('面板按钮不用 type=submit，取消不会误触发宿主表单保存', () => {
    const host = document.createElement('div')
    const outer = document.createElement('form')
    outer.append(host)
    document.body.append(outer)
    let bubbled = 0
    outer.addEventListener('submit', (event) => { event.preventDefault(); bubbled++ })
    const editor = window.DawnMomentEditor.create(host, {
        locale: 'zh', render: (text) => text
    })
    const input = host.querySelector('[data-editor-input]')
    // 浮层由内核生成并挂到宿主层（后台是 body），不在编辑器 root 内。
    const link = [...document.querySelectorAll('[data-editor-panel="link"]')].at(-1)
    assert.ok(link, '内核没有生成链接浮层')
    // 嵌在宿主 form 内时内层 form 会被解析器丢弃，所以面板一律不依赖 form
    assert.equal(link.querySelector('form'), null)
    const cancel = [...link.querySelectorAll('[data-panel-action="cancel"]')].find((b) => b.textContent.includes('取消'))
    const insertBtn = link.querySelector('[data-panel-action="default"]')
    assert.equal(cancel.type, 'button')
    assert.equal(insertBtn.type, 'button')

    host.querySelector('[data-editor-command="link"]').click()
    assert.equal(link.hidden, false)
    cancel.click()
    assert.equal(bubbled, 0, '取消不应提交宿主表单（否则后台会误报修改成功）')
    assert.equal(link.hidden, true, '取消应关闭浮层')
    assert.equal(input.value, '', '取消不应插入内容')

    host.querySelector('[data-editor-command="link"]').click()
    link.querySelector('[data-link-text]').value = '官网'
    link.querySelector('[data-link-url]').value = 'https://example.com'
    link.querySelector('[data-link-url]').dispatchEvent(new window.Event('input', { bubbles: true }))
    insertBtn.click()
    assert.equal(bubbled, 0, '插入同样不能提交宿主表单')
    assert.equal(link.hidden, true)
    assert.match(input.value, /\[官网\]\(https:\/\/example\.com\)/)
    editor.destroy()
    assert.equal(document.body.contains(link), false, '销毁时应移除注入的浮层')
    outer.remove()
})

test('插入表格生成合法分隔行且预览渲染成 table', () => {
    $('format-post').click()
    tool('table').click()
    panel('table').querySelector('[data-table-grid]').querySelector('[data-row="2"][data-column="2"]').click()
    assert.match(input().value, /\|\s*---\s*\|\s*---\s*\|/)
    ed().querySelector('[data-editor-view="live"]').click()
    assert.ok(preview().querySelector('table'), '预览里没有 table')
})

test('四个下拉菜单点击后处于打开状态', () => {
    $('format-post').click()
    for (const key of ['heading', 'code', 'list', 'align']) {
        const button = ed().querySelector(`[data-editor-menu="${key}"]`)
        button.click()
        assert.ok(dialog.querySelector(`[data-editor-menu-panel="${key}"]`).classList.contains('open'), key)
    }
})

test('三视图、全屏、快捷键与草稿', () => {
    $('format-post').click()
    ed().querySelector('[data-editor-view="live"]').click()
    assert.ok(ed().classList.contains('view-live'))
    assert.equal(preview().hidden, false)
    ed().querySelector('[data-editor-view="preview"]').click()
    assert.equal(input().hidden, true)
    ed().querySelector('[data-editor-view="write"]').click()
    assert.equal(input().hidden, false)
    const fullBtn = editorNode('fullscreen')
    const enterIcon = fullBtn.innerHTML
    editorNode('fullscreen').click()
    assert.ok(ed().classList.contains('fullscreen'))
    const exitIcon = fullBtn.innerHTML
    assert.notEqual(exitIcon, enterIcon, '进入/退出全屏不能用同一个图标')
    assert.match(exitIcon, /<rect/, '退出全屏应有居中的小窗口框，便于一眼分辨')
    assert.doesNotMatch(enterIcon, /<rect/, '进入全屏不应带小窗口框')
    assert.match(fullBtn.title, /退出全屏|Exit fullscreen/)
    editorNode('fullscreen').click()
    assert.equal(fullBtn.innerHTML, enterIcon, '再次点击应还原为进入图标')
    assert.match(fullBtn.title, /进入全屏|Fullscreen/)

    input().value = '内容'
    input().setSelectionRange(0, 2)
    input().dispatchEvent(new window.Event('select', { bubbles: true }))
    input().dispatchEvent(new window.KeyboardEvent('keydown', { key: 'b', ctrlKey: true, bubbles: true, cancelable: true }))
    assert.match(input().value, /\*\*/)
    assert.ok(window.localStorage.getItem('dawn.moment.drafts'))
})

test('行号随正文行数同步', () => {
    $('format-post').click()
    input().value = '一\n二\n三'
    input().dispatchEvent(new window.Event('input', { bubbles: true }))
    assert.equal(editorNode('lines').textContent, '1\n2\n3')
})

test('表格网格默认未选中，悬停整体变色、移出复原、点击直接插入', () => {
    $('format-post').click()
    tool('table').click()
    const grid = panel('table').querySelector('[data-table-grid]')
    const cells = [...grid.querySelectorAll('.editor-table-cell')]
    assert.equal(cells.length, 48)
    assert.match(panel('table').querySelector('[data-table-hint]').textContent, /选择表格尺寸|Choose a size/)
    assert.equal(cells.filter((cell) => cell.classList.contains('selected')).length, 0, '原版默认没有预选 3×3')
    grid.querySelector('[data-row="3"][data-column="3"]').dispatchEvent(new window.MouseEvent('mouseover', { bubbles: true }))
    assert.match(panel('table').querySelector('[data-table-hint]').textContent, /3 行 3 列|3 × 3/)
    assert.equal(panel('table').querySelector('[data-table-size]').textContent, '3 × 3')
    const target = grid.querySelector('[data-row="5"][data-column="4"]')
    target.dispatchEvent(new window.MouseEvent('mouseover', { bubbles: true }))
    assert.match(panel('table').querySelector('[data-table-hint]').textContent, /5 行 4 列|5 × 4/)
    assert.equal(cells.filter((cell) => cell.classList.contains('selected')).length, 20)
    grid.dispatchEvent(new window.MouseEvent('mouseleave', { bubbles: false }))
    assert.equal(cells.filter((cell) => cell.classList.contains('selected')).length, 0, '移出网格应复原')
    assert.match(panel('table').querySelector('[data-table-hint]').textContent, /选择表格尺寸|Choose a size/)
    target.dispatchEvent(new window.MouseEvent('mouseover', { bubbles: true }))
    target.click()
    assert.equal(panel('table').hidden, true)
    assert.match(input().value, /\| --- \| --- \| --- \| --- \|/)
})

test('工具栏提示样式统一，H1 不再是深色大卡片', () => {
    const css = read('site.css'), md = read('moment-markdown.css')
    for (const source of [css, md]) {
        assert.ok(!source.includes('.editor-heading-tip'), '仍存在深色特例样式')
        assert.ok(!source.includes('.editor-tip-shortcuts'), '仍存在特例快捷键样式')
    }
    $('format-post').click()
    const heading = ed().querySelector('[data-editor-menu="heading"]')
    heading.dispatchEvent(new window.MouseEvent('mouseenter', { bubbles: false }))
    const tip = document.body.querySelector('.editor-tooltip')
    assert.ok(tip, '未生成提示气泡')
    assert.ok(!tip.classList.contains('editor-heading-tip'))
    assert.match(tip.innerHTML, /<strong>.*<\/strong><kbd>/)
    assert.ok($('moment-dialog').contains(tip), '气泡必须挂在弹窗内，否则被 top layer 遮挡')
    heading.dispatchEvent(new window.MouseEvent('mouseleave', { bubbles: false }))
})

test('实时预览左右分栏、预览视图单栏，下拉菜单不被裁剪', () => {
    const css = read('moment-markdown.css')
    assert.match(css, /\.compose-editor\.post \{[^}]*grid-template-columns: minmax\(0, 1fr\) minmax\(0, 1fr\)/)
    assert.match(css, /\.compose-editor-shell\.view-preview \.compose-editor\.post \{ display: block; \}/)
    assert.doesNotMatch(css.match(/\.compose-editor-shell \{[^}]*/)?.[0] || '', /overflow:\s*hidden/)
    assert.match(css, /\.editor-menu\.open \{ display: grid/)
    assert.match(css, /\.editor-toolbar \{[^}]*flex-wrap: nowrap[^}]*overflow-x: auto/)
    assert.ok(css.includes('@media (max-width: 1023px)'))
})

test('滚动归属：弹窗定高，滚动只在输入区与预览区内部', () => {
    const css = read('site.css')
    for (const rule of [
        'height: min(calc(100dvh - 32px), 760px)',
        '.compose-dialog > form { display: flex; flex: 1; flex-direction: column; min-width: 0; min-height: 0; }',
    ]) assert.ok(css.includes(rule), rule)
    assert.doesNotMatch(css, /\.editor-toolbar\s*\{/)
    assert.match(read('moment-markdown.css'), /\.markdown-preview \{ height: 100%; min-height: 0; overflow-y: auto;/)
})

test('高亮配色分亮暗两套，暗色不再是亮黄底配浅色字', () => {
    const css = read('site.css'), md = read('moment-markdown.css')
    assert.ok(!css.includes('#ffe58f'))
    assert.ok(md.includes('--md-mark-bg: #ffe58f'))
    assert.ok(md.includes('--md-mark-bg: #6a5212'))
    assert.ok(md.includes('--md-mark-fg: #ffe9a8'))
    assert.match(md, /\.markdown-body mark \{[^}]*var\(--md-mark-bg\)[^}]*var\(--md-mark-fg\)/)
})

test('遮罩用点阵覆盖，亮暗各一套灰点，不再用 currentColor 刷出白块', () => {
    const css = read('site.css'), md = read('moment-markdown.css')
    assert.ok(!css.includes('.moment-mask'))
    assert.ok(!md.includes('background: currentColor'))
    assert.ok(md.includes('--md-mask-dot: #a7b0c0'))
    assert.ok(md.includes('--md-mask-dot: #8b96ab'))
    assert.match(md, /\.markdown-body \.md-mask \{[^}]*radial-gradient[^}]*color: transparent/)
    assert.match(md, /\.md-mask\.revealed \{[^}]*var\(--md-fg\)/)
})

test('待办项覆盖官网全局 label/input 样式，复选框与文字同行', () => {
    const md = read('moment-markdown.css')
    assert.match(md, /\.markdown-body \.task-list-item label \{[^}]*display: inline[^}]*margin: 0[^}]*font-size: inherit/)
    assert.match(md, /\.markdown-body \.task-list-item input \{[^}]*padding: 0[^}]*background: none/)
})

test('未打开的 dialog 保持隐藏，移动端弹窗内收起客服悬浮按钮', () => {
    const css = read('site.css')
    assert.match(css, /dialog:not\(\[open\]\) \{ display: none; \}/)
    assert.match(css, /body:has\(dialog\[open\]\) \[data-dawn-support\] \{ display: none; \}/)
    assert.equal($('moment-dialog').open, false)
})

test('浮层点外部或 ESC 收起，ESC 不会连带关掉宿主发帖框', () => {
    $('format-post').click()
    if (!$('moment-dialog').open) $('moment-dialog').showModal()
    tool('link').click()
    assert.equal(panel('link').hidden, false)
    document.body.dispatchEvent(new window.MouseEvent('click', { bubbles: true }))
    assert.equal(panel('link').hidden, true, '点浮层外面应收起')
    tool('table').click()
    const dialogOpen = $('moment-dialog').open
    document.dispatchEvent(new window.KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }))
    assert.equal(panel('table').hidden, true, 'ESC 应收起浮层')
    assert.equal($('moment-dialog').open, dialogOpen, 'ESC 不应关闭发帖弹窗')
    $('moment-dialog').close()
})

test('链接按选区预填，媒体按地址推断类型，剧透选项点一下就插入', () => {
    $('format-post').click()
    input().value = 'https://example.com/page'
    input().setSelectionRange(0, input().value.length)
    input().dispatchEvent(new window.Event('select', { bubbles: true }))
    tool('link').click()
    assert.equal(panel('link').querySelector('[data-link-text]').value, '')
    assert.equal(panel('link').querySelector('[data-link-url]').value, 'https://example.com/page')
    assert.equal(panel('link').querySelector('[data-panel-action="default"]').disabled, false)

    tool('media').click()
    assert.equal(panel('media').querySelector('[data-panel-action="default"]').disabled, true)
    const mediaUrl = panel('media').querySelector('[data-media-url]')
    mediaUrl.value = 'https://cdn.example.com/clip.mp4'
    mediaUrl.dispatchEvent(new window.Event('input', { bubbles: true }))
    assert.match(panel('media').querySelector('[data-media-hint]').textContent, /视频|video/i)
    assert.equal(panel('media').querySelector('[data-panel-action="default"]').disabled, false)
    panel('media').querySelector('[data-panel-action="default"]').click()
    assert.match(input().value, /MEDIA::video::/)

    input().value = ''
    tool('spoiler').click()
    const spoiler = panel('spoiler')
    assert.equal(spoiler.querySelector('[data-panel-action]'), null, '原版剧透浮层没有 × / 取消')
    assert.equal(spoiler.querySelectorAll('[data-editor-command]').length, 2)
    spoiler.querySelector('[data-editor-command="spoiler-block"]').click()
    assert.match(input().value, /:::spoiler/)
    assert.equal(spoiler.hidden, true)
})

test('四项是锚定浮层，选区条跟随选区，按钮有激活态样式', () => {
    const css = read('site.css'), md = read('moment-markdown.css')
    for (const source of [md]) {
        assert.match(source, /\.editor-panel \{ position: fixed/)
        assert.match(source, /\.editor-selection-bar \{ position: fixed/)
        assert.match(source, /\.editor-tool\.active/)
        assert.doesNotMatch(source, /\.editor-table-badge/)
    }
    assert.match(md, /\.editor-popover-head \{/)
})

// ── 图片 provider、秒传、自定义按钮 ──────────────────────────────────────
function isolatedEditor(options) {
    const host = document.createElement('div')
    document.body.append(host)
    const editor = window.DawnMomentEditor.create(host, { render: (text) => text, ...options })
    return { host, editor, input: host.querySelector('textarea') }
}
const settle = () => new Promise(resolve => setTimeout(resolve, 40))
function dropImage(input, file) {
    const event = new window.Event('drop', { bubbles: true, cancelable: true })
    Object.defineProperty(event, 'dataTransfer', { value: { files: [file], items: [] } })
    input.dispatchEvent(event)
}

test('无图片 provider 时隐藏插图入口', () => {
    const { host, editor } = isolatedEditor({})
    assert.equal(host.querySelector('[data-editor-command="image"]').hidden, true)
    assert.equal(host.querySelector('.editor-attachments'), null)
    editor.destroy()
    host.remove()
})

test('秒传命中插入标准 Markdown，图库状态由使用方维护', async () => {
    let uploads = 0
    const { host, editor, input } = isolatedEditor({
        imageUploader: {
            check: async () => [{ id: 42, url: '/img/42.png', name: 'shot.png' }],
            upload: async () => { uploads++; return [] }
        }
    })
    dropImage(input, new window.File(['x'], 'shot.png', { type: 'image/png' }))
    await settle()
    assert.equal(uploads, 0)
    assert.match(editor.getValue(), /!\[shot\.png\]\(<\/img\/42\.png>\)/)
    assert.equal(host.querySelector('.editor-attachments'), null)
    editor.destroy()
    host.remove()
})

test('粘贴图片可独立启用，业务可以定义正文标记', async () => {
    let uploads = 0
    const { host, editor, input } = isolatedEditor({
        config: () => ({ hidden: ['image'] }),
        imageUploader: {
            upload: async () => { uploads++; return [{ id: 9, url: '/img/9.png' }] },
            toMarkdown: image => `[attachment://${image.id}]`
        }
    })
    const paste = () => {
        const file = new window.File(['x'], '', { type: 'image/png' })
        const event = new window.Event('paste', { bubbles: true, cancelable: true })
        Object.defineProperty(event, 'clipboardData', { value: { files: [], items: [{ kind: 'file', type: 'image/png', getAsFile: () => file }] } })
        input.dispatchEvent(event)
        return event.defaultPrevented
    }
    assert.equal(paste(), false)
    assert.equal(uploads, 0)
    editor.setOptions({ pasteImages: true })
    assert.equal(paste(), true)
    await settle()
    assert.equal(uploads, 1)
    assert.equal(editor.getValue(), '[attachment://9]')
    editor.destroy()
    host.remove()
})

test('自定义工具包裹选区，渲染器接收原始业务标记', () => {
    const { host, editor, input } = isolatedEditor({
        value: '正文',
        render: text => text.replace(/\[login\](.*?)\[\/login\]/, '<b>$1</b>'),
        customTools: [{ key: 'login', title: '登录可见', onClick: ({ editor }) =>
            editor.insertText(selected => `[login]${selected}[/login]`) }]
    })
    input.setSelectionRange(0, 2)
    host.querySelector('[data-editor-tool="login"]').click()
    assert.equal(editor.getValue(), '[login]正文[/login]')
    host.querySelector('[data-editor-view="preview"]').click()
    assert.equal(host.querySelector('[data-editor-preview] b').textContent, '正文')
    editor.setOptions({ config: () => ({ hidden: ['login'] }) })
    assert.equal(host.querySelector('[data-editor-tool="login"]'), null)
    editor.destroy()
    host.remove()
})

test('官网动态格式隐藏媒体按钮、字数 1024；帖子格式恢复', () => {
    $('format-moment').click()
    assert.equal(tool('image').hidden, true)
    assert.equal(tool('media').hidden, true)
    assert.equal(editorNode('input').maxLength, 1024)
    $('format-post').click()
    assert.equal(tool('image').hidden, false)
    assert.equal(editorNode('input').maxLength, 10000)
})

test('多个编辑器实例的 Markdown 帮助弹窗互不串台（动态模板不再生成重复 id）', () => {
    const first = isolatedEditor({})
    const second = isolatedEditor({})
    const help1 = first.host.querySelector('[data-editor-help]')
    const help2 = second.host.querySelector('[data-editor-help]')
    assert.ok(help1 && help2, '每个实例都应有自己的帮助弹窗')
    assert.notEqual(help1, help2)
    assert.equal(!!help1.id || !!help2.id, false)

    first.host.querySelector('[data-editor-command="help"]').click()
    assert.equal(help1.open, true, '第一个实例应打开自己的帮助')
    assert.equal(help2.open, false)
    assert.match(help1.querySelector('[data-help-content]').innerHTML, /Markdown/)
    help1.querySelector('[data-help-close]').click()
    assert.equal(help1.open, false, '关闭按钮应作用于自己的弹窗')

    second.host.querySelector('[data-editor-command="help"]').click()
    assert.equal(help2.open, true)
    assert.equal(help1.open, false)
    first.editor.destroy(); second.editor.destroy()
    first.host.remove(); second.host.remove()

    // 官网主编辑器接管的是静态那份，仍然能正常打开
    tool('help').click()
    assert.equal(help().open, true)
    help().close()
})

test('详情页仅展示一次正文图片，编辑仍保留图库 ID 和文件原名', async () => {
    const row = { id: 14, format: 'post', title: '图文', content: '[attachment://3]', images: '[3,8]',
        image_names: { 3: '原图.png' }, can_manage: true, comments: 0, created_at: 1000 };
    const fetch = window.fetch
    let sent
    window.fetch = async (path, options) => {
        if (String(path).startsWith('/api/moments/14/comments')) {
            return { ok: true, json: async () => ({ data: [], meta: { page_size: 10, total: 0 } }) }
        }
        if (path === '/api/moments/14') {
            if (options.method === 'PUT') sent = JSON.parse(options.body)
            return { ok: true, json: async () => ({ data: row }) }
        }
        return fetch(path, options)
    }
    try {
        window.history.pushState(null, '', '/moments/14')
        window.dispatchEvent(new window.PopStateEvent('popstate'))
        await settle()
        const detail = document.querySelector('.moment-detail')
        assert.equal(detail.querySelector('.moment-content img')?.alt, '原图.png')
        assert.deepEqual(Array.from(detail.querySelectorAll('.moment-images img'), img => new URL(img.src).pathname), ['/api/moments/14/images/8'])
        Array.from(detail.querySelectorAll('button')).find(button => button.textContent === '编辑').click()
        await settle()
        assert.equal($('moment-dialog-title').textContent, '编辑动态')
        assert.equal($('moment-submit').textContent, '保存修改')
        assert.equal($('moment-image-preview').querySelectorAll('img').length, 2)
        input().value = '移除正文图'
        $('moment-form').dispatchEvent(new window.Event('submit', { bubbles: true, cancelable: true }))
        await settle()
        assert.deepEqual(sent.images, [3, 8])
        assert.equal(sent.content, '移除正文图')
    } finally {
        window.fetch = fetch
    }
})
