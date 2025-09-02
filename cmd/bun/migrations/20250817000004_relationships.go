package migrations

import (
	"context"

	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// 創建關係表
		_, err := db.NewCreateTable().
			Model((*dbmodels.RelationshipDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 創建索引
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_relationships_user_id ON relationships(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_relationships_character_id ON relationships(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_relationships_affection ON relationships(affection)",
			"CREATE INDEX IF NOT EXISTS idx_relationships_updated_at ON relationships(updated_at)",
			"CREATE INDEX IF NOT EXISTS idx_relationships_chat_id ON relationships(chat_id) WHERE chat_id IS NOT NULL",
			"CREATE INDEX IF NOT EXISTS idx_relationships_user_char_chat ON relationships(user_id, character_id, chat_id) WHERE chat_id IS NOT NULL",
			// 保持原有的 (user_id, character_id) 唯一約束，用於全域關係（當 chat_id 為 NULL 時）
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_relationships_unique_user_char_global ON relationships(user_id, character_id) WHERE chat_id IS NULL",
			// 新的約束：每個 chat 只能有一個 relationship 記錄
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_relationships_unique_chat ON relationships(chat_id) WHERE chat_id IS NOT NULL",
			// JSONB 索引，支援情感數據搜尋
			"CREATE INDEX IF NOT EXISTS idx_relationships_emotion_data ON relationships USING GIN((emotion_data))",
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
			Model((*dbmodels.RelationshipDB)(nil)).
			IfExists().
			Exec(ctx)
		return err
	})
}
