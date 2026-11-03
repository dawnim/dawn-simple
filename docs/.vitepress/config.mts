import { readFileSync } from 'node:fs'
import { slug } from 'github-slugger'
import { defineConfig, type DefaultTheme } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

const prismgoSidebar: DefaultTheme.SidebarItem[] = []
const prismgoIndex = readFileSync(new URL('../prismgo/README.md', import.meta.url), 'utf8')
for (const [, section, text, link] of prismgoIndex.matchAll(
  /^## (.+)$|^- \[([^\]]+)\]\((zh_CN\/[^)]+\.md)\)$/gm
)) {
  if (section) prismgoSidebar.push({ text: section, items: [] })
  else prismgoSidebar.at(-1)!.items!.push({ text, link: `/prismgo/${link.slice(0, -3)}` })
}

export default withMermaid(
  defineConfig({
    lang: 'zh-CN',
    title: 'Dawn',
    description: '像写 Laravel 一样，轻松构建 Go 管理系统',
    base: '/dawn-simple/',
    head: [
      [
        'link',
        {
          rel: 'icon',
          type: 'image/svg+xml',
          href: '/dawn-simple/logo.svg'
        }
      ]
    ],
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
        md.core.ruler.after('inline', 'prismgo-links', state => {
          if (!state.env.relativePath.startsWith('prismgo/')) return
          const locale = state.env.relativePath.startsWith('prismgo/en/') ? 'en' : 'zh_CN'
          for (const token of state.tokens.flatMap(token => token.children || [])) {
            if (token.type !== 'link_open') continue
            const href = decodeURI(token.attrGet('href')!)
            if (!href.startsWith('/docs/')) continue
            const path = href
              .replace(/^\/docs\/(?:(?:\{\{version\}\}|zh_CN|en)\/)?/, '')
              .replace(/^providers(?=[.#/]|$)/, 'service-provider')
            token.attrSet('href', `/prismgo/${locale}/${path}`)
          }
        })
      }
    },
    themeConfig: {
      logo: '/logo.svg',
      siteTitle: 'Dawn 文档',
      nav: [
        { text: '开发文档', link: '/quickstart' },
        { text: 'PrismGo', link: '/prismgo/README' },
        {
          text: '最佳实践',
          link: 'https://github.com/dawnim/dawn-simple'
        }
      ],
      sidebar: {
        '/prismgo/': [
          { text: 'PrismGo 文档', link: '/prismgo/README' },
          ...prismgoSidebar,
          { text: 'Dawn 文档', link: '/README' }
        ],
        '/': [
          { text: '文档导航', link: '/README' },
          {
            text: '入门',
            items: [
              { text: '快速开始', link: '/quickstart' },
              { text: '创建空白项目', link: '/create-project' },
              { text: '架构与目录', link: '/architecture' }
            ]
          },
          {
            text: '业务开发',
            items: [
              { text: '后端开发', link: '/backend' },
              { text: '前端开发', link: '/frontend' },
              { text: '权限与数据范围', link: '/authorization' },
              { text: '配置', link: '/configuration' },
              { text: '聊天与实时通信', link: '/chat' }
            ]
          },
          {
            text: '运行与交付',
            items: [
              { text: '服务运行', link: '/runtime' },
              { text: '测试与发布', link: '/testing' },
              { text: '开发工具', link: '/lens' }
            ]
          }
        ]
      },
      editLink: {
        text: '查看源文件',
        pattern: ({ filePath }) =>
          filePath.startsWith('prismgo/')
            ? `https://github.com/prismgo/docs/blob/main/${filePath.slice(8)}`
            : `https://github.com/dawnim/dawn-simple/blob/main/docs/${filePath}`
      },
      outline: { level: [2, 3], label: '本页目录' },
      docFooter: { prev: '上一篇', next: '下一篇' },
      sidebarMenuLabel: '菜单',
      returnToTopLabel: '返回顶部',
      skipToContentLabel: '跳转至正文',
      darkModeSwitchLabel: '切换主题',
      lightModeSwitchTitle: '切换至亮色',
      darkModeSwitchTitle: '切换至暗色',
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
                  .filter(part => part.isWordLike)
                  .map(part => part.segment)
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
                  footer: {
                    selectText: '选择',
                    navigateText: '切换',
                    closeText: '关闭'
                  }
                }
              }
            }
          }
        }
      },
      footer: {
        message:
          '<a href="https://github.com/dawnim/dawn-simple/blob/main/LICENSE">MIT License</a> · <a href="https://github.com/dawnim/dawn-simple#鸣谢">鸣谢</a>',
        copyright: 'Copyright © 2026 dawnim'
      }
    }
  })
)
