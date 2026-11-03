package exceptions

import (
	sdk "github.com/dawnim/dawn-server/app/admin"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/foundation"
)

func Configure(e *foundation.Exceptions) {
	e.Context(func(c *gin.Context) map[string]any {
		if actor, ok := c.Get("dawn.actor"); ok {
			return map[string]any{"admin_id": actor.(*sdk.Actor).User.ID}
		}
		return nil
	})
}
