package migrations

import (
	"context"

	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/uptrace/bun"
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
			"CREATE INDEX IF NOT EXISTS idx_users_registration_ip ON users(registration_ip)",
			"CREATE INDEX IF NOT EXISTS idx_users_last_login_ip ON users(last_login_ip)",
			"CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL",
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
