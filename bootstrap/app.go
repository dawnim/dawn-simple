package bootstrap

import (
	"github.com/dawnim/dawn-simple/app/cmd"
	"github.com/dawnim/dawn-simple/app/exceptions"
	"github.com/dawnim/dawn-simple/app/schedule"
	_ "github.com/dawnim/dawn-simple/config"
	_ "github.com/dawnim/dawn-simple/database/migrations"
	"github.com/dawnim/dawn-simple/routes"
	"github.com/prismgo/framework/foundation"
)

func NewApplication(basePath ...string) *foundation.Application {
	return foundation.Configure(basePath...).WithExtensionProviders(Extensions()...).WithProviders(Providers()...).WithRouting(func(r *foundation.Routing) {
		r.MigrationPaths("database/migrations")
		r.SeedPaths("database/seeders")
		r.Commands(cmd.CommandFactories()...)
		r.Schedules(schedule.Register)
		r.Routes(routes.Register)
	}).WithExceptions(exceptions.Configure).Create()
}
