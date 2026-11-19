package migrations

import (
	"context"
	"testing"

	"github.com/dawnim/dawn-server/database/schema"
	"github.com/dawnim/dawn-simple/app/models"
	"github.com/glebarez/sqlite"
	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

func TestMomentMigrationMatchesSchema(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	migration, ok := database.MigrationByName("202611140000_create_moments")
	if !ok {
		t.Fatal("moment migration not registered")
	}
	if err := migration.Up(db); err != nil {
		t.Fatal(err)
	}
	target := schema.Definition{Tables: map[string]schema.Table{
		"moments":          {Model: &models.Moment{}},
		"moment_comments":  {Model: &models.MomentComment{}},
		"moment_reactions": {Model: &models.MomentReaction{}},
	}, Keep: []string{"sqlite_*"}}
	plan, err := target.Check(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Changes) != 0 {
		t.Fatalf("migration differs from target: %+v", plan.Changes)
	}
	if err := migration.Down(db); err != nil {
		t.Fatal(err)
	}
}
