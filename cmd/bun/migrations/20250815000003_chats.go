package migrations

import (
	"context"

	"github.com/uptrace/bun"
	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// 1. 創建聊天會話表
		_, err := db.NewCreateTable().
			Model((*dbmodels.ChatSessionDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 2. 創建消息表
		_, err = db.NewCreateTable().
			Model((*dbmodels.MessageDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 創建索引
		indexes := []string{
			// chat_sessions 表索引
			"CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_id ON chat_sessions(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_chat_sessions_character_id ON chat_sessions(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_chat_sessions_status ON chat_sessions(status)",
			"CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated_at ON chat_sessions(updated_at)",
			"CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_character ON chat_sessions(user_id, character_id)",

			// messages 表索引
			"CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id)",
			"CREATE INDEX IF NOT EXISTS idx_messages_role ON messages(role)",
			"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)",
			"CREATE INDEX IF NOT EXISTS idx_messages_nsfw_level ON messages(nsfw_level)",
			"CREATE INDEX IF NOT EXISTS idx_messages_session_created ON messages(session_id, created_at DESC)",
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
			(*dbmodels.MessageDB)(nil),
			(*dbmodels.ChatSessionDB)(nil),
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