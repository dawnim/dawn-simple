package migrations

import (
	"github.com/dawnim/dawn-server/database/schema"
	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// These types are a snapshot of the initial moment tables.
type MomentSnapshotBase struct {
	ID        uint                  `gorm:"not null;primaryKey"`
	CreatedAt int64                 `gorm:"not null;autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"not null;autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"not null;softDelete:milli;index;default:0"`
}
type momentSnapshot struct {
	MomentSnapshotBase
	UserID      uint   `gorm:"not null;index;comment:发布用户ID"`
	Format      string `gorm:"not null;size:16;default:moment;index;comment:内容形式:moment动态 post帖子"`
	Title       string `gorm:"not null;size:100;default:'';comment:主题标题"`
	Board       string `gorm:"not null;size:32;default:general;index;comment:板块general交流 help求助 showcase分享 ai趣闻"`
	Content     string `gorm:"not null;type:text;comment:动态内容"`
	Images      string `gorm:"not null;type:text;comment:图片附件ID列表JSON"`
	Location    string `gorm:"not null;size:200;default:'';comment:用户选择的地点"`
	Region      string `gorm:"not null;size:100;default:'';comment:发布时IP所在地区"`
	IsAnonymous bool   `gorm:"not null;default:false;comment:是否匿名发布"`
}
type momentCommentSnapshot struct {
	MomentSnapshotBase
	MomentID uint   `gorm:"not null;index;comment:动态ID"`
	UserID   uint   `gorm:"not null;index;comment:评论用户ID"`
	ReplyID  uint   `gorm:"not null;default:0;comment:回复的评论ID"`
	Content  string `gorm:"not null;type:text;comment:评论内容"`
	Region   string `gorm:"not null;size:100;default:'';comment:评论时IP所在地区"`
}
type momentReactionSnapshot struct {
	ID        uint  `gorm:"primaryKey"`
	MomentID  uint  `gorm:"not null;uniqueIndex:idx_moment_reaction;comment:动态ID"`
	UserID    uint  `gorm:"not null;uniqueIndex:idx_moment_reaction;comment:用户ID"`
	Flowers   uint  `gorm:"not null;default:0;comment:鲜花次数"`
	Eggs      uint  `gorm:"not null;default:0;comment:鸡蛋次数"`
	CreatedAt int64 `gorm:"not null;autoCreateTime:milli"`
	UpdatedAt int64 `gorm:"not null;autoUpdateTime:milli"`
}

func init() {
	d := schema.Definition{Tables: map[string]schema.Table{
		"moments":          {Model: &momentSnapshot{}, Comment: "网友动态"},
		"moment_comments":  {Model: &momentCommentSnapshot{}, Comment: "动态评论"},
		"moment_reactions": {Model: &momentReactionSnapshot{}, Comment: "动态互动"},
	}}
	database.RegisterMigration(func(db *gorm.DB) error { return d.Create(db) }, func(db *gorm.DB) error { return d.Drop(db) })
}
