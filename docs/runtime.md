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
| `status` | 查看状态、PID、重启次数、启动时间和最近错误 |

相同运行目录的命令连接已有管理器，控制范围为本机当前项目。`all stop` 停止所有服务，包含 local 的前端，并退出管理器。当前控制终端的 Ctrl+C 或 SIGTERM 同样触发收尾。

异常服务按 1、2、4 秒递增重试，最大间隔 30 秒，稳定运行一分钟后重置。Horizon 自行管理内部 worker，单实例停止等待任务收尾，遵循原生租约恢复。

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

`status` 的 `terminal_pid` 表示当前控制进程，日志写入 `storage/framework/local.log`。更新管理器代码或监听地址时，先执行 `go run . dawn:start local stop`，再启动。

## 平滑重载

HTTP/WS 的监听资源由管理器持有。新进程继承监听资源并报告就绪后，旧进程开始排空；新进程启动失败时保留旧服务。

HTTP 等待在途请求完成，超时使用 `SERVER_SHUTDOWN_TIMEOUT`。WS 分散发送 `1012` 关闭通知，客户端重连后通过 Redis 恢复会话、订阅与确认位置；消息恢复受保留时间和条数限制。

进程守护与监听器交接面向 Linux、macOS。生产更新固定路径的二进制后执行 reload 或 restart；`go run` 使用本次编译的临时二进制。监听地址变化需重启管理器。

## 前端代理

local 向前端传入 `DAWN_HTTP_TARGET`、`DAWN_WS_TARGET`、`DAWN_WS_PATH`、`DAWN_ADMIN_PREFIX`。项目的 [Vite 配置](https://github.com/dawnim/dawn-simple/blob/main/admin/vite.config.ts) 已读取这些值，浏览器经前端端口访问 API 与 WS。

`DAWN_FRONTEND_DIRECTORY`、`DAWN_FRONTEND_PORT` 可调整前端目录和端口。Vite 使用严格端口，端口占用时返回错误。

## 生产部署

在项目根目录构建：

```sh
mkdir -p bin storage/logs
go build -o bin/dawn-simple .
pnpm --dir admin install --frozen-lockfile
pnpm --dir admin build
./bin/dawn-simple dawn:install --force
./bin/dawn-simple dawn:start all
```

部署保留与二进制同版的 `database/migrations`、`database/seeders` 源码，由一个实例执行安装。设置生产环境变量，配置数据库、Redis 与 `APP_KEY`；多个实例共享连接、应用密钥和文件存储。

Nginx 提供 `admin/dist`，将 `/admin/`、`/api/` 转发至 HTTP，`/ws` 转发至 WS 并启用 Upgrade。默认 WS 服务路径为 `/`，代理需将 `/ws` 改写为 `/`。公开附件通过 `public/storage` 提供，私密附件通过授权接口读取。

在站点的 `server` 中添加反向代理：

```nginx
location ~ ^/(admin|api)(/|$) {
    proxy_pass http://127.0.0.1:5123;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
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
