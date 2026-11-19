# Frontend

Business frontend code lives in `admin/src`. Use `@dawnim/admin` for layouts, system pages, requests, permissions, and shared components.

## Register pages

Add pages to `createAdmin` in `admin/src/main.ts`:

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

The backend menu's `Component` must match the registered path. Business pages override built-in pages and load on demand. See [Customize generated CRUD](backend.md#customize-generated-crud) for menus and permissions.

## Business lists

Start with the [code generator or CLI](backend.md#generate-business-code). The following example shows how to customize the generated API and columns:

```ts
import { createResourceApi, type Column } from '@dawnim/admin'

export default createResourceApi('/records')

export const columns: Column[] = [
  { prop: 'name', label: '名称', required: true, maxlength: 100 }
]
```

Use the standard list in `admin/src/views/records/index.vue`:

```vue
<template>
  <CrudPage permission="app.records" :api="api" :columns="columns" :search="['name']" />
</template>

<script setup lang="ts">
  import { CrudPage } from '@dawnim/admin'
  import api, { columns } from '../../api/records'
</script>
```

`CrudPage` provides filters, column settings, pagination, editing, deletion, and CSV import/export. For custom layouts, compose `ArtTable`, `ArtTableHeader`, `DawnSearchBar`, `DawnButton`, `useTable`, `useAdmin`, and `useAuth`.

The generator's reference mode reuses `CrudPage`. Source mode emits list, filter, and edit-dialog components for deeper customization. Both modes share fields, permissions, and interactions.

`Column.dict` reuses framework dictionaries: tinted labels in lists, `DawnSelect` for filters, and `DawnRadio` for editing. Set `dictValueType: 'string'` for string values. Use `options` for fixed choices. Columns also support `readonly`, `immutable`, `nullable`, `required`, and `maxlength`. Filters after the first three are collapsed by default and keep their values when expanded. Reference and source modes behave identically.

```ts
{ prop: 'status', label: t('状态'), type: 'number', dict: 'data_status' }
```

### Table density

The table toolbar shares three density settings across lists and persists the selection in the browser. Set `ArtTable.size` to override it for one table. Paginated lists, trees, and plain tables use the same dimensions:

| Density | `size` | Base row height | Body font | Action buttons, avatars, and images |
|---|---|---|---|---|
| Compact | `small` | 32px | 12px | 24px |
| Default | `default` | 36px | 13px | 28px |
| Spacious | `large` | 40px | 14px | 32px |

Standard lists default to 10 rows per page. With extra filters collapsed, default density targets 10 visible rows at a 1024×768 viewport; spacious density targets 10 rows at a typical 14-inch MacBook Pro viewport. Search panels and table cards retain 16px horizontal padding, with pagination displayed separately. Plain long text uses ellipsis and a hover tooltip; set a column's `showOverflowTooltip: false` to disable it. Multiline slots, expanded rows, and extra filters can still increase height. Use column `minWidth` to keep content readable on narrow screens with horizontal scrolling inside the table.

## Requests and permissions

Use shared `request` for business APIs:

```ts
import { request } from '@dawnim/admin'

const record = await request.get<{ id: number; name: string }>({ url: '/records/1' })
await request.patch({ url: `/records/${record.id}`, data: { name: '新名称' } })
```

Resource APIs expose `list/get/create/update/replace/remove/removeMany`. `update` sends PATCH; `replace` sends PUT. When implementing `list(query, config?)`, forward the second argument to `request` so tables can cancel outdated requests. See [API conventions](backend.md#api-conventions) for pagination and responses.

```vue
<ElButton v-permission="'app.records.create'">新增</ElButton>
```

Permission directives control visibility; APIs enforce authorization. Requests carry Bearer tokens and consolidate concurrent refreshes. Cross-origin requests default to `withCredentials=false`; same-origin iframes and downloads can use the login Cookie. See [CORS and proxies](configuration.md#cors-and-proxies).

## Public APIs

| API | Purpose |
|---|---|
| `createAdmin` | Register layout, routes, themes, and permission directives |
| `CrudPage`, `createResourceApi` | Standard lists and resource requests |
| `DawnSelect`, `DawnRadio`, `DawnDict` | Dictionary selects, radio groups, and labels |
| `Upload` | Uploads, chunks, digest preflight, and deduplication |
| `request` | Authentication, refresh, errors, and downloads |
| `realtime`, `openChat` | Shared realtime connection and admin chat |
| `StatisticsDashboard`, `WorkDashboard` | Default statistics and employee dashboards |
| `ArtBarChart`, `ArtLineChart` | Theme-aware business charts |
| `ArtCountTo`, `ArtSvgIcon` | Number displays and icons |

Import components and types from `@dawnim/admin`. See [Chat and realtime](chat.md#http-and-realtime-protocol) for realtime APIs.

## Dashboards and header

Override either dashboard through `pages`:

```ts
pages: {
  '/dashboard/statistics': () => import('./views/dashboard/index.vue'),
  '/dashboard/work': () => import('./views/dashboard/work.vue')
}
```

Set home page in admin user management selects the dashboard type. Reuse default components and fetch business statistics from your own APIs. dawn-simple's pages and charts live in `admin/src/views/dashboard`; requests live in `admin/src/api/dashboard.ts`.

Extend the header with components:

```ts
import HeaderTools from './components/HeaderTools.vue'

// createAdmin 配置
header: { before: HeaderTools }
```

`before` appears before search; `after` appears after the avatar; `right` replaces the entire right-hand tool area. Components receive `user`, the application `timezone`, and `serverTimeOffset`. They can use Router, i18n, and shared requests. Release timers and listeners when unmounting.

## Themes and languages

Light/dark modes and brand colors use global settings. The default logo and favicon follow the theme; pass image URLs through `logo` and `favicon` for your brand. Saved site names and logos update the login page and layout; `title` and `logo` provide initial values.

Admin defaults are system theme, vertical menu, the first menu style, the third brand color, bordered boxes, full-width containers, default tabs, a 200px menu, fade transitions, and 0.75 corner radius. Basic display options and accordion mode are enabled; color-weakness mode and watermark are disabled. Existing browsers retain personal settings. Use Reset configuration in theme settings to apply defaults. Dark mode preserves the selected menu style.

Use the original Chinese text as the key for fixed interface labels:

```ts
const text = t('欢迎{username}', { username: 'Dawn' })
```

Chinese falls back to the original. Register English through `createAdmin.messages`:

```ts
messages: {
  en: { '欢迎{username}': 'Welcome, {username}' }
}
```

Chinese language files maintain only technical identifier mappings such as `fields`. Generated pages load zh/en field files by module. API fields remain `snake_case`; pass dynamic values as named placeholders.

## Integrate an existing frontend

Install the component library in an existing Vue application. See [Requirements](quickstart.md#requirements) for Node versions.

```sh
pnpm add @dawnim/admin vue vue-router pinia element-plus vue-i18n
pnpm add -D vite@^7 @vitejs/plugin-vue@^6 typescript@~5.6 vue-tsc@~2.1
```

Use the `createAdmin` entry above. Provide `#app` in HTML and load `src/main.ts`. Set `type: module`, use `vite` for `dev`, and `vue-tsc --noEmit && vite build` for `build` in `package.json`.

Deduplicate shared runtime dependencies and proxy the backend:

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

Proxy public attachments under `/storage/` to the backend too. For cross-origin integration, relative image URLs use the origin of `apiBase`; absolute CDN URLs stay unchanged. Shared image components fetch private images with Bearer authentication.

TypeScript uses `moduleResolution: Bundler` and includes `vite/client` types. See the [admin directory](https://github.com/dawnim/dawn-simple/tree/main/admin) for complete configuration, and [Production deployment](runtime.md#production-deployment) for builds and proxies.

## Website and embedded documentation

The backend HTTP service's `/` serves the community feed with posts, images, comments, flowers/eggs, AI drafting, and chat. `/dawn` keeps the Dawn introduction, account, and support page. The admin frontend uses its own server or domain. Website assets live in `app/web/assets` and are embedded in the Go binary; rebuild the backend after editing them.

`//go:embed assets/*` only matches top-level files. Adding JS/CSS takes four steps: place it flat in `app/web/assets/`, register it in the allowlist and sha256 version list in `app/web/web.go`, link it from `index.html`, and rebuild the Go binary. The post Markdown assets come from the self-contained npm project in `tools/`: `npm --prefix tools ci` installs dependencies, `sh tools/build_moment_markdown.sh` bundles the render rules with the extension's `enhanceMarkdown` into `moment-markdown.js` and copies `moment-markdown.css` from the extension. Both artifacts are committed; `npm --prefix tools test` checks syntax, XSS, and artifact integrity. `dev` resolves the extension source through the `link:` dependency to rebuild, while `main` uses the public package and committed artifacts. See [Community example](moments.md).

The website serves Dawn and PrismGo documentation under `/docs/`, with search and light/dark themes. Click the documentation logo to return to the website. Build from the dawn-simple repository:

```sh
cd docs
pnpm install --frozen-lockfile
pnpm build:site
```

Deploy the generated `public/docs` with the application. HTTP exposes only the build output, not documentation sources. Rebuild after editing docs; missing output returns 404. This command uses `/docs/` as its resource base. `pnpm build` retains GitHub Pages' `/dawn-simple/` base and documentation home. The builds have separate output directories.

A persistent website shell displays docs in a borderless full-page iframe, without a second navigation bar or nested scroll container. Support stays mounted on the shell: chapter navigation, returning home, and browser back/forward retain the chat, draft, and connection. The address bar follows document paths and anchors. Deep links can be opened or refreshed directly. HTTP injects the navigation, title, theme, and language bridge into HTML; PrismGo sources and generated files are unchanged. `__content=1` is an internal iframe parameter, omitted from the outer URL.

### Shared website language

The website, Dawn docs, PrismGo docs, and visitor support share a `zh/en` preference stored in `localStorage` as `dawn.site.locale`. First visits follow the browser language; explicit choices persist and synchronize across tabs. Both the website and docs use language dropdowns. Switching docs opens the same article in the other language; sidebar, search, outline, and navigation also switch. Existing PrismGo links remain accessible. Backend admin retains its own language settings.

Chinese Dawn docs retain their root paths; English docs use `en/`. PrismGo uses matching routes such as `prismgo/session.html` and `en/prismgo/session.html`. The build maps upstream sources and retains redirects for old `prismgo/zh_CN/`, `prismgo/en/`, and README links. PrismGo's own language links update the shared preference too.

The public support widget does not read this website's storage key. simple passes `locale` on initialization and calls `DawnSupport.setLocale()` when it changes. The visitor iframe is retained, preserving drafts, history, scroll, and the connection. Message content, application branding, and custom greetings remain authored content; changing UI language does not translate existing messages.

Standard paginated lists default to 10 rows and shared compact spacing to fit one page where possible. Toolbars still allow compact, normal, or spacious density. Short screens and multiline rows retain internal table scrolling; mobile uses normal page scrolling.
