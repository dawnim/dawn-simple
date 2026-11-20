# 服务运行

本地开发使用 `dawn:start local`，生产使用 `dawn:start all`。HTTP、WS、Horizon 和 cron 分进程运行，由 Dawn 管理启动、停止与异常恢复。

## 启动与控制

在项目根目录执行：

```sh
go run . dawn:start local
```

`local` 启动前端开发服务器与 Go 源码监听，默认前端目录 `admin`、端口 `5173`。`all` 运行后端全部服务，也可用逗号选择服务，如 `http,ws,horizon`。

另开终端执行控制命令：

```sh
go run . dawn:start all status
go run . dawn:start horizon restart
go run . dawn:start http reload
go run . dawn:start ws reload
go run . dawn:start cron stop
go run . dawn:start cron
go run . dawn:start all stop
```

| 操作 | 行为 |
|---|---|
| `start` | 启动服务并保持守护，省略操作时使用 start |
| `stop` | 停止指定服务并保持停止 |
| `restart` | 先停止，再启动 |
| `reload` | HTTP/WS 先启动新进程，就绪后交接 |
| `status` | 查看主子进程、PID/PPID、RSS内存、CPU、协程数、运行用户、线程、运行时长、重启次数和最近错误 |

相同运行目录的命令连接已有管理器，控制范围为本机当前项目。`all stop` 停止所有服务，包含 local 的前端，并退出管理器。当前控制终端的 Ctrl+C 或 SIGTERM 同样触发收尾。

异常服务按 1、2、4 秒递增重试，最大间隔 30 秒，稳定运行一分钟后重置。Horizon 自行管理内部 worker，单实例停止等待任务收尾，遵循原生租约恢复。

启动和控制命令采用统一的服务表与日志样式。Go进程通过项目本地控制通道上报协程数，其他进程显示`—`。CPU按相邻采样计算，单核满载为100%，多核可以超过100%；首次采样和不可用的指标显示`—`。RSS为操作系统进程驻留内存，不等于Go堆内存。指标每2秒采样，`status`读取最新快照。宽终端直接展示详情列，普通宽度在服务详情行展示。

使用`--no-ansi`或`NO_COLOR=1`关闭颜色；重定向默认使用纯文本。脚本读取结构化状态：

```sh
go run . dawn:start all status --json
```

### 结构未就绪

Go服务启动前只读检查应用完整目标中的必需表和列，不执行DDL。数据库不可达或缺少结构时显示`等待就绪`，保持当前进程和管理器存活，每5秒重新检查；同一服务的相同原因不重复刷屏。此时服务尚未就绪，不能对外接收业务请求。修复后原进程继续启动，不增加重启次数；`status`与`stop`仍可使用。

首次安装另开终端执行`go run . dawn:install`；版本升级执行未运行迁移。缺失项位于已执行迁移内部时，重新运行`migrate`不会重放该文件，执行`dawn:schema check`预览、`dawn:schema sync`同步目标结构。重载候选进程结构检查未通过时，保留原服务，不切换到等待中的新进程。

进程存活不表示服务就绪，自动部署仍应检查服务状态与实际健康接口。结构未就绪时保留运行控制能力，不能将其判作成功发布。

## 源码监听

`local` 监听应用及本地 `replace` 模块的 Go 源码与模块文件，合并连续保存后重新编译。编译成功后，HTTP/WS 平滑交接，Horizon 与 cron 重启，前端继续运行；编译失败会输出错误，当前服务继续运行。主动停止的服务保持停止。

测试文件、文档、前端构建产物与运行目录不触发编译。设置 `DAWN_GO_WATCH=false` 可关闭监听；监听地址和管理器自身的变更需停止后重新启动。

## 本地终端接管

需要将控制权切换到新终端时执行：

```sh
go run . dawn:start local
```

新命令加载本次编译的服务，接管日志与 Ctrl+C 控制，原终端命令退出。管理器持续运行，HTTP/WS 平滑交接，其他服务依次停止并启动。

```mermaid
flowchart LR
    A[新终端编译应用] --> B[更新服务]
    B --> C[确认就绪]
    C --> D[交接日志与控制权]
    D --> E[原终端命令退出]
    B -->|接管失败| F[原终端继续控制]
```

`status` 的 `terminal_pid` 表示当前控制进程，日志写入 `storage/framework/local-cli.log`。服务输出携带时间与服务名，完整保留多行编译和运行错误。更新管理器代码或监听地址时，先执行 `go run . dawn:start local stop`，再启动。

旧管理器正在收尾时，新终端最多等待30秒释放锁；启动锁直接传给新管理器，避免并发启动冲突。管理器启动失败时，终端会先输出具体错误，再退出。

## 平滑重载

HTTP/WS 的监听资源由管理器持有。新进程继承监听资源并报告就绪后，旧进程开始排空；新进程启动失败时保留旧服务。

HTTP 等待在途请求完成，超时使用 `SERVER_SHUTDOWN_TIMEOUT`。WS 分散发送 `1012` 关闭通知，客户端重连后通过 Redis 恢复会话、订阅与确认位置；消息恢复受保留时间和条数限制。

进程守护与监听器交接面向 Linux、macOS。生产更新固定路径的二进制后执行 reload 或 restart；`go run` 使用本次编译的临时二进制。监听地址变化需重启管理器。

