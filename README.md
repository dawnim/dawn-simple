# Dawn 示例应用

提供开箱即用的后台和可替换的仪表盘，业务功能从 app/admin 和 admin/src 开始开发。

## 分支与框架联调

`main`使用正式发布的Go模块和npm包；`dev`使用相邻目录的`dawn-server`和`dawn-admin`。开发框架功能或修复问题时使用dev；框架发版后切回main，升级正式依赖并验证。有问题再回dev修复、发版、重复验证。

切换分支后执行`go mod download`和`pnpm --dir admin install --frozen-lockfile`，避免沿用另一分支的前端依赖。业务变更单独提交，本地依赖配置不合入main。

## 准备

使用已发布的`dawn-server v0.1.0`与`@dawnim/admin 0.1.0`，直接安装项目依赖：

```sh
go mod download
pnpm --dir admin install --frozen-lockfile
cp .env.example .env
printf 'base64:'; openssl rand -base64 32
```

把完整密钥写入 `.env` 的 `APP_KEY`。数据库默认值在 `config/database.go`，本机MySQL使用 `dawn-admin` 库；MySQL通过 `DB_HOST`、`DB_PORT`、`DB_DATABASE`、`DB_USERNAME`、`DB_PASSWORD` 配置，`DB_DSN` 可覆盖连接串。Redis连接通过 `REDIS_URL` 配置，HTTP监听使用 `SERVER_HOST`、`SERVER_PORT`。

```sh
go run . dawn:install
```

首次安装默认创建 `admin`，密码为 `123456`；设置 `DAWN_ADMIN_PASSWORD` 可在安装时使用自定义密码。

`dawn:install`执行系统与项目迁移，补齐基础数据并保留已有配置和管理员密码。

## 方案一：组合运行

```sh
go run . dawn:start all
```

一次启动并守护HTTP5123、WS8851、Horizon和cron。服务异常退出时单独拉起，Ctrl+C或SIGTERM统一停止。Horizon继续管理自己的队列worker。

前后端一起开发：

```sh
cd admin
pnpm install --frozen-lockfile
cd ..
go run . dawn:start local
```

访问 `http://127.0.0.1:5173`，接口与WS通过前端开发服务器代理。前端支持热更新；修改Go代码后重新执行启动命令。

另开终端控制服务：

```sh
go run . dawn:start all status
go run . dawn:start horizon restart
go run . dawn:start http reload
go run . dawn:start ws reload
go run . dawn:start cron stop
go run . dawn:start cron
go run . dawn:start all stop
```

stop保持停止，restart停机后重新启动。HTTP重载等待在途请求完成；WS重载通知客户端重连，并恢复Redis中的会话与消息确认位置。应用日志默认 `storage/logs/dawn.log`。

单机生产构建：

```sh
mkdir -p bin storage/logs
go build -o bin/dawn-simple .
cd admin
pnpm install --frozen-lockfile
pnpm build
cd ..
./bin/dawn-simple dawn:start all
```

[supervisord.conf](supervisord.conf) 托管这一条组合命令：

```sh
supervisord -c supervisord.conf
supervisorctl -c supervisord.conf status
supervisorctl -c supervisord.conf restart 'dawn-simple:*'
```

## 方案二：独立运行

```sh
DAWN_RUNTIME_DIRECTORY=storage/framework/http ./bin/dawn-simple dawn:start http
DAWN_RUNTIME_DIRECTORY=storage/framework/ws ./bin/dawn-simple dawn:start ws
DAWN_RUNTIME_DIRECTORY=storage/framework/horizon ./bin/dawn-simple dawn:start horizon
DAWN_RUNTIME_DIRECTORY=storage/framework/cron ./bin/dawn-simple dawn:start cron
```

使用 [supervisord.split.conf](supervisord.split.conf) 分别托管：

```sh
supervisord -c supervisord.split.conf
supervisorctl -c supervisord.split.conf status
supervisorctl -c supervisord.split.conf restart dawn-simple:dawn-http
```

同目录分开托管时，用 `DAWN_RUNTIME_DIRECTORY` 区分管理器；控制命令使用相同值。两份配置二选一。组合模式也支持 `dawn:start http,ws,horizon`、`dawn:start cron,http`。前端生产使用 `admin/dist`，Nginx按根目录配置代理HTTP和WS。

## 队列与跨域

Horizon入口位于“运维管理 → 队列监控”，页面和API共用 `dawn.horizon.list` 权限。

| 组 | 任务 | 进程 |
|---|---|---|
| supervisor | default及需要顺序处理的任务 | 固定1个worker |
| multiples | parallel及可并发任务 | 按积压量伸缩，默认最多4个 |

新队列加入对应组，不为单个任务新建supervisor。串行组同一部署只启用一份；其他Horizon实例设置 `HORIZON_SERIAL=false`，并发上限通过 `HORIZON_CONCURRENCY` 调整。cron同一部署只启动一个实例。

`ALLOWED_ORIGINS=*` 允许任意来源；白名单用空格分隔域名。HTTP和WS使用同一份来源配置，接口仍执行登录与权限校验。前端跨域使用Bearer认证。

## 开发入口

| 路径 | 内容 |
|---|---|
| bootstrap/ | 应用启动、扩展和服务提供者 |
| app/admin/menu.go、routes.go | 本项目菜单、权限和后台路由 |
| app/admin/controllers/、middleware/ | 项目接口与中间件 |
| app/models/ | 实体和固定选项 |
| app/constants/ | 缓存、队列名称 |
| app/services/、events/、listeners/、jobs/ | 共享操作、事件和异步任务（按需添加） |
| app/cmd/、schedule/、exceptions/ | 命令、调度、业务异常 |
| database/ | 项目迁移与种子 |
| admin/src/ | 业务页面和自定义仪表盘 |

`admin/src/main.ts` 中的 `/dashboard/console` 映射完全替换系统仪表盘。完整页面和模块位于`admin/src/views/dashboard`，请求位于`admin/src/api/dashboard.ts`，可以直接修改卡片、图表和项目入口。顶栏时钟使用后端`APP_TIMEZONE`，同时展示时区和UTC偏移。目录约定见[架构说明](docs/architecture.md)，开发工具见[工具配置](docs/lens.md)。

## 检查与部署

```sh
go test -race ./...
go vet ./...
cd admin
pnpm build
```

测试使用临时SQLite和隔离Redis。生产部署保留与二进制同版的迁移文件，多个实例共享数据库、Redis、APP_KEY和文件磁盘。每个独立应用使用自己的数据库和缓存命名空间。

## 创建自己的项目

修改go.mod模块名、源码导入、config/dawn.go中的module和前端包名；按业务调整模型、菜单、队列与页面。公共后台能力继续通过依赖升级。
