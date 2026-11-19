package admin

import (
	"context"
	sdk "github.com/dawnim/dawn-server/app/admin"
	"github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-server/support/customer"
	appconst "github.com/dawnim/dawn-simple/app/constants"
)

// Configure appends business definitions without duplicating framework constants.
func Configure(o *sdk.Options) {
	o.Customer = func(m *customer.Manager) {
		m.RegisterTool("account_profile", func(ctx context.Context, actor customer.ToolActor, args map[string]any) (any, error) {
			if !actor.Verified || actor.UserType != models.UserMember || actor.Public || actor.IdentityProvider != "" {
				return nil, customer.ErrForbidden
			}
			var user models.User
			err := m.DB.WithContext(ctx).Select("id", "unique_id", "nickname", "email", "mobile", "locale").First(&user, actor.UserID).Error
			return user.Profile(), err
		})
	}
	o.Settings = append(o.Settings, sdk.SettingDefinition{Key: "moments.daily_limit", Title: "每日发布动态上限", Group: "网友动态", Kind: "number", Input: "number", Default: 2, Remark: "每位用户每天可发布的动态数量"})
	o.Settings = append(o.Settings, sdk.SettingDefinition{Key: "moments.ai_daily_limit", Title: "每日AI创作总次数", Group: "网友动态", Kind: "number", Input: "number", Default: 23, Remark: "每个IP每天最多使用2次"})
	o.Settings = append(o.Settings, sdk.SettingDefinition{Key: "moments.auto_daily_limit", Title: "每日AI自动动态上限", Group: "网友动态", Kind: "number", Input: "number", Default: 1, Remark: "设为0关闭，依据近期文档阅读生成"})
	o.Settings = append(o.Settings, sdk.SettingDefinition{Key: "moments.brave_key", Title: "Brave搜索密钥", Group: "网友动态", Kind: "string", Input: "password", Default: "", Secret: true, Remark: "选填，为AI自动动态提供近期趣闻素材"})
	o.Settings = append(o.Settings,
		sdk.SettingDefinition{Key: "moments.toolbar_hidden", Title: "隐藏的编辑器工具", Group: "网友动态", Kind: "string", Input: "textarea", Default: "", Remark: "逗号分隔的工具名，如 media,alignment；留空显示全部"},
		sdk.SettingDefinition{Key: "moments.emoji_shortcodes", Title: "自定义表情短码", Group: "网友动态", Kind: "string", Input: "textarea", Default: "party=🎉\nclap=👏\nidea=💡\nbug=🐛\ndone=✅", Remark: "每行一个，格式 名称=表情，正文中用 :名称: 引用"},
	)
	o.Menus = append(o.Menus, sdk.Menu{Key: "app.moments", ParentKey: "dawn.support", Title: "网友动态", Path: "/support/moments", Component: "/support/moments", Permission: "app.moments.list", Icon: "ri:article-line", Sort: 40})
	o.Permissions = append(o.Permissions,
		sdk.Permission{Key: "app.moments.list", Title: "查看", ParentKey: "app.moments"},
		sdk.Permission{Key: "app.moments.create", Title: "新增", ParentKey: "app.moments"},
		sdk.Permission{Key: "app.moments.update", Title: "修改", ParentKey: "app.moments"},
		sdk.Permission{Key: "app.moments.delete", Title: "删除", ParentKey: "app.moments"},
		sdk.Permission{Key: "app.moment_comments.list", Title: "查看评论", ParentKey: "app.moments"},
		sdk.Permission{Key: "app.moment_comments.create", Title: "新增评论", ParentKey: "app.moments"},
		sdk.Permission{Key: "app.moment_comments.update", Title: "修改评论", ParentKey: "app.moments"},
		sdk.Permission{Key: "app.moment_comments.delete", Title: "删除评论", ParentKey: "app.moments"},
	)
	o.Routes = RegisterRoutes
	o.CacheKeys = append(o.CacheKeys, appconst.TempMomentAI, appconst.TempMomentReading)
}
