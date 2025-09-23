package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// CreateChatSession godoc
// @Summary      創建聊天會話
// @Description  創建新的聊天會話
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat body models.CreateChatRequest true "聊天信息"
// @Success      201 {object} models.APIResponse{data=models.ChatResponse} "創建成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chats [post]
func CreateChatSession(c *gin.Context) {

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var req models.CreateChatRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 驗證角色是否存在
	characterExists, err := GetDB().NewSelect().
		Model((*db.CharacterDB)(nil)).
		Where("id = ? AND is_active = ?", req.CharacterID, true).
		Exists(context.Background())

	if err != nil || !characterExists {
		utils.Logger.WithError(err).WithField("character_id", req.CharacterID).Error("Character not found")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在或已停用",
			},
		})
		return
	}

	// 使用 character service 獲取完整角色信息
	characterService := services.GetCharacterService()
	character, err := characterService.GetCharacter(context.Background(), req.CharacterID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_SERVICE_ERROR",
				Message: "無法獲取角色信息",
			},
		})
		return
	}

	// 創建新的聊天會話
	chatID := utils.GenerateChatID()
	var chat models.Chat
	chatTitle := req.Title
	if chatTitle == "" {
		chatTitle = "與 " + character.Name + " 的對話"
	}

	// 創建 DB 模型並插入
	chatDB := db.ChatDB{
		ID:          chatID,
		UserID:      userID.(string),
		CharacterID: req.CharacterID,
		Title:       chatTitle,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 使用事務同時創建聊天和關係記錄
	// 重要：必須同時創建關係記錄，否則關係端點會返回404錯誤
	// 這確保了多會話架構中每個對話都有獨立的關係狀態
	ctx := context.Background()

	// 使用 RunInTx 進行事務處理 - Bun ORM 最佳實踐
	err = GetDB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 插入聊天記錄
		_, err := tx.NewInsert().Model(&chatDB).Exec(ctx)
		if err != nil {
			utils.Logger.WithError(err).Error("Failed to create chat session")
			return err
		}

		// 初始化關係記錄 - 為多會話架構設置默認值
		// 注意：ChatID 必須是指針類型 (*string)，因為數據庫模型定義為可選字段
		relationshipDB := db.RelationshipDB{
			ID:                utils.GenerateRelationshipID(),
			UserID:            userID.(string),
			CharacterID:       req.CharacterID,
			ChatID:            &chatID,
			Affection:         50, // 默認好感度
			Mood:              "neutral",
			Relationship:      "stranger",
			IntimacyLevel:     "casual",
			TotalInteractions: 0,
			LastInteraction:   time.Now(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		_, err = tx.NewInsert().Model(&relationshipDB).Exec(ctx)
		if err != nil {
			utils.Logger.WithError(err).Error("Failed to create relationship record")
			return err
		}

		return nil
	})

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to create chat and relationship")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "創建聊天會話失敗",
			},
		})
		return
	}

	// 轉換到 domain 模型
	chat = models.Chat{
		ID:              chatDB.ID,
		UserID:          chatDB.UserID,
		CharacterID:     chatDB.CharacterID,
		Title:           chatDB.Title,
		Status:          chatDB.Status,
		ChatMode:        chatDB.ChatMode,
		MessageCount:    chatDB.MessageCount,
		TotalCharacters: chatDB.TotalCharacters,
		LastMessageAt:   chatDB.LastMessageAt,
		CreatedAt:       chatDB.CreatedAt,
		UpdatedAt:       chatDB.UpdatedAt,
	}

	// 關聯角色信息
	chat.Character = character

	// 準備響應數據
	response := chat.ToResponse()

	// 生成歡迎消息作為第一條消息
	chatService := services.GetChatService()
	welcomeMessage, err := chatService.GenerateWelcomeMessage(ctx, userID.(string), req.CharacterID, chatID)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to generate welcome message")
		// 不阻塞會話創建，繼續返回成功響應
	} else if welcomeMessage != nil {
		// 將歡迎消息作為 LastMessage 包含在響應中
		response.LastMessage = &models.MessageResponse{
			ID:             welcomeMessage.MessageID,
			ChatID:         welcomeMessage.ChatID,
			Role:           "assistant",
			Dialogue:       welcomeMessage.Content,
			EmotionalState: map[string]interface{}{"affection": welcomeMessage.Affection},
			AIEngine:       welcomeMessage.AIEngine,
			ResponseTimeMs: int(welcomeMessage.ResponseTime.Milliseconds()),
			NSFWLevel:      welcomeMessage.NSFWLevel,
			CreatedAt:      time.Now(),
		}

		// 更新會話計數和時間
		response.MessageCount = 1
		now := time.Now()
		response.LastMessageAt = &now
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "聊天會話創建成功",
		Data:    response,
	})
}

