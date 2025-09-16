package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/clarencetw/thewavess-ai-core/models"
	db "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// CharacterService 角色業務邏輯服務
type CharacterService struct {
	store *CharacterStore
}

// NewCharacterService 創建角色服務
func NewCharacterService(store *CharacterStore) *CharacterService {
	return &CharacterService{
		store: store,
	}
}

// GetCharacter 獲取角色詳情
func (s *CharacterService) GetCharacter(ctx context.Context, id string) (*models.Character, error) {
	if utils.Logger != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
		}).Debug("獲取角色詳情")
	}

	character, err := s.store.GetAggregate(ctx, id)
	if err != nil {
		if utils.Logger != nil {
			utils.Logger.WithFields(logrus.Fields{
				"character_id": id,
				"error":        err.Error(),
			}).Error("獲取角色失敗")
		}
		return nil, err
	}

	return character, nil
}

// ListCharacters 獲取角色列表
func (s *CharacterService) ListCharacters(ctx context.Context, query *models.CharacterListQuery) ([]*models.Character, *models.PaginationResponse, error) {
	if utils.Logger != nil {
		utils.Logger.WithFields(logrus.Fields{
			"page":      query.Page,
			"page_size": query.PageSize,
			"type":      query.Type,
			"search":    query.Search,
		}).Debug("查詢角色列表")
	}

	// 構建資料庫查詢
	dbQuery := s.store.GetDB().NewSelect().
		Model((*db.CharacterDB)(nil)).
		Where("is_active = ?", true)

	// 應用篩選條件
	if query.Type != "" {
		dbQuery = dbQuery.Where("type = ?", query.Type)
	}

	if query.IsActive != nil {
		dbQuery = dbQuery.Where("is_active = ?", *query.IsActive)
	}

	if len(query.Tags) > 0 {
		dbQuery = dbQuery.Where("tags && ?", query.Tags)
	}

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("(name ILIKE ? OR user_description ILIKE ?)", searchPattern, searchPattern)
	}

	// 計算總數
	totalCount, err := dbQuery.Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count characters: %w", err)
	}

	// 應用排序
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	orderBy := fmt.Sprintf("%s %s", sortBy, strings.ToUpper(query.SortOrder))
	dbQuery = dbQuery.Order(orderBy)

	// 應用分頁
	offset := (query.Page - 1) * query.PageSize
	dbQuery = dbQuery.Limit(query.PageSize).Offset(offset)

	// 執行查詢
	var characterDBs []db.CharacterDB
	err = dbQuery.Scan(ctx, &characterDBs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch characters: %w", err)
	}

	// 轉換為領域模型
	characters := make([]*models.Character, len(characterDBs))
	for i, charDB := range characterDBs {
		characters[i] = models.CharacterFromDB(&charDB)
	}

	// 構建分頁響應
	totalPages := (totalCount + query.PageSize - 1) / query.PageSize
	pagination := &models.PaginationResponse{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalCount: int64(totalCount),
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}

	return characters, pagination, nil
}

// SearchCharacters 搜尋角色
func (s *CharacterService) SearchCharacters(ctx context.Context, keyword string, limit int) ([]*models.Character, error) {
	if s.store == nil || s.store.db == nil {
		return nil, fmt.Errorf("character store not initialized")
	}

	var dbCharacters []db.CharacterDB

	// 使用 PostgreSQL 全文搜索
	query := s.store.db.NewSelect().
		Model(&dbCharacters).
		Where("is_active = ?", true).
		Where("name ILIKE ? OR user_description ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%").
		Order("name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("搜尋角色失敗")
		return nil, fmt.Errorf("database search failed: %w", err)
	}

	// 轉換為業務模型
	characters := make([]*models.Character, len(dbCharacters))
	for i, dbChar := range dbCharacters {
		characters[i] = &models.Character{
			ID:              dbChar.ID,
			Name:            dbChar.Name,
			Type:            models.CharacterType(dbChar.Type),
			Locale:          models.Locale(dbChar.Locale),
			IsActive:        dbChar.IsActive,
			UserDescription: dbChar.UserDescription,
			Metadata: models.CharacterMetadata{
				Tags: dbChar.Tags,
			},
			CreatedAt: dbChar.CreatedAt,
			UpdatedAt: dbChar.UpdatedAt,
		}
	}

	utils.Logger.WithFields(logrus.Fields{
		"keyword": keyword,
		"limit":   limit,
		"found":   len(characters),
	}).Info("角色搜尋完成")

	return characters, nil
}

