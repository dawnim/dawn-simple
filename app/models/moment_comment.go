package models

import sdk "github.com/dawnim/dawn-server/app/models"

// MomentComment retains replies when a parent comment is removed.
type MomentComment struct {
	sdk.Base
	MomentID uint   `json:"moment_id" gorm:"not null;index;comment:动态ID"`
	UserID   uint   `json:"user_id" gorm:"not null;index;comment:评论用户ID"`
	ReplyID  uint   `json:"reply_id" gorm:"not null;default:0;comment:回复的评论ID"`
	Content  string `json:"content" gorm:"not null;type:text;comment:评论内容"`
	Region   string `json:"region" gorm:"not null;size:100;default:'';comment:评论时IP所在地区"`
}
