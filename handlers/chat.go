package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

var (
	chatMapper = models.NewChatMapper()
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
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 驗證角色是否存在
	characterExists, err := GetDB().NewSelect().
		Model((*db.CharacterDB)(nil)).
		Where("id = ? AND is_active = ?", req.CharacterID, true).
		Exists(ctx)
	
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
	
	// 為了後續使用，創建簡單的 character 物件
	// TODO: 應該使用 character service 或 mapper 來正確轉換
	var character models.Character
	character.ID = req.CharacterID
	character.Name = "角色" // 簡化版本，後續應該正確查詢

	// 檢查是否已存在用戶-角色會話（一對一架構）
	sessionExists, err := GetDB().NewSelect().
		Model((*db.ChatSessionDB)(nil)).
		Where("user_id = ? AND character_id = ? AND status != ?", userID.(string), req.CharacterID, "deleted").
		Exists(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to check existing session")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "檢查會話失敗",
			},
		})
		return
	}

	var session models.ChatSession
	if sessionExists {
		// 如果已存在會話，返回現有會話
		var sessionDB db.ChatSessionDB
		err = GetDB().NewSelect().
			Model(&sessionDB).
			Where("user_id = ? AND character_id = ? AND status != ?", userID.(string), req.CharacterID, "deleted").
			Scan(ctx)
		
		if err == nil {
			// 轉換 DB 模型到 domain 模型
			session = models.ChatSession{
				ID:              sessionDB.ID,
				UserID:          sessionDB.UserID,
				CharacterID:     sessionDB.CharacterID,
				Title:           sessionDB.Title,
				Status:          sessionDB.Status,
				MessageCount:    sessionDB.MessageCount,
				TotalCharacters: sessionDB.TotalCharacters,
				LastMessageAt:   sessionDB.LastMessageAt,
				CreatedAt:       sessionDB.CreatedAt,
				UpdatedAt:       sessionDB.UpdatedAt,
			}
		}

		if err != nil {
			utils.Logger.WithError(err).Error("Failed to get existing session")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "獲取現有會話失敗",
				},
			})
			return
		}

		// 如果提供了新標題，更新會話標題
		if req.Title != "" && req.Title != session.Title {
			session.Title = req.Title
			session.UpdatedAt = time.Now()
			_, err = GetDB().NewUpdate().
				Model(&session).
				Column("title", "updated_at").
				Where("id = ?", session.ID).
				Exec(ctx)
			if err != nil {
				utils.Logger.WithError(err).Error("Failed to update session title")
			}
		}

		// 關聯角色信息
		session.Character = &character

		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "獲取現有聊天會話成功",
			Data:    session.ToResponse(),
		})
		return
	}

	// 檢查是否有已刪除的會話可以重新啟用
	deletedExists, err := GetDB().NewSelect().
		Model((*db.ChatSessionDB)(nil)).
		Where("user_id = ? AND character_id = ? AND status = ?", userID.(string), req.CharacterID, "deleted").
		Exists(ctx)
	
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to check deleted session")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "檢查已刪除會話失敗",
			},
		})
		return
	}

	var deletedSession models.ChatSession
	if deletedExists {
		// 重新啟用已刪除的會話
		var deletedSessionDB db.ChatSessionDB
		err = GetDB().NewSelect().
			Model(&deletedSessionDB).
			Where("user_id = ? AND character_id = ? AND status = ?", userID.(string), req.CharacterID, "deleted").
			Scan(ctx)
		
		if err == nil {
			// 轉換 DB 模型到 domain 模型
			deletedSession = models.ChatSession{
				ID:              deletedSessionDB.ID,
				UserID:          deletedSessionDB.UserID,
				CharacterID:     deletedSessionDB.CharacterID,
				Title:           deletedSessionDB.Title,
				Status:          deletedSessionDB.Status,
				MessageCount:    deletedSessionDB.MessageCount,
				TotalCharacters: deletedSessionDB.TotalCharacters,
				LastMessageAt:   deletedSessionDB.LastMessageAt,
				CreatedAt:       deletedSessionDB.CreatedAt,
				UpdatedAt:       deletedSessionDB.UpdatedAt,
			}
		}

		if err != nil {
			utils.Logger.WithError(err).Error("Failed to get deleted session")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "獲取已刪除會話失敗",
				},
			})
			return
		}

		// 更新會話資訊並重新啟用
		deletedSession.Status = "active"
		deletedSession.UpdatedAt = time.Now()
		if req.Title != "" {
			deletedSession.Title = req.Title
		}

		_, err = GetDB().NewUpdate().
			Model(&deletedSession).
			Column("status", "title", "updated_at").
			Where("id = ?", deletedSession.ID).
			Exec(ctx)

		if err != nil {
			utils.Logger.WithError(err).Error("Failed to reactivate deleted session")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "重新啟用會話失敗",
				},
			})
			return
		}

		// 關聯角色信息
		deletedSession.Character = &character

		c.JSON(http.StatusCreated, models.APIResponse{
			Success: true,
			Message: "聊天會話重新啟用成功",
			Data:    deletedSession.ToResponse(),
		})
		return
	}

	// 如果不存在會話，創建新的聊天會話
	sessionID := utils.GenerateSessionID()
	sessionTitle := req.Title
	if sessionTitle == "" {
		sessionTitle = "與 " + character.Name + " 的對話"
	}

	// 創建 DB 模型並插入
	sessionDB := db.ChatSessionDB{
		ID:          sessionID,
		UserID:      userID.(string),
		CharacterID: req.CharacterID,
		Title:       sessionTitle,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 插入數據庫
	_, err = GetDB().NewInsert().Model(&sessionDB).Exec(ctx)
	
	if err == nil {
		// 轉換到 domain 模型
		session = models.ChatSession{
			ID:              sessionDB.ID,
			UserID:          sessionDB.UserID,
			CharacterID:     sessionDB.CharacterID,
			Title:           sessionDB.Title,
			Status:          sessionDB.Status,
			MessageCount:    sessionDB.MessageCount,
			TotalCharacters: sessionDB.TotalCharacters,
			LastMessageAt:   sessionDB.LastMessageAt,
			CreatedAt:       sessionDB.CreatedAt,
			UpdatedAt:       sessionDB.UpdatedAt,
		}
	}
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

	var sessionDB db.ChatSessionDB
	err := GetDB().NewSelect().
		Model(&sessionDB).
		Where("id = ? AND user_id = ? AND status != ?", sessionID, userID, "deleted").
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

	// 轉換為領域模型
	session := chatMapper.ChatSessionFromDB(&sessionDB)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取聊天會話成功",
		Data:    session.ToResponse(),
	})
}

