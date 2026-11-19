# 测试与发布

## 应用检查

在 dawn-simple 根目录执行：

```sh
go test -race ./...
go vet ./...
pnpm --dir admin build
```

测试使用临时 SQLite 与隔离 Redis。接口变更检查认证、操作权限、数据范围和响应；页面变更检查主题、权限显示、请求与交互。

## 框架检查

在 dawn-server 根目录执行：

```sh
go test -race ./...
go vet ./...
```

测试覆盖授权、参数加密、附件、缓存、CSV、CORS、聊天投递和进程管理。WS 检查消息确认、跨实例恢复与补发，终端接管检查控制权交接与收尾。

在 dawn-admin 根目录执行：

```sh
pnpm install --frozen-lockfile
pnpm test
pnpm lint
pnpm build
npm pack --pack-destination /tmp
```

公开入口 `src/index.ts` 与 `types/index.d.ts` 同步维护。修改公共组件、类型、认证或菜单后，在独立业务工程安装实际打包产物并构建。

## 数据库矩阵

在 dawn-server 使用明确指定的专用空库：

```sh
DAWN_TEST_DRIVER=pgsql DAWN_TEST_DSN='专用测试库DSN' \
  go test -run 'Test(Application|DatabaseMatrix|DatabaseMetadata|RESTResources)$' -count=1 ./tests
```

驱动可选 `mysql`、`mariadb`、`pgsql`。测试创建固定账号和菜单，按测试清理逻辑重建表；连接必须指向专用测试库。未指定连接时，外部矩阵跳过，常规测试使用临时 SQLite。

## 框架联调

`dawn-server`、`dawn-admin`、`dawn-simple`放在同级目录。`main`使用已发布的Go模块和npm包；`dev`通过Go `replace`和npm `link`引用相邻框架源码。切换到`dev`后，确认相邻仓库包含要联调的源码，再安装依赖：

```sh
git switch dev
go mod download
pnpm --dir ../dawn-admin install --frozen-lockfile
pnpm --dir admin install --frozen-lockfile
```

在`dev`运行`pnpm --dir admin dev:lib`监听前端库构建，另开终端运行`go run . dawn:start local`。Go源码由`replace`或显式启用的`go.work`决定，不能仅凭分支名判断；即使在`main`，启用本地`go.work`也会加载相邻`dawn-server`源码。前端由`admin/package.json`中的`link:`决定，未链接时仍使用发布包。用`go env GOWORK`和`realpath admin/node_modules/@dawnim/admin`确认实际依赖来源。发布构建使用`main`，关闭本地`go.work`并安装锁定的发布依赖。

## 运行验收

| 范围 | 检查内容 |
|---|---|
| 服务 | HTTP 健康、WS 认证、队列消费、定时任务 |
| 进程 | SIGTERM 收尾、异常恢复、主动停止、重载失败保留旧进程 |
| 前端 | 跨域预检、Bearer、iframe、下载、权限显示与页面覆盖 |
| 实时消息 | 双端收发、未读与已读、断线恢复、重复消息合并 |
| 部署 | 共享存储、Redis 故障恢复、SMTP 与实际容量 |

## 文档站

在 dawn-simple 根目录执行：

```sh
git submodule update --init docs/prismgo
pnpm --dir docs install --frozen-lockfile
pnpm --dir docs dev
```

构建使用 `pnpm --dir docs build`，本地预览使用 `pnpm --dir docs preview`，产物位于 `docs/.vitepress/dist`。官网使用`pnpm --dir docs build:site`，产物位于`public/docs`，资源路径为`/docs/`。

文档更新后执行`pnpm --dir docs build:all`，一次更新两套产物。官网与GitHub Pages共用`docs/.vitepress/config.mts`，Dawn侧栏和上下篇直接从[中文目录](README.md)与[英文目录](en/README.md)生成；新增、删除或改名教程时同步更新两个目录，无需另外维护菜单。PrismGo目录从子模块的中英文README读取。GitHub Actions也执行两套构建检查，发布时只上传Pages产物；官网部署使用`public/docs`。

Dawn英文文档放在`docs/en`，与中文保持相同文件名，修改教程时同步维护两种语言。导航、搜索及上下篇使用VitePress语言配置；官网与文档共用语言偏好，见[官网语言同步](frontend.md#官网语言同步)。两种构建均检查内部页面链接。

GitHub Pages 使用`/dawn-simple/`路径发布静态文档，语言菜单可在中英文页面间切换并记住选择。浏览器按域名隔离语言偏好，因此 Pages 与官网分别保存各自的语言设置。

PrismGo 文档通过 `docs/prismgo` 子模块连接官方仓库，构建时读取上游最新内容。本地更新执行 `git submodule update --remote docs/prismgo`。

仓库的 Settings → Pages → Source 选择 **GitHub Actions**。推送 `main` 后，Docs 工作流自动构建并发布到 `https://dawnim.github.io/dawn-simple/`；PR 仅构建检查，也可从 Actions 手动运行。

## 升级与发布

应用升级时固定框架版本、同步锁文件，备份数据库，执行 `dawn:install --force`，重新构建前端并更新服务。菜单与权限 key 保持稳定，变更任务类型或编码前排空旧任务。

dawn-server 使用版本标签发布 Go Module，CI 通过 Go 测试与数据库矩阵后创建 Release。在 dawn-server 执行，例如：

```sh
git tag -a v0.1.2 -m v0.1.2
git push origin main v0.1.2
```

dawn-admin 的标签须与 `package.json` 版本一致，CI 检查后通过 npm 可信发布上传包并创建 Release。在 dawn-admin 执行：

```sh
npm version patch -m "chore: release v%s"
git push origin main --follow-tags
```

发布后使用正式 Go Module 和 npm 包验证 dawn-simple 的类型、构建与主要业务路径。
