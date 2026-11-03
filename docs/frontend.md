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

后端菜单的 `Component` 与注册路径一致。业务页面优先于内置页面解析，按访问需要加载。菜单与权限注册见[后端开发](backend.md#手写-crud)。

## 业务列表

在 `admin/src/api/records.ts` 定义资源 API 与字段：

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
      '/ws': { target: 'ws://127.0.0.1:8851', ws: true, rewrite: () => '/' }
    }
  }
})
```

TypeScript 使用 `moduleResolution: Bundler` 并包含 `vite/client` 类型。完整配置可参考 [admin 目录](https://github.com/dawnim/dawn-simple/tree/main/admin)。生产构建与代理见[服务运行](runtime.md#生产部署)。
