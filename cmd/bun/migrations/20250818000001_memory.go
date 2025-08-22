package migrations

import (
	"context"

	"github.com/uptrace/bun"
	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// 1. 創建長期記憶主表
		_, err := db.NewCreateTable().
			Model((*dbmodels.LongTermMemoryModelDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 2. 創建偏好表
		_, err = db.NewCreateTable().
			Model((*dbmodels.MemoryPreferenceDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 3. 創建稱呼表
		_, err = db.NewCreateTable().
			Model((*dbmodels.MemoryNicknameDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 4. 創建里程碑表
		_, err = db.NewCreateTable().
			Model((*dbmodels.MemoryMilestoneDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 5. 創建禁忌表
		_, err = db.NewCreateTable().
			Model((*dbmodels.MemoryDislikeDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 6. 創建個人信息表
		_, err = db.NewCreateTable().
			Model((*dbmodels.MemoryPersonalInfoDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 創建索引和唯一約束
		indexes := []string{
			// long_term_memories 主表唯一約束 (修復 ON CONFLICT 支持)
			"ALTER TABLE long_term_memories ADD CONSTRAINT unique_user_character_memory UNIQUE (user_id, character_id)",

			// 相關表索引
			"CREATE INDEX IF NOT EXISTS idx_memory_preferences_memory_id ON memory_preferences(memory_id)",
			"CREATE INDEX IF NOT EXISTS idx_memory_nicknames_memory_id ON memory_nicknames(memory_id)",
			"CREATE INDEX IF NOT EXISTS idx_memory_milestones_memory_id ON memory_milestones(memory_id)",
			"CREATE INDEX IF NOT EXISTS idx_memory_dislikes_memory_id ON memory_dislikes(memory_id)",
			"CREATE INDEX IF NOT EXISTS idx_memory_personal_info_memory_id ON memory_personal_info(memory_id)",
		}

		for _, idx := range indexes {
			if _, err := db.ExecContext(ctx, idx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// 回滾 - 按相反順序刪除表
		tables := []interface{}{
			(*dbmodels.MemoryPersonalInfoDB)(nil),
			(*dbmodels.MemoryDislikeDB)(nil),
			(*dbmodels.MemoryMilestoneDB)(nil),
			(*dbmodels.MemoryNicknameDB)(nil),
			(*dbmodels.MemoryPreferenceDB)(nil),
			(*dbmodels.LongTermMemoryModelDB)(nil),
		}

		for _, table := range tables {
			_, err := db.NewDropTable().
				Model(table).
				IfExists().
				Exec(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}