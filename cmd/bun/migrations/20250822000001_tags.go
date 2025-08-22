package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// 標籤主表
		_, err := db.NewCreateTable().
			Model((*Tag)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 角色標籤關聯表
		_, err = db.NewCreateTable().
			Model((*CharacterTag)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 添加索引
		_, err = db.NewCreateIndex().
			Model((*CharacterTag)(nil)).
			Index("idx_character_tags_character_id").
			Column("character_id").
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		_, err = db.NewCreateIndex().
			Model((*CharacterTag)(nil)).
			Index("idx_character_tags_tag_id").
			Column("tag_id").
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		// 插入初始標籤數據
		initialTags := []Tag{
			{ID: "tag_001", Name: "甜寵", Category: "genre", Color: "#FF69B4", CreatedAt: time.Now()},
			{ID: "tag_002", Name: "腹黑", Category: "personality", Color: "#8B008B", CreatedAt: time.Now()},
			{ID: "tag_003", Name: "霸總", Category: "role", Color: "#4169E1", CreatedAt: time.Now()},
			{ID: "tag_004", Name: "古風", Category: "style", Color: "#8B4513", CreatedAt: time.Now()},
			{ID: "tag_005", Name: "現代", Category: "style", Color: "#00CED1", CreatedAt: time.Now()},
			{ID: "tag_006", Name: "溫柔", Category: "personality", Color: "#98FB98", CreatedAt: time.Now()},
			{ID: "tag_007", Name: "高冷", Category: "personality", Color: "#87CEEB", CreatedAt: time.Now()},
			{ID: "tag_008", Name: "陽光", Category: "personality", Color: "#FFD700", CreatedAt: time.Now()},
			{ID: "tag_009", Name: "神秘", Category: "personality", Color: "#9370DB", CreatedAt: time.Now()},
			{ID: "tag_010", Name: "浪漫", Category: "genre", Color: "#FFC0CB", CreatedAt: time.Now()},
			{ID: "tag_011", Name: "青春", Category: "genre", Color: "#90EE90", CreatedAt: time.Now()},
			{ID: "tag_012", Name: "都市", Category: "style", Color: "#20B2AA", CreatedAt: time.Now()},
			{ID: "tag_013", Name: "校園", Category: "style", Color: "#87CEFA", CreatedAt: time.Now()},
			{ID: "tag_014", Name: "職場", Category: "style", Color: "#DDA0DD", CreatedAt: time.Now()},
			{ID: "tag_015", Name: "醫生", Category: "role", Color: "#F0E68C", CreatedAt: time.Now()},
			{ID: "tag_016", Name: "律師", Category: "role", Color: "#D2B48C", CreatedAt: time.Now()},
			{ID: "tag_017", Name: "軍人", Category: "role", Color: "#808080", CreatedAt: time.Now()},
			{ID: "tag_018", Name: "老師", Category: "role", Color: "#F5DEB3", CreatedAt: time.Now()},
			{ID: "tag_019", Name: "成熟", Category: "personality", Color: "#CD853F", CreatedAt: time.Now()},
			{ID: "tag_020", Name: "幽默", Category: "personality", Color: "#FFA500", CreatedAt: time.Now()},
		}

		_, err = db.NewInsert().Model(&initialTags).Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// Down migration
		_, err := db.NewDropTable().Model((*CharacterTag)(nil)).IfExists().Exec(ctx)
		if err != nil {
			return err
		}

		_, err = db.NewDropTable().Model((*Tag)(nil)).IfExists().Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}

// Tag 標籤表
type Tag struct {
	bun.BaseModel `bun:"table:tags,alias:t"`

	ID          string    `bun:"id,pk" json:"id"`
	Name        string    `bun:"name,notnull" json:"name"`
	Category    string    `bun:"category,notnull" json:"category"` // genre, personality, role, style
	Color       string    `bun:"color" json:"color"`
	Description string    `bun:"description" json:"description"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// CharacterTag 角色標籤關聯表
type CharacterTag struct {
	bun.BaseModel `bun:"table:character_tags,alias:ct"`

	ID          string    `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	CharacterID string    `bun:"character_id,notnull" json:"character_id"`
	TagID       string    `bun:"tag_id,notnull" json:"tag_id"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`

	// 外鍵約束
	Character *Character `bun:"rel:belongs-to,join:character_id=id"`
	Tag       *Tag       `bun:"rel:belongs-to,join:tag_id=id"`
}

// Character reference for foreign key
type Character struct {
	bun.BaseModel `bun:"table:characters,alias:c"`
	ID            string `bun:"id,pk" json:"id"`
}