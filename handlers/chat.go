package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// CreateChatSession godoc
// @Summary      創建聊天會話
// @Description  創建新的聊天會話
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session body models.CreateSessionRequest true "會話信息"
// @Success      201 {object} models.APIResponse{data=models.ChatSessionResponse} "創建成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chat/session [post]
func CreateChatSession(c *gin.Context) {
	ctx := context.Background()

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

	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "輸入參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 驗證角色是否存在
	var character models.Character
	exists, err := database.DB.NewSelect().
		Model(&character).
		Where("id = ? AND is_active = ?", req.CharacterID, true).
		Exists(ctx)

	if err != nil || !exists {
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

	// 創建聊天會話
	session := &models.ChatSession{
		ID:          utils.GenerateID(16),
		UserID:      userID.(string),
		CharacterID: req.CharacterID,
		Title:       req.Title,
		Mode:        req.Mode,
		Status:      "active",
		Tags:        req.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 設置默認模式
	if session.Mode == "" {
		session.Mode = "normal"
	}

	// 設置默認標題
	if session.Title == "" {
		session.Title = "與 " + character.Name + " 的對話"
	}

	// 插入數據庫
	_, err = database.DB.NewInsert().Model(session).Exec(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to create chat session")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "創建聊天會話失敗",
			},
		})
		return
	}

	// 關聯角色信息
	session.Character = &character

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "聊天會話創建成功",
		Data:    session.ToResponse(),
	})
}

// GetChatSession godoc
// @Summary      獲取聊天會話詳情
// @Description  獲取特定聊天會話的詳細信息
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話ID"
// @Success      200 {object} models.APIResponse{data=models.ChatSessionResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id} [get]
func GetChatSession(c *gin.Context) {
	ctx := context.Background()

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

	sessionID := c.Param("session_id")

	var session models.ChatSession
	err := database.DB.NewSelect().
		Model(&session).
		Relation("Character").
		Where("cs.id = ? AND cs.user_id = ? AND cs.status != ?", sessionID, userID, "deleted").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("session_id", sessionID).Error("Failed to query chat session")
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
		Message: "獲取聊天會話成功",
		Data:    session.ToResponse(),
	})
}

// GetChatSessions godoc
// @Summary      獲取用戶聊天會話列表
// @Description  獲取用戶的聊天會話列表，支援分頁
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Param        status query string false "會話狀態篩選"
// @Success      200 {object} models.APIResponse{data=models.ChatSessionListResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chat/sessions [get]
func GetChatSessions(c *gin.Context) {
	ctx := context.Background()

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
	query := database.DB.NewSelect().
		Model((*models.ChatSession)(nil)).
		Relation("Character").
		Where("cs.user_id = ? AND cs.status != ?", userID, "deleted")

	// 應用狀態篩選
	if status := c.Query("status"); status != "" {
		query = query.Where("cs.status = ?", status)
	}

	// 獲取總數
	totalCount, err := query.Count(ctx)
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
	var sessions []*models.ChatSession
	err = query.
		Order("cs.updated_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(ctx, &sessions)

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

	// 轉換為響應格式
	sessionResponses := make([]*models.ChatSessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = session.ToResponse()
	}

	// 計算分頁信息
	totalPages := (totalCount + limit - 1) / limit

	response := &models.ChatSessionListResponse{
		Sessions: sessionResponses,
		Pagination: models.PaginationResponse{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  totalCount,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
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
// @Description  發送新消息到聊天會話
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        message body models.SendMessageRequest true "消息內容"
// @Success      201 {object} models.APIResponse{data=models.MessageResponse} "發送成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chat/message [post]
func SendMessage(c *gin.Context) {
	ctx := context.Background()

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

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "輸入參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 驗證會話是否存在且屬於當前用戶
	var session models.ChatSession
	err := database.DB.NewSelect().
		Model(&session).
		Where("id = ? AND user_id = ? AND status = ?", req.SessionID, userID, "active").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("session_id", req.SessionID).Error("Session not found")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "聊天會話不存在或已結束",
			},
		})
		return
	}

	// 創建用戶消息
	userMessage := &models.Message{
		ID:        utils.GenerateID(16),
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Message,
		CreatedAt: time.Now(),
	}

	// 插入用戶消息
	_, err = database.DB.NewInsert().Model(userMessage).Exec(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to create user message")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "保存消息失敗",
			},
		})
		return
	}

	// 這裡可以添加 AI 回應生成邏輯
	// 目前先創建一個簡單的自動回應
	aiMessage := &models.Message{
		ID:        utils.GenerateID(16),
		SessionID: req.SessionID,
		Role:      "assistant",
		Content:   "這是一個自動回應。AI 集成功能正在開發中...",
		AIEngine:  "placeholder",
		CreatedAt: time.Now(),
	}

	// 插入 AI 消息
	_, err = database.DB.NewInsert().Model(aiMessage).Exec(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to create AI message")
		// 不中斷流程，用戶消息已成功保存
	}

	// 更新會話統計
	now := time.Now()
	_, err = database.DB.NewUpdate().
		Model((*models.ChatSession)(nil)).
		Set("message_count = message_count + ?", 2). // 用戶消息 + AI 回應
		Set("last_message_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", req.SessionID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to update session stats")
		// 不中斷流程
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "消息發送成功",
		Data:    aiMessage.ToResponse(),
	})
}

