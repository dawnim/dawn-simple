package config

import (
	"path/filepath"
	"slices"
	"testing"

	cfg "github.com/prismgo/framework/config"
	"github.com/prismgo/horizon"
)

func TestHorizonGroups(t *testing.T) {
	for _, serial := range []string{"true", "false"} {
		t.Run(serial, func(t *testing.T) {
			t.Setenv("HORIZON_SERIAL", serial)
			t.Setenv("HORIZON_CONCURRENCY", "6")
			configuration, err := cfg.NewFromFile(filepath.Join(t.TempDir(), ".env"))
			if err != nil {
				t.Fatal(err)
			}
			h, err := horizon.LoadConfigFrom(configuration)
			if err != nil {
				t.Fatal(err)
			}
			sequential, enabled := h.Supervisors["supervisor"]
			if enabled != (serial == "true") {
				t.Fatalf("串行组启用状态错误: %+v", h.Supervisors)
			}
			if enabled && (sequential.Balance != "false" || sequential.MinProcesses != 1 || sequential.MaxProcesses != 1 || !slices.Equal(sequential.Queues, []string{"default"})) {
				t.Fatalf("串行组必须使用一个worker: %+v", sequential)
			}
			groups := 1
			if enabled {
				groups++
			}
			parallel := h.Supervisors["multiples"]
			if parallel.Balance != "auto" || parallel.MaxProcesses != 6 || !slices.Equal(parallel.Queues, []string{"parallel"}) || len(h.Supervisors) != groups {
				t.Fatalf("并发组配置错误: %+v", h.Supervisors)
			}
		})
	}
}
