# 前端开发

业务前端位于 `admin/src`，通过 `@dawnim/admin` 使用布局、系统页面、请求、权限与公共组件。

## 注册页面

在 `admin/src/main.ts` 的 `createAdmin` 配置中追加页面：

```ts
import { createAdmin } from '@dawnim/admin'
import '@dawnim/admin/style.css'

createAdmin({
  apiBase: '/admin',
  pages: {
    '/records': () => import('./views/records/index.vue')
  }
}).mount('#app')
```

后端菜单的 `Component` 与注册路径一致。业务页面优先于内置页面解析，按访问需要加载。菜单与权限注册见[后端开发](backend.md#调整生成的crud)。

## 业务列表

先通过[代码生成器或CLI](backend.md#生成业务代码)生成业务模块，以下示例说明如何调整生成的资源API与字段：

```ts
import { createResourceApi, type Column } from '@dawnim/admin'

export default createResourceApi('/records')

export const columns: Column[] = [
  { prop: 'name', label: '名称', required: true, maxlength: 100 }
]
```

在 `admin/src/views/records/index.vue` 使用标准列表：

```vue
<template>
  <CrudPage permission="app.records" :api="api" :columns="columns" :search="['name']" />
</template>

<script setup lang="ts">
  import { CrudPage } from '@dawnim/admin'
  import api, { columns } from '../../api/records'
</script>
```

`CrudPage` 提供搜索、列设置、分页、编辑、删除与 CSV 导入导出。复杂页面可使用 `ArtTable`、`ArtTableHeader`、`DawnSearchBar`、`DawnButton`、`useTable`、`useAdmin` 和 `useAuth` 自行组合。

代码生成器的引用模式复用 `CrudPage`；源码模式输出列表、搜索栏和编辑弹窗，适合深度定制。两种模式的字段、权限与交互一致。

`Column.dict`复用框架字典：列表显示淡色标签，搜索使用`DawnSelect`，编辑使用`DawnRadio`；字符串字典设置`dictValueType: 'string'`。固定选项可用`options`。字段可设置`readonly`、`immutable`、`nullable`、`required`和`maxlength`；搜索超过3项默认折叠其余项，展开后保留输入值。引用与源码模式保持同样行为。

```ts
{ prop: 'status', label: t('状态'), type: 'number', dict: 'data_status' }
```

### 列表密度

表格工具栏提供全局共享、浏览器持久保存的三档密度，`ArtTable`的`size`属性可单独覆盖。分页列表、树表与普通列表使用同一套尺寸：

| 密度 | `size` | 基础行高 | 正文字号 | 操作按钮、头像和图片 |
|---|---|---|---|---|
| 紧凑 | `small` | 32px | 12px | 24px |
| 默认 | `default` | 36px | 13px | 28px |
| 宽松 | `large` | 40px | 14px | 32px |

标准列表默认每页10条。收起额外筛选项时，默认密度按1024×768视口容纳10条设计，宽松密度按14寸MacBook Pro常用视口容纳10条设计。搜索栏和表格卡片保留16px水平留白；分页独立显示。普通长文本省略并支持悬浮查看，列配置可用`showOverflowTooltip: false`关闭；多行插槽、展开行和额外筛选项仍可增加高度。窄屏通过列的`minWidth`保持可读性，并在表格内部横向滚动。

## 请求与权限

普通业务请求使用公共 `request`：

```ts
import { request } from '@dawnim/admin'

const record = await request.get<{ id: number; name: string }>({ url: '/records/1' })
await request.patch({ url: `/records/${record.id}`, data: { name: '新名称' } })
```

资源 API 采用 `list/get/create/update/replace/remove/removeMany`。`update` 发送 PATCH，`replace` 发送 PUT。自定义 `list(query, config?)` 时将第二个参数透传到 `request`，表格据此取消过期请求。分页与响应格式见[接口约定](backend.md#接口约定)。

```vue
<ElButton v-permission="'app.records.create'">新增</ElButton>
```

权限指令控制界面显示，操作授权由接口校验。请求默认携带 Bearer Token，并合并并发刷新；跨域请求默认 `withCredentials=false`，同源 iframe 和下载可使用登录 Cookie。跨域策略见[配置](configuration.md#跨域与代理)。

## 公共接口

| 接口 | 用途 |
|---|---|
| `createAdmin` | 注册布局、路由、主题与权限指令 |
| `CrudPage`、`createResourceApi` | 标准业务列表与资源请求 |
| `DawnSelect`、`DawnRadio`、`DawnDict` | 字典下拉、单选与标签 |
| `Upload` | 文件上传、分片、摘要预检与秒传 |
| `request` | 认证、刷新、错误处理与下载 |
| `realtime`、`openChat` | 共享实时连接与管理员聊天 |
| `StatisticsDashboard`、`WorkDashboard` | 默认统计页与员工工作台 |
| `ArtBarChart`、`ArtLineChart` | 随主题变化的业务图表 |
| `ArtCountTo`、`ArtSvgIcon` | 数字展示与图标 |

组件与类型从 `@dawnim/admin` 导入。实时接口见[聊天与实时通信](chat.md#前端调用)。

## 工作台与顶栏

`pages` 可分别覆盖两种工作台：

```ts
pages: {
  '/dashboard/statistics': () => import('./views/dashboard/index.vue'),
  '/dashboard/work': () => import('./views/dashboard/work.vue')
}
```

管理员管理中的“设置首页”决定工作台类型。默认组件可直接复用，业务统计从自己的 API 获取。dawn-simple 的图表与页面位于 `admin/src/views/dashboard`，请求位于 `admin/src/api/dashboard.ts`。

顶栏通过组件扩展：

```ts
import HeaderTools from './components/HeaderTools.vue'

// createAdmin 配置
header: { before: HeaderTools }
```

`before` 放在搜索前，`after` 放在头像后，`right` 替换整个右侧工具区。组件接收当前用户 `user`、程序时区 `timezone` 和时间差 `serverTimeOffset`，可使用 Router、i18n 与公共请求。定时器与监听器随组件卸载释放。

## 主题与语言

亮暗模式和主题色使用全局设置。默认 Logo、浏览器图标随主题变化；自定义品牌通过 `logo`、`favicon` 传入图片地址。后台保存的站点名称与 Logo 会更新登录页和布局，`title`、`logo` 用作启动初始值。

后台默认跟随系统主题，使用垂直菜单、第一种菜单风格、第三种主题色、边框盒子、铺满容器、默认标签页、200px菜单宽度、淡入淡出动画和0.75圆角。基础显示项与手风琴模式默认开启，色弱模式与水印关闭。已有浏览器保留个人设置，可在主题设置中点击“重置配置”应用默认值；暗色显示不会覆盖所选菜单风格。

固定文案直接使用中文原文 key：

```ts
const text = t('欢迎{username}', { username: 'Dawn' })
```

中文回退原文，英文通过 `createAdmin.messages` 注册：

```ts
messages: {
  en: { '欢迎{username}': 'Welcome, {username}' }
}
```

zh 语言包只维护 `fields` 等技术标识的显示映射。生成页面按模块加载 zh/en 字段包，接口字段名保持 `snake_case`，动态内容通过命名占位符传入。

## 独立前端接入

已有 Vue 应用可直接安装组件库。Node 版本要求见[快速开始](quickstart.md#环境)。

```sh
pnpm add @dawnim/admin vue vue-router pinia element-plus vue-i18n
pnpm add -D vite@^7 @vitejs/plugin-vue@^6 typescript@~5.6 vue-tsc@~2.1
```

入口使用本页的 `createAdmin` 示例。HTML 提供 `#app` 节点，并加载 `src/main.ts`；`package.json` 设置 `type: module`，`dev` 使用 `vite`，`build` 使用 `vue-tsc --noEmit && vite build`。

Vite 配置共享运行时依赖并代理后端：

```ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  resolve: { dedupe: ['vue', 'vue-router', 'pinia', 'element-plus', 'vue-i18n'] },
  server: {
    proxy: {
      '/admin': 'http://127.0.0.1:5123',
      '/storage': 'http://127.0.0.1:5123',
      '/ws': { target: 'ws://127.0.0.1:8851', ws: true, rewrite: () => '/' }
    }
  }
})
```

公开附件的`/storage/`也需代理到后端；跨域接入时，相对图片地址按`apiBase`的来源解析，CDN绝对地址保持原样。私密图片由公共图片组件携带Bearer读取。

TypeScript 使用 `moduleResolution: Bundler` 并包含 `vite/client` 类型。完整配置可参考 [admin 目录](https://github.com/dawnim/dawn-simple/tree/main/admin)。生产构建与代理见[服务运行](runtime.md#生产部署)。

## 示例官网与同站文档

后端HTTP服务的`/`是网友动态首页，提供发帖、图片、评论、鲜花/鸡蛋互动、AI快捷创作和聊天；`/dawn`保留Dawn介绍页、登录注册及在线客服。后台前端使用自己的前端服务或域名。官网源码在`app/web/assets`，随Go二进制内嵌，修改后重新构建后端。

官网资源由`//go:embed assets/*`内嵌，只匹配顶层文件。新增JS/CSS需要四步：平铺放进`app/web/assets/`、在`app/web/web.go`的资源白名单与sha256版本列表登记、在`index.html`引入、重新编译Go二进制。帖子Markdown资源由`tools/`自包含npm工程生成：`npm --prefix tools ci`安装依赖，`sh tools/build_moment_markdown.sh`把渲染规则和扩展的`enhanceMarkdown`打包成`moment-markdown.js`，并从扩展复制`moment-markdown.css`；两份产物提交入库，`npm --prefix tools test`校验语法、XSS与产物完整性。`dev`分支通过`link:`解析扩展源码重建，`main`分支使用公共包与已提交产物。详见[网站动态开发实例](moments.md)。

官网的开发文档入口在同站`/docs/`，同时提供Dawn与PrismGo文档、搜索和亮暗主题；点击文档左上角品牌返回官网。构建命令在dawn-simple仓库执行：

```sh
cd docs
pnpm install --frozen-lockfile
pnpm build:site
```

将生成的`public/docs`随应用部署，HTTP服务只公开该构建目录，不公开文档源码。修改文档后重新构建即可；缺少构建产物时文档入口返回404。此命令使用`/docs/`作为资源路径；`pnpm build`仍使用GitHub Pages的`/dawn-simple/`路径与原文档首页，两种构建互不覆盖。

官网使用常驻外壳，无边框iframe铺满页面展示文档，不增加导航栏或内层滚动容器。客服挂载在外壳，切换章节、返回首页及浏览器前进后退保留聊天窗口、输入草稿和连接；地址栏同步当前文档路径与章节锚点，深层链接可直接打开或刷新。HTTP层为文档HTML注入导航、标题、主题和语言桥接脚本，不修改文档源码及构建文件；`__content=1`是iframe内部加载参数，不显示在官网地址栏。

### 官网语言同步

官网、Dawn文档、PrismGo文档和访客客服共用`zh/en`语言偏好，保存在`localStorage`的`dawn.site.locale`。首次访问跟随浏览器，手动选择后记住设置，并在标签页间同步。官网和文档使用语言下拉菜单；切换到当前文章的另一语言版本，导航、侧栏、搜索和目录同时切换。后台管理保留自己的语言设置。

Dawn中文文档保持根路径，英文放在`en/`。PrismGo使用对应路径，例如`prismgo/session.html`与`en/prismgo/session.html`；构建时映射上游源文件，同时保留旧`prismgo/zh_CN/`、`prismgo/en/`及README链接的跳转，不修改子模块。PrismGo正文中的语言链接也更新全局偏好。

客服公共脚本不读取官网的存储key；simple在初始化时传入`locale`，切换时调用`DawnSupport.setLocale()`。访客iframe不会重建，草稿、历史、滚动位置和连接保持不变；已有消息、应用名称及自定义欢迎语不会被翻译。

标准分页列表默认每页10条，复用公共表格的紧凑行距以尽量在一屏展示完整页；工具栏仍可切换紧凑、默认和宽松。短屏或包含多行内容的行保留表格内部滚动，手机端保持正常页面滚动。
