import { mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, posix, resolve } from 'node:path'
import { slug } from 'github-slugger'
import { defineConfig, type DefaultTheme } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

const localSite = process.env.DAWN_DOCS_SITE === '1'
const base = localSite ? '/docs/' : '/dawn-simple/'

// Both languages share the same route suffix, so VitePress switches the current article.
function rewrite(path: string) {
  return path
    .replace(/^prismgo\/README_en\.md$/, 'en/prismgo/index.md')
    .replace(/^prismgo\/README\.md$/, 'prismgo/index.md')
    .replace(/^prismgo\/en\//, 'en/prismgo/')
    .replace(/^prismgo\/zh_CN\//, 'prismgo/')
}

// README是唯一目录来源，官网与GitHub Pages共用同一份导航。
function sidebar(file: string): DefaultTheme.SidebarItem[] {
  const items: DefaultTheme.SidebarItem[] = []
  const index = readFileSync(new URL(`../${file}`, import.meta.url), 'utf8')
  for (const [, section, label, link] of index.matchAll(
    /^## (.+)$|^(?:-|\|) \[([^\]]+)\]\(([^)]+\.md)\)/gm
  )) {
    if (section) items.push({ text: section, items: [] })
    else items.at(-1)!.items!.push({ text: label, link: `/${rewrite(posix.join(posix.dirname(file), link)).slice(0, -3)}` })
  }
  return items
}

function navigation(english = false): DefaultTheme.Config {
  const prefix = english ? '/en' : ''
  const text = (zh: string, en: string) => english ? en : zh
  const page = (zh: string, en: string, path: string) => ({ text: text(zh, en), link: `${prefix}/${path}` })
  return {
    siteTitle: text('Dawn 文档', 'Dawn Docs'),
    nav: [page('开发文档', 'Documentation', 'quickstart'), page('PrismGo', 'PrismGo', 'prismgo/'), {
      text: text('最佳实践', 'Example application'), link: 'https://github.com/dawnim/dawn-simple'
    }],
    sidebar: {
      [`${prefix}/prismgo/`]: [page('PrismGo 文档', 'PrismGo Docs', 'prismgo/'), ...sidebar(`prismgo/${english ? 'README_en' : 'README'}.md`), page('Dawn 文档', 'Dawn Docs', 'README')],
      [`${prefix}/`]: [
        page('文档导航', 'Documentation', 'README'),
        ...sidebar(`${english ? 'en/' : ''}README.md`)
      ]
    },
    editLink: {
      text: text('查看源文件', 'View source'),
      pattern: ({ filePath }) => filePath.startsWith('prismgo/')
        ? `https://github.com/prismgo/docs/blob/main/${filePath.slice(8)}`
        : `https://github.com/dawnim/dawn-simple/blob/main/docs/${filePath}`
    },
    outline: { level: [2, 3], label: text('本页目录', 'On this page') },
    docFooter: { prev: text('上一篇', 'Previous page'), next: text('下一篇', 'Next page') },
    sidebarMenuLabel: text('菜单', 'Menu'), returnToTopLabel: text('返回顶部', 'Return to top'), skipToContentLabel: text('跳转至正文', 'Skip to content'),
    darkModeSwitchLabel: text('切换主题', 'Appearance'), lightModeSwitchTitle: text('切换至亮色', 'Switch to light theme'), darkModeSwitchTitle: text('切换至暗色', 'Switch to dark theme'),
    langMenuLabel: text('切换语言', 'Change language'),
    footer: {
      message: `<a href="https://github.com/dawnim/dawn-simple/blob/main/LICENSE">MIT License</a> · <a href="https://github.com/dawnim/dawn-simple#鸣谢">${text('鸣谢', 'Acknowledgements')}</a>`,
      copyright: 'Copyright © 2026 dawnim'
    }
  }
}

export default withMermaid(
  defineConfig({
    lang: 'zh-CN',
    title: 'Dawn',
    description: '像写 Laravel 一样，轻松构建 Go 管理系统',
    base,
    rewrites: rewrite,
    locales: {
      root: { label: '简体中文', lang: 'zh-CN', themeConfig: navigation() },
      en: { label: 'English', lang: 'en', description: 'Build Go admin applications with the simplicity of Laravel', themeConfig: navigation(true) }
    },
    // Keep published PrismGo links working without changing its submodule sources.
    buildEnd(config) {
      for (const [source, target] of Object.entries(config.rewrites.map)) {
        const path = resolve(config.outDir, source.replace(/\.md$/, '.html'))
        const href = `${base}${target.replace(/(?:index)?\.md$/, target.endsWith('index.md') ? '' : '.html')}`
        mkdirSync(dirname(path), { recursive: true })
        writeFileSync(path, `<!doctype html><meta charset="utf-8"><title>Dawn</title><a href="${href}">Continue</a><script>location.replace(${JSON.stringify(href)}+location.search+location.hash)</script>`)
      }
    },
    head: [
      [
        'link',
        {
          rel: 'icon',
          type: 'image/svg+xml',
          href: `${base}logo.svg`
        }
      ]
    ],
    mermaidPlugin: { class: 'docs-mermaid' },
    mermaid: {
      flowchart: { useMaxWidth: false },
      sequence: { useMaxWidth: false }
    },
    markdown: {
      anchor: { slugify: slug },
      toc: { slugify: slug },
      languageAlias: { env: 'dotenv' },
      config(md) {
        const fence = md.renderer.rules.fence!
        md.renderer.rules.fence = (...args) => {
          const html = fence(...args)
          return args[0][args[1]].info.trim() === 'mermaid' ? `<Diagram>${html}</Diagram>` : html
        }
        md.core.ruler.after('inline', 'prismgo-links', (state) => {
          const source = posix.relative(state.env.srcDir || new URL('..', import.meta.url).pathname, state.env.realPath || state.env.path)
          if (!source.startsWith('prismgo/')) return
          const english = state.env.relativePath.startsWith('en/')
          for (const token of state.tokens.flatMap((token) => token.children || [])) {
            if (token.type !== 'link_open') continue
            const href = decodeURI(token.attrGet('href')!)
            if (/^(?:[a-z]+:|#|\/\/)/i.test(href)) continue
            const path = href.startsWith('/docs/')
              ? `prismgo/${english ? 'en' : 'zh_CN'}/${href.replace(/^\/docs\/(?:(?:\{\{version\}\}|zh_CN|en)\/)?/, '').replace(/^providers(?=[.#/]|$)/, 'service-provider')}`
              : posix.resolve('/', posix.dirname(source), href).slice(1)
            token.attrSet('href', `/${rewrite(path)}`)
          }
        })
      }
    },
    themeConfig: {
      logo: '/logo.svg',
      ...(localSite ? { logoLink: { link: '/', target: '_self' } } : {}),
      socialLinks: [
        {
          icon: 'github',
          link: 'https://github.com/dawnim/dawn-simple'
        }
      ],
      search: {
        provider: 'local',
        options: {
          miniSearch: {
            options: {
              tokenize(text) {
                return Array.from(
                  new Intl.Segmenter('zh-CN', {
                    granularity: 'word'
                  }).segment(text)
                )
                  .filter((part) => part.isWordLike)
                  .map((part) => part.segment)
              }
            }
          },
          locales: {
            root: {
              translations: {
                button: {
                  buttonText: '搜索文档',
                  buttonAriaLabel: '搜索文档'
                },
                modal: {
                  displayDetails: '显示详情',
                  resetButtonTitle: '清除搜索',
                  backButtonTitle: '关闭搜索',
                  noResultsText: '没有找到相关内容',
                  footer: { selectText: '选择', navigateText: '切换', closeText: '关闭' }
                }
              }
            }
          }
        }
      }
    }
  })
)
