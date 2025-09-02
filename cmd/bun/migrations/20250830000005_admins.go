package migrations

import (
	"context"

	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, bunDB *bun.DB) error {
		// 創建管理員表
		_, err := bunDB.NewCreateTable().
			Model((*db.AdminDB)(nil)).
			IfNotExists().
			Exec(ctx)

		if err != nil {
			return err
		}

		// 創建索引
		queries := []string{
			"CREATE INDEX IF NOT EXISTS idx_admins_username ON admins(username)",
			"CREATE INDEX IF NOT EXISTS idx_admins_email ON admins(email)",
			"CREATE INDEX IF NOT EXISTS idx_admins_role ON admins(role)",
			"CREATE INDEX IF NOT EXISTS idx_admins_status ON admins(status)",
			"CREATE INDEX IF NOT EXISTS idx_admins_created_at ON admins(created_at)",
		}

		for _, query := range queries {
			_, err := bunDB.ExecContext(ctx, query)
			if err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, bunDB *bun.DB) error {
		// 回滾：刪除管理員表
		_, err := bunDB.NewDropTable().
			Model((*db.AdminDB)(nil)).
			IfExists().
			Exec(ctx)
		return err
	})
}
