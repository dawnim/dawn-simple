package routes

import (
	dawn "github.com/dawnim/dawn-server"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/foundation"
)

func Register(app *foundation.Application, engine *gin.Engine) error {
	if err := dawn.RegisterRoutes(app, engine); err != nil {
		return err
	}
	engine.GET("/api", func(c *gin.Context) { c.JSON(200, gin.H{"name": "dawn-simple", "admin": "/admin"}) })
	return nil
}
