package migrations

import (
	sdk "github.com/dawnim/dawn-server/database/migrations"
	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(func(db *gorm.DB) error { return sdk.Up(db) }, func(db *gorm.DB) error { return sdk.Down(db) })
}