// GetChatSession godoc
// @Summary      獲取會話詳情
// @Description  特定聊天會話的詳細信息
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "會話ID"
// @Success      200 {object} models.APIResponse{data=models.ChatResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chats/{chat_id} [get]
func GetChatSession(c *gin.Context) {

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	sessionID := c.Param("chat_id")

	var chatDB db.ChatDB
	err := GetDB().NewSelect().
		Model(&chatDB).
		Where("id = ? AND user_id = ? AND status != ?", sessionID, userID, "deleted").
		Scan(context.Background())

	if err != nil {
		utils.Logger.WithError(err).WithField("chat_id", sessionID).Error("Failed to query chat session")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "聊天會話不存在",
			},
		})
		return
	}

	// 轉換為領域模型
	chat := models.ChatFromDB(&chatDB)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取聊天會話成功",
		Data:    chat.ToResponse(),
	})
}

// GetChatSessions godoc
// @Summary      獲取會話列表
// @Description  支援分頁和角色篩選的會話列表
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Param        status query string false "會話狀態篩選"
// @Param        character_id query string false "角色ID篩選"
// @Success      200 {object} models.APIResponse{data=models.ChatListResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chats [get]
func GetChatSessions(c *gin.Context) {

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	// 解析查詢參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// 構建查詢
	query := GetDB().NewSelect().
		Model((*db.ChatDB)(nil)).
		Where("user_id = ? AND status != ?", userID, "deleted")

	// 應用狀態篩選
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 應用角色篩選
	if characterID := c.Query("character_id"); characterID != "" {
		query = query.Where("character_id = ?", characterID)
	}

	// 獲取總數
	totalCount, err := query.Count(context.Background())
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to count chat sessions")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "無法查詢會話數量",
			},
		})
		return
	}

	// 分頁查詢
	var chatsDB []*db.ChatDB
	err = query.
		Order("updated_at DESC").
		Limit(limit).
		Offset((page-1)*limit).
		Scan(context.Background(), &chatsDB)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to query chat sessions")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "無法查詢聊天會話列表",
			},
		})
		return
	}

	// 轉換為領域模型並生成響應格式
	chatResponses := make([]*models.ChatResponse, len(chatsDB))
	for i, chatDB := range chatsDB {
		chat := models.ChatFromDB(chatDB)
		chatResponses[i] = chat.ToResponse()
	}

	// 計算分頁信息
	totalPages := (totalCount + limit - 1) / limit

	response := &models.ChatListResponse{
		Chats: chatResponses,
		Pagination: models.PaginationResponse{
			Page:       page,
			PageSize:   limit,
			TotalPages: totalPages,
			TotalCount: int64(totalCount),
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取聊天會話列表成功",
		Data:    response,
	})
}

