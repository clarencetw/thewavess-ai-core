package migrations

import (
	"github.com/uptrace/bun/migrate"
)

// Migrations 全局遷移集合
var Migrations = migrate.NewMigrations()

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
}
