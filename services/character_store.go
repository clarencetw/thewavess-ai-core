package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

// CharacterStore 角色存儲聚合服務（DB 驅動）
type CharacterStore struct {
	db     *bun.DB
	mapper *models.CharacterMapper
}

// NewCharacterStore 創建角色存儲服務
func NewCharacterStore() *CharacterStore {
	db := GetDB()
	// Note: db might be nil during initialization, methods should handle this gracefully
	return &CharacterStore{
		db:     db,
		mapper: models.NewCharacterMapper(),
	}
}

// GetAggregate 獲取完整的角色聚合
func (s *CharacterStore) GetAggregate(ctx context.Context, id string) (*models.Character, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}
	
	// 先嘗試從快照表獲取
	var snapshot db.CharacterSnapshotDB
	err := s.db.NewSelect().
		Model(&snapshot).
		Where("character_id = ?", id).
		Scan(ctx)

	if err == nil {
		// 從快照恢復角色
		character, err := s.restoreFromSnapshot(&snapshot)
		if err == nil {
			return character, nil
		}
		if utils.Logger != nil {
			utils.Logger.WithError(err).Warn("快照恢復失敗，使用聚合查詢")
		}
	}

	// 從各個表聚合查詢
	return s.getAggregateFromTables(ctx, id)
}

// CreateAggregate 創建角色聚合
func (s *CharacterStore) CreateAggregate(ctx context.Context, character *models.Character) error {
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

	// 轉換為資料庫模型群組
	dbGroup, err := s.mapper.ToDB(character)
	if err != nil {
		return fmt.Errorf("轉換角色模型失敗: %w", err)
	}

	// 使用事務執行創建操作
	err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 插入各個表
		if err := s.insertCharacterTables(ctx, tx, dbGroup); err != nil {
			return err
		}

		// 創建快照
		if err := s.createSnapshot(ctx, tx, character); err != nil {
			return fmt.Errorf("創建快照失敗: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if utils.Logger != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": character.ID,
			"name":         character.Name,
		}).Info("角色聚合創建成功")
	}

	return nil
}

// UpdateAggregate 更新角色聚合
func (s *CharacterStore) UpdateAggregate(ctx context.Context, character *models.Character) error {
	if err := character.Validate(); err != nil {
		return err
	}

	character.UpdatedAt = time.Now()

	// 轉換為資料庫模型群組
	dbGroup, err := s.mapper.ToDB(character)
	if err != nil {
		return fmt.Errorf("轉換角色模型失敗: %w", err)
	}

	// 使用 RunInTx 處理事務
	err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 檢查角色是否存在
		existingCount, err := tx.NewSelect().
			Model((*db.CharacterDB)(nil)).
			Where("id = ?", character.ID).
			Count(ctx)
		if err != nil {
			return fmt.Errorf("檢查角色存在性失敗: %w", err)
		}
		if existingCount == 0 {
			return models.NewCharacterNotFoundError(character.ID)
		}

		// 更新各個表
		if err := s.updateCharacterTables(ctx, tx, dbGroup); err != nil {
			return err
		}

		// 刷新快照
		if err := s.refreshSnapshot(ctx, tx, character); err != nil {
			return fmt.Errorf("刷新快照失敗: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if utils.Logger != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": character.ID,
			"name":         character.Name,
		}).Info("角色聚合更新成功")
	}

	return nil
}

// DeleteAggregate 刪除角色聚合
func (s *CharacterStore) DeleteAggregate(ctx context.Context, id string) error {
	// 使用 RunInTx 處理事務
	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 檢查角色是否存在
		existingCount, err := tx.NewSelect().
			Model((*db.CharacterDB)(nil)).
			Where("id = ?", id).
			Count(ctx)
		if err != nil {
			return fmt.Errorf("檢查角色存在性失敗: %w", err)
		}
		if existingCount == 0 {
			return models.NewCharacterNotFoundError(id)
		}

		// 刪除角色（由於外鍵約束，相關表會級聯刪除）
		_, err = tx.NewDelete().
			Model((*db.CharacterDB)(nil)).
			Where("id = ?", id).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("刪除角色失敗: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if utils.Logger != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
		}).Info("角色聚合刪除成功")
	}

	return nil
}