// SendMessage godoc
// @Summary      發送聊天消息
// @Description  發送新消息到指定的聊天會話
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "會話ID"
// @Param        message body models.SendMessageRequest true "消息內容"
// @Success      201 {object} models.APIResponse{data=models.SendMessageResponse} "發送成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chats/{chat_id}/messages [post]
func SendMessage(c *gin.Context) {

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	// 從URL路徑參數獲取chat_id
	chatID := c.Param("chat_id")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_PARAMETERS",
				Message: "缺少會話ID參數",
			},
		})
		return
	}

	var req models.SendMessageRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 驗證會話是否存在且屬於當前用戶
	var chatDB db.ChatDB
	err := GetDB().NewSelect().
		Model(&chatDB).
		Where("id = ? AND user_id = ? AND status = ?", chatID, userID, "active").
		Scan(context.Background())

	var chat models.Chat
	if err == nil {
		// 轉換 DB 模型到 domain 模型
		chat = models.Chat{
			ID:              chatDB.ID,
			UserID:          chatDB.UserID,
			CharacterID:     chatDB.CharacterID,
			Title:           chatDB.Title,
			Status:          chatDB.Status,
			ChatMode:        chatDB.ChatMode,
			MessageCount:    chatDB.MessageCount,
			TotalCharacters: chatDB.TotalCharacters,
			LastMessageAt:   chatDB.LastMessageAt,
			CreatedAt:       chatDB.CreatedAt,
			UpdatedAt:       chatDB.UpdatedAt,
		}
	}

	if err != nil {
		utils.Logger.WithError(err).WithField("chat_id", chatID).Error("Session not found")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "聊天會話不存在或已結束",
			},
		})
		return
	}

	// 整合女性向AI聊天服務
	chatService := services.GetChatService()

	// 構建處理請求
	processRequest := &services.ProcessMessageRequest{
		ChatID:      chatID,
		UserMessage: req.Message,
		CharacterID: chat.CharacterID, // 從會話獲取角色ID
		UserID:      userID.(string),
		ChatMode:    chat.ChatMode, // 從會話獲取聊天模式
		Metadata:    map[string]interface{}{},
	}

	// 處理女性向AI對話
	chatResponse, err := chatService.ProcessMessage(context.Background(), processRequest)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"user_id":      userID,
			"character_id": chat.CharacterID,
			"chat_id":      chatID,
		}).WithError(err).Error("女性向AI對話處理失敗")

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "AI_GENERATION_FAILED",
				Message: "AI 回應生成失敗",
			},
		})
		return
	}

	// 確保回應有效性
	if chatResponse == nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "AI_GENERATION_FAILED",
				Message: "AI 回應生成失敗",
			},
		})
		return
	}

	// 確保好感度存在
	if chatResponse.Affection == 0 {
		chatResponse.Affection = 50
	}

	// ChatService 已經處理了 AI 消息插入和會話統計更新
	// 這裡不需要重複操作

	// 構建 SendMessage API 專用響應格式
	response := &models.SendMessageResponse{
		ChatID:         chatResponse.ChatID,
		MessageID:      chatResponse.MessageID,
		Content:        chatResponse.Content,
		Affection:      chatResponse.Affection,
		AIEngine:       chatResponse.AIEngine,
		NSFWLevel:      chatResponse.NSFWLevel,
		Confidence:     chatResponse.Confidence,
		ChatMode:       chat.ChatMode,
		ResponseTimeMs: chatResponse.ResponseTime.Milliseconds(),
		Timestamp:      time.Now(),
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "對話回應生成成功",
		Data:    response,
	})
}

// GetMessageHistory godoc
// @Summary      獲取聊天歷史
// @Description  獲取聊天會話的消息歷史
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "會話ID"
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(50)
// @Success      200 {object} models.APIResponse{data=models.MessageHistoryResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chats/{chat_id}/history [get]
func GetMessageHistory(c *gin.Context) {

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	sessionID := c.Param("chat_id")

	// 驗證會話是否存在且屬於當前用戶
	var chatDB db.ChatDB
	err := GetDB().NewSelect().
		Model(&chatDB).
		Where("id = ? AND user_id = ? AND status != ?", sessionID, userID, "deleted").
		Scan(context.Background())

	if err != nil {
		utils.Logger.WithError(err).WithField("chat_id", sessionID).Error("Session not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "聊天會話不存在",
			},
		})
		return
	}

	// 轉換為領域模型
	chat := models.ChatFromDB(&chatDB)

	// 解析查詢參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 查詢消息歷史 (最新消息在前，符合用戶期望)
	var messagesDB []*db.MessageDB
	err = GetDB().NewSelect().
		Model(&messagesDB).
		Where("chat_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(context.Background())

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to query message history")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "無法查詢聊天歷史",
			},
		})
		return
	}

	// 轉換為領域模型並生成響應格式
	messageResponses := make([]*models.DetailedMessageResponse, len(messagesDB))
	for i, messageDB := range messagesDB {
		message := models.MessageFromDB(messageDB)
		messageResponses[i] = message.ToResponse()
	}

	// 獲取總消息數
	totalCount := chat.MessageCount

	// 計算分頁信息
	totalPages := (totalCount + limit - 1) / limit

	response := &models.MessageHistoryResponse{
		ChatID:   sessionID,
		Messages: messageResponses,
		Pagination: models.PaginationResponse{
			Page:       page,
			PageSize:   limit,
			TotalPages: totalPages,
			TotalCount: int64(totalCount),
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取聊天歷史成功",
		Data:    response,
	})
}

// DeleteChatSession godoc
// @Summary      刪除聊天會話
// @Description  軟刪除聊天會話
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "會話ID"
// @Success      200 {object} models.APIResponse "刪除成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chats/{chat_id} [delete]
func DeleteChatSession(c *gin.Context) {

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	sessionID := c.Param("chat_id")

	// 軟刪除會話
	result, err := GetDB().NewUpdate().
		Model((*db.ChatDB)(nil)).
		Set("status = ?", "deleted").
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND user_id = ? AND status != ?", sessionID, userID, "deleted").
		Exec(context.Background())

	if err != nil {
		utils.Logger.WithError(err).WithField("chat_id", sessionID).Error("Failed to delete chat session")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "刪除聊天會話失敗",
			},
		})
		return
	}

	// 檢查是否有行被更新
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "聊天會話不存在",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "聊天會話已刪除",
	})
}

