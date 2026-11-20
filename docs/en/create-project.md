# Create a project

Install dawn-server and dawn-admin independently to keep a standard application structure with built-in administration. Your project owns its business models, menus, APIs, and pages.

## Create the base application

See [Requirements](quickstart.md#requirements). Generate the base structure with the application installer:

```sh
go run github.com/prismgo/installer/cmd/prismgo@latest new my-admin --module example.com/my-admin --no-install
cd my-admin
go get github.com/dawnim/dawn-server@latest
```

Replace `example.com/my-admin` with your module name and use it consistently in imports and `config/dawn.go`. Remove the base application's welcome page and example user migration, then create the business directories:

```sh
rm -rf app/http routes
rm database/migrations/*_create_users_table.go
mkdir -p app/admin/controllers app/admin/middleware app/models app/constants database/seeders admin/src
```

Keep the installer's `app/providers`, `app/cmd`, `app/schedule`, and core configuration files.

## Integrate the backend

Replace the following entry points.

`main.go`:

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

`bootstrap/app.go`:

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
    "github.com/dawnim/dawn-server/support/clientip"
    "github.com/prismgo/framework/encryption"
    "github.com/prismgo/framework/foundation"
)

func NewApplication(basePath ...string) *foundation.Application {
    return foundation.Configure(basePath...).
        WithExtensionProviders(encryption.ServiceProvider{}, dawn.ServiceProvider{Configure: admin.Configure}).
        WithProviders(Providers()...).
        WithMiddleware(func(m *foundation.Middleware) {
            m.Prepend(clientip.Register)
        }).
        WithRouting(func(r *foundation.Routing) {
            r.MigrationPaths("database/migrations")
            r.SeedPaths("database/seeders")
            r.Commands(cmd.CommandFactories()...)
            r.Schedules(schedule.Register)
            r.Routes(dawn.RegisterRoutes)
        }).Create()
}
```

`app/admin/menu.go`:

```go
package admin

import sdk "github.com/dawnim/dawn-server/app/admin"

func Configure(o *sdk.Options) {
    // Append admin menus, permissions, and routes here.
}
```

## Configure connections and queues

Create `config/dawn.go` and `config/horizon.go`, and replace `config/database.go`. Keep the other configuration files supplied by the base application.

`config/dawn.go`:

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

`config/database.go`:

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

`config/horizon.go`:

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
			// Serial tasks use one worker and run once per deployment.
			"supervisor": map[string]any{
				"connection": "redis", "queue": []string{constants.QueueDefault}, "balance": "false",
				"min_processes": 1, "max_processes": 1, "tries": 3, "timeout": Env("HORIZON_WORKER_TIMEOUT", 60), "sleep": 1,
			},
			// Autoscale concurrent workers; add business queues to this group.
			"multiples": map[string]any{
				"connection": "redis", "queue": []string{constants.QueueParallel}, "balance": "auto",
				"min_processes": 1, "max_processes": Env("HORIZON_CONCURRENCY", 4),
				"tries": 3, "timeout": Env("HORIZON_WORKER_TIMEOUT", 60), "sleep": 1, "auto_scaling_strategy": "size",
			},
		}
		// Enable the serial group only on the primary Horizon instance.
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

For AI support and indexing jobs, use the current timeout and retry settings from dawn-simple's [Horizon](https://github.com/dawnim/dawn-simple/blob/main/config/horizon.go) and [queue](https://github.com/dawnim/dawn-simple/blob/main/config/queue.go) configuration.

## Register installation migrations and seeders

The extension provides built-in schemas and base data. Register them in the application.

`database/migrations/202611030000_install_dawn.go`:

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

`database/seeders/database_seeder.go`:

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

## Integrate the frontend

Create the following files.

`admin/package.json`:

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

`admin/index.html`:

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

`admin/src/main.ts`:

```ts
import { createAdmin } from '@dawnim/admin'
import '@dawnim/admin/style.css'

createAdmin({ apiBase: import.meta.env.VITE_ADMIN_PREFIX || '/admin' }).mount('#app')
```

`admin/vite.config.ts`:

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

`admin/tsconfig.json`:

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

Install frontend dependencies:

```sh
pnpm --dir admin add @dawnim/admin vue vue-router pinia element-plus vue-i18n
pnpm --dir admin add -D vite@^7 @vitejs/plugin-vue@^6 typescript@~5.6 vue-tsc@~2.1
```

Add `/admin/node_modules/`, `/admin/dist/`, and `/storage/framework/` to `.gitignore`. Commit the generated Go and pnpm lockfiles.

## Initialize and run

The installer creates `.env`. Update its corresponding settings and create the MariaDB database first:

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

Connection names must match cache and queue configuration. Give each application its own database and namespace. Update the prefixes when changing `APP_ENV` between environments.

```sh
go mod tidy
go run . key:generate
go run . dawn:install
go run . dawn:start local
```

Open `http://127.0.0.1:5173`. The initial credentials are `admin / 123456`. The extension loads administration, permissions, chat, and developer tools. Start adding business features in `app/admin` and `admin/src`.

## Project structure

The resulting directories are:

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

Continue with [Backend](backend.md) and [Frontend](frontend.md). For deployment, see [Running services](runtime.md#production-deployment).