// List 列表查詢角色
func (s *CharacterStore) List(ctx context.Context, query *models.CharacterListQuery) ([]*models.Character, *models.PaginationResponse, error) {
	if s.db == nil {
		return nil, nil, fmt.Errorf("database connection is not available")
	}
	
	// 構建基本查詢
	baseQuery := s.db.NewSelect().Model((*db.CharacterDB)(nil))

	// 應用篩選條件
	if query.Search != "" {
		baseQuery = baseQuery.Where("name ILIKE ? OR tags::text ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	if query.Type != "" {
		baseQuery = baseQuery.Where("type = ?", query.Type)
	}

	if query.Locale != "" {
		baseQuery = baseQuery.Where("locale = ?", query.Locale)
	}

	if query.IsActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *query.IsActive)
	}

	if len(query.Tags) > 0 {
		baseQuery = baseQuery.Where("tags && ?", query.Tags)
	}

	// 統計總數
	totalCount, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("統計角色總數失敗: %w", err)
	}

	// 應用排序
	sortField := "created_at"
	if query.SortBy != "" {
		sortField = query.SortBy
	}
	sortOrder := "DESC"
	if query.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	baseQuery = baseQuery.Order(sortField + " " + sortOrder)

	// 應用分頁
	offset := (query.Page - 1) * query.PageSize
	baseQuery = baseQuery.Offset(offset).Limit(query.PageSize)

	// 執行查詢
	var characters []db.CharacterDB
	err = baseQuery.Scan(ctx, &characters)
	if err != nil {
		return nil, nil, fmt.Errorf("查詢角色列表失敗: %w", err)
	}

	// 轉換為領域模型（簡化版，不加載完整聚合）
	result := make([]*models.Character, len(characters))
	for i, char := range characters {
		result[i] = &models.Character{
			ID:        char.ID,
			Name:      char.Name,
			Type:      models.CharacterType(char.Type),
			Locale:    models.Locale(char.Locale),
			IsActive:  char.IsActive,
			CreatedAt: char.CreatedAt,
			UpdatedAt: char.UpdatedAt,
			Metadata: models.CharacterMetadata{
				AvatarURL:  char.AvatarURL,
				Tags:       char.Tags,
				Popularity: char.Popularity,
			},
		}
	}

	// 構建分頁響應
	totalPages := (int64(totalCount) + int64(query.PageSize) - 1) / int64(query.PageSize)
	pagination := &models.PaginationResponse{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalCount: int64(totalCount),
		TotalPages: int(totalPages),
		HasNext:    query.Page < int(totalPages),
		HasPrev:    query.Page > 1,
	}

	return result, pagination, nil
}

// GetActive 獲取活躍角色
func (s *CharacterStore) GetActive(ctx context.Context) ([]*models.Character, error) {
	query := &models.CharacterListQuery{
		IsActive: &[]bool{true}[0],
		PageSize: 100, // 假設活躍角色不會太多
		Page:     1,
	}
	
	characters, _, err := s.List(ctx, query)
	return characters, err
}

// Search 搜索角色
func (s *CharacterStore) Search(ctx context.Context, keyword string, limit int) ([]*models.Character, error) {
	query := &models.CharacterListQuery{
		Search:   keyword,
		PageSize: limit,
		Page:     1,
	}
	
	characters, _, err := s.List(ctx, query)
	return characters, err
}

// Count 統計角色總數
func (s *CharacterStore) Count(ctx context.Context) (int64, error) {
	count, err := s.db.NewSelect().
		Model((*db.CharacterDB)(nil)).
		Count(ctx)
	return int64(count), err
}

// CountByType 按類型統計角色數量
func (s *CharacterStore) CountByType(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Type  string `bun:"type"`
		Count int64  `bun:"count"`
	}

	err := s.db.NewSelect().
		Model((*db.CharacterDB)(nil)).
		Column("type").
		ColumnExpr("COUNT(*) as count").
		Group("type").
		Scan(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("按類型統計角色失敗: %w", err)
	}

	counts := make(map[string]int64)
	for _, result := range results {
		counts[result.Type] = result.Count
	}

	return counts, nil
}

// CountByLocale 按語言統計角色數量
func (s *CharacterStore) CountByLocale(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Locale string `bun:"locale"`
		Count  int64  `bun:"count"`
	}

	err := s.db.NewSelect().
		Model((*db.CharacterDB)(nil)).
		Column("locale").
		ColumnExpr("COUNT(*) as count").
		Group("locale").
		Scan(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("按語言統計角色失敗: %w", err)
	}

	counts := make(map[string]int64)
	for _, result := range results {
		counts[result.Locale] = result.Count
	}

	return counts, nil
}