// UpdateSessionMode godoc
// @Summary      切換會話模式
// @Description  切換聊天會話的對話模式，支援兩種模式：chat（簡潔對話）和 novel（小說敘述）。不同模式會影響 AI 的回應風格和提示詞策略。
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "會話ID"
// @Param        request body object{mode=string} true "模式設定。mode: 'chat' | 'novel'"
// @Success      200 {object} models.APIResponse{data=object{chat_id=string,current_mode=string,mode_description=string,previous_mode=string,updated_at=string,user_id=string}} "切換成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤或無效的模式"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chats/{chat_id}/mode [put]
func UpdateSessionMode(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	sessionID := c.Param("chat_id")

	var req struct {
		Mode string `json:"mode" binding:"required"`
	}

	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 支援聊天模式和小說模式
	validModes := []string{"chat", "novel"}
	isValid := false
	for _, validMode := range validModes {
		if req.Mode == validMode {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_MODE",
				Message: "無效的對話模式",
			},
		})
		return
	}

	// 獲取當前會話
	var chat db.ChatDB
	err := database.GetApp().DB().NewSelect().
		Model(&chat).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Scan(context.Background())

	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 更新會話模式
	previousMode := "chat"

	// 更新會話模式
	_, err = database.GetApp().DB().NewUpdate().
		Model(&chat).
		Set("chat_mode = ?", req.Mode).
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Exec(context.Background())

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to update chat mode")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "更新對話模式失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "對話模式已切換",
		Data: gin.H{
			"chat_id":       sessionID,
			"user_id":       userID,
			"previous_mode": previousMode,
			"current_mode":  req.Mode,
			"updated_at":    time.Now(),
			"mode_description": map[string]string{
				"chat":  "簡潔對話模式 - 1-2句話的日常聊天風格",
				"novel": "小說敘述模式 - 包含細緻描寫的文學風格",
			}[req.Mode],
		},
	})
}

// ExportChatSession godoc
// @Summary      匯出會話記錄
// @Description  匯出聊天會話的完整記錄
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "會話ID"
// @Param        format query string false "匯出格式" Enums(json,txt,pdf) default(json)
// @Success      200 {object} models.APIResponse "匯出成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chats/{chat_id}/export [get]
func ExportChatSession(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	sessionID := c.Param("chat_id")
	format := c.DefaultQuery("format", "json")

	// 驗證格式
	if format != "json" && format != "txt" && format != "pdf" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_FORMAT",
				Message: "不支援的匯出格式",
			},
		})
		return
	}

	// 查詢會話資訊
	var chatDB db.ChatDB
	err := database.GetApp().DB().NewSelect().
		Model(&chatDB).
		Relation("Character").
		Where("chat_db.id = ? AND chat_db.user_id = ?", sessionID, userID).
		Scan(context.Background(), &chatDB)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 查詢會話消息
	var messagesDB []db.MessageDB
	err = database.GetApp().DB().NewSelect().
		Model(&messagesDB).
		Where("chat_id = ?", sessionID).
		Order("created_at ASC").
		Scan(context.Background(), &messagesDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "QUERY_ERROR",
				Message: "查詢消息失敗",
			},
		})
		return
	}

	// 轉換為domain模型
	chat := models.ChatFromDB(&chatDB)

	var messages []*models.Message
	for _, msgDB := range messagesDB {
		messages = append(messages, models.MessageFromDB(&msgDB))
	}

	// 計算會話統計
	messageCount := len(messages)
	var duration time.Duration
	if messageCount > 0 {
		duration = messages[messageCount-1].CreatedAt.Sub(messages[0].CreatedAt)
	}

	// 構建匯出數據
	characterName := "未知角色"
	if chat.Character != nil {
		characterName = chat.Character.Name
	}

	exportData := gin.H{
		"chat_id":       sessionID,
		"user_id":       userID,
		"export_format": format,
		"generated_at":  time.Now(),
		"file_info": gin.H{
			"filename": fmt.Sprintf("chat_session_%s.%s", sessionID, format),
			"size":     fmt.Sprintf("%.1fKB", float64(len(fmt.Sprintf("%v", messages))*2)/1024),
		},
		"session_summary": gin.H{
			"title":         fmt.Sprintf("與%s的對話", characterName),
			"message_count": messageCount,
			"duration":      formatDuration(duration),
			"characters":    []string{characterName},
			"created_at":    chat.CreatedAt,
		},
		"messages": func() interface{} {
			if format == "json" {
				return messages
			}
			// 為txt格式準備純文本
			var textMessages []string
			for _, msg := range messages {
				timestamp := msg.CreatedAt.Format("2006-01-02 15:04:05")
				role := "用戶"
				if msg.Role == "assistant" {
					role = characterName
				}
				textMessages = append(textMessages, fmt.Sprintf("[%s] %s: %s", timestamp, role, msg.Dialogue))
			}
			return textMessages
		}(),
		"export_id": utils.GenerateUUID(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "會話匯出成功",
		Data:    exportData,
	})
}

