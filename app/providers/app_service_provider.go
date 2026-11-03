package providers

import (
	"github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/translation"
)

type AppServiceProvider struct{}

func (AppServiceProvider) Register(provider.Application) error { return nil }
func (AppServiceProvider) Boot(provider.Application) error {
	translation.AddLines(map[string]any{
		"messages.实时示例":     "Realtime example",
		"messages.实时示例问候缓存": "Realtime greeting cache",
	}, "en", "admin")
	return nil
}
