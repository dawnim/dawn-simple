package config

func init() {
	Add("dawn", func() map[string]any {
		return map[string]any{
			"runtime_directory": Env("DAWN_RUNTIME_DIRECTORY", "storage/framework"),
			"module":            "github.com/dawnim/dawn-simple",
			"ws_addr":           Env("WS_ADDR", ":8851"),
			"ws_path":           Env("WS_PATH", "/"),
			"ws_url":            Env("WS_PUBLIC_URL", ""),
			"admin_prefix":      Env("ADMIN_PREFIX", "/admin"),
			"allowed_origins":   Env("ALLOWED_ORIGINS", "*"),
			"disk":              Env("ADMIN_DISK", ""),
			"admin_username":    Env("DAWN_ADMIN_USERNAME", "admin"),
			"admin_password":    Env("DAWN_ADMIN_PASSWORD", "123456"),
			"access_ttl":        Env("DAWN_ACCESS_TTL", "720h"),
			"refresh_ttl":       Env("DAWN_REFRESH_TTL", "720h"),
			"monitor": map[string]any{
				"disk_path": Env("DAWN_MONITOR_DISK_PATH", ""),
			},
			"geoip": map[string]any{
				"enabled":   Env("GEOIP_ENABLED", true),
				"directory": Env("GEOIP_DIRECTORY", "storage/app/geoip"),
				"sources": map[string]string{
					"GeoLite2-City.mmdb":    Env("GEOIP_CITY_URL", "https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-City.mmdb").(string),
					"GeoLite2-Country.mmdb": Env("GEOIP_COUNTRY_URL", "https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-Country.mmdb").(string),
					"GeoLite2-ASN.mmdb":     Env("GEOIP_ASN_URL", "https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-ASN.mmdb").(string),
					"ip2region_v4.xdb":      Env("GEOIP_IPV4_URL", "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/ip2region_v4.xdb").(string),
					"ip2region_v6.xdb":      Env("GEOIP_IPV6_URL", "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/ip2region_v6.xdb").(string),
				},
			},
			"embedding": map[string]any{
				"enabled":    Env("EMBEDDING_ENABLED", false),
				"directory":  Env("EMBEDDING_DIRECTORY", "storage/support/embedding"),
				"python":     Env("EMBEDDING_PYTHON", "python3"),
				"host":       Env("EMBEDDING_HOST", "127.0.0.1"),
				"port":       Env("EMBEDDING_PORT", 8852),
				"model":      Env("EMBEDDING_MODEL", "Qwen/Qwen3-Embedding-0.6B"),
				"threads":    Env("EMBEDDING_THREADS", 4),
				"max_length": Env("EMBEDDING_MAX_LENGTH", 2048),
				"device":     Env("EMBEDDING_DEVICE", "cpu"),
			},
			"local": map[string]any{
				"watch":     Env("DAWN_GO_WATCH", true),
				"directory": Env("DAWN_FRONTEND_DIRECTORY", "admin"),
				"port":      Env("DAWN_FRONTEND_PORT", 5173),
			},
		}
	})
}
