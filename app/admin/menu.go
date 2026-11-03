package admin

import (
	sdk "github.com/dawnim/dawn-server/app/admin"
	appconst "github.com/dawnim/dawn-simple/app/constants"
)

// Configure appends business definitions without duplicating framework constants.
func Configure(o *sdk.Options) {
	o.Routes = RegisterRoutes
	o.Realtime = RegisterRealtime
	o.CacheKeys = append(o.CacheKeys, appconst.TempGreeting)
	o.Menus = append(o.Menus, sdk.Menu{Key: "app.realtime_example", ParentKey: "dawn.development", Title: "实时示例", Path: "/examples/realtime", Component: "/examples/realtime", Permission: "app.realtime_example.view", Icon: "ri:chat-3-line"})
	o.Permissions = append(o.Permissions, sdk.Permission{Key: "app.realtime_example.view", Title: "实时示例", ParentKey: "dawn.development"})
}
