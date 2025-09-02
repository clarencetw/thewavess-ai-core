package migrations

import (
	"context"

	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// 1. 創建聊天表
		_, err := db.NewCreateTable().
			Model((*dbmodels.ChatDB)(nil)).
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
			// chats 表索引
			"CREATE INDEX IF NOT EXISTS idx_chats_user_id ON chats(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_chats_character_id ON chats(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_chats_status ON chats(status)",
			"CREATE INDEX IF NOT EXISTS idx_chats_updated_at ON chats(updated_at)",
			"CREATE INDEX IF NOT EXISTS idx_chats_user_character ON chats(user_id, character_id)",
			"CREATE INDEX IF NOT EXISTS idx_chats_chat_mode ON chats(chat_mode)",

			// messages 表索引
			"CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id)",
			"CREATE INDEX IF NOT EXISTS idx_messages_role ON messages(role)",
			"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)",
			"CREATE INDEX IF NOT EXISTS idx_messages_nsfw_level ON messages(nsfw_level)",
			"CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON messages(chat_id, created_at DESC)",
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
			(*dbmodels.ChatDB)(nil),
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