// GetChatSessions godoc
// @Summary      獲取用戶聊天會話列表
// @Description  獲取用戶的聊天會話列表，支援分頁和角色篩選
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Param        status query string false "會話狀態篩選"
// @Param        character_id query string false "角色ID篩選"
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
	query := GetDB().NewSelect().
		Model((*db.ChatSessionDB)(nil)).
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
	var sessionsDB []*db.ChatSessionDB
	err = query.
		Order("updated_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(ctx, &sessionsDB)

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
	sessionResponses := make([]*models.ChatSessionResponse, len(sessionsDB))
	for i, sessionDB := range sessionsDB {
		session := chatMapper.ChatSessionFromDB(sessionDB)
		sessionResponses[i] = session.ToResponse()
	}

	// 計算分頁信息
	totalPages := (totalCount + limit - 1) / limit

	response := &models.ChatSessionListResponse{
		Sessions: sessionResponses,
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
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 驗證會話是否存在且屬於當前用戶
	var sessionDB db.ChatSessionDB
	err := GetDB().NewSelect().
		Model(&sessionDB).
		Where("id = ? AND user_id = ? AND status = ?", req.SessionID, userID, "active").
		Scan(ctx)
	
	var session models.ChatSession
	if err == nil {
		// 轉換 DB 模型到 domain 模型
		session = models.ChatSession{
			ID:              sessionDB.ID,
			UserID:          sessionDB.UserID,
			CharacterID:     sessionDB.CharacterID,
			Title:           sessionDB.Title,
			Status:          sessionDB.Status,
			MessageCount:    sessionDB.MessageCount,
			TotalCharacters: sessionDB.TotalCharacters,
			LastMessageAt:   sessionDB.LastMessageAt,
			CreatedAt:       sessionDB.CreatedAt,
			UpdatedAt:       sessionDB.UpdatedAt,
		}
	}

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

    // 整合女性向AI聊天服務
    chatService := services.NewChatService()
	
	// 構建處理請求
	processRequest := &services.ProcessMessageRequest{
		SessionID:   req.SessionID,
		UserMessage: req.Message,
		CharacterID: session.CharacterID, // 從會話獲取角色ID
		UserID:      userID.(string),
		Metadata:    map[string]interface{}{},
	}
	
    // 處理女性向AI對話
    chatResponse, err := chatService.ProcessMessage(ctx, processRequest)
    if err != nil {
        utils.Logger.WithError(err).Error("女性向AI對話處理失敗")
    }
    // 構建保底回應（避免 nil 導致 500）
    if chatResponse == nil {
        chatResponse = &services.ChatResponse{
            SessionID:         req.SessionID,
            MessageID:         utils.GenerateMessageID(),
            CharacterDialogue: "抱歉，我現在有些困惑...能再說一遍嗎？",
            SceneDescription:  "房間裡的氣氛有些緊張",
            CharacterAction:   "他皺了皺眉，似乎在思考什麼",
            EmotionState: &services.EmotionState{
                Affection:     50,
                Mood:          "concerned",
                Relationship:  "friend",
                IntimacyLevel: "friendly",
            },
            AIEngine:     "fallback",
            NSFWLevel:    1,
            ResponseTime: 0,
        }
    } else if chatResponse.EmotionState == nil {
        // 確保情感狀態存在
        chatResponse.EmotionState = &services.EmotionState{
            Affection:     50,
            Mood:          "neutral",
            Relationship:  "friend",
            IntimacyLevel: "friendly",
        }
    }
	
	// ChatService 已經處理了 AI 消息插入和會話統計更新
	// 這裡不需要重複操作

	// 構建完整的女性向聊天回應
    // 安全構建情感狀態
    emotion := map[string]interface{}{}
    if chatResponse.EmotionState != nil {
        emotion = map[string]interface{}{
            "affection":      chatResponse.EmotionState.Affection,
            "mood":           chatResponse.EmotionState.Mood,
            "relationship":   chatResponse.EmotionState.Relationship,
            "intimacy_level": chatResponse.EmotionState.IntimacyLevel,
        }
    }

    response := map[string]interface{}{
        "session_id":         chatResponse.SessionID,
        "message_id":         chatResponse.MessageID,
        "content":            chatResponse.CharacterDialogue,
        "scene_description":  chatResponse.SceneDescription,
        "character_action":   chatResponse.CharacterAction,
        "emotion_state":      emotion,
        "ai_engine":          chatResponse.AIEngine,
        "nsfw_level":         chatResponse.NSFWLevel,
        "response_time":      chatResponse.ResponseTime.Milliseconds(),
        "special_event":      chatResponse.SpecialEvent,
        "timestamp":          time.Now(),
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
	var sessionDB db.ChatSessionDB
	err := GetDB().NewSelect().
		Model(&sessionDB).
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

	// 轉換為領域模型
	session := chatMapper.ChatSessionFromDB(&sessionDB)

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
	var messagesDB []*db.MessageDB
	err = GetDB().NewSelect().
		Model(&messagesDB).
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

	// 轉換為領域模型並生成響應格式
	messageResponses := make([]*models.MessageResponse, len(messagesDB))
	for i, messageDB := range messagesDB {
		message := chatMapper.MessageFromDB(messageDB)
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
	result, err := GetDB().NewUpdate().
		Model((*db.ChatSessionDB)(nil)).
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

	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
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

	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
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
				"tag_id":     utils.GenerateTagID(),
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

	// 查詢會話資訊
	var sessionDB db.ChatSessionDB
	err := database.GetApp().DB().NewSelect().
		Model(&sessionDB).
		Relation("Character").
		Where("cs.id = ? AND cs.user_id = ?", sessionID, userID).
		Scan(context.Background(), &sessionDB)
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
		Where("session_id = ?", sessionID).
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
	mapper := models.NewChatMapper()
	session := mapper.ChatSessionFromDB(&sessionDB)
	
	var messages []*models.Message
	for _, msgDB := range messagesDB {
		messages = append(messages, mapper.MessageFromDB(&msgDB))
	}

	// 計算會話統計
	messageCount := len(messages)
	var duration time.Duration
	if messageCount > 0 {
		duration = messages[messageCount-1].CreatedAt.Sub(messages[0].CreatedAt)
	}

	// 構建匯出數據
	characterName := "未知角色"
	if session.Character != nil {
		characterName = session.Character.Name
	}

	exportData := gin.H{
		"session_id":    sessionID,
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
			"created_at":    session.CreatedAt,
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
				textMessages = append(textMessages, fmt.Sprintf("[%s] %s: %s", timestamp, role, msg.Content))
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

	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 靜態回應 - 模擬重新生成
	newMessage := gin.H{
		"message_id":       utils.GenerateUUID(),
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

