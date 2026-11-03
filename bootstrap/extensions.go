package bootstrap

import (
	dawn "github.com/dawnim/dawn-server"
	"github.com/dawnim/dawn-simple/app/admin"
	"github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/encryption"
)

func Extensions() []provider.ServiceProvider {
	return []provider.ServiceProvider{encryption.ServiceProvider{}, dawn.ServiceProvider{Configure: admin.Configure}}
}
