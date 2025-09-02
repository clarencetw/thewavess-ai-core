package migrations

import (
	"context"

	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, bunDB *bun.DB) error {
		// 使用 Bun ORM 模型定義來創建表結構

		// 1. 創建角色主表
		_, err := bunDB.NewCreateTable().Model((*db.CharacterDB)(nil)).IfNotExists().Exec(ctx)
		if err != nil {
			return err
		}

		// 4. 創建索引以優化查詢
		indices := []string{
			"CREATE INDEX IF NOT EXISTS idx_characters_active ON characters(is_active)",
			"CREATE INDEX IF NOT EXISTS idx_characters_type ON characters(type)",
			"CREATE INDEX IF NOT EXISTS idx_characters_locale ON characters(locale)",
			"CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name)",
			"CREATE INDEX IF NOT EXISTS idx_characters_popularity ON characters(popularity DESC)",
			"CREATE INDEX IF NOT EXISTS idx_characters_tags ON characters USING GIN(tags)",
			// 用戶追蹤和軟刪除索引
			"CREATE INDEX IF NOT EXISTS idx_characters_created_by ON characters(created_by)",
			"CREATE INDEX IF NOT EXISTS idx_characters_updated_by ON characters(updated_by)", 
			"CREATE INDEX IF NOT EXISTS idx_characters_deleted_at ON characters(deleted_at)",
			"CREATE INDEX IF NOT EXISTS idx_characters_is_public ON characters(is_public)",
			"CREATE INDEX IF NOT EXISTS idx_characters_is_system ON characters(is_system)",
		}

		for _, idx := range indices {
			_, err = bunDB.ExecContext(ctx, idx)
			if err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, bunDB *bun.DB) error {
		// 回滾 - 刪除表
		tables := []string{
			"character_scenes",
			"character_speech_styles",
			"characters",
		}

		for _, table := range tables {
			_, err := bunDB.ExecContext(ctx, "DROP TABLE IF EXISTS "+table)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