// CreateCharacter 創建角色
func (s *CharacterService) CreateCharacter(ctx context.Context, req *models.CharacterCreateRequest) (*models.Character, error) {
	utils.Logger.WithFields(logrus.Fields{
		"name": req.Name,
		"type": req.Type,
	}).Info("創建新角色")

	// 生成角色ID
	characterID := utils.GenerateCharacterID()

	// 創建角色實例
	character := &models.Character{
		ID: characterID,
	}

	// 從請求填充數據
	character.FromCreateRequest(req)

	// 驗證數據
	if err := character.Validate(); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"error":        err.Error(),
		}).Error("角色數據驗證失敗")
		return nil, err
	}

	// 保存到存儲
	if err := s.store.CreateAggregate(ctx, character); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"error":        err.Error(),
		}).Error("創建角色失敗")
		return nil, err
	}

	utils.Logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"name":         character.Name,
	}).Info("角色創建成功")

	return character, nil
}

// UpdateCharacter 更新角色
func (s *CharacterService) UpdateCharacter(ctx context.Context, id string, req *models.CharacterUpdateRequest) (*models.Character, error) {
	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
	}).Info("更新角色")

	// 獲取現有角色
	character, err := s.store.GetAggregate(ctx, id)
	if err != nil {
		return nil, err
	}

	// 應用更新
	character.ApplyUpdateRequest(req)

	// 驗證數據
	if err := character.Validate(); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"error":        err.Error(),
		}).Error("角色更新數據驗證失敗")
		return nil, err
	}

	// 保存更新
	if err := s.store.UpdateAggregate(ctx, character); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"error":        err.Error(),
		}).Error("更新角色失敗")
		return nil, err
	}

	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
	}).Info("角色更新成功")

	return character, nil
}

// DeleteCharacter 刪除角色
func (s *CharacterService) DeleteCharacter(ctx context.Context, id string) error {
	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
	}).Info("刪除角色")

	err := s.store.DeleteAggregate(ctx, id)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"error":        err.Error(),
		}).Error("刪除角色失敗")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
	}).Info("角色刪除成功")

	return nil
}

