package migrations

import (
	"fmt"

	sdkmigrations "github.com/dawnim/dawn-server/database/migrations"
	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration(func(db *gorm.DB) error { return sdkmigrations.Install(db) }, func(db *gorm.DB) error {
		return fmt.Errorf("内置系统表不支持自动回滚，请通过备份恢复")
	})
}
