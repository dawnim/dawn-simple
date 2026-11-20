# 配置

启动参数放在 `config/*.go`，通过 `.env` 覆盖；后台可编辑的参数通过 `Options.Settings` 注册，按组保存到数据库。

## 启动参数

配置使用 `config.Add`、`config.Env` 注册，启动包空导入配置包。完整环境变量见 [.env.example](https://github.com/dawnim/dawn-simple/blob/main/.env.example)。

| 文件 | 内容 |
|---|---|
| `config/app.go` | 应用名称、环境、密钥、时区与 HTTP 服务 |
| `config/database.go` | SQL 与 Redis 连接 |
| `config/dawn.go` | 后台前缀、WS、登录、地理库与本地运行参数 |
| `config/cache.go`、`session.go` | 缓存与框架 Session |
| `config/queue.go`、`horizon.go` | 队列、失败记录、批次与 worker |
| `config/filesystem.go` | 本地与云端磁盘 |
| `config/logging.go`、`encoding.go` | 日志与编码 |

缓存、Session、队列与 Horizon 使用独立 Redis 连接。应用名称作为 Redis key 命名空间，`APP_ENV` 不参与前缀。

## 数据库与 Redis

默认使用 MariaDB，`DB_CONNECTION` 选择连接：

| 连接 | 配置 |
|---|---|
| `mysql`、`mariadb` | `DB_HOST`、`DB_PORT`、`DB_DATABASE`、`DB_USERNAME`、`DB_PASSWORD`；`DB_DSN` 可覆盖连接串 |
| `pgsql` | `DB_DSN`，例如 `host=127.0.0.1 port=5432 user=postgres dbname=dawn_admin sslmode=disable` |
| `sqlite` | `DB_DATABASE` 指定文件，例如 `storage/database.sqlite?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)` |

SQLite 父目录须存在，连接池默认 1 个连接。其他连接可用 `DB_MAX_OPEN_CONNS`、`DB_MAX_IDLE_CONNS`、`DB_CONN_MAX_LIFETIME`、`DB_CONN_MAX_IDLE_TIME` 调整连接池，TLS 与驱动参数保留在对应连接下。

Redis 优先使用 URL：

```text
redis://127.0.0.1:6379/0
redis://:password@127.0.0.1:6379/0
redis://username:password@127.0.0.1:6379/0
rediss://default:password@cache.example.com:6379/0
```

设置 `REDIS_URL` 或共用的主机和认证字段。默认连接用于应用状态；`cache`、`session`、`queue`、`horizon` 使用独立 Redis DB，分别由 `REDIS_CACHE_DB`、`REDIS_SESSION_DB`、`REDIS_QUEUE_DB`、`REDIS_HORIZON_DB` 指定。失败任务与批次记录使用队列连接。只有组件需要不同 Redis 服务时才单独设置该组件的 URL。URL 中的特殊字符需编码，如 `@` 写为 `%40`。`rediss://` 启用 TLS 并校验证书，使用服务提供的 DNS 端点。

也可使用 `REDIS_HOST`、`REDIS_PORT`、`REDIS_USERNAME`、`REDIS_PASSWORD`；非空分项覆盖 URL 中的对应值。当前客户端配置面向单节点或关闭 Cluster Mode 的实例。

## 跨域与代理

| 参数 | 默认值与用途 |
|---|---|
| `SERVER_HOST`、`SERVER_PORT` | HTTP 监听，默认 `127.0.0.1:5123` |
| `WS_ADDR`、`WS_PATH` | WS 监听，默认 `:8851` 与 `/` |
| `ADMIN_PREFIX` | 后台 API 前缀，默认 `/admin` |
| `ALLOWED_ORIGINS` | 默认 `*`；白名单用空格分隔完整来源 |
| `SERVER_CLIENT_IP_MODE` | `trusted` 或 `simple`，默认 `trusted` |
| `SERVER_CLIENT_IP_SIMPLE_HEADERS` | 简单模式请求头优先级，逗号分隔 |
| `SERVER_CLIENT_IP_HEADERS` | 可信模式请求头优先级，逗号分隔 |
| `SERVER_CLOUDFLARE_PROXIES`、`SERVER_CLOUDFRONT_PROXIES` | 启用官方 CDN 网段自动更新，默认 `false` |
| `SERVER_TRUSTED_PROXIES` | 可信代理 IP 或网段，用逗号分隔 |
| `SERVER_SHUTDOWN_TIMEOUT` | 在途请求关闭等待，默认 `15s` |


简单模式依次查找请求头中的公网地址，再使用直连地址或私网地址兜底，请求头可以由客户端伪造。可信模式从直连地址开始，按配置顺序选择首个存在的转发头，从右向左查找第一个非可信地址，遇到无效地址时停止。`SERVER_TRUSTED_PROXIES` 需要包含应用与 CDN 之间的每一层代理，包括本机 Nginx 地址。启用 CDN 后，Go 进程启动时及每天下载官方网段，最近一次成功的结果保存在 `storage/app/clientip`。尚无有效网段时只信任显式配置的代理。应限制源站直接访问，并确保可信代理保留转发链。修改环境配置后重启服务生效。

官网和后台域名的 Nginx 示例位于 `deploy/nginx`。`/ws` 使用独立监听，当前服务自身没有使用客户端 IP 的处理。

HTTP 与 WS 使用相同的来源配置。任意来源模式使用 Bearer 认证；白名单模式允许指定来源携带凭证。同源 iframe 和下载使用登录 Cookie。接口始终执行身份与权限校验。

## 系统参数

在 `Configure` 中追加业务配置：

```go
o.Settings = append(o.Settings, sdk.SettingDefinition{
	Key: "business.order_timeout", Title: "订单超时分钟数", Group: "业务配置",
	Kind: "number", Input: "number", Default: 30,
})
```

执行 `go run . db:seed` 补齐分组与新增字段，读取使用 `a.Setting(ctx, "business.order_timeout")`。数据库每组一行，值为 JSON；类型、控件、选项和说明由代码声明，已有配置值保留。

站点名称、Logo、版权、简介与备案用于界面展示，`GET /admin/site` 返回公开展示字段、程序时区与服务器时间。个人资料通过 `PATCH /admin/auth/profile` 局部更新，日志按当前账号读取。

`APP_KEY` 使用 `base64:` 加 32 字节随机值。保密参数设置 `Secret: true`，加密存储，接口返回掩码；保存时留空保留原值。轮换密钥时安排数据转换，并保留所需旧密钥。

## 文件存储

后台上传配置支持本地、阿里云 OSS、七牛云、腾讯云 COS 与 Amazon S3。云端提前准备 Bucket，并填写 Endpoint、区域、凭证和访问域名；七牛与 COS 使用 S3 兼容端点，S3 支持 Path Style。

| 访问方式 | 本地目录 | 读取方式 |
|---|---|---|
| 公开 | `storage/app/public` | `/storage/` 或配置的访问域名 |
| 私密 | `storage/app/private` | 带 Bearer 或登录 Cookie 的附件接口 |

公开目录通过 `go run . storage:link` 建立 `public/storage` 链接，Nginx 可直接托管。`FILESYSTEM_DISK` 用于临时上传与分片；完成附件使用后台上传配置，多主机使用共享磁盘或对象存储。

附件记录上传时的驱动、Bucket 和路径，切换模式后仍按原位置读取。轮换凭证时保留旧 Bucket 的访问权限。前端通过 MD5、SHA-256 预检秒传，服务端复核大小、扩展名与摘要；文件按内容摘要及扩展名存储，支持六种内置分类与自定义分类。

## 缓存管理

业务缓存通过 `Definition` 注册，示例见[业务缓存](backend.md#业务缓存)。

| 策略 | 含义 |
|---|---|
| `TTL` | 临时缓存有效期 |
| `Readable` | 允许在管理页查看 |
| `Editable` | 允许编辑 string 或 hash 单字段 |
| `Clearable` | 纳入顶栏“清理缓存”的扫描范围 |

管理页删除要求 `dawn.cache.delete` 权限，且 Key 属于当前应用、环境与已注册定义；删除权限独立于 `Clearable`。编辑需要 `dawn.cache.update`，保存时保留原 TTL。hash 按游标读取，单字段更新。

系统配置统一复用`ForeverSettings`，每组使用独立 Redis Hash，字段保留 JSON 类型，`__revision` 防止并发旧值回写；同一配置不另建缓存副本。以`__`开头的内部配置组仅供服务端读取，不参与后台查看、编辑、删除或批量清理。验证码、会话、锁等基础设施 Key 默认关闭查看、编辑与顶栏批量清理。

## 日志与监控

`LOG_CHANNEL` 选择日志通道，`APP_LOGGER_FILE` 默认 `storage/logs/dawn.log`，`APP_LOGGER_LEVEL` 控制等级。daily 通道按日期归档，错误日志独立写入 `storage/logs/error/`。

`APP_DEBUG=false` 使用数据库警告级别日志，记录慢查询与错误；开启调试时输出普通 SQL。HTTP 访问日志由 `SERVER_ACCESS_LOG` 控制。

操作日志记录写请求与 JSON 响应，正文上限 16 KiB；认证、账号写入、配置、上传、缓存值与聊天接口省略正文，凭证与文件内容不采集。

服务监控展示当前主机与进程资源，`DAWN_MONITOR_DISK_PATH` 可指定磁盘路径。监控和队列页在可见时自动刷新；数据表统计保留数据库原始口径，缺失指标显示为空。Horizon 页面与只读 API 使用 `dawn.horizon.list` 权限。

## 邮件与地理位置

SMTP 支持 STARTTLS 和 SSL/TLS，默认 UTF-8。测试邮件使用当前配置，结果写入邮件日志。

IP地理位置结合ip2region的IPv4/IPv6与GeoLite2 City、Country、ASN五类离线文件。`dawn:install`准备数据，cron每日03:30更新；也可运行`go run . dawn:geoip`。文件校验后原子替换，失败保留现有库，长期进程自动重载。

使用`GEOIP_DIRECTORY`、`GEOIP_*_URL`与`GEOIP_ENABLED`配置。保存目录、下载源及公开接口见[IP地域查询与更新](geoip.md)。各主机使用独立本地库时分别安排更新，数据来源与许可见[README鸣谢](https://github.com/dawnim/dawn-simple#鸣谢)。
