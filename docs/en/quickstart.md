# Quick start

Use dawn-simple to create an application with administration, chat, and developer tools included.

## Requirements

- Go 1.26+
- Node.js 20.19+ (20.x) or 22.12+, and pnpm
- Redis
- MySQL, MariaDB, PostgreSQL, or SQLite

MariaDB is the default. See [Configuration](configuration.md) for connection options.

## Get the project

```sh
git clone https://github.com/dawnim/dawn-simple.git my-app
cd my-app
cp .env.example .env
go mod download
go run . dawn:rename example.com/my-app
pnpm --dir admin install --frozen-lockfile
```

Replace `my-app` and `example.com/my-app` with your directory name and Go module path. The server extension's `dawn:rename` command uses the final module-path segment as the project name. It updates `go.mod`, Go imports, `config/dawn.go`, application-name defaults, the frontend package name, Supervisor examples, and application names and original cache/session/queue prefixes in `.env.example` and an existing `.env`. Database connections and keys keep their values. Choose the directory name when cloning. Run this before installation; no database or Redis connection is required.

`main` uses released framework dependencies. Use `dev` for framework development; see [Local framework integration](testing.md#local-framework-integration).

`dev` uses Go `replace` and npm `link` for adjacent framework sources. A local `go.work` can also override Go dependency sources.

To integrate the framework into a base application, see [Create a project](create-project.md).

## Configure connections

Create an application database first, then set the connections in `.env`:

```dotenv
DB_CONNECTION=mariadb
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=dawn-admin
DB_USERNAME=your_user
DB_PASSWORD=your_password
REDIS_URL=redis://127.0.0.1:6379/0
```

Generate the application key:

```sh
go run . key:generate
```

The command generates a key and writes it to `APP_KEY` in `.env`. All instances of the same application must use the same key.

## Install and start

```sh
go run . dawn:install
go run . dawn:start local
```

Installation runs migrations, seeds missing base data, and prepares IP databases. Running it again preserves existing accounts and settings.

Open <http://127.0.0.1:5173>. The initial credentials are `admin / 123456`; set `DAWN_ADMIN_USERNAME` and `DAWN_ADMIN_PASSWORD` before initialization to change them. Change the password after your first sign-in.

| Service | Default address |
|---|---|
| Frontend | `http://127.0.0.1:5173` |
| Admin API | `http://127.0.0.1:5123/admin` |
| WebSocket | `ws://127.0.0.1:8851/` |

The browser accesses APIs and WebSocket through Vite's proxy. Frontend changes update immediately; Go changes trigger compilation and service updates. Run `go run . dawn:start local` in a new terminal to take over logs and Ctrl+C control.

## Start developing

Put admin business APIs in `app/admin` and pages in `admin/src`. Read [Architecture](architecture.md), then follow [Backend](backend.md) and [Frontend](frontend.md).

The earlier `dawn:rename` command handles project naming. Upgrade framework capabilities through Go Modules and npm packages. See [Running services](runtime.md) for deployment.