// IsHealthy 健康檢查
func (s *CharacterStore) IsHealthy(ctx context.Context) bool {
	// 簡單的資料庫連接檢查
	var result int
	err := s.db.NewSelect().
		ColumnExpr("1").
		Scan(ctx, &result)
	return err == nil
}

// 內部輔助方法

// getAggregateFromTables 從各個表聚合查詢角色
func (s *CharacterStore) getAggregateFromTables(ctx context.Context, id string) (*models.Character, error) {
	// 查詢主要角色資訊
	var charDB db.CharacterDB
	err := s.db.NewSelect().
		Model(&charDB).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, models.NewCharacterNotFoundError(id)
		}
		return nil, fmt.Errorf("查詢角色失敗: %w", err)
	}

	// 查詢角色檔案
	var profileDB db.CharacterProfileDB
	err = s.db.NewSelect().
		Model(&profileDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return nil, fmt.Errorf("查詢角色檔案失敗: %w", err)
	}

	// 查詢本地化資訊
	var localizationsDB []*db.CharacterLocalizationDB
	err = s.db.NewSelect().
		Model(&localizationsDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("查詢角色本地化失敗: %w", err)
	}

	// 查詢對話風格
	var speechStylesDB []*db.CharacterSpeechStyleDB
	err = s.db.NewSelect().
		Model(&speechStylesDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("查詢角色對話風格失敗: %w", err)
	}

	// 查詢場景
	var scenesDB []*db.CharacterSceneDB
	err = s.db.NewSelect().
		Model(&scenesDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("查詢角色場景失敗: %w", err)
	}

	// 查詢狀態
	var statesDB []*db.CharacterStateDB
	err = s.db.NewSelect().
		Model(&statesDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("查詢角色狀態失敗: %w", err)
	}

	// 查詢情感配置
	var emotionalConfigDB db.CharacterEmotionalConfigDB
	err = s.db.NewSelect().
		Model(&emotionalConfigDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return nil, fmt.Errorf("查詢角色情感配置失敗: %w", err)
	}

	// 查詢NSFW配置
	var nsfwConfigDB db.CharacterNSFWConfigDB
	err = s.db.NewSelect().
		Model(&nsfwConfigDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return nil, fmt.Errorf("查詢角色NSFW配置失敗: %w", err)
	}

	// 查詢NSFW等級
	var nsfwLevelsDB []*db.CharacterNSFWLevelDB
	err = s.db.NewSelect().
		Model(&nsfwLevelsDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("查詢角色NSFW等級失敗: %w", err)
	}

	// 查詢互動規則
	var interactionRulesDB []*db.CharacterInteractionRuleDB
	err = s.db.NewSelect().
		Model(&interactionRulesDB).
		Where("character_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("查詢角色互動規則失敗: %w", err)
	}

	// 使用映射器轉換為領域模型
	var profilePtr *db.CharacterProfileDB
	if profileDB.CharacterID != "" {
		profilePtr = &profileDB
	}
	var emotionalPtr *db.CharacterEmotionalConfigDB
	if emotionalConfigDB.CharacterID != "" {
		emotionalPtr = &emotionalConfigDB
	}
	var nsfwPtr *db.CharacterNSFWConfigDB
	if nsfwConfigDB.CharacterID != "" {
		nsfwPtr = &nsfwConfigDB
	}

	return s.mapper.FromDB(
		&charDB,
		profilePtr,
		localizationsDB,
		speechStylesDB,
		scenesDB,
		statesDB,
		emotionalPtr,
		nsfwPtr,
		nsfwLevelsDB,
		interactionRulesDB,
	)
}

// restoreFromSnapshot 從快照恢復角色
func (s *CharacterStore) restoreFromSnapshot(snapshot *db.CharacterSnapshotDB) (*models.Character, error) {
	// 這裡應該實現從 JSONB 快照恢復角色的邏輯
	// 目前先返回錯誤，讓它回退到聚合查詢
	return nil, fmt.Errorf("快照恢復功能尚未實現")
}

