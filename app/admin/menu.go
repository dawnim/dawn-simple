package admin

import sdk "github.com/dawnim/dawn-server/app/admin"

// Configure 在这里追加项目菜单、权限和后台路由。
func Configure(o *sdk.Options) { o.Routes = RegisterRoutes }
