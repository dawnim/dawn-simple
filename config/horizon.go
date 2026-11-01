package config

import (
	"fmt"

	"github.com/dawnim/dawn-simple/app/constants"
)

func init() {
	Add("horizon", func() map[string]any {
		prefix := fmt.Sprintf("%v:%v", Env("APP_NAME", "dawn-simple"), Env("APP_ENV", "local"))
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
