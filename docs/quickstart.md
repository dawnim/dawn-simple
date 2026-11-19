# 快速开始

使用 dawn-simple 创建应用，系统管理、聊天和开发工具随项目启动。

## 环境

- Go 1.26+
- Node.js 20.19+（20.x）或 22.12+，pnpm
- Redis
- MySQL、MariaDB、PostgreSQL 或 SQLite

默认使用 MariaDB。连接选项见[配置](configuration.md)。

## 获取项目

```sh
git clone https://github.com/dawnim/dawn-simple.git my-app
cd my-app
cp .env.example .env
go mod download
go run . dawn:rename example.com/my-app
pnpm --dir admin install --frozen-lockfile
```

将`my-app`和`example.com/my-app`换成自己的目录名与Go模块路径。`dawn:rename`由server扩展提供，以模块路径最后一段作为项目名，更新`go.mod`、Go源码导入、`config/dawn.go`、应用默认名、前端包名、Supervisor示例，以及`.env.example`和已存在`.env`中的应用名与原项目缓存、会话、队列前缀。数据库连接和密钥保持原值，目录名在克隆时指定。此命令在首次安装前执行，无需先配置数据库或Redis。

`main` 使用已发布的框架依赖；参与框架开发时使用 `dev`，见[框架联调](testing.md#框架联调)。

`dev`通过Go `replace`和npm `link`引用相邻框架源码；本地`go.work`也可能覆盖Go依赖来源。

从基础应用独立接入框架，见[创建空白项目](create-project.md)。

## 配置连接

先创建应用数据库，在 `.env` 填写连接：

```dotenv
DB_CONNECTION=mariadb
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=dawn-admin
DB_USERNAME=your_user
DB_PASSWORD=your_password
REDIS_URL=redis://127.0.0.1:6379/0
```

生成应用密钥：

```sh
go run . key:generate
```

命令自动生成密钥并写入 `.env` 的 `APP_KEY`。同一应用的多个实例使用相同密钥。

## 初始化与启动

```sh
go run . dawn:install
go run . dawn:start local
```

安装命令执行数据库迁移、基础数据填充和 IP 地理库准备。重复安装保留已有账号与配置。

打开 <http://127.0.0.1:5173>。首次账号为 `admin / 123456`，可在初始化前通过 `DAWN_ADMIN_USERNAME`、`DAWN_ADMIN_PASSWORD` 设置。首次登录后修改密码。

| 服务 | 默认地址 |
|---|---|
| 前端 | `http://127.0.0.1:5173` |
| 后台 API | `http://127.0.0.1:5123/admin` |
| WebSocket | `ws://127.0.0.1:8851/` |

浏览器经 Vite 代理访问 API 与 WebSocket。前端支持热更新；修改 Go 后自动编译并更新服务；也可在新终端再次运行 `go run . dawn:start local`，接管日志与 Ctrl+C 控制。

## 开始开发

后台管理业务接口放在 `app/admin`，页面放在 `admin/src`。先阅读[架构与目录](architecture.md)，再按[后端开发](backend.md)和[前端开发](frontend.md)添加功能。

项目命名由前面的`dawn:rename`统一完成。框架能力通过Go Module与npm包升级。部署方式见[服务运行](runtime.md)。