// GetCharacterStats 獲取角色統計
func (s *CharacterService) GetCharacterStats(ctx context.Context, id string) (map[string]interface{}, error) {
	character, err := s.store.GetAggregate(ctx, id)
	if err != nil {
		return nil, err
	}

	// 從資料庫獲取實際統計數據

	// 統計真實數據
	chatCount := 0
	messageCount := 0
	activeUsers := 0
	avgAffection := 0.0
	totalRelationships := 0
	relationshipTypes := make(map[string]int)
	intimacyLevels := make(map[string]int)

	// 如果資料庫可用，則統計真實數據
	if s.store.db != nil {
		// 統計聊天數量
		chatCount, _ = s.store.db.NewSelect().
			Model((*db.ChatDB)(nil)).
			Where("character_id = ? AND status != 'deleted'", id).
			Count(ctx)

		// 統計活躍用戶數（有對話的不同用戶）
		var result struct {
			Count int `bun:"count"`
		}
		err = s.store.db.NewSelect().
			ColumnExpr("COUNT(DISTINCT user_id) as count").
			TableExpr("chats").
			Where("character_id = ? AND status = 'active'", id).
			Scan(ctx, &result)
		if err == nil {
			activeUsers = result.Count
		}

		// 統計消息數量
		if chatCount > 0 {
			messageCount, err = s.store.db.NewSelect().
				Model((*db.MessageDB)(nil)).
				Join("JOIN chats ON chats.id = messages.chat_id").
				Where("chats.character_id = ?", id).
				Count(ctx)
			if err != nil {
				utils.Logger.WithError(err).Debug("Failed to count messages")
				messageCount = 0
			}
		}

		// 統計核心關係數據
		var relationshipStats struct {
			AvgAffection   *float64 `bun:"avg_affection"`
			TotalRelations int      `bun:"total_relations"`
		}
		err = s.store.db.NewSelect().
			Model((*db.RelationshipDB)(nil)).
			ColumnExpr("AVG(affection) AS avg_affection, COUNT(*) AS total_relations").
			Where("character_id = ?", id).
			Scan(ctx, &relationshipStats)

		if err == nil {
			if relationshipStats.AvgAffection != nil {
				avgAffection = *relationshipStats.AvgAffection
			}
			totalRelationships = relationshipStats.TotalRelations

			// 統計關係類型分佈（重要業務指標）
			var relationshipTypeResults []struct {
				Relationship string `bun:"relationship"`
				Count        int    `bun:"count"`
			}
			s.store.db.NewSelect().
				Model((*db.RelationshipDB)(nil)).
				ColumnExpr("relationship, COUNT(*) as count").
				Where("character_id = ?", id).
				Group("relationship").
				Scan(ctx, &relationshipTypeResults)

			for _, rt := range relationshipTypeResults {
				relationshipTypes[rt.Relationship] = rt.Count
			}

			// 統計親密度分佈（重要業務指標）
			var intimacyResults []struct {
				IntimacyLevel string `bun:"intimacy_level"`
				Count         int    `bun:"count"`
			}
			s.store.db.NewSelect().
				Model((*db.RelationshipDB)(nil)).
				ColumnExpr("intimacy_level, COUNT(*) as count").
				Where("character_id = ?", id).
				Group("intimacy_level").
				Scan(ctx, &intimacyResults)

			for _, il := range intimacyResults {
				intimacyLevels[il.IntimacyLevel] = il.Count
			}
		}
	}

	// 創建完整統計資料，使用真實數據
	stats := map[string]interface{}{
		"character_id": character.ID,
		"basic_info": map[string]interface{}{
			"name":       character.Name,
			"type":       string(character.Type),
			"is_active":  character.IsActive,
			"created_at": character.CreatedAt,
		},
		"interaction_stats": map[string]interface{}{
			"total_conversations": chatCount,
			"total_messages":      messageCount,
			"total_users":         activeUsers,
			"last_interaction":    character.UpdatedAt,
		},
		"relationship_stats": map[string]interface{}{
			"avg_affection_level": avgAffection,       // 核心指標：平均好感度
			"total_relationships": totalRelationships, // 核心指標：關係總數
			"relationship_types":  relationshipTypes,  // 重要業務指標：關係類型分佈
			"intimacy_levels":     intimacyLevels,     // 重要業務指標：親密度分佈
		},
		"performance_stats": map[string]interface{}{
			"avg_response_time_ms": 0,   // TODO: 實際計算回應時間
			"success_rate":         1.0, // TODO: 實際計算成功率
		},
		"generated_at": utils.Now(),
	}

	return stats, nil
}

// 內部輔助方法

// GetCharacterDB 獲取角色資料庫物件
func (s *CharacterService) GetCharacterDB(ctx context.Context, characterID string) (*db.CharacterDB, error) {
	return s.store.GetCharacterDB(ctx, characterID)
}

// GetCharacterService 獲取角色服務實例
func GetCharacterService() *CharacterService {
	store := NewCharacterStore()
	return NewCharacterService(store)
}