// formatDuration 格式化時間長度為中文
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0分鐘"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d小時%d分鐘", hours, minutes)
	}
	return fmt.Sprintf("%d分鐘", minutes)
}

// RegenerateResponse godoc
// @Summary      重新生成回應
// @Description  重新生成指定消息的 AI 回應
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "會話ID"
// @Param        message_id path string true "消息ID"
// @Success      200 {object} models.APIResponse{data=models.RegenerateMessageResponse} "生成成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "消息不存在"
// @Router       /chats/{chat_id}/messages/{message_id}/regenerate [post]
func RegenerateResponse(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	// 從URL路徑參數獲取ID
	chatID := c.Param("chat_id")
	messageID := c.Param("message_id")

	if chatID == "" || messageID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_PARAMETERS",
				Message: "缺少必要的路徑參數",
			},
		})
		return
	}

	// 獲取原始消息
	var originalMessage db.MessageDB
	err := database.GetApp().DB().NewSelect().
		Model(&originalMessage).
		Where("id = ? AND chat_id = ?", messageID, chatID).
		Scan(context.Background())

	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MESSAGE_NOT_FOUND",
				Message: "消息不存在",
			},
		})
		return
	}

	// 獲取會話信息
	var chat db.ChatDB
	err = database.GetApp().DB().NewSelect().
		Model(&chat).
		Where("id = ? AND user_id = ?", chatID, userID).
		Scan(context.Background())

	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 獲取會話歷史記錄（用於上下文）
	var messages []db.MessageDB
	err = database.GetApp().DB().NewSelect().
		Model(&messages).
		Where("chat_id = ? AND created_at <= ?", chatID, originalMessage.CreatedAt).
		Order("created_at ASC").
		Limit(10). // 最近10條消息作為上下文
		Scan(context.Background())

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to retrieve message history")
	}

	// 找到對應的用戶消息
	var previousUserMsg db.MessageDB
	err = database.GetApp().DB().NewSelect().
		Model(&previousUserMsg).
		Where("chat_id = ? AND role = ? AND created_at < ?", chatID, "user", originalMessage.CreatedAt).
		Order("created_at DESC").
		Limit(1).
		Scan(context.Background())

	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_MESSAGE_NOT_FOUND",
				Message: "找不到對應的用戶消息",
			},
		})
		return
	}

	// 重新生成回應
	chatService := services.GetChatService()
	processReq := &services.ProcessMessageRequest{
		ChatID:      chatID,
		UserMessage: previousUserMsg.Dialogue,
		CharacterID: chat.CharacterID,
		UserID:      userID.(string),
	}

	response, err := chatService.ProcessMessage(context.Background(), processReq)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to regenerate message")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "REGENERATION_FAILED",
				Message: "重新生成失敗",
			},
		})
		return
	}

	// 標記原消息為已替換
	_, err = database.GetApp().DB().NewUpdate().
		Model(&originalMessage).
		Set("is_regenerated = true").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", messageID).
		Exec(context.Background())

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to mark original message as regenerated")
	}

	// 構建詳細 MessageResponse 用於 RegenerateMessageResponse
	now := time.Now()
	messageResponse := &models.DetailedMessageResponse{
		MessageResponse: models.MessageResponse{
			ID:             response.MessageID,
			ChatID:         response.ChatID,
			Role:           "assistant",
			Dialogue:       response.Content,
			AIEngine:       response.AIEngine,
			ResponseTimeMs: int(response.ResponseTime.Milliseconds()),
			NSFWLevel:      response.NSFWLevel,
			CreatedAt:      now,
		},
		IsRegenerated: true, // 標記為重新生成
	}

	// 使用專用的 RegenerateMessageResponse 結構
	regenerateResponse := &models.RegenerateMessageResponse{
		Message:           messageResponse,
		PreviousMessageID: messageID,
		Regenerated:       true,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "回應重新生成成功",
		Data:    regenerateResponse,
	})
}
