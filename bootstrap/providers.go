package bootstrap

import (
	"github.com/dawnim/dawn-simple/app/providers"
	"github.com/prismgo/framework/contracts/provider"
)

func Providers() []provider.ServiceProvider {
	return []provider.ServiceProvider{providers.AppServiceProvider{}}
}
