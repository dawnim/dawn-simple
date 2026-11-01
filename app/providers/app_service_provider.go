package providers

import "github.com/prismgo/framework/contracts/provider"

type AppServiceProvider struct{}

func (AppServiceProvider) Register(provider.Application) error { return nil }
func (AppServiceProvider) Boot(provider.Application) error     { return nil }