## 前端代理

local 向前端传入 `DAWN_HTTP_TARGET`、`DAWN_WS_TARGET`、`DAWN_WS_PATH`、`DAWN_ADMIN_PREFIX`。项目的 [Vite 配置](https://github.com/dawnim/dawn-simple/blob/main/admin/vite.config.ts) 已读取这些值，浏览器经前端端口访问 API 与 WS。

`DAWN_FRONTEND_DIRECTORY`、`DAWN_FRONTEND_PORT` 可调整前端目录和端口。Vite 使用严格端口，端口占用时返回错误。

## 生产部署

在构建机器的项目根目录运行 `./deploy.sh root@server /srv/dawn-simple`。脚本构建 Linux 程序、后台和文档，通过 SSH 上传产物。目标服务器需要 Bash、rsync、flock、Python 及 venv，以及可用的项目 `.env`。默认按 Linux amd64 构建，应用以 `www` 运行，可通过 `DEPLOY_ARCH` 和 `DEPLOY_USER` 调整。

部署依次执行安装、完整数据库结构同步和服务更新。脚本安装并启动本地向量服务，模型就绪后同步 Dawn 与 PrismGo 的中文文档，再通过 `app:support-docs --index` 重建官网客服索引。

在项目根目录构建：

```sh
mkdir -p bin storage/logs
git submodule update --init --recursive
pnpm --dir docs install --frozen-lockfile
pnpm --dir docs build:site
pnpm --dir admin install --frozen-lockfile
pnpm --dir admin build
go build -o bin/dawn-simple .
./bin/dawn-simple dawn:install --force
./bin/dawn-simple dawn:start all
```

部署保留与二进制同版的 `database/migrations`、`database/seeders` 源码，由一个实例执行安装。设置生产环境变量，配置数据库、Redis 与 `APP_KEY`；多个实例共享连接、应用密钥和文件存储。

Go HTTP 服务负责 `/` 官网、`/docs/` 文档外壳、`/site-assets/`官网资源、`/favicon.ico`、`/chat/`访客客服和`/support-assets/`客服资源；这些路径必须代理到HTTP，不能由Nginx绕过，否则文档导航和客服注入不会生效。Nginx将`/admin/`、`/api/`转发至HTTP，将`/ws`转发至WS并启用Upgrade。默认WS服务路径为`/`，代理需将`/ws`改写为`/`。公开附件通过`public/storage`提供，私密附件通过授权接口读取。

使用宝塔等会生成 CSS、JS 扩展名匹配规则的面板时，为 `/site-assets/`、`/support-assets/`、`/docs/`、`/storage/` 分别设置 `location ^~`，让请求进入 Go HTTP 服务。

在站点的 `server` 中添加反向代理：

```nginx
location ~ ^/(admin|api|chat|support-assets|site-assets|docs|storage)(/|$) {
    proxy_pass http://127.0.0.1:5123;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

location = /favicon.ico {
    proxy_pass http://127.0.0.1:5123;
    proxy_set_header Host $host;
}

location = / {
    proxy_pass http://127.0.0.1:5123;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
}

location = /ws {
    proxy_pass http://127.0.0.1:8851/;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 75s;
}
```

生产不需要运行Node/Vite；保留Go二进制、`public/docs`、`admin/dist`和配置文件即可。`admin/dist`构建不会复制官网文档，官网资源和favicon由Go服务提供。

同机代理设置 `SERVER_TRUSTED_PROXIES=127.0.0.1`；分开部署时填写实际代理 IP。调整接口前缀或 WS 路径时同步修改对应 `location`。

### 组合托管

使用 [supervisord.conf](https://github.com/dawnim/dawn-simple/blob/main/supervisord.conf) 托管 Dawn 主进程：

```sh
supervisord -c supervisord.conf
supervisorctl -c supervisord.conf status
supervisorctl -c supervisord.conf restart 'dawn-simple:*'
```

### 独立托管

需要分别部署时，使用 [supervisord.split.conf](https://github.com/dawnim/dawn-simple/blob/main/supervisord.split.conf)。同目录的各管理器使用独立运行目录：

```sh
DAWN_RUNTIME_DIRECTORY=storage/framework/http ./bin/dawn-simple dawn:start http
DAWN_RUNTIME_DIRECTORY=storage/framework/ws ./bin/dawn-simple dawn:start ws
DAWN_RUNTIME_DIRECTORY=storage/framework/http ./bin/dawn-simple dawn:start http reload
```

组合与独立托管按部署方式择一使用。同一部署保留一个 cron 和一个串行队列组，其他 Horizon 实例设置 `HORIZON_SERIAL=false`，按需扩容并发 worker。

## 日志

应用日志默认 `storage/logs/dawn.log`，错误日志位于 `storage/logs/error/`，终端输出由外部进程管理器收集。日志通道、SQL 日志与监控配置见[配置](configuration.md#日志与监控)。

## 可选向量服务

先执行`go run . dawn:embedding install`，再执行`go run . dawn:start embedding`。设置`EMBEDDING_ENABLED=true`后随all/local启动。启动、停止、重启、状态均由现有管理器处理；下载及加载权重时显示等待就绪，普通启动不安装Python依赖。详见[本地向量服务](ai.md#本地向量服务)。
