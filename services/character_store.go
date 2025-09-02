package services

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// CharacterStore 角色存儲服務
type CharacterStore struct {
	db *bun.DB
}

// GetDB 獲取資料庫連接
func (s *CharacterStore) GetDB() *bun.DB {
	return s.db
}

// NewCharacterStore 創建角色存儲服務
func NewCharacterStore() *CharacterStore {
	db := GetDB()
	return &CharacterStore{
		db: db,
	}
}

// GetCharacterDB 獲取角色DB對象
func (s *CharacterStore) GetCharacterDB(ctx context.Context, id string) (*db.CharacterDB, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	var characterDB db.CharacterDB
	err := s.db.NewSelect().
		Model(&characterDB).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get character from database: %w", err)
	}

	return &characterDB, nil
}

// GetAggregate 獲取完整的角色聚合
func (s *CharacterStore) GetAggregate(ctx context.Context, id string) (*models.Character, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	// 獲取主角色資料
	var characterDB db.CharacterDB
	err := s.db.NewSelect().
		Model(&characterDB).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("character not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 角色轉換
	character := &models.Character{
		ID:              characterDB.ID,
		Name:            characterDB.Name,
		Type:            models.CharacterType(characterDB.Type),
		Locale:          models.Locale(characterDB.Locale),
		IsActive:        characterDB.IsActive,
		UserDescription: characterDB.UserDescription,
		CreatedAt:       characterDB.CreatedAt,
		UpdatedAt:       characterDB.UpdatedAt,
	}

	// 設置基本 metadata
	character.Metadata = models.CharacterMetadata{
		AvatarURL:  characterDB.AvatarURL,
		Tags:       characterDB.Tags,
		Popularity: characterDB.Popularity,
	}

	return character, nil
}

// CreateAggregate 創建角色聚合
func (s *CharacterStore) CreateAggregate(ctx context.Context, character *models.Character) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	if err := character.Validate(); err != nil {
		return err
	}

	// 生成ID（如果沒有的話）
	if character.ID == "" {
		character.ID = utils.GenerateCharacterID()
	}

	now := time.Now()
	character.CreatedAt = now
	character.UpdatedAt = now

	// 創建角色記錄
	characterDB := &db.CharacterDB{
		ID:              character.ID,
		Name:            character.Name,
		Type:            string(character.Type),
		Locale:          string(character.Locale),
		IsActive:        character.IsActive,
		AvatarURL:       character.Metadata.AvatarURL,
		Tags:            character.Metadata.Tags,
		Popularity:      character.Metadata.Popularity,
		UserDescription: character.UserDescription,
		CreatedAt:       character.CreatedAt,
		UpdatedAt:       character.UpdatedAt,
	}

	_, err := s.db.NewInsert().Model(characterDB).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// UpdateAggregate 更新角色聚合
func (s *CharacterStore) UpdateAggregate(ctx context.Context, character *models.Character) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	if err := character.Validate(); err != nil {
		return err
	}

	character.UpdatedAt = time.Now()

	// 更新角色記錄
	characterDB := &db.CharacterDB{
		ID:              character.ID,
		Name:            character.Name,
		Type:            string(character.Type),
		Locale:          string(character.Locale),
		IsActive:        character.IsActive,
		AvatarURL:       character.Metadata.AvatarURL,
		Tags:            character.Metadata.Tags,
		Popularity:      character.Metadata.Popularity,
		UserDescription: character.UserDescription,
		UpdatedAt:       character.UpdatedAt,
	}

	_, err := s.db.NewUpdate().
		Table("characters").
		Set("name = ?", characterDB.Name).
		Set("type = ?", characterDB.Type).
		Set("locale = ?", characterDB.Locale).
		Set("is_active = ?", characterDB.IsActive).
		Set("avatar_url = ?", characterDB.AvatarURL).
		Set("popularity = ?", characterDB.Popularity).
		Set("tags = ?", pgdialect.Array(characterDB.Tags)).
		Set("user_description = ?", characterDB.UserDescription).
		Set("updated_at = ?", characterDB.UpdatedAt).
		Where("id = ?", character.ID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	return nil
}

// DeleteAggregate 刪除角色聚合
func (s *CharacterStore) DeleteAggregate(ctx context.Context, id string) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	_, err := s.db.NewDelete().
		Model((*db.CharacterDB)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	return nil
}

// ===== 帶用戶追蹤的新方法 =====

// CreateAggregateWithUser 創建角色聚合（帶用戶追蹤）
func (s *CharacterStore) CreateAggregateWithUser(ctx context.Context, character *models.Character) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	// 轉換為DB模型
	characterDB := &db.CharacterDB{
		ID:              character.ID,
		Name:            character.Name,
		Type:            string(character.Type),
		Locale:          string(character.Locale),
		IsActive:        character.IsActive,
		AvatarURL:       character.Metadata.AvatarURL,
		Popularity:      character.Metadata.Popularity,
		Tags:            character.Metadata.Tags,
		UserDescription: character.UserDescription,
		CreatedBy:       character.CreatedBy,
		UpdatedBy:       character.UpdatedBy,
		IsPublic:        character.IsPublic,
		IsSystem:        character.IsSystem,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 插入主表
	_, err := s.db.NewInsert().
		Model(characterDB).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// UpdateAggregateWithUser 更新角色聚合（帶用戶追蹤）
func (s *CharacterStore) UpdateAggregateWithUser(ctx context.Context, character *models.Character) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	// 更新主表
	_, err := s.db.NewUpdate().
		Model((*db.CharacterDB)(nil)).
		Set("name = ?", character.Name).
		Set("type = ?", character.Type).
		Set("locale = ?", character.Locale).
		Set("is_active = ?", character.IsActive).
		Set("avatar_url = ?", character.Metadata.AvatarURL).
		Set("popularity = ?", character.Metadata.Popularity).
		Set("tags = ?", pgdialect.Array(character.Metadata.Tags)).
		Set("user_description = ?", character.UserDescription).
		Set("updated_by = ?", character.UpdatedBy).
		Set("is_public = ?", character.IsPublic).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", character.ID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	return nil
}

// SoftDeleteAggregateWithUser 軟刪除角色聚合（帶用戶追蹤）
func (s *CharacterStore) SoftDeleteAggregateWithUser(ctx context.Context, id string, userID string) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	// 檢查角色是否存在且未被刪除
	var count int
	count, err := s.db.NewSelect().
		Model((*db.CharacterDB)(nil)).
		Where("id = ?", id).
		Count(ctx)

	if err != nil {
		return fmt.Errorf("failed to check character existence: %w", err)
	}

	if count == 0 {
		return models.CharacterError{
			Type:    "CHARACTER_NOT_FOUND",
			Message: "角色不存在或已被刪除",
		}
	}

	// 執行軟刪除 - 使用 Bun 標準方法
	_, err = s.db.NewDelete().
		Model((*db.CharacterDB)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to soft delete character: %w", err)
	}

	return nil
}
