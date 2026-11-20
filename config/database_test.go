package config

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	dawn "github.com/dawnim/dawn-server"
	mysql "github.com/go-sql-driver/mysql"
	cfg "github.com/prismgo/framework/config"
	"github.com/prismgo/framework/database"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/redis"
	goredis "github.com/redis/go-redis/v9"
)

func TestRedisConfiguration(t *testing.T) {
	for _, test := range []struct {
		name, url, cacheURL, address, user, password string
		db                                           int
		tls                                          bool
		fields                                       map[string]string
	}{
		{name: "local", address: "127.0.0.1:6379"},
		{name: "url", url: "redis://cache.example.com:6380/5", address: "cache.example.com:6380", db: 5},
		{name: "password", url: "redis://:p%40ssword@cache.example.com:6380/5", address: "cache.example.com:6380", password: "p@ssword", db: 5},
		{name: "tls", url: "rediss://default:p%40ssword@cache.example.com:6380/5", cacheURL: "rediss://cache.example.com:6381/2", address: "cache.example.com:6380", user: "default", password: "p@ssword", db: 5, tls: true},
		{name: "independent-cache", cacheURL: "redis://cache.example.com:6381/2", address: "127.0.0.1:6380", user: "admin", password: "secret", fields: map[string]string{"REDIS_HOST": "127.0.0.1", "REDIS_PORT": "6380", "REDIS_USERNAME": "admin", "REDIS_PASSWORD": "secret"}},
		{name: "fields", address: "cache.example.com:6380", user: "admin", password: "p@ssword", db: 5,
			fields: map[string]string{"REDIS_HOST": "cache.example.com", "REDIS_PORT": "6380", "REDIS_USERNAME": "admin", "REDIS_PASSWORD": "p@ssword", "REDIS_MAIN_DB": "5", "REDIS_TIMEOUT": "3s"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, key := range []string{"REDIS_HOST", "REDIS_PORT", "REDIS_USERNAME", "REDIS_PASSWORD", "REDIS_MAIN_DB", "REDIS_SCHEME", "REDIS_CACHE_SCHEME", "REDIS_CACHE_DB", "REDIS_SESSION_DB", "REDIS_QUEUE_DB", "REDIS_HORIZON_DB", "REDIS_SESSION_URL", "REDIS_QUEUE_URL", "REDIS_HORIZON_URL"} {
				t.Setenv(key, "")
			}
			t.Setenv("REDIS_URL", test.url)
			t.Setenv("REDIS_CACHE_URL", test.cacheURL)
			for key, value := range test.fields {
				t.Setenv(key, value)
			}
			configuration, err := cfg.NewFromFile(filepath.Join(t.TempDir(), ".env"))
			if err != nil {
				t.Fatal(err)
			}
			manager, err := redis.NewManager(redis.ConfigFromRepository(configuration))
			if err != nil {
				t.Fatal(err)
			}
			defer manager.Close(context.Background())
			connection, err := manager.Connection()
			if err != nil {
				t.Fatal(err)
			}
			options := connection.Client().(*goredis.Client).Options()
			if options.Addr != test.address || options.DB != test.db || options.Username != test.user || options.Password != test.password {
				t.Fatal("Redis 地址、数据库或认证信息未按配置生效")
			}
			if test.tls {
				if options.TLSConfig == nil || options.TLSConfig.InsecureSkipVerify || options.TLSConfig.ServerName != "cache.example.com" || options.TLSConfig.MinVersion < tls.VersionTLS12 {
					t.Fatal("未启用安全 TLS")
				}
				cache, err := manager.Connection("cache")
				if err != nil {
					t.Fatal(err)
				}
				options := cache.Client().(*goredis.Client).Options()
				if options.Addr != "cache.example.com:6381" || options.DB != 2 || options.Password != "" || options.TLSConfig == nil {
					t.Fatal("独立缓存 URL 未生效")
				}
			}
			if test.name == "fields" && options.DialTimeout != 3*time.Second {
				t.Fatal("连接超时未生效")
			}
		})
	}
}

