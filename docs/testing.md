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

`dawn-server`、`dawn-admin`、`dawn-simple` 放在同级目录。dawn-simple 的 `main` 使用发布依赖，`dev` 通过 Go `replace` 和 npm `link` 引用相邻框架仓库。

在 dawn-simple 执行：

```sh
git switch dev
go mod download
pnpm --dir ../dawn-admin install --frozen-lockfile
pnpm --dir admin install --frozen-lockfile
pnpm --dir admin dev:lib
```

保持前端库监听构建，另开终端运行 `go run . dawn:start local`。切换分支后重新安装对应依赖；框架发布后回到 main，升级版本与锁文件，再验证业务工程。

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

构建使用 `pnpm --dir docs build`，本地预览使用 `pnpm --dir docs preview`，产物位于 `docs/.vitepress/dist`。

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
