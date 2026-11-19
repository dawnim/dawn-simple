package bootstrap

import (
	dawn "github.com/dawnim/dawn-server"
	"github.com/dawnim/dawn-server/database/schema"
	"github.com/dawnim/dawn-simple/app/admin"
	"github.com/dawnim/dawn-simple/app/models"
	"github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/encryption"
)

func Extensions() []provider.ServiceProvider {
	return []provider.ServiceProvider{encryption.ServiceProvider{}, dawn.ServiceProvider{Configure: admin.Configure, Schema: func(d *schema.Definition) {
		d.Tables["moments"] = schema.Table{Model: &models.Moment{}, Comment: "网友动态"}
		d.Tables["moment_comments"] = schema.Table{Model: &models.MomentComment{}, Comment: "动态评论"}
		d.Tables["moment_reactions"] = schema.Table{Model: &models.MomentReaction{}, Comment: "动态互动"}
	}}}
}
