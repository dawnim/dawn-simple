package models

import sdk "github.com/dawnim/dawn-server/app/models"

// Moment is a public post created by a member or the site's AI account.
type Moment struct {
	sdk.Base
	UserID      uint   `json:"user_id" gorm:"not null;index;comment:发布用户ID"`
	Format      string `json:"format" gorm:"not null;size:16;default:moment;index;comment:内容形式:moment动态 post帖子"`
	Title       string `json:"title" gorm:"not null;size:100;default:'';comment:主题标题"`
	Board       string `json:"board" gorm:"not null;size:32;default:general;index;comment:板块general交流 help求助 showcase分享 ai趣闻"`
	Content     string `json:"content" gorm:"not null;type:text;comment:动态内容"`
	Images      string `json:"images" gorm:"not null;type:text;comment:图片附件ID列表JSON"`
	Location    string `json:"location" gorm:"not null;size:200;default:'';comment:用户选择的地点"`
	Region      string `json:"region" gorm:"not null;size:100;default:'';comment:发布时IP所在地区"`
	IsAnonymous bool   `json:"is_anonymous" gorm:"not null;default:false;comment:是否匿名发布"`
}
