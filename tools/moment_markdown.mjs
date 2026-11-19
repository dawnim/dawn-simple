import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js/lib/common'
import DOMPurify from 'dompurify'
import abbr from 'markdown-it-abbr'
import container from 'markdown-it-container'
import deflist from 'markdown-it-deflist'
import footnote from 'markdown-it-footnote'
import ins from 'markdown-it-ins'
import mark from 'markdown-it-mark'
import sub from 'markdown-it-sub'
import sup from 'markdown-it-sup'
import taskLists from 'markdown-it-task-lists'
import katex from '@traptitech/markdown-it-katex'
import { full as emoji } from 'markdown-it-emoji'

const EMBED_HOSTS = new Set([
    'player.bilibili.com', 'www.bilibili.com', 'music.163.com',
    'www.youtube.com', 'youtu.be', 'v.qq.com', 'player.youku.com',
])
const CALLOUTS = ['info', 'tip', 'warning', 'danger', 'success']
const esc = MarkdownIt().utils.escapeHtml

const md = new MarkdownIt({
    html: true,
    linkify: true,
    typographer: true,
    breaks: true,
    highlight(code, lang) {
        const name = (lang || '').toLowerCase().split(/\s+/)[0]
        if (name === 'mermaid') return `<pre class="md-mermaid">${esc(code)}</pre>`
        const body = name && hljs.getLanguage(name)
            ? hljs.highlight(code, { language: name, ignoreIllegals: true }).value
            : esc(code)
        const label = esc(name || 'text')
        return `<pre class="md-code-block"><div class="md-code-header"><span>${label}</span></div><code class="language-${label}">${body}</code></pre>`
    },
})

md.linkify.set({ fuzzyLink: false })
for (const plugin of [abbr, deflist, footnote, ins, mark, sub, sup, katex, emoji]) md.use(plugin)
md.use(taskLists, { enabled: true, label: true, labelAfter: true })

for (const type of CALLOUTS) {
    md.use(container, type, {
        render(tokens, index) {
            const title = esc(tokens[index].info.slice(type.length).trim() || type)
            return tokens[index].nesting === 1
                ? `<div class="md-callout md-callout-${type}"><div class="md-callout-title">${title}</div><div class="md-callout-body">`
                : '</div></div>'
        },
    })
}
md.use(container, 'spoiler', {
    render(tokens, index) {
        const title = esc(tokens[index].info.slice(7).trim() || '剧透内容')
        return tokens[index].nesting === 1
            ? `<details class="md-spoiler"><summary>${title}<span class="md-spoiler-hint">点击展开</span></summary><div class="md-spoiler-body">`
            : '</div></details>'
    },
})

const textRule = md.renderer.rules.text
md.renderer.rules.text = (tokens, index, options, env, self) => {
    let html = textRule(tokens, index, options, env, self)
    html = html.replace(/:([a-z0-9_+-]+):/gi, (raw, name) => {
        const value = env.emojis?.[name.toLowerCase()]
        return value ? `<span class="md-emoji">${esc(value)}</span>` : raw
    })
    html = html.replace(/(^|[^\w@])@([\w\u4e00-\u9fa5-]{1,30})/g,
        (_all, lead, name) => `${lead}<a class="md-mention" href="/?author=${encodeURIComponent(name)}">@${name}</a>`)
    return html.replace(/!!([^!\n]+)!!/g, '<span class="md-mask" data-md-mask>$1</span>')
}

md.renderer.rules.heading_open = (tokens, index, options, env, self) => {
    const base = (tokens[index + 1]?.content || '').toLowerCase()
        .replace(/[^\w\u4e00-\u9fa5]+/g, '-').replace(/^-|-$/g, '') || `h${index}`
    const used = (env.slugs ||= new Map())
    const seen = used.get(base) || 0
    used.set(base, seen + 1)
    tokens[index].attrSet('id', seen ? `${base}-${seen}` : base)
    return self.renderToken(tokens, index, options)
}

md.renderer.rules.image = (tokens, index, options, _env, self) => {
    const src = tokens[index].attrGet('src') || ''
    if (!/^(\/|blob:|https?:\/\/)/i.test(src)) return esc(tokens[index].attrGet('alt') || src)
    tokens[index].attrSet('loading', 'lazy')
    tokens[index].attrSet('class', 'md-image')
    return self.renderToken(tokens, index, options)
}

const linkRule = md.renderer.rules.link_open || ((tokens, index, options, _env, self) => self.renderToken(tokens, index, options))
md.renderer.rules.link_open = (tokens, index, options, env, self) => {
    if (/^https?:\/\//i.test(tokens[index].attrGet('href') || '')) {
        tokens[index].attrSet('target', '_blank')
        tokens[index].attrSet('rel', 'noopener noreferrer')
        tokens[index].attrJoin('class', 'md-link-external')
    }
    return linkRule(tokens, index, options, env, self)
}

DOMPurify.addHook('uponSanitizeAttribute', (node, data) => {
    if (node.nodeName === 'IFRAME' && data.attrName === 'src') {
        const ok = URL.canParse(data.attrValue) && EMBED_HOSTS.has(new URL(data.attrValue).hostname)
        if (!ok) data.keepAttr = false
    }
})

function preprocess(source) {
    return source.replace(/MEDIA::(audio|video|iframe)::(https?:\/\/[^\s<>]+)/gi, (_all, type, url) => {
        if (!URL.canParse(url)) return esc(url)
        if (type.toLowerCase() === 'iframe' && !EMBED_HOSTS.has(new URL(url).hostname)) return esc(url)
        const safe = esc(url)
        return type.toLowerCase() === 'iframe'
            ? `<iframe loading="lazy" src="${safe}" title="media" sandbox="allow-scripts allow-same-origin allow-presentation"></iframe>`
            : `<${type.toLowerCase()} controls preload="metadata" src="${safe}"></${type.toLowerCase()}>`
    })
}

/** 帖子 Markdown -> 安全 HTML。emojis 为后台配置的短码表。 */
export function renderMarkdown(source, emojis = {}) {
    const html = md.render(preprocess(source || ''), { emojis, slugs: new Map() })
    return DOMPurify.sanitize(html, {
        ADD_TAGS: ['center', 'ruby', 'rt', 'rp', 'iframe', 'details', 'summary', 'input'],
        ADD_ATTR: ['align', 'target', 'loading', 'checked', 'disabled', 'sandbox', 'allowfullscreen', 'data-md-mask'],
        ALLOWED_URI_REGEXP: /^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix|blob):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i,
        FORBID_TAGS: ['script', 'style', 'form', 'object', 'embed'],
        FORBID_ATTR: ['style', 'onerror', 'onload'],
    })
}

export { EMBED_HOSTS }
