package providers

import (
	"github.com/prismgo/framework/contracts/provider"
	"github.com/prismgo/framework/translation"
)

type AppServiceProvider struct{}

func (AppServiceProvider) Register(provider.Application) error { return nil }
func (AppServiceProvider) Boot(provider.Application) error {
	translation.AddLines(map[string]any{
		"messages.网友动态":               "Community",
		"messages.每日发布动态上限":           "Daily post limit",
		"messages.每日AI创作总次数":          "Daily AI draft limit",
		"messages.每日AI自动动态上限":         "Daily automated AI post limit",
		"messages.每位用户每天可发布的动态数量":     "Number of posts each user may publish per day",
		"messages.每个IP每天最多使用2次":       "Each IP can generate up to 2 drafts per day",
		"messages.设为0关闭，依据近期文档阅读生成":   "Set to 0 to disable; based on recent documentation visits",
		"messages.Brave搜索密钥":          "Brave Search key",
		"messages.选填，为AI自动动态提供近期趣闻素材": "Optional; provides recent stories for automated AI posts",
		"messages.动态AI创作每日额度":         "Daily community AI quota",
		"messages.近期文档阅读":             "Recent documentation visits",
	}, "en", "admin")
	return nil
}
