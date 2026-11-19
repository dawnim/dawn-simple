package seeders

import (
	sdk "github.com/dawnim/dawn-server/database/seeders"
	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

func init() { database.RegisterSeeder(Seed) }

func Seed(db *gorm.DB) error {
	if err := sdk.Seed(db); err != nil {
		return err
	}
	return seedSupport(db)
}