// ===== 帶用戶追蹤的新方法 =====

// CreateCharacterWithUser 創建角色（帶用戶追蹤）
func (s *CharacterService) CreateCharacterWithUser(ctx context.Context, req *models.CharacterCreateRequest, userID string) (*models.Character, error) {
	utils.Logger.WithFields(logrus.Fields{
		"name":    req.Name,
		"type":    req.Type,
		"user_id": userID,
	}).Info("用戶創建新角色")

	// 生成角色ID
	characterID := utils.GenerateCharacterID()

	// 創建角色實例
	character := &models.Character{
		ID: characterID,
	}

	// 從請求填充數據
	character.FromCreateRequest(req)

	// 設定用戶追蹤資訊
	character.CreatedBy = &userID
	character.UpdatedBy = &userID
	character.IsPublic = true  // 創建的角色默認為公開
	character.IsSystem = false // 用戶創建的角色非系統角色

	// 驗證數據
	if err := character.Validate(); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userID,
			"error":        err.Error(),
		}).Error("角色數據驗證失敗")
		return nil, err
	}

	// 保存到存儲
	if err := s.store.CreateAggregateWithUser(ctx, character); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"user_id":      userID,
			"error":        err.Error(),
		}).Error("創建角色失敗")
		return nil, err
	}

	utils.Logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"name":         character.Name,
		"user_id":      userID,
	}).Info("用戶角色創建成功")

	return character, nil
}

// UpdateCharacterWithUser 更新角色（帶用戶追蹤）
func (s *CharacterService) UpdateCharacterWithUser(ctx context.Context, id string, req *models.CharacterUpdateRequest, userID string) (*models.Character, error) {
	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
		"user_id":      userID,
	}).Info("用戶更新角色")

	// 獲取現有角色
	character, err := s.store.GetAggregate(ctx, id)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"user_id":      userID,
			"error":        err.Error(),
		}).Error("獲取角色失敗")
		return nil, err
	}

	// 檢查權限：系統角色只能管理員修改
	if character.IsSystem {
		return nil, models.CharacterError{
			Type:    "SYSTEM_CHARACTER_PROTECTED",
			Message: "系統角色無法由一般用戶修改",
		}
	}

	// 從請求更新數據
	character.ApplyUpdateRequest(req)
	character.UpdatedBy = &userID

	// 驗證更新的數據
	if err := character.Validate(); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"user_id":      userID,
			"error":        err.Error(),
		}).Error("角色更新數據驗證失敗")
		return nil, err
	}

	// 保存更新
	if err := s.store.UpdateAggregateWithUser(ctx, character); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"user_id":      userID,
			"error":        err.Error(),
		}).Error("更新角色失敗")
		return nil, err
	}

	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
		"user_id":      userID,
	}).Info("用戶角色更新成功")

	return character, nil
}

// SoftDeleteCharacterWithUser 軟刪除角色（帶用戶追蹤）
func (s *CharacterService) SoftDeleteCharacterWithUser(ctx context.Context, id string, userID string) error {
	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
		"user_id":      userID,
	}).Info("用戶軟刪除角色")

	// 獲取現有角色檢查權限
	character, err := s.store.GetAggregate(ctx, id)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"user_id":      userID,
			"error":        err.Error(),
		}).Error("獲取角色失敗")
		return err
	}

	// 檢查權限：系統角色不能被刪除
	if character.IsSystem {
		return models.CharacterError{
			Type:    "SYSTEM_CHARACTER_PROTECTED",
			Message: "系統角色無法刪除",
		}
	}

	// 執行軟刪除
	err = s.store.SoftDeleteAggregateWithUser(ctx, id, userID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"user_id":      userID,
			"error":        err.Error(),
		}).Error("軟刪除角色失敗")
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
		"user_id":      userID,
	}).Info("用戶角色軟刪除成功")

	return nil
}
