package models

// MomentReaction accumulates both reactions in one row per user and post.
type MomentReaction struct {
	ID        uint  `json:"id" gorm:"primaryKey"`
	MomentID  uint  `json:"moment_id" gorm:"not null;uniqueIndex:idx_moment_reaction;comment:动态ID"`
	UserID    uint  `json:"user_id" gorm:"not null;uniqueIndex:idx_moment_reaction;comment:用户ID"`
	Flowers   uint  `json:"flowers" gorm:"not null;default:0;comment:鲜花次数"`
	Eggs      uint  `json:"eggs" gorm:"not null;default:0;comment:鸡蛋次数"`
	CreatedAt int64 `json:"created_at" gorm:"not null;autoCreateTime:milli"`
	UpdatedAt int64 `json:"updated_at" gorm:"not null;autoUpdateTime:milli"`
}
