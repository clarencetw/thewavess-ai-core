package services

import (
	"context"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// stringValue 安全地解引用字符串指針
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// stringPtr 將字符串轉換為指針
func stringPtr(s string) *string {
	return &s
}

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
				"error":       err.Error(),
			}).Error("獲取角色失敗")
		}
		return nil, err
	}
	
	return character, nil
}


// GetBestSpeechStyle 獲取最適合的對話風格
func (s *CharacterService) GetBestSpeechStyle(ctx context.Context, characterID string, nsfwLevel int, conversationContext *ConversationContext) (*models.CharacterSpeechStyle, error) {
	character, err := s.store.GetAggregate(ctx, characterID)
	if err != nil {
		return nil, err
	}
	
	affection := 50 // 預設好感度
	if conversationContext != nil && conversationContext.EmotionState != nil {
		affection = conversationContext.EmotionState.Affection
	}
	
	style := character.GetBestSpeechStyle(models.NSFWLevel(nsfwLevel), affection)
	if style == nil {
		// 回傳預設風格
		if len(character.Behavior.SpeechStyles) > 0 {
			return &character.Behavior.SpeechStyles[0], nil
		}
		return nil, models.NewCharacterConfigurationError("角色沒有可用的對話風格")
	}
	
	return style, nil
}

// GetNSFWGuideline 獲取NSFW指引
func (s *CharacterService) GetNSFWGuideline(ctx context.Context, level int, locale string, engine string) (*models.CharacterNSFWLevel, error) {
	// 這個方法現在返回通用的NSFW配置，不依賴特定角色
	// 可以根據需要擴展為按角色返回不同配置
	
	nsfwLevel := models.NSFWLevel(level)
	engineType := models.EngineType(engine)
	
	if !nsfwLevel.IsValid() {
		return nil, models.NewCharacterValidationError("level", "無效的NSFW等級")
	}
	
	if !engineType.IsValid() {
		return nil, models.NewCharacterValidationError("engine", "無效的引擎類型")
	}
	
	// 返回通用NSFW配置
	guideline := &models.CharacterNSFWLevel{
		Level:       nsfwLevel,
		Engine:      engineType,
		Title:       stringPtr(s.getNSFWTitle(nsfwLevel)),
		Description: stringPtr(s.getNSFWDescription(nsfwLevel)),
		Guidelines:  stringPtr(s.getNSFWGuidelines(nsfwLevel)),
		PositiveKeywords: s.getNSFWPositiveKeywords(nsfwLevel),
		NegativeKeywords: s.getNSFWNegativeKeywords(nsfwLevel),
		IsActive:    true,
	}
	
	if nsfwLevel >= models.NSFWLevelAdult {
		temp := 0.8
		guideline.Temperature = &temp
	}
	
	return guideline, nil
}

// GetCharacterScenes 獲取角色場景
func (s *CharacterService) GetCharacterScenes(ctx context.Context, characterID, sceneType, timeOfDay string, affection, nsfwLevel int) ([]models.CharacterScene, error) {
	character, err := s.store.GetAggregate(ctx, characterID)
	if err != nil {
		return nil, err
	}
	
	scenes := character.GetActiveScenes(sceneType, timeOfDay, affection, models.NSFWLevel(nsfwLevel))
	return scenes, nil
}

// ListCharacters 獲取角色列表
func (s *CharacterService) ListCharacters(ctx context.Context, query *models.CharacterListQuery) ([]*models.Character, *models.PaginationResponse, error) {
	if utils.Logger != nil {
		utils.Logger.WithFields(logrus.Fields{
			"page":      query.Page,
			"page_size": query.PageSize,
			"type":      query.Type,
			"locale":    query.Locale,
			"search":    query.Search,
		}).Debug("查詢角色列表")
	}
	
	characters, pagination, err := s.store.List(ctx, query)
	if err != nil {
		if utils.Logger != nil {
			utils.Logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("查詢角色列表失敗")
		}
		return nil, nil, err
	}
	
	if utils.Logger != nil {
		utils.Logger.WithFields(logrus.Fields{
			"total_count": pagination.TotalCount,
			"result_count": len(characters),
		}).Debug("角色列表查詢成功")
	}
	
	return characters, pagination, nil
}

// GetActiveCharacters 獲取活躍角色
func (s *CharacterService) GetActiveCharacters(ctx context.Context) ([]*models.Character, error) {
	return s.store.GetActive(ctx)
}

// SearchCharacters 搜尋角色
func (s *CharacterService) SearchCharacters(ctx context.Context, keyword string, limit int) ([]*models.Character, error) {
	return s.store.Search(ctx, keyword, limit)
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
			"error":       err.Error(),
		}).Error("角色數據驗證失敗")
		return nil, err
	}
	
	// 保存到存儲
	if err := s.store.CreateAggregate(ctx, character); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"error":       err.Error(),
		}).Error("創建角色失敗")
		return nil, err
	}
	
	utils.Logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"name":        character.Name,
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
			"error":       err.Error(),
		}).Error("角色更新數據驗證失敗")
		return nil, err
	}
	
	// 保存更新
	if err := s.store.UpdateAggregate(ctx, character); err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": id,
			"error":       err.Error(),
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
			"error":       err.Error(),
		}).Error("刪除角色失敗")
		return err
	}
	
	utils.Logger.WithFields(logrus.Fields{
		"character_id": id,
	}).Info("角色刪除成功")
	
	return nil
}

