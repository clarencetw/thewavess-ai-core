package migrations

import (
	"context"

	"github.com/uptrace/bun"
	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// 1. 創建情感狀態表
		_, err := db.NewCreateTable().
			Model((*dbmodels.EmotionStateDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 2. 創建情感歷史表
		_, err = db.NewCreateTable().
			Model((*dbmodels.EmotionHistoryDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 3. 創建情感里程碑表
		_, err = db.NewCreateTable().
			Model((*dbmodels.EmotionMilestoneDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 創建索引
		indexes := []string{
			// emotion_states 表索引
			"CREATE INDEX IF NOT EXISTS idx_emotion_states_user_id ON emotion_states(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_emotion_states_character_id ON emotion_states(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_emotion_states_affection ON emotion_states(affection)",
			"CREATE INDEX IF NOT EXISTS idx_emotion_states_updated_at ON emotion_states(updated_at)",

			// emotion_history 表索引
			"CREATE INDEX IF NOT EXISTS idx_emotion_history_user_id ON emotion_history(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_emotion_history_character_id ON emotion_history(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_emotion_history_created_at ON emotion_history(created_at)",
			"CREATE INDEX IF NOT EXISTS idx_emotion_history_trigger_type ON emotion_history(trigger_type)",

			// emotion_milestones 表索引
			"CREATE INDEX IF NOT EXISTS idx_emotion_milestones_user_id ON emotion_milestones(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_emotion_milestones_character_id ON emotion_milestones(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_emotion_milestones_achieved_at ON emotion_milestones(achieved_at)",
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
			(*dbmodels.EmotionMilestoneDB)(nil),
			(*dbmodels.EmotionHistoryDB)(nil),
			(*dbmodels.EmotionStateDB)(nil),
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