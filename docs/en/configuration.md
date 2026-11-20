# Configuration

Startup settings live in `config/*.go` and can be overridden through `.env`. Register admin-editable settings through `Options.Settings`; SQL stores one JSON value per group.

## Startup configuration

Register settings with `config.Add` and `config.Env`, and blank-import the configuration package during startup. See [.env.example](https://github.com/dawnim/dawn-simple/blob/main/.env.example) for all environment variables.

| File | Contents |
|---|---|
| `config/app.go` | Application name, environment, key, timezone, and HTTP server |
| `config/database.go` | SQL and Redis connections |
| `config/dawn.go` | Admin prefix, WS, authentication, GeoIP, and local runtime |
| `config/cache.go`, `session.go` | Cache and framework Session |
| `config/queue.go`, `horizon.go` | Queues, failed jobs, batches, and workers |
| `config/filesystem.go` | Local and cloud disks |
| `config/logging.go`, `encoding.go` | Logs and encoding |

Cache, queues, Session, and Horizon reference named `database` connections. The application name and environment form the cache namespace; keep them consistent across instances.

## SQL and Redis

MariaDB is the default. Select a connection through `DB_CONNECTION`:

| Connection | Configuration |
|---|---|
| `mysql`, `mariadb` | `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD`; `DB_DSN` overrides the connection string |
| `pgsql` | `DB_DSN`, such as `host=127.0.0.1 port=5432 user=postgres dbname=dawn_admin sslmode=disable` |
| `sqlite` | File in `DB_DATABASE`, such as `storage/database.sqlite?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)` |

Create SQLite's parent directory first; its default pool size is 1. For other connections, adjust `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, and `DB_CONN_MAX_IDLE_TIME`. Keep TLS and driver options under their connection definitions.

Prefer Redis URLs:

```text
redis://127.0.0.1:6379/0
redis://:password@127.0.0.1:6379/0
redis://username:password@127.0.0.1:6379/0
rediss://default:password@cache.example.com:6379/0
```

Set `REDIS_URL`. For a separate cache connection, set `REDIS_CACHE_URL` and `CACHE_REDIS_CONNECTION=cache`. Encode special characters in URLs, such as `%40` for `@`. `rediss://` enables TLS and certificate verification; use the provider's DNS endpoint.

Alternatively, use `REDIS_HOST`, `REDIS_PORT`, `REDIS_USERNAME`, `REDIS_PASSWORD`, and `REDIS_MAIN_DB`. Nonempty individual settings override the corresponding URL fields. The current client supports single-node or Cluster Mode Disabled instances.

## CORS and proxies

| Setting | Default and purpose |
|---|---|
| `SERVER_HOST`, `SERVER_PORT` | HTTP listener, `127.0.0.1:5123` |
| `WS_ADDR`, `WS_PATH` | WS listener, `:8851` and `/` |
| `ADMIN_PREFIX` | Admin API prefix, `/admin` |
| `ALLOWED_ORIGINS` | `*`; separate explicit full origins with spaces |
| `SERVER_CLIENT_IP_MODE` | `trusted` or `simple`, defaults to `trusted` |
| `SERVER_CLIENT_IP_SIMPLE_HEADERS` | Comma-separated header priority for simple mode |
| `SERVER_CLIENT_IP_HEADERS` | Comma-separated trusted proxy header priority |
| `SERVER_CLOUDFLARE_PROXIES`, `SERVER_CLOUDFRONT_PROXIES` | Enable automatic official CDN range updates, default `false` |
| `SERVER_TRUSTED_PROXIES` | Trusted proxy IPs or networks, comma-separated |
| `SERVER_SHUTDOWN_TIMEOUT` | In-flight shutdown wait, `15s` |

In simple mode the first public address from the configured headers is used, followed by the direct address and private-address fallback. Clients can forge these headers. Trusted mode starts at the direct peer and walks the first configured forwarding header from right to left until the first untrusted address. Invalid entries stop the walk. Configure every proxy between the application and CDN, including a local Nginx address, in `SERVER_TRUSTED_PROXIES`. CDN networks are downloaded by the Go process on startup and daily, with the last successful list cached under `storage/app/clientip`. Until a list is available, only explicitly configured proxies are trusted. Restrict direct origin access and ensure trusted proxies preserve the forwarding chain. Restart the service after changing environment settings.

Nginx examples for the site and admin domains are available in `deploy/nginx`. The `/ws` service has a separate listener and does not currently use client IP for its own processing.

HTTP and WS share allowed origins. Wildcard mode uses Bearer authentication; explicit allowlists permit credentials from those origins. Same-origin iframes and downloads use the login Cookie. APIs always check identity and permissions.

## System settings

Append business settings in `Configure`:

```go
o.Settings = append(o.Settings, sdk.SettingDefinition{
	Key: "business.order_timeout", Title: "订单超时分钟数", Group: "业务配置",
	Kind: "number", Input: "number", Default: 30,
})
```

Run `go run . db:seed` to fill missing groups and fields. Read values through `a.Setting(ctx, "business.order_timeout")`. SQL stores one JSON row per group. Code declares types, controls, options, and descriptions; existing values are preserved.

Site name, logo, copyright, description, and registration details appear in the interface. `GET /admin/site` returns public display fields, application timezone, and server time. `PATCH /admin/auth/profile` updates submitted profile fields. Logs are scoped to the current account.

`APP_KEY` is `base64:` followed by 32 random bytes in Base64. Set `Secret: true` for encrypted values; responses mask them, and blank edits preserve their existing values. Plan data conversion and retain required old keys when rotating encryption keys.

## File storage

Upload settings support local storage, Alibaba Cloud OSS, Qiniu, Tencent COS, and Amazon S3. Prepare the bucket, endpoint, region, credentials, and access domain. Qiniu and COS use S3-compatible endpoints; S3 supports Path Style.

| Access | Local directory | Read path |
|---|---|---|
| Public | `storage/app/public` | `/storage/` or the configured domain |
| Private | `storage/app/private` | Attachment API with Bearer or login Cookie |

Run `go run . storage:link` to link `public/storage`; Nginx can serve it directly. `FILESYSTEM_DISK` handles temporary uploads and chunks. Completed attachments use admin upload settings. Use shared disks or object storage across hosts.

Attachments retain their original driver, bucket, and path when settings change. Preserve access to old buckets during credential rotation. The frontend checks MD5 and SHA-256 before deduplicated uploads; the server verifies size, extension, and digests. Files are stored by content digest and extension. Six built-in categories and custom categories are supported.

## Cache management

Register business keys through `Definition`; see [Business cache](backend.md#business-cache).

| Policy | Meaning |
|---|---|
| `TTL` | Temporary-cache lifetime |
| `Readable` | Visible in cache management |
| `Editable` | Editable string or individual hash field |
| `Clearable` | Included in header Clear cache scans |

Deletion requires `dawn.cache.delete` and a key belonging to the current application, environment, and registered definition; deletion is independent of `Clearable`. Editing requires `dawn.cache.update` and preserves TTL. Hashes use cursor reads and individual field updates.

All settings reuse `ForeverSettings`, with one Redis Hash per group and no separate cache copies of the same configuration. Fields retain JSON types, and `__revision` prevents stale concurrent writes. Internal groups starting with `__` are server-only and excluded from admin viewing, editing, deletion, and bulk clearing. Captchas, sessions, locks, and other infrastructure keys disable viewing, editing, and header bulk clearing by default.

## Logs and monitoring

`LOG_CHANNEL` selects the channel. `APP_LOGGER_FILE` defaults to `storage/logs/dawn.log`; `APP_LOGGER_LEVEL` controls severity. The daily channel rotates by date; errors also go to `storage/logs/error/`.

`APP_DEBUG=false` uses database warning-level logs for slow queries and errors. Debug mode also logs ordinary SQL. `SERVER_ACCESS_LOG` controls HTTP access logs.

Operation logs capture write requests and JSON responses up to 16 KiB. Authentication, account writes, settings, uploads, cache values, and chat omit bodies; credentials and file contents are not collected.

Server monitoring shows current host and process resources. `DAWN_MONITOR_DISK_PATH` selects the disk path. Monitoring and queue pages refresh while visible. Table statistics retain database semantics; unavailable metrics stay empty. Horizon pages and read-only APIs require `dawn.horizon.list`.

## Mail and geolocation

SMTP supports STARTTLS and SSL/TLS, with UTF-8 by default. Test mail uses current settings and records its outcome in mail logs.

GeoIP combines local ip2region IPv4/IPv6 data with GeoLite2 City, Country, and ASN. `dawn:install` prepares the files; cron updates them daily at 03:30. Run `go run . dawn:geoip` to update manually. Downloads are validated and atomically replaced; failures preserve the existing files, and long-running processes reload changes.

Use `GEOIP_DIRECTORY`, the `GEOIP_*_URL` settings, and `GEOIP_ENABLED`. See [IP geolocation](geoip.md) for paths, sources, and public APIs. Arrange updates on every host with separate local data. Data sources and licenses are listed in [Acknowledgements](https://github.com/dawnim/dawn-simple#鸣谢).