func TestRedisSubsystemConnections(t *testing.T) {
	for key, value := range map[string]string{
		"REDIS_URL":              "redis://127.0.0.1:6379/0",
		"REDIS_HOST":             "127.0.0.1",
		"REDIS_PORT":             "6379",
		"REDIS_PASSWORD":         "test-only",
		"REDIS_CACHE_DB":         "2",
		"REDIS_SESSION_DB":       "4",
		"REDIS_QUEUE_DB":         "3",
		"REDIS_HORIZON_DB":       "3",
		"CACHE_REDIS_CONNECTION": "cache",
	} {
		t.Setenv(key, value)
	}
	configuration, err := cfg.NewFromFile(filepath.Join(t.TempDir(), ".env"))
	if err != nil {
		t.Fatal(err)
	}
	manager, err := redis.NewManager(redis.ConfigFromRepository(configuration))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close(context.Background())
	for name, want := range map[string]int{"default": 0, "cache": 2, "session": 4, "queue": 3, "horizon": 3} {
		connection, err := manager.Connection(name)
		if err != nil {
			t.Fatal(err)
		}
		options := connection.Client().(*goredis.Client).Options()
		if options.DB != want || options.Password != "test-only" {
			t.Fatalf("%s Redis DB=%d password=%t", name, options.DB, options.Password != "")
		}
	}
	for key, want := range map[string]string{
		"cache.stores.redis.connection":      "cache",
		"session.connection":                 "session",
		"queue.connections.redis.connection": "queue",
		"queue.failed.store":                 "queue",
		"queue.batching.store":               "queue",
		"horizon.connection":                 "horizon",
	} {
		if got := configuration.GetString(key); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestPrefixesExpandApplicationName(t *testing.T) {
	t.Setenv("APP_NAME", "dawn-test")
	t.Setenv("APP_ENV", "production")
	path := filepath.Join(t.TempDir(), ".env")
	data := []byte(`APP_NAME=dawn-test
APP_ENV=production
CACHE_PREFIX="${APP_NAME}"
SESSION_PREFIX="${APP_NAME}:session"
QUEUE_FAILED_PREFIX="${APP_NAME}:failed"
QUEUE_BATCHING_PREFIX="${APP_NAME}:batches"
QUEUE_RESTART_KEY="${APP_NAME}:queue:restart"
REDIS_QUEUE_PREFIX="${APP_NAME}:queue"
HORIZON_PREFIX="${APP_NAME}:horizon:"
`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	configuration, err := cfg.NewFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"cache.prefix":                   "dawn-test",
		"session.prefix":                 "dawn-test:session",
		"queue.failed.prefix":            "dawn-test:failed",
		"queue.batching.prefix":          "dawn-test:batches",
		"queue.restart.key":              "dawn-test:queue:restart",
		"queue.connections.redis.prefix": "dawn-test:queue",
		"horizon.prefix":                 "dawn-test:horizon:",
	} {
		if got := configuration.GetString(key); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestSQLiteConfiguration(t *testing.T) {
	t.Setenv("APP_KEY", "base64:MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	directory := t.TempDir()
	path := filepath.Join(directory, "database.sqlite")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_DATABASE", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	t.Setenv("DB_DSN", "")
	t.Setenv("DB_MAX_OPEN_CONNS", "1")
	app := foundation.Configure(directory).WithExtensionProviders(dawn.ServiceProvider{}).Create()
	defer app.Close()
	if err := app.Boot(); err != nil {
		t.Fatal(err)
	}
	db, err := database.OpenDefaultConnection()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		t.Fatal(err)
	}
	var foreignKeys int
	if err := db.Raw("PRAGMA foreign_keys").Scan(&foreignKeys).Error; err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 || sqlDB.Stats().MaxOpenConnections != 1 {
		t.Fatal("SQLite 约束或连接池未生效")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

// TestSQLConnection 只接受显式提供的专用测试库。
func TestSQLConnection(t *testing.T) {
	driver, dsn := os.Getenv("DAWN_TEST_DRIVER"), os.Getenv("DAWN_TEST_DSN")
	if driver == "" || dsn == "" {
		t.Skip("未配置专用数据库")
	}
	for _, mode := range []string{"dsn", "fields"} {
		if driver == "pgsql" && mode == "fields" {
			continue
		}
		t.Run(mode, func(t *testing.T) {
			t.Setenv("DB_CONNECTION", driver)
			t.Setenv("DB_DSN", dsn)
			t.Setenv("DB_MAX_OPEN_CONNS", "7")
			if mode == "fields" {
				connection, err := mysql.ParseDSN(dsn)
				if err != nil {
					t.Fatal(err)
				}
				host, port, err := net.SplitHostPort(connection.Addr)
				if err != nil {
					t.Fatal(err)
				}
				for key, value := range map[string]string{"DB_DSN": "", "DB_HOST": host, "DB_PORT": port, "DB_USERNAME": connection.User, "DB_PASSWORD": connection.Passwd, "DB_DATABASE": connection.DBName} {
					t.Setenv(key, value)
				}
			}
			app := foundation.Configure(t.TempDir()).WithExtensionProviders(dawn.ServiceProvider{}).Create()
			defer app.Close()
			if err := app.Boot(); err != nil {
				t.Fatal(err)
			}
			db, err := database.OpenDefaultConnection()
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			defer sqlDB.Close()
			if err := sqlDB.Ping(); err != nil {
				t.Fatal(err)
			}
			if sqlDB.Stats().MaxOpenConnections != 7 {
				t.Fatal("连接池配置未生效")
			}
		})
	}
}