// GetMessageHistory godoc
// @Summary      獲取聊天歷史
// @Description  獲取聊天會話的消息歷史
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話ID"
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(50)
// @Success      200 {object} models.APIResponse{data=models.MessageHistoryResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id}/history [get]
func GetMessageHistory(c *gin.Context) {
	ctx := context.Background()

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

	sessionID := c.Param("session_id")

	// 驗證會話是否存在且屬於當前用戶
	var session models.ChatSession
	err := database.DB.NewSelect().
		Model(&session).
		Where("id = ? AND user_id = ? AND status != ?", sessionID, userID, "deleted").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("session_id", sessionID).Error("Session not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "聊天會話不存在",
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

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 查詢消息歷史
	var messages []*models.Message
	err = database.DB.NewSelect().
		Model(&messages).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(ctx)

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

	// 轉換為響應格式
	messageResponses := make([]*models.MessageResponse, len(messages))
	for i, message := range messages {
		messageResponses[i] = message.ToResponse()
	}

	// 獲取總消息數
	totalCount := session.MessageCount

	// 計算分頁信息
	totalPages := (totalCount + limit - 1) / limit

	response := &models.MessageHistoryResponse{
		SessionID: sessionID,
		Messages:  messageResponses,
		Pagination: models.PaginationResponse{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  totalCount,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
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
// @Param        session_id path string true "會話ID"
// @Success      200 {object} models.APIResponse "刪除成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id} [delete]
func DeleteChatSession(c *gin.Context) {
	ctx := context.Background()

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

	sessionID := c.Param("session_id")

	// 軟刪除會話
	result, err := database.DB.NewUpdate().
		Model((*models.ChatSession)(nil)).
		Set("status = ?", "deleted").
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND user_id = ? AND status != ?", sessionID, userID, "deleted").
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("session_id", sessionID).Error("Failed to delete chat session")
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
// @Description  切換聊天會話的對話模式
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話ID"
// @Param        mode body object true "模式設定"
// @Success      200 {object} models.APIResponse "切換成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chat/session/{session_id}/mode [put]
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

	sessionID := c.Param("session_id")

	var req struct {
		Mode string `json:"mode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "請求參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 驗證模式
	validModes := []string{"normal", "romantic", "adventure", "roleplay", "novel"}
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

	// 靜態回應 - 模擬模式切換
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "對話模式已切換",
		Data: gin.H{
			"session_id":   sessionID,
			"user_id":      userID,
			"previous_mode": "normal",
			"current_mode":  req.Mode,
			"updated_at":   time.Now(),
			"mode_description": map[string]string{
				"normal":   "日常對話模式",
				"romantic": "浪漫互動模式", 
				"adventure": "冒險探索模式",
				"roleplay": "角色扮演模式",
				"novel":    "小說敘述模式",
			}[req.Mode],
		},
	})
}

// AddSessionTag godoc
// @Summary      為會話添加標籤
// @Description  為聊天會話添加分類標籤
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話ID"
// @Param        tag body object true "標籤信息"
// @Success      200 {object} models.APIResponse "添加成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chat/session/{session_id}/tag [post]
func AddSessionTag(c *gin.Context) {
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

	sessionID := c.Param("session_id")

	var req struct {
		Tag   string `json:"tag" binding:"required"`
		Color string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "請求參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 靜態回應 - 模擬標籤添加
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "標籤添加成功",
		Data: gin.H{
			"session_id": sessionID,
			"user_id":    userID,
			"tag": gin.H{
				"name":       req.Tag,
				"color":      req.Color,
				"created_at": time.Now(),
				"tag_id":     utils.GenerateID(8),
			},
			"current_tags": []gin.H{
				{"name": "浪漫", "color": "#ff69b4"},
				{"name": "日常", "color": "#87ceeb"},
				{"name": req.Tag, "color": req.Color},
			},
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
// @Param        session_id path string true "會話ID"
// @Param        format query string false "匯出格式" Enums(json,txt,pdf) default(json)
// @Success      200 {object} models.APIResponse "匯出成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id}/export [get]
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

	sessionID := c.Param("session_id")
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

	// 靜態回應 - 模擬匯出
	exportData := gin.H{
		"session_id":   sessionID,
		"user_id":      userID,
		"export_format": format,
		"generated_at": time.Now(),
		"file_info": gin.H{
			"filename": "chat_session_" + sessionID + "." + format,
			"size":     "2.3MB",
			"url":      "https://example.com/exports/chat_session_" + sessionID + "." + format,
		},
		"session_summary": gin.H{
			"title":         "與陸燁銘的對話",
			"message_count": 45,
			"duration":      "2小時35分鐘",
			"characters":    []string{"陸燁銘"},
			"tags":          []string{"浪漫", "日常"},
		},
		"export_id": utils.GenerateID(16),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "會話匯出成功",
		Data:    exportData,
	})
}

// RegenerateResponse godoc
// @Summary      重新生成回應
// @Description  重新生成最後一個 AI 回應
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "重新生成請求"
// @Success      200 {object} models.APIResponse "生成成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chat/regenerate [post]
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

	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		MessageID string `json:"message_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "請求參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 靜態回應 - 模擬重新生成
	newMessage := gin.H{
		"message_id":       utils.GenerateID(16),
		"session_id":       req.SessionID,
		"user_id":          userID,
		"role":             "assistant",
		"content":          "讓我重新組織一下思緒...其實我想說的是，和你在一起的時光總是過得特別快，就像時間也捨不得打擾我們的對話一樣。",
		"character_name":   "陸燁銘",
		"emotion":          "溫柔",
		"scene_description": "夕陽西下，辦公室裡只剩下溫暖的燈光",
		"created_at":       time.Now(),
		"regenerated":      true,
		"previous_message_id": req.MessageID,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "回應重新生成成功",
		Data:    newMessage,
	})
}