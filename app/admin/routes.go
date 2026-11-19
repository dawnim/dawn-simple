package admin

import (
	sdk "github.com/dawnim/dawn-server/app/admin"
	"github.com/dawnim/dawn-simple/app/admin/controllers"
	"github.com/prismgo/framework/route"
)

func RegisterRoutes(a *sdk.App, g *route.Registrar) { controllers.RegisterMoments(a, g) }
