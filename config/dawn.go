package config

func init() {
	Add("dawn", func() map[string]any {
		return map[string]any{
			"runtime_directory": Env("DAWN_RUNTIME_DIRECTORY", "storage/framework"),
			"module":            "github.com/dawnim/dawn-simple",
			"ws_addr":           Env("WS_ADDR", ":8851"),
			"ws_path":           Env("WS_PATH", "/"),
			"admin_prefix":      Env("ADMIN_PREFIX", "/admin"),
			"allowed_origins":   Env("ALLOWED_ORIGINS", "*"),
			"disk":              Env("ADMIN_DISK", ""),
			"admin_password":    Env("DAWN_ADMIN_PASSWORD", "123456"),
			"access_ttl":        Env("DAWN_ACCESS_TTL", "720h"),
			"refresh_ttl":       Env("DAWN_REFRESH_TTL", "720h"),
			"monitor": map[string]any{
				"disk_path": Env("DAWN_MONITOR_DISK_PATH", ""),
			},
			"geoip": map[string]any{
				"enabled": Env("GEOIP_ENABLED", true),
				"path":    Env("GEOIP_PATH", "storage/app/geoip/GeoLite2-City.mmdb"),
				"url":     Env("GEOIP_URL", "https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-City.mmdb"),
			},
			"local": map[string]any{
				"directory": Env("DAWN_FRONTEND_DIRECTORY", "admin"),
				"port":      Env("DAWN_FRONTEND_PORT", 5173),
			},
		}
	})
}
