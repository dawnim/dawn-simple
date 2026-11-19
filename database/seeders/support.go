package seeders

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-server/support/customer"
	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/support"
	"gorm.io/gorm"
)

const supportInstructions = "你服务于Dawn CMS。默认回答Dawn使用与开发问题，PrismGo是Dawn使用的底层框架。未明确提及PrismGo时，安装问题优先查询Dawn快速开始；区分两个产品的命令与入口。"
const officialDocsLocation = "https://dawnim.github.io/dawn-simple/"

// seedSupport 只初始化官方演示应用，不覆盖管理员后续配置和知识编辑。
func seedSupport(db *gorm.DB) error {
	var existing models.SupportApp
	result := db.Unscoped().Where("public_key = ?", "dawn").Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	manager, err := container.Make[*customer.Manager]("dawn.customer")
	if err != nil {
		return err
	}
	var admin models.AdminUser
	if err = db.Where("super = ? AND status = ?", true, models.Enabled).Order("id").First(&admin).Error; err != nil {
		return err
	}
	cfg := customer.Config{AgentIDs: []uint{admin.ID}, Capacity: 5, AllowLocalUsers: true, Greeting: "您好，欢迎咨询Dawn使用问题", Instructions: supportInstructions, EmbeddingURL: "http://127.0.0.1:8852/v1/", EmbeddingModel: "Qwen/Qwen3-Embedding-0.6B", EmbeddingDimensions: 1024, IndexDriver: "portable", Tools: []customer.Tool{{Name: "account_profile", Description: "查询当前已登录的本站用户资料", Driver: "go", RequireAccount: true, Skill: "account-profile", Parameters: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}}}, Skills: []customer.Skill{{Name: "account-profile", Description: "查看当前登录用户的本站资料", AllowedTools: "account_profile", Instructions: "只使用account_profile查询当前用户资料，不接受客户提交的用户ID，未验证身份时提示先登录"}}}
	raw, _ := json.Marshal(cfg)
	app := models.SupportApp{Name: "Dawn在线客服", PublicKey: "dawn", CreatedBy: admin.ID, Config: raw}
	ctx := context.Background()
	if err = manager.CreateApp(ctx, &app, customer.Secrets{}); err != nil {
		return err
	}
	return SyncSupportDocs(ctx, db, app.ID)
}

// SyncSupportDocs同步官方演示文档并补全空的产品说明，保留模型配置与聊天记录。
func SyncSupportDocs(ctx context.Context, db *gorm.DB, appID uint) error {
	manager := customer.Resolve()
	app, cfg, _, err := manager.App(ctx, appID)
	if err != nil {
		return err
	}
	if cfg.Instructions == "" {
		cfg.Instructions = supportInstructions
		result := db.Model(&app).Where("revision = ?", app.Revision).Updates(map[string]any{"config": customer.Raw(cfg), "revision": gorm.Expr("revision + 1")})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return customer.ErrConflict
		}
	}
	root := support.BasePath("docs")
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	seen := map[uint]bool{}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" || entry.Name() == "public" || strings.HasPrefix(relative, "prismgo/") && !strings.HasPrefix(relative, "prismgo/zh_CN") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		if strings.Contains(relative, "/") && !strings.HasPrefix(relative, "prismgo/zh_CN/") {
			return nil
		}
		content, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		for line := range strings.SplitSeq(string(content), "\n") {
			if strings.HasPrefix(line, "# ") {
				name = strings.TrimSpace(line[2:])
				break
			}
		}
		brand := "Dawn"
		if strings.HasPrefix(filepath.ToSlash(relative), "prismgo/") {
			brand = "PrismGo"
		}
		if !strings.HasPrefix(name, brand) {
			name = brand + "-" + name
		}
		location := officialDocsLocation + strings.TrimSuffix(filepath.ToSlash(relative), ".md")
		var source models.KnowledgeSource
		var sourceID uint
		if e = db.Model(&models.KnowledgeEntry{}).Where("app_id = ? AND source_id > 0 AND location IN ?", appID, []string{location, location + ".html"}).Order("id DESC").Limit(1).Pluck("source_id", &sourceID).Error; e != nil {
			return e
		}
		query := db.Where("app_id = ? AND kind = ?", appID, "markdown")
		if sourceID != 0 {
			query = query.Where("id = ?", sourceID)
		} else {
			query = query.Where("name = ?", name)
		}
		if e = query.Limit(1).Find(&source).Error; e != nil {
			return e
		}
		source.AppID, source.Name, source.Kind, source.Content = appID, name, "markdown", string(content)
		if e = db.Save(&source).Error; e != nil {
			return e
		}
		seen[source.ID] = true
		if e = manager.ImportSource(ctx, source.ID); e != nil {
			return e
		}
		return db.Model(&models.KnowledgeEntry{}).Where("app_id = ? AND source_id = ? AND status IN ?", appID, source.ID, []string{"draft", "published"}).Updates(map[string]any{"status": "published", "location": location + ".html"}).Error
	}); err != nil {
		return err
	}
	var officialSourceIDs []uint
	if err := db.Model(&models.KnowledgeEntry{}).Where("app_id = ? AND source_id > 0 AND location LIKE ?", appID, officialDocsLocation+"%").Distinct("source_id").Pluck("source_id", &officialSourceIDs).Error; err != nil {
		return err
	}
	for _, sourceID := range officialSourceIDs {
		if seen[sourceID] {
			continue
		}
		if err := db.Model(&models.KnowledgeEntry{}).Where("app_id = ? AND source_id = ? AND status IN ?", appID, sourceID, []string{"draft", "published"}).Updates(map[string]any{"status": "disabled"}).Error; err != nil {
			return err
		}
	}
	return nil
}
