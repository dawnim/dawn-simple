# 创建空白项目

独立安装 dawn-server 与 dawn-admin，保留标准应用结构和后台基础能力。业务模型、菜单、接口与页面由自己的项目维护。

## 创建基础应用

环境要求见[快速开始](quickstart.md#环境)。使用应用安装器生成基础骨架：

```sh
go run github.com/prismgo/installer/cmd/prismgo@latest new my-admin --module example.com/my-admin --no-install
cd my-admin
go get github.com/dawnim/dawn-server@latest
```

将 `example.com/my-admin` 换成自己的模块名，下文导入与 `config/dawn.go` 保持一致。清理基础应用的欢迎页和用户表示例，再创建业务目录：

```sh
rm -rf app/http routes
rm database/migrations/*_create_users_table.go
mkdir -p app/admin/controllers app/admin/middleware app/models app/constants database/seeders admin/src
```

`app/providers`、`app/cmd`、`app/schedule` 与核心 `config` 文件由安装器创建，保留使用。

## 接入服务端

替换以下入口：

`main.go`：

```go
package main

import (
    "context"
    "os"

    "example.com/my-admin/bootstrap"
    dawn "github.com/dawnim/dawn-server"
    "github.com/prismgo/framework/console"
)

func main() {
    if err := dawn.HandleCommand(context.Background(), bootstrap.NewApplication(), os.Args); err != nil {
        console.Exit(err.Error())
    }
}
```

`bootstrap/app.go`：

```go
package bootstrap

import (
    "example.com/my-admin/app/admin"
    "example.com/my-admin/app/cmd"
    "example.com/my-admin/app/schedule"
    _ "example.com/my-admin/config"
    _ "example.com/my-admin/database/migrations"
    _ "example.com/my-admin/database/seeders"
    dawn "github.com/dawnim/dawn-server"
    "github.com/prismgo/framework/encryption"
    "github.com/prismgo/framework/foundation"
)

func NewApplication(basePath ...string) *foundation.Application {
    return foundation.Configure(basePath...).
        WithExtensionProviders(encryption.ServiceProvider{}, dawn.ServiceProvider{Configure: admin.Configure}).
        WithProviders(Providers()...).
        WithRouting(func(r *foundation.Routing) {
            r.MigrationPaths("database/migrations")
            r.SeedPaths("database/seeders")
            r.Commands(cmd.CommandFactories()...)
            r.Schedules(schedule.Register)
            r.Routes(dawn.RegisterRoutes)
        }).Create()
}
```

`app/admin/menu.go`：

```go
package admin

import sdk "github.com/dawnim/dawn-server/app/admin"

func Configure(o *sdk.Options) {
    // 在此追加后台管理业务菜单、权限和路由
}
```

## 配置连接与队列

创建 `config/dawn.go`、`config/horizon.go`，替换 `config/database.go`。其他配置文件保留基础应用的定义。

`config/dawn.go`：

```go
package config

func init() {
    Add("dawn", func() map[string]any {
        return map[string]any{
            "module": "example.com/my-admin",
            "admin_prefix": Env("ADMIN_PREFIX", "/admin"),
            "ws_addr": Env("WS_ADDR", ":8851"),
            "ws_path": Env("WS_PATH", "/"),
            "allowed_origins": Env("ALLOWED_ORIGINS", "*"),
            "local": map[string]any{
                "directory": Env("DAWN_FRONTEND_DIRECTORY", "admin"),
                "port": Env("DAWN_FRONTEND_PORT", 5173),
                "watch": Env("DAWN_GO_WATCH", true),
            },
        }
    })
}
```

`config/database.go`：

```go
package config

func init() {
    Add("database", func() map[string]any {
        return map[string]any{
            "default": Env("DB_CONNECTION", "mariadb"),
            "connections": map[string]any{
                "mariadb": map[string]any{
                    "driver": "mysql",
                    "host": Env("DB_HOST", "127.0.0.1"),
                    "port": Env("DB_PORT", 3306),
                    "database": Env("DB_DATABASE", "my_admin"),
                    "username": Env("DB_USERNAME", "root"),
                    "password": Env("DB_PASSWORD", ""),
                    "charset": "utf8mb4", "parse_time": true,
                    "max_open_conns": 30, "max_idle_conns": 10,
                },
            },
            "redis": map[string]any{
                "client": "go",
                "default": map[string]any{
                    "url": Env("REDIS_URL", "redis://127.0.0.1:6379/0"),
                },
            },
        }
    })
}
```

`config/horizon.go`：

```go
package config

import (
	"fmt"

	"github.com/dawnim/dawn-server/constants"
)

func init() {
	Add("horizon", func() map[string]any {
		prefix := fmt.Sprintf("%v:%v", Env("APP_NAME", "my-admin"), Env("APP_ENV", "local"))
		supervisors := map[string]any{
			// 串行任务：一个worker按队列优先级处理；同一部署只运行一份。
			"supervisor": map[string]any{
				"connection": "redis", "queue": []string{constants.QueueDefault}, "balance": "false",
				"min_processes": 1, "max_processes": 1, "tries": 3, "timeout": Env("HORIZON_WORKER_TIMEOUT", 60), "sleep": 1,
			},
			// 并发任务：按积压量调整worker数量，新并发队列追加到此处。
			"multiples": map[string]any{
				"connection": "redis", "queue": []string{constants.QueueParallel}, "balance": "auto",
				"min_processes": 1, "max_processes": Env("HORIZON_CONCURRENCY", 4),
				"tries": 3, "timeout": Env("HORIZON_WORKER_TIMEOUT", 60), "sleep": 1, "auto_scaling_strategy": "size",
			},
		}
		// 增加Horizon实例时，只有主实例启用串行组。
		if !Env("HORIZON_SERIAL", true).(bool) {
			delete(supervisors, "supervisor")
		}
		return map[string]any{
			"path":                  Env("HORIZON_PATH", fmt.Sprintf("%v/horizon", Env("ADMIN_PREFIX", "/admin"))),
			"heartbeat_ttl_seconds": Env("HORIZON_HEARTBEAT_TTL", 60),
			"use":                   "redis", "prefix": Env("HORIZON_PREFIX", prefix+":horizon:"),
			"environments": map[string]any{"*": supervisors},
		}
	})
}
```

## 注册安装迁移与基础数据

内置表结构与初始化数据由扩展提供，应用登记迁移与 Seeder：

`database/migrations/202611030000_install_dawn.go`：

```go
package migrations

import (
    sdk "github.com/dawnim/dawn-server/database/migrations"
    "github.com/prismgo/framework/database"
    "gorm.io/gorm"
)

func init() {
    database.RegisterMigration(
        func(db *gorm.DB) error { return sdk.Up(db) },
        func(db *gorm.DB) error { return sdk.Down(db) },
    )
}
```

`database/seeders/database_seeder.go`：

```go
package seeders

import (
    sdk "github.com/dawnim/dawn-server/database/seeders"
    "github.com/prismgo/framework/database"
    "gorm.io/gorm"
)

func init() { database.RegisterSeeder(Seed) }

func Seed(db *gorm.DB) error { return sdk.Seed(db) }
```

## 接入前端

创建以下文件：

`admin/package.json`：

```json
{
  "name": "my-admin-ui",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite --host 127.0.0.1",
    "build": "vue-tsc --noEmit && vite build"
  }
}
```

`admin/index.html`：

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>My Admin</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
```

`admin/src/main.ts`：

```ts
import { createAdmin } from '@dawnim/admin'
import '@dawnim/admin/style.css'

createAdmin({ apiBase: import.meta.env.VITE_ADMIN_PREFIX || '/admin' }).mount('#app')
```

`admin/vite.config.ts`：

```ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const httpTarget = process.env.DAWN_HTTP_TARGET || 'http://127.0.0.1:5123'
const wsTarget = process.env.DAWN_WS_TARGET || 'ws://127.0.0.1:8851'
const adminPrefix = process.env.DAWN_ADMIN_PREFIX || '/admin'
const wsPath = process.env.DAWN_WS_PATH || '/'

export default defineConfig({
  plugins: [vue()],
  define: { 'import.meta.env.VITE_ADMIN_PREFIX': JSON.stringify(adminPrefix) },
  publicDir: '../public',
  resolve: { dedupe: ['vue', 'vue-router', 'pinia', 'element-plus', 'vue-i18n'] },
  server: {
    host: '127.0.0.1',
    strictPort: true,
    proxy: {
      [adminPrefix]: httpTarget,
      '/api': httpTarget,
      '/ws': { target: wsTarget, ws: true, rewrite: () => wsPath }
    }
  }
})
```

`admin/tsconfig.json`：

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "strict": true,
    "skipLibCheck": true,
    "lib": ["ES2022", "DOM"],
    "types": ["vite/client"]
  },
  "include": ["src/**/*.ts", "src/**/*.vue"]
}
```

安装前端依赖：

```sh
pnpm --dir admin add @dawnim/admin vue vue-router pinia element-plus vue-i18n
pnpm --dir admin add -D vite@^7 @vitejs/plugin-vue@^6 typescript@~5.6 vue-tsc@~2.1
```

在 `.gitignore` 中追加 `/admin/node_modules/`、`/admin/dist/`、`/storage/framework/`，提交生成的 Go 与 pnpm 锁文件。

## 初始化与运行

安装器已创建 `.env`。修改其中对应项，先创建 MariaDB 数据库：

```dotenv
APP_NAME=my-admin
APP_ENV=local
APP_KEY=
APP_DEBUG=false
APP_LOCALE=zh
APP_URL=http://127.0.0.1:5123
SERVER_HOST=127.0.0.1
SERVER_PORT=5123
DB_CONNECTION=mariadb
DB_DATABASE=my_admin
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USERNAME=your_user
DB_PASSWORD=your_password
REDIS_URL=redis://127.0.0.1:6379/0
CACHE_STORE=redis
CACHE_PREFIX=my-admin:local
SESSION_DRIVER=redis
SESSION_PREFIX=my-admin:local:session
QUEUE_CONNECTION=redis
QUEUE_FAILED_DRIVER=redis
QUEUE_FAILED_PREFIX=my-admin:local:failed
QUEUE_BATCHING_DRIVER=redis
QUEUE_BATCHING_PREFIX=my-admin:local:batches
QUEUE_RESTART_KEY=my-admin:local:queue:restart
REDIS_QUEUE_PREFIX=my-admin:local:queue
```

连接名与缓存、队列配置保持一致。每个独立应用使用自己的数据库和命名空间，多环境修改 `APP_ENV` 时同步调整上述前缀。

```sh
go mod tidy
go run . key:generate
go run . dawn:install
go run . dawn:start local
```

打开 `http://127.0.0.1:5173`，初始账号为 `admin / 123456`。系统管理、权限、聊天和开发工具由扩展加载，业务页面从 `app/admin` 与 `admin/src` 开始添加。

## 项目目录

接入后的主要目录：

```text
my-admin/
├── main.go
├── bootstrap/
│   ├── app.go
│   └── providers.go
├── app/
│   ├── admin/
│   │   ├── menu.go
│   │   ├── controllers/
│   │   └── middleware/
│   ├── cmd/
│   ├── constants/
│   ├── models/
│   ├── providers/
│   └── schedule/
├── config/
│   ├── app.go
│   ├── database.go
│   ├── dawn.go
│   ├── horizon.go
│   └── ...
├── database/
│   ├── migrations/
│   └── seeders/
├── admin/
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── src/main.ts
├── public/
└── storage/
```

业务开发见[后端开发](backend.md)与[前端开发](frontend.md)，部署见[服务运行](runtime.md#生产部署)。
