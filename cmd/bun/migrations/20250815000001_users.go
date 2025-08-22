package migrations

import (
	"context"

	"github.com/uptrace/bun"
	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// 創建用戶表
		_, err := db.NewCreateTable().
			Model((*dbmodels.UserDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 創建索引
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
			"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
			"CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)",
			"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
		}

		for _, idx := range indexes {
			if _, err := db.ExecContext(ctx, idx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// 回滾 - 刪除表
		_, err := db.NewDropTable().
			Model((*dbmodels.UserDB)(nil)).
			IfExists().
			Exec(ctx)
		return err
	})
}