// insertCharacterTables 插入角色相關表格
func (s *CharacterStore) insertCharacterTables(ctx context.Context, tx bun.Tx, dbGroup *models.CharacterDBGroup) error {
	// 插入主角色表
	_, err := tx.NewInsert().
		Model(dbGroup.Character).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("插入角色失敗: %w", err)
	}

	// 插入檔案表
	if dbGroup.Profile != nil {
		_, err = tx.NewInsert().
			Model(dbGroup.Profile).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入角色檔案失敗: %w", err)
		}
	}

	// 插入其他相關表
	if len(dbGroup.Localizations) > 0 {
		_, err = tx.NewInsert().
			Model(&dbGroup.Localizations).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入本地化失敗: %w", err)
		}
	}

	if len(dbGroup.SpeechStyles) > 0 {
		_, err = tx.NewInsert().
			Model(&dbGroup.SpeechStyles).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入對話風格失敗: %w", err)
		}
	}

	if len(dbGroup.Scenes) > 0 {
		_, err = tx.NewInsert().
			Model(&dbGroup.Scenes).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入場景失敗: %w", err)
		}
	}

	if len(dbGroup.States) > 0 {
		_, err = tx.NewInsert().
			Model(&dbGroup.States).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入狀態失敗: %w", err)
		}
	}

	if dbGroup.EmotionalConfig != nil {
		_, err = tx.NewInsert().
			Model(dbGroup.EmotionalConfig).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入情感配置失敗: %w", err)
		}
	}

	if dbGroup.NSFWConfig != nil {
		_, err = tx.NewInsert().
			Model(dbGroup.NSFWConfig).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入NSFW配置失敗: %w", err)
		}
	}

	if len(dbGroup.NSFWLevels) > 0 {
		_, err = tx.NewInsert().
			Model(&dbGroup.NSFWLevels).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入NSFW等級失敗: %w", err)
		}
	}

	if len(dbGroup.InteractionRules) > 0 {
		_, err = tx.NewInsert().
			Model(&dbGroup.InteractionRules).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("插入互動規則失敗: %w", err)
		}
	}

	return nil
}

// updateCharacterTables 更新角色相關表格
func (s *CharacterStore) updateCharacterTables(ctx context.Context, tx bun.Tx, dbGroup *models.CharacterDBGroup) error {
	// 更新主角色表
	_, err := tx.NewUpdate().
		Model(dbGroup.Character).
		Where("id = ?", dbGroup.Character.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新角色失敗: %w", err)
	}

	// 對於關聯表，簡單起見先刪除再插入
	characterID := dbGroup.Character.ID

	// 清理並重新插入關聯數據
	tables := []string{
		"character_localizations",
		"character_speech_styles", 
		"character_scenes",
		"character_states",
		"character_nsfw_levels",
		"character_interaction_rules",
	}

	for _, table := range tables {
		_, err = tx.NewDelete().
			Table(table).
			Where("character_id = ?", characterID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("清理表 %s 失敗: %w", table, err)
		}
	}

	// 重新插入數據（除了主表和配置表）
	if len(dbGroup.Localizations) > 0 {
		_, err = tx.NewInsert().
			Model(&dbGroup.Localizations).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("重新插入本地化失敗: %w", err)
		}
	}

	// 繼續插入其他表...
	return s.insertCharacterTables(ctx, tx, dbGroup)
}

// createSnapshot 創建快照
func (s *CharacterStore) createSnapshot(ctx context.Context, tx bun.Tx, character *models.Character) error {
	// 簡單實現：暫時跳過快照創建
	// 實際實現需要將角色序列化為 JSONB
	if utils.Logger != nil {
		utils.Logger.WithField("character_id", character.ID).Debug("跳過快照創建（待實現）")
	}
	return nil
}

// refreshSnapshot 刷新快照
func (s *CharacterStore) refreshSnapshot(ctx context.Context, tx bun.Tx, character *models.Character) error {
	// 簡單實現：暫時跳過快照刷新
	if utils.Logger != nil {
		utils.Logger.WithField("character_id", character.ID).Debug("跳過快照刷新（待實現）")
	}
	return nil
}

// 全域服務實例
var globalCharacterStore *CharacterStore

// GetCharacterStore 獲取全域角色存儲實例
func GetCharacterStore() *CharacterStore {
	if globalCharacterStore == nil {
		globalCharacterStore = NewCharacterStore()
	}
	return globalCharacterStore
}

// safeStringValue 安全地從字符串指針獲取值
func (s *CharacterStore) safeStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}