// GetCharacterStats 獲取角色統計
func (s *CharacterService) GetCharacterStats(ctx context.Context, id string) (*models.CharacterStats, error) {
	character, err := s.store.GetAggregate(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 這裡應該從統計服務或資料庫獲取實際統計數據
	// 目前返回模擬數據
	stats := &models.CharacterStats{
		CharacterID: character.ID,
		BasicInfo: models.CharacterBasicInfo{
			Name:        character.Name,
			Type:        string(character.Type),
			Description: stringValue(character.Metadata.Description),
			Tags:        character.Metadata.Tags,
			Popularity:  character.Metadata.Popularity,
			IsActive:    character.IsActive,
			CreatedAt:   character.CreatedAt,
			Version:     "1.0.0",
		},
		// 其他統計數據可以根據實際需求填充
		GeneratedAt: utils.Now(),
	}
	
	return stats, nil
}

// HealthCheck 健康檢查
func (s *CharacterService) HealthCheck(ctx context.Context) bool {
	return s.store.IsHealthy(ctx)
}

// 內部輔助方法

func (s *CharacterService) getNSFWTitle(level models.NSFWLevel) string {
	titles := map[models.NSFWLevel]string{
		models.NSFWLevelSafe:     "日常對話（安全）",
		models.NSFWLevelRomantic: "浪漫內容（含情感與曖昧）",
		models.NSFWLevelIntimate: "親密內容（牽手/擁抱/親吻/貼近）",
		models.NSFWLevelAdult:    "成人內容（中式優雅版）",
		models.NSFWLevelExplicit: "深度親密內容（中式進階版）",
	}
	
	if title, exists := titles[level]; exists {
		return title
	}
	return "未知等級"
}

func (s *CharacterService) getNSFWDescription(level models.NSFWLevel) string {
	descriptions := map[models.NSFWLevel]string{
		models.NSFWLevelSafe:     "安全的日常對話內容",
		models.NSFWLevelRomantic: "包含浪漫情感的對話內容",
		models.NSFWLevelIntimate: "親密但不涉及明確成人內容",
		models.NSFWLevelAdult:    "成人向內容但保持優雅品味",
		models.NSFWLevelExplicit: "明確的成人內容，維持文學美感",
	}
	
	if desc, exists := descriptions[level]; exists {
		return desc
	}
	return "未知描述"
}

func (s *CharacterService) getNSFWGuidelines(level models.NSFWLevel) string {
	guidelines := map[models.NSFWLevel]string{
		models.NSFWLevelSafe:     "輕微的浪漫暗示，保持優雅含蓄",
		models.NSFWLevelRomantic: "適度的親密描述，注重情感細節",
		models.NSFWLevelIntimate: "更直接的親密內容，但要有品味",
		models.NSFWLevelAdult:    "明確的身體接觸和情感表達，保持中式含蓄美學",
		models.NSFWLevelExplicit: "深度親密內容，維持中式文學美感",
	}
	
	if guideline, exists := guidelines[level]; exists {
		return guideline
	}
	return "未知指引"
}

func (s *CharacterService) getNSFWPositiveKeywords(level models.NSFWLevel) []string {
	keywords := map[models.NSFWLevel][]string{
		models.NSFWLevelSafe:     {"溫柔", "關懷", "溫暖", "體貼"},
		models.NSFWLevelRomantic: {"愛意", "深情", "浪漫", "心動"},
		models.NSFWLevelIntimate: {"親密", "擁抱", "親吻", "愛撫"},
		models.NSFWLevelAdult:    {"熱情", "渴望", "迷戀", "沉醉"},
		models.NSFWLevelExplicit: {"激情", "纏綿", "銷魂", "極致"},
	}
	
	if words, exists := keywords[level]; exists {
		return words
	}
	return []string{}
}

func (s *CharacterService) getNSFWNegativeKeywords(level models.NSFWLevel) []string {
	keywords := map[models.NSFWLevel][]string{
		models.NSFWLevelSafe:     {"露骨", "挑逗", "誘惑", "慾望"},
		models.NSFWLevelRomantic: {"粗俗", "下流", "猥褻", "淫穢"},
		models.NSFWLevelIntimate: {"粗暴", "強迫", "侵犯", "羞辱"},
		models.NSFWLevelAdult:    {"暴力", "虐待", "痛苦", "恐懼"},
		models.NSFWLevelExplicit: {"殘忍", "變態", "病態", "極端"},
	}
	
	if words, exists := keywords[level]; exists {
		return words
	}
	return []string{}
}

// GetCharacterService 獲取角色服務實例
func GetCharacterService() *CharacterService {
	store := GetCharacterStore()
	return NewCharacterService(store)
}