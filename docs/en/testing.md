# Testing and releases

## Application checks

Run from the dawn-simple root:

```sh
go test -race ./...
go vet ./...
pnpm --dir admin build
```

Tests use temporary SQLite and isolated Redis. For API changes, verify authentication, permissions, scopes, and responses. For page changes, check themes, permission visibility, requests, and interactions.

## Framework checks

From the dawn-server root:

```sh
go test -race ./...
go vet ./...
```

Tests cover authorization, encrypted settings, attachments, cache, CSV, CORS, chat delivery, and process management. WS checks cover acknowledgements and cross-instance recovery/replay. Terminal checks cover control transfer and cleanup.

From the dawn-admin root:

```sh
pnpm install --frozen-lockfile
pnpm test
pnpm lint
pnpm build
npm pack --pack-destination /tmp
```

Keep `src/index.ts` and `types/index.d.ts` synchronized. After changing public components, types, authentication, or menus, install the actual packed artifact in an independent application and build it.

## Database matrix

In dawn-server, explicitly specify a dedicated empty test database:

```sh
DAWN_TEST_DRIVER=pgsql DAWN_TEST_DSN='dedicated-test-database-DSN' \
  go test -run 'Test(Application|DatabaseMatrix|DatabaseMetadata|RESTResources)$' -count=1 ./tests
```

Supported drivers are `mysql`, `mariadb`, and `pgsql`. Tests create fixed accounts and menus and rebuild tables during cleanup. Never point them at a development or production database. Without explicit connection settings, external matrix tests are skipped and ordinary tests use temporary SQLite.

## Local framework integration

Place `dawn-server`, `dawn-admin`, and `dawn-simple` side by side. `main` uses released Go modules and npm packages. `dev` uses Go `replace` and npm `link` to load adjacent sources. After switching to `dev`, confirm those repositories contain the intended sources and install dependencies:

```sh
git switch dev
go mod download
pnpm --dir ../dawn-admin install --frozen-lockfile
pnpm --dir admin install --frozen-lockfile
```

On `dev`, run `pnpm --dir admin dev:lib` to watch the frontend library and `go run . dawn:start local` in another terminal. Go sources come from `replace` or an explicitly active `go.work`; the branch name alone is insufficient. Even `main` loads adjacent dawn-server sources if a local workspace enables them. Frontend linking comes from `link:` in `admin/package.json`; otherwise it uses a released package. Check actual sources with `go env GOWORK` and `realpath admin/node_modules/@dawnim/admin`. Release builds use `main`, disable local `go.work`, and install locked released dependencies.

## Runtime acceptance

| Area | Checks |
|---|---|
| Services | HTTP health, WS authentication, queue consumption, scheduling |
| Processes | SIGTERM cleanup, failure recovery, manual stop, failed reload retaining old processes |
| Frontend | CORS preflight, Bearer, iframe, downloads, permissions, page overrides |
| Realtime | Bidirectional delivery, unread/read state, reconnect recovery, duplicate merging |
| Deployment | Shared storage, Redis recovery, SMTP, measured capacity |

## Documentation site

From the dawn-simple root:

```sh
git submodule update --init docs/prismgo
pnpm --dir docs install --frozen-lockfile
pnpm --dir docs dev
```

Build with `pnpm --dir docs build` and preview with `pnpm --dir docs preview`. Output goes to `docs/.vitepress/dist`. For the example website, use `pnpm --dir docs build:site`, which writes `public/docs` with `/docs/` as its base.

After documentation changes, run `pnpm --dir docs build:all` to update both outputs. The website and GitHub Pages share `docs/.vitepress/config.mts`. Dawn sidebars and previous/next links come directly from the [Chinese index](../README.md) and [English index](README.md). Update both indexes when adding, removing, or renaming a guide; no separate menu list is needed. PrismGo indexes come from the submodule's two READMEs. GitHub Actions checks both builds and uploads only the Pages output; deploy `public/docs` with the website.

Dawn English sources live in `docs/en`, with the same filenames as Chinese documents. Update both languages together. VitePress provides language-specific navigation and search. The website and documentation share a language preference; see [Shared website language](frontend.md#shared-website-language). Both builds check internal page links.

GitHub Pages publishes static docs under `/dawn-simple/`. Its language menu switches between Chinese and English and remembers the selection. Browsers isolate `localStorage` by origin, so Pages and the official website keep separate language preferences.

PrismGo documentation is included through the official `docs/prismgo` submodule. Update it locally with `git submodule update --remote docs/prismgo`. Dawn's build maps both upstream languages and retains old URL redirects without modifying the submodule.

In Settings → Pages → Source, select **GitHub Actions**. Pushes to `main` build and publish documentation at `https://dawnim.github.io/dawn-simple/`. PRs run build checks only. You can also run the Docs workflow manually.

## Upgrades and releases

Pin framework versions and synchronize lockfiles. Back up SQL, run `dawn:install --force`, rebuild the frontend, and update services. Menu and permission keys stay stable. Drain old jobs before changing task types or encoding.

dawn-server publishes Go Modules with version tags. CI runs Go checks and the database matrix before creating a Release. Example from dawn-server:

```sh
git tag -a v0.1.2 -m v0.1.2
git push origin main v0.1.2
```

dawn-admin tags must match `package.json`. After checks, CI publishes through npm trusted publishing and creates a Release. From dawn-admin:

```sh
npm version patch -m "chore: release v%s"
git push origin main --follow-tags
```

After publishing, verify dawn-simple types, builds, and main business paths using the released Go Module and npm package.
