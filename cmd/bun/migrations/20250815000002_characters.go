package migrations

import (
	"context"

	"github.com/uptrace/bun"
	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// 1. 創建主角色表
		_, err := db.NewCreateTable().
			Model((*dbmodels.CharacterDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 2. 創建角色檔案表 (1:1)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterProfileDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 3. 創建角色本地化表 (1:N)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterLocalizationDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 4. 創建對話風格表 (1:N)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterSpeechStyleDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 5. 創建場景表 (1:N)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterSceneDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 6. 創建狀態表 (1:N)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterStateDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 7. 創建情感配置表 (1:1)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterEmotionalConfigDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 8. 創建 NSFW 配置表 (1:1)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterNSFWConfigDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 9. 創建 NSFW 等級表 (1:N)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterNSFWLevelDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 10. 創建互動規則表 (1:N)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterInteractionRuleDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 11. 創建快照表 (CQRS)
		_, err = db.NewCreateTable().
			Model((*dbmodels.CharacterSnapshotDB)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 創建索引
		indexes := []string{
			// characters 表索引
			"CREATE INDEX IF NOT EXISTS idx_characters_active ON characters(is_active)",
			"CREATE INDEX IF NOT EXISTS idx_characters_type ON characters(type)",
			"CREATE INDEX IF NOT EXISTS idx_characters_locale ON characters(locale)",
			"CREATE INDEX IF NOT EXISTS idx_characters_tags ON characters USING GIN (tags)",
			"CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name)",
			"CREATE INDEX IF NOT EXISTS idx_characters_popularity ON characters(popularity DESC)",

			// speech_styles 索引
			"CREATE INDEX IF NOT EXISTS idx_speech_styles_char ON character_speech_styles(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_speech_styles_active ON character_speech_styles(is_active)",
			"CREATE INDEX IF NOT EXISTS idx_speech_styles_type ON character_speech_styles(style_type)",

			// scenes 索引
			"CREATE INDEX IF NOT EXISTS idx_scenes_char ON character_scenes(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_scenes_type ON character_scenes(scene_type)",
			"CREATE INDEX IF NOT EXISTS idx_scenes_active ON character_scenes(is_active)",

			// states 索引
			"CREATE INDEX IF NOT EXISTS idx_states_char ON character_states(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_states_key ON character_states(state_key)",
			"CREATE INDEX IF NOT EXISTS idx_states_active ON character_states(is_active)",

			// nsfw_levels 索引
			"CREATE INDEX IF NOT EXISTS idx_nsfw_levels_char ON character_nsfw_levels(character_id)",
			"CREATE INDEX IF NOT EXISTS idx_nsfw_levels_level ON character_nsfw_levels(level)",
			"CREATE INDEX IF NOT EXISTS idx_nsfw_levels_engine ON character_nsfw_levels(engine)",
			"CREATE INDEX IF NOT EXISTS idx_nsfw_levels_active ON character_nsfw_levels(is_active)",

			// interaction_rules 索引
			"CREATE INDEX IF NOT EXISTS idx_rules_char ON character_interaction_rules(character_id)",

			// snapshots 索引
			"CREATE INDEX IF NOT EXISTS idx_snapshots_refreshed ON character_snapshots(refreshed_at)",
			"CREATE INDEX IF NOT EXISTS idx_snapshots_version ON character_snapshots(version)",
		}

		for _, idx := range indexes {
			if _, err := db.ExecContext(ctx, idx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// 回滾 - 按相反順序刪除表（避免外鍵約束問題）
		tables := []interface{}{
			(*dbmodels.CharacterSnapshotDB)(nil),
			(*dbmodels.CharacterInteractionRuleDB)(nil),
			(*dbmodels.CharacterNSFWLevelDB)(nil),
			(*dbmodels.CharacterNSFWConfigDB)(nil),
			(*dbmodels.CharacterEmotionalConfigDB)(nil),
			(*dbmodels.CharacterStateDB)(nil),
			(*dbmodels.CharacterSceneDB)(nil),
			(*dbmodels.CharacterSpeechStyleDB)(nil),
			(*dbmodels.CharacterLocalizationDB)(nil),
			(*dbmodels.CharacterProfileDB)(nil),
			(*dbmodels.CharacterDB)(nil),
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