package cmd

import (
	"github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-server/support/customer"
	"github.com/dawnim/dawn-server/support/i18n"
	"github.com/dawnim/dawn-simple/database/seeders"
	"github.com/prismgo/framework/console"
)

// SupportDocsCommand刷新官方演示知识，适用于文档更新后的显式发布。
type SupportDocsCommand struct{}

func (*SupportDocsCommand) Definition() *console.Definition {
	return console.MustDefinition("app:support-docs", i18n.Text("", "同步官网客服文档"))
}

func (*SupportDocsCommand) Handle(command console.CommandContext) error {
	manager := customer.Resolve()
	var app models.SupportApp
	if err := manager.DB.WithContext(command.Context()).Where("public_key = ?", "dawn").First(&app).Error; err != nil {
		return err
	}
	return seeders.SyncSupportDocs(command.Context(), manager.DB, app.ID)
}
