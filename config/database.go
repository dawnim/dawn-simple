package config

func init() {
	Add("database", func() map[string]interface{} {
		return map[string]interface{}{
			"default": Env("DB_CONNECTION", "mysql"),
			"connections": map[string]interface{}{
				// SQLite 文件连接：storage/database.sqlite?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)
				"sqlite": map[string]interface{}{
					"driver":             "sqlite",
					"database":           Env("DB_DATABASE", "storage/database.sqlite?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"),
					"max_open_conns":     Env("DB_MAX_OPEN_CONNS", 1),
					"max_idle_conns":     Env("DB_MAX_IDLE_CONNS", 1),
					"conn_max_lifetime":  Env("DB_CONN_MAX_LIFETIME", "1h"),
					"conn_max_idle_time": Env("DB_CONN_MAX_IDLE_TIME", "10m"),
				},
				// MySQL/MariaDB：分项配置，或用 DB_DSN 覆盖连接串。
				// DSN: user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=true&loc=UTC
				"mysql": map[string]interface{}{
					"driver":             "mysql",
					"host":               Env("DB_HOST", "127.0.0.1"),
					"port":               Env("DB_PORT", 3306),
					"database":           Env("DB_DATABASE", "dawn-admin"),
					"username":           Env("DB_USERNAME", "root"),
					"password":           Env("DB_PASSWORD", "123456"),
					"charset":            Env("DB_CHARSET", "utf8mb4"),
					"timezone":           Env("DB_TIMEZONE", ""),
					"collation":          Env("DB_COLLATION", ""),
					"parse_time":         Env("DB_PARSE_TIME", true),
					"loc":                Env("DB_LOC", "Local"),
					"unix_socket":        Env("DB_UNIX_SOCKET", ""),
					"prefix":             Env("DB_PREFIX", ""),
					"strict":             Env("DB_STRICT", false),
					"modes":              Env("DB_MODES", ""),
					"isolation_level":    Env("DB_ISOLATION_LEVEL", ""),
					"engine":             Env("DB_ENGINE", ""),
					"dsn":                Env("DB_DSN", ""),
					"max_open_conns":     Env("DB_MAX_OPEN_CONNS", 30),
					"max_idle_conns":     Env("DB_MAX_IDLE_CONNS", 10),
					"conn_max_lifetime":  Env("DB_CONN_MAX_LIFETIME", "1h"),
					"conn_max_idle_time": Env("DB_CONN_MAX_IDLE_TIME", "10m"),
					"ssl": map[string]interface{}{
						"ca":                   Env("DB_SSL_CA", ""),
						"cert":                 Env("DB_SSL_CERT", ""),
						"key":                  Env("DB_SSL_KEY", ""),
						"insecure_skip_verify": Env("DB_SSL_INSECURE_SKIP_VERIFY", false),
					},
					"options": map[string]interface{}{
						"allow_native_passwords": Env("DB_ALLOW_NATIVE_PASSWORDS", false),
						"timeout":                Env("DB_OPTION_TIMEOUT", ""),
						"read_timeout":           Env("DB_OPTION_READ_TIMEOUT", ""),
						"write_timeout":          Env("DB_OPTION_WRITE_TIMEOUT", ""),
					},
				},
				// MariaDB 使用同一 MySQL 驱动与连接选项。
				"mariadb": map[string]interface{}{
					"driver":             "mysql",
					"host":               Env("DB_HOST", "127.0.0.1"),
					"port":               Env("DB_PORT", 3306),
					"database":           Env("DB_DATABASE", "dawn-admin"),
					"username":           Env("DB_USERNAME", "root"),
					"password":           Env("DB_PASSWORD", "123456"),
					"charset":            Env("DB_CHARSET", "utf8mb4"),
					"timezone":           Env("DB_TIMEZONE", ""),
					"collation":          Env("DB_COLLATION", ""),
					"parse_time":         Env("DB_PARSE_TIME", true),
					"loc":                Env("DB_LOC", "Local"),
					"unix_socket":        Env("DB_UNIX_SOCKET", ""),
					"prefix":             Env("DB_PREFIX", ""),
					"strict":             Env("DB_STRICT", false),
					"modes":              Env("DB_MODES", ""),
					"isolation_level":    Env("DB_ISOLATION_LEVEL", ""),
					"engine":             Env("DB_ENGINE", ""),
					"dsn":                Env("DB_DSN", ""),
					"max_open_conns":     Env("DB_MAX_OPEN_CONNS", 30),
					"max_idle_conns":     Env("DB_MAX_IDLE_CONNS", 10),
					"conn_max_lifetime":  Env("DB_CONN_MAX_LIFETIME", "1h"),
					"conn_max_idle_time": Env("DB_CONN_MAX_IDLE_TIME", "10m"),
					"ssl": map[string]interface{}{
						"ca":                   Env("DB_SSL_CA", ""),
						"cert":                 Env("DB_SSL_CERT", ""),
						"key":                  Env("DB_SSL_KEY", ""),
						"insecure_skip_verify": Env("DB_SSL_INSECURE_SKIP_VERIFY", false),
					},
					"options": map[string]interface{}{
						"allow_native_passwords": Env("DB_ALLOW_NATIVE_PASSWORDS", true),
						"timeout":                Env("DB_OPTION_TIMEOUT", ""),
						"read_timeout":           Env("DB_OPTION_READ_TIMEOUT", ""),
						"write_timeout":          Env("DB_OPTION_WRITE_TIMEOUT", ""),
					},
				},
				// PostgreSQL 扩展驱动使用 DSN。
				// DSN: host=127.0.0.1 port=5432 user=postgres password=password dbname=database sslmode=disable
				// URL: postgres://user:password@127.0.0.1:5432/database?sslmode=disable
				"pgsql": map[string]interface{}{
					"driver":             "pgsql",
					"dsn":                Env("DB_DSN", "host=127.0.0.1 port=5432 user=postgres dbname=dawn-admin sslmode=disable"),
					"max_open_conns":     Env("DB_MAX_OPEN_CONNS", 30),
					"max_idle_conns":     Env("DB_MAX_IDLE_CONNS", 10),
					"conn_max_lifetime":  Env("DB_CONN_MAX_LIFETIME", "1h"),
					"conn_max_idle_time": Env("DB_CONN_MAX_IDLE_TIME", "10m"),
				},
			},

			// URL 与分项配置择一使用；分项字段非空时会覆盖 URL 中的对应值。
			// 无密码：redis://127.0.0.1:6379/0
			// 有密码：redis://:password@127.0.0.1:6379/0
			// ACL：redis://username:password@127.0.0.1:6379/0
			// TLS 无认证：rediss://cache.example.com:6379/0
			// TLS 有密码：rediss://:password@cache.example.com:6379/0
			// AWS TLS/ACL：rediss://default:password@primary.example.cache.amazonaws.com:6379/0
			// TLS 使用 DNS 端点并校验证书；密码中的特殊字符须 URL 编码，如 @ 写成 %40。
			"redis": map[string]interface{}{
				"client":  Env("REDIS_CLIENT", "go"),
				"options": map[string]interface{}{},
				// 不设 URL 和地址时，客户端默认连接 127.0.0.1:6379，数据库 0。
				"default": map[string]interface{}{
					"url":           Env("REDIS_URL", ""),
					"scheme":        Env("REDIS_SCHEME", ""),
					"host":          Env("REDIS_HOST", ""),
					"port":          Env("REDIS_PORT", ""),
					"username":      Env("REDIS_USERNAME", ""),
					"password":      Env("REDIS_PASSWORD", ""),
					"database":      Env("REDIS_MAIN_DB", ""),
					"name":          Env("REDIS_NAME", ""),
					"timeout":       Env("REDIS_TIMEOUT", ""),
					"read_timeout":  Env("REDIS_READ_TIMEOUT", ""),
					"write_timeout": Env("REDIS_WRITE_TIMEOUT", ""),
					"max_retries":   Env("REDIS_MAX_RETRIES", ""),
				},
				"cache": map[string]interface{}{
					"url":           Env("REDIS_CACHE_URL", ""),
					"scheme":        Env("REDIS_CACHE_SCHEME", ""),
					"host":          Env("REDIS_HOST", ""),
					"port":          Env("REDIS_PORT", ""),
					"username":      Env("REDIS_USERNAME", ""),
					"password":      Env("REDIS_PASSWORD", ""),
					"database":      Env("REDIS_CACHE_DB", ""),
					"name":          Env("REDIS_CACHE_NAME", ""),
					"timeout":       Env("REDIS_CACHE_TIMEOUT", ""),
					"read_timeout":  Env("REDIS_CACHE_READ_TIMEOUT", ""),
					"write_timeout": Env("REDIS_CACHE_WRITE_TIMEOUT", ""),
					"max_retries":   Env("REDIS_CACHE_MAX_RETRIES", ""),
				},
			},
		}
	})
}
