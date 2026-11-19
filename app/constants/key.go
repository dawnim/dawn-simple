package constants

import (
	dawn "github.com/dawnim/dawn-server/constants"
	"time"
)

// TempMomentAI stores daily AI generation usage per IP and globally.
var TempMomentAI = dawn.Definition{Name: "TempMomentAI", Pattern: "temp:moment:ai:%s:%s", Type: "string", Description: "动态AI创作每日额度", TTL: 24 * time.Hour}

// TempMomentReading stores recent documentation paths and coarse regions for AI posts.
var TempMomentReading = dawn.Definition{Name: "TempMomentReading", Pattern: "temp:moment:reading", Type: "list", Description: "近期文档阅读", TTL: 48 * time.Hour}
