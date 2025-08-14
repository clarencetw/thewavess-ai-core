package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// CreateChatSession godoc
// @Summary      創建新對話會話
// @Description  創建新的對話會話
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateSessionRequest true "會話創建參數"
// @Success      201 {object} models.APIResponse{data=models.ChatSession} "創建成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chat/session [post]
func CreateChatSession(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "請求參數驗證失敗",
				Details: err.Error(),
			},
		})
		return
	}

	// 驗證角色ID是否有效
	validCharacters := []string{"char_001", "char_002"}
	if !utils.StringInSlice(req.CharacterID, validCharacters) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CHARACTER",
				Message: "無效的角色ID",
			},
		})
		return
	}

	// 驗證模式是否有效
	validModes := []string{"normal", "novel", "nsfw"}
	if req.Mode != "" && !utils.StringInSlice(req.Mode, validModes) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_MODE",
				Message: "無效的對話模式，支援: normal, novel, nsfw",
			},
		})
		return
	}

	// 設置默認值
	if req.Mode == "" {
		req.Mode = "normal"
	}
	if req.Title == "" {
		characterName := "陸寒淵"
		if req.CharacterID == "char_002" {
			characterName = "沈言墨"
		}
		req.Title = "與" + characterName + "的對話"
	}

	// 模擬獲取用戶ID
	userID := "user_alice123_001"
	if authHeader == "Bearer demo_token" {
		userID = "user_demo_001"
	}

	// 創建新會話
	sessionID := "session_" + utils.IntToString(int(time.Now().Unix()))
	session := models.ChatSession{
		BaseModel: models.BaseModel{
			ID:        sessionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:        userID,
		CharacterID:   req.CharacterID,
		Title:         req.Title,
		Mode:          req.Mode,
		Status:        "active",
		Tags:          req.Tags,
		MessageCount:  0,
		LastMessageAt: time.Time{}, // 零值，表示還沒有消息
	}

	utils.Logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"session_id":   sessionID,
		"character_id": req.CharacterID,
		"mode":         req.Mode,
	}).Info("Chat session created successfully")

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "對話會話創建成功",
		Data:    session,
	})
}

// GetChatSession godoc
// @Summary      獲取對話會話資訊
// @Description  獲取指定對話會話的詳細資訊
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話 ID"
// @Success      200 {object} models.APIResponse{data=models.ChatSession} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id} [get]
func GetChatSession(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_SESSION_ID",
				Message: "缺少會話 ID",
			},
		})
		return
	}

	// 模擬會話數據
	mockSessions := map[string]models.ChatSession{
		"test_session": {
			BaseModel: models.BaseModel{
				ID:        "test_session",
				CreatedAt: time.Now().AddDate(0, 0, -1),
				UpdatedAt: time.Now(),
			},
			UserID:        "user_alice123_001",
			CharacterID:   "char_001",
			Title:         "與陸寒淵的對話",
			Mode:          "normal",
			Status:        "active",
			Tags:          []string{"工作", "日常"},
			MessageCount:  5,
			LastMessageAt: time.Now().Add(-2 * time.Hour),
		},
		"session_001": {
			BaseModel: models.BaseModel{
				ID:        "session_001",
				CreatedAt: time.Now().AddDate(0, 0, -3),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			UserID:        "user_alice123_001",
			CharacterID:   "char_002",
			Title:         "與沈言墨的對話",
			Mode:          "normal",
			Status:        "active",
			Tags:          []string{"醫學", "溫柔"},
			MessageCount:  12,
			LastMessageAt: time.Now().Add(-1 * time.Hour),
		},
		"demo_session": {
			BaseModel: models.BaseModel{
				ID:        "demo_session",
				CreatedAt: time.Now().AddDate(0, 0, -7),
				UpdatedAt: time.Now().Add(-3 * time.Hour),
			},
			UserID:        "user_demo_001",
			CharacterID:   "char_001",
			Title:         "測試對話",
			Mode:          "nsfw",
			Status:        "paused",
			Tags:          []string{"測試", "NSFW"},
			MessageCount:  8,
			LastMessageAt: time.Now().Add(-3 * time.Hour),
		},
	}

	session, exists := mockSessions[sessionID]
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 驗證會話所有權 (在真實環境中會檢查JWT中的用戶ID)
	expectedUserID := "user_alice123_001"
	if authHeader == "Bearer demo_token" {
		expectedUserID = "user_demo_001"
	}

	if session.UserID != expectedUserID {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ACCESS_DENIED",
				Message: "無權限存取此會話",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取會話資訊成功",
		Data:    session,
	})
}

// GetChatSessions godoc
// @Summary      獲取用戶對話會話列表
// @Description  獲取當前用戶的所有對話會話，支援分頁
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Param        status query string false "會話狀態篩選" Enums(active,ended,paused)
// @Success      200 {object} models.APIResponse{data=models.SessionListResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /chat/sessions [get]
func GetChatSessions(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	// 解析查詢參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := utils.ParseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := utils.ParseInt(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	status := c.Query("status")

	// 模擬用戶ID
	userID := "user_alice123_001"
	if authHeader == "Bearer demo_token" {
		userID = "user_demo_001"
	}

	// 模擬會話列表數據
	allSessions := []models.ChatSession{
		{
			BaseModel: models.BaseModel{
				ID:        "session_1",
				CreatedAt: time.Now().AddDate(0, 0, -1),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			UserID:        userID,
			CharacterID:   "char_001",
			Title:         "與陸寒淵的深夜對話",
			Mode:          "normal",
			Status:        "active",
			Tags:          []string{"工作", "深夜"},
			MessageCount:  15,
			LastMessageAt: time.Now().Add(-1 * time.Hour),
		},
		{
			BaseModel: models.BaseModel{
				ID:        "session_2",
				CreatedAt: time.Now().AddDate(0, 0, -3),
				UpdatedAt: time.Now().Add(-3 * time.Hour),
			},
			UserID:        userID,
			CharacterID:   "char_002",
			Title:         "醫學諮詢",
			Mode:          "normal",
			Status:        "active",
			Tags:          []string{"健康", "醫學"},
			MessageCount:  8,
			LastMessageAt: time.Now().Add(-3 * time.Hour),
		},
		{
			BaseModel: models.BaseModel{
				ID:        "session_3",
				CreatedAt: time.Now().AddDate(0, 0, -7),
				UpdatedAt: time.Now().Add(-2 * 24 * time.Hour),
			},
			UserID:        userID,
			CharacterID:   "char_001",
			Title:         "親密時光",
			Mode:          "nsfw",
			Status:        "ended",
			Tags:          []string{"親密", "NSFW"},
			MessageCount:  25,
			LastMessageAt: time.Now().Add(-2 * 24 * time.Hour),
		},
	}

	// 應用狀態篩選
	var filteredSessions []models.ChatSession
	for _, session := range allSessions {
		if status != "" && session.Status != status {
			continue
		}
		filteredSessions = append(filteredSessions, session)
	}

	// 分頁處理
	totalCount := len(filteredSessions)
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	if endIndex > totalCount {
		endIndex = totalCount
	}

	var paginatedSessions []models.ChatSession
	if startIndex < totalCount {
		paginatedSessions = filteredSessions[startIndex:endIndex]
	}

	totalPages := (totalCount + limit - 1) / limit

	response := map[string]interface{}{
		"sessions": paginatedSessions,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  totalCount,
			"has_next":     page < totalPages,
			"has_prev":     page > 1,
		},
		"summary": map[string]interface{}{
			"active_count": countSessionsByStatus(filteredSessions, "active"),
			"total_messages": calculateTotalMessages(filteredSessions),
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取會話列表成功",
		Data:    response,
	})
}

// SendMessage godoc
// @Summary      發送對話訊息
// @Description  向 AI 角色發送訊息並獲取回應
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.SendMessageRequest true "對話訊息"
// @Success      200 {object} models.APIResponse{data=models.ChatResponse} "發送成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/message [post]
func SendMessage(c *gin.Context) {
	startTime := time.Now()
	requestID := c.GetString("request_id")
	
	var request models.SendMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.LogError(err, "request validation failed", logrus.Fields{
			"request_id": requestID,
			"path":       c.Request.URL.Path,
		})
		utils.HandleError(c, utils.ErrValidation.WithDetails(err.Error()))
		return
	}

	// 驗證必填字段
	if err := utils.ValidateRequired(map[string]interface{}{
		"session_id": request.SessionID,
		"message":    request.Message,
	}); err != nil {
		utils.HandleError(c, err)
		return
	}

	// 獲取用戶 ID（通常從 JWT token 中獲取）
	// 現在先使用模擬的用戶 ID
	userID := "user_demo_001"
	characterID := "char_001" // TODO: 從會話中獲取角色 ID

	utils.Logger.WithFields(logrus.Fields{
		"request_id":   requestID,
		"user_id":      userID,
		"session_id":   request.SessionID,
		"character_id": characterID,
		"message_len":  len(request.Message),
	}).Info("Processing chat message")

	// 創建 ChatService 實例
	chatService := services.NewChatService()

	// 處理消息
	processRequest := &services.ProcessMessageRequest{
		SessionID:   request.SessionID,
		UserMessage: request.Message,
		CharacterID: characterID,
		UserID:      userID,
		Metadata:    map[string]interface{}{},
	}

	response, err := chatService.ProcessMessage(c.Request.Context(), processRequest)
	if err != nil {
		utils.LogChatMessage(request.SessionID, userID, characterID, "unknown", 0, false)
		utils.HandleError(c, utils.ErrProcessingFailed.WithDetails(err.Error()).WithContext(map[string]interface{}{
			"session_id":   request.SessionID,
			"user_id":      userID,
			"character_id": characterID,
		}))
		return
	}

	// 記錄成功的對話處理
	duration := time.Since(startTime)
	utils.LogChatMessage(request.SessionID, userID, characterID, response.AIEngine, duration.Milliseconds(), true)
	
	// 記錄 API 請求
	utils.LogAPIRequest(c.Request.Method, c.Request.URL.Path, userID, request.SessionID, http.StatusOK, duration.Milliseconds())

	// 構建完整的回應，包含場景描述
	fullResponse := map[string]interface{}{
		"session_id":  response.SessionID,
		"message_id":  response.MessageID,
		"character_response": map[string]interface{}{
			"message":           response.CharacterDialogue,
			"emotion":           response.EmotionState.Mood,
			"affection_change":  1, // TODO: 計算實際的好感度變化
			"engine_used":       response.AIEngine,
			"response_time_ms":  response.ResponseTime.Milliseconds(),
		},
		"scene_description": response.SceneDescription,
		"character_action":  response.CharacterAction,
		"emotional_state": map[string]interface{}{
			"affection":      response.EmotionState.Affection,
			"mood":           response.EmotionState.Mood,
			"relationship":   response.EmotionState.Relationship,
			"intimacy_level": response.EmotionState.IntimacyLevel,
		},
		"ai_engine":   response.AIEngine,
		"nsfw_level":  response.NSFWLevel,
		"novel_choices": []models.NovelChoice{},
		"special_event": nil,
	}

	utils.Logger.WithFields(logrus.Fields{
		"request_id":     requestID,
		"session_id":     request.SessionID,
		"user_id":        userID,
		"character_id":   characterID,
		"response_time":  duration.Milliseconds(),
		"ai_engine":      response.AIEngine,
		"affection":      response.EmotionState.Affection,
	}).Info("Chat message processed successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "訊息發送成功",
		Data:    fullResponse,
	})
}

// RegenerateMessage godoc
// @Summary      重新生成回應
// @Description  重新生成 AI 角色的最後一個回應
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.RegenerateRequest true "重新生成參數"
// @Success      200 {object} models.APIResponse{data=models.ChatResponse} "重新生成成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "訊息不存在"
// @Router       /chat/regenerate [post]
func RegenerateMessage(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	var req models.RegenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "請求參數驗證失敗",
				Details: err.Error(),
			},
		})
		return
	}

	// 驗證必要參數
	if req.MessageID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_MESSAGE_ID",
				Message: "缺少訊息 ID",
			},
		})
		return
	}

	// 模擬驗證訊息是否存在和屬於當前用戶
	validMessages := []string{"msg_test_session_1", "msg_test_session_2", "msg_session_001_1", "msg_demo_session_1"}
	messageExists := false
	for _, validMsg := range validMessages {
		if req.MessageID == validMsg {
			messageExists = true
			break
		}
	}

	if !messageExists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MESSAGE_NOT_FOUND",
				Message: "訊息不存在",
			},
		})
		return
	}

	// 模擬用戶ID獲取
	userID := "user_alice123_001"
	if authHeader == "Bearer demo_token" {
		userID = "user_demo_001"
	}

	// 模擬重新生成的回應
	newMessageID := "msg_regen_" + utils.IntToString(int(time.Now().Unix()))
	
	regeneratedResponse := map[string]interface{}{
		"message_id":       newMessageID,
		"original_message_id": req.MessageID,
		"regeneration_reason": req.RegenerationReason,
		"character_response": map[string]interface{}{
			"message":           "讓我重新組織一下語言...你剛才說的話讓我想到了很多。",
			"emotion":           "thoughtful",
			"affection_change":  1,
			"engine_used":       "openai",
			"response_time_ms":  1450,
		},
		"scene_description": "他沉思了片刻，眼中閃過一絲思索的光芒，然後重新看向你...",
		"character_action":  "他調整了一下坐姿，手指輕敲桌面，似乎在重新考慮回應",
		"emotional_state": map[string]interface{}{
			"affection":      55,
			"mood":           "thoughtful",
			"relationship":   "friend",
			"intimacy_level": "friendly",
		},
		"ai_engine":        "openai",
		"nsfw_level":       1,
		"regenerated_at":   time.Now(),
		"regeneration_count": 1,
	}

	utils.Logger.WithFields(logrus.Fields{
		"user_id":              userID,
		"original_message_id":  req.MessageID,
		"new_message_id":       newMessageID,
		"regeneration_reason":  req.RegenerationReason,
	}).Info("Message regenerated successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "訊息重新生成成功",
		Data:    regeneratedResponse,
	})
}

// UpdateSessionMode godoc
// @Summary      切換對話模式
// @Description  切換對話會話的模式（普通/小說/NSFW）
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話 ID"
// @Param        request body models.UpdateModeRequest true "模式切換參數"
// @Success      200 {object} models.APIResponse{data=models.ChatSession} "切換成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id}/mode [put]
func UpdateSessionMode(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_SESSION_ID",
				Message: "缺少會話 ID",
			},
		})
		return
	}

	var req models.UpdateModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "請求參數驗證失敗",
				Details: err.Error(),
			},
		})
		return
	}

	// 驗證模式是否有效
	validModes := []string{"normal", "novel", "nsfw"}
	if !utils.StringInSlice(req.Mode, validModes) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_MODE",
				Message: "無效的對話模式，支援: normal, novel, nsfw",
			},
		})
		return
	}

	// 模擬檢查會話是否存在
	validSessions := []string{"test_session", "session_001", "demo_session", "session_1", "session_2", "session_3"}
	if !utils.StringInSlice(sessionID, validSessions) {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 模擬用戶ID獲取
	userID := "user_alice123_001"
	if authHeader == "Bearer demo_token" {
		userID = "user_demo_001"
	}

	// 模擬模式切換的系統消息
	var transitionMessage string
	var characterReaction string
	
	switch req.Mode {
	case "normal":
		transitionMessage = "我們回到正常的對話模式吧"
		characterReaction = "好的，我們可以聊些輕鬆的話題"
	case "novel":
		transitionMessage = "讓我們進入小說模式，開始一段故事"
		characterReaction = "有趣，讓我為你描繪一個美妙的故事場景..."
	case "nsfw":
		transitionMessage = "切換到成人模式"
		characterReaction = "我明白了...讓我們進入更私密的空間"
	}

	if req.TransitionMessage != "" {
		transitionMessage = req.TransitionMessage
	}

	// 模擬更新後的會話數據
	updatedSession := models.ChatSession{
		BaseModel: models.BaseModel{
			ID:        sessionID,
			CreatedAt: time.Now().AddDate(0, 0, -1),
			UpdatedAt: time.Now(),
		},
		UserID:        userID,
		CharacterID:   "char_001",
		Title:         "與陸寒淵的對話",
		Mode:          req.Mode,
		Status:        "active",
		Tags:          []string{"模式切換", req.Mode},
		MessageCount:  5,
		LastMessageAt: time.Now(),
	}

	response := map[string]interface{}{
		"session":             updatedSession,
		"mode_changed":        true,
		"previous_mode":       "normal",
		"new_mode":            req.Mode,
		"transition_message":  transitionMessage,
		"character_reaction":  characterReaction,
		"mode_change_time":    time.Now(),
		"available_features": getModeFeatures(req.Mode),
	}

	utils.Logger.WithFields(logrus.Fields{
		"user_id":            userID,
		"session_id":         sessionID,
		"new_mode":           req.Mode,
		"transition_message": transitionMessage,
	}).Info("Session mode updated successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "對話模式切換成功",
		Data:    response,
	})
}

// getModeFeatures 返回不同模式下可用的功能
func getModeFeatures(mode string) map[string]interface{} {
	features := map[string]interface{}{
		"normal": map[string]interface{}{
			"scene_description": true,
			"emotion_tracking":  true,
			"character_action":  true,
			"nsfw_content":      false,
			"story_choices":     false,
		},
		"novel": map[string]interface{}{
			"scene_description": true,
			"emotion_tracking":  true,
			"character_action":  true,
			"nsfw_content":      true,
			"story_choices":     true,
			"save_progress":     true,
		},
		"nsfw": map[string]interface{}{
			"scene_description": true,
			"emotion_tracking":  true,
			"character_action":  true,
			"nsfw_content":      true,
			"story_choices":     false,
			"intimate_actions":  true,
		},
	}
	
	if modeFeatures, exists := features[mode]; exists {
		return modeFeatures.(map[string]interface{})
	}
	
	return features["normal"].(map[string]interface{})
}

// GetMessageHistory godoc
// @Summary      獲取對話歷史
// @Description  獲取指定會話的對話訊息歷史
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話 ID"
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(50)
// @Param        before query string false "獲取該訊息 ID 之前的歷史"
// @Param        after query string false "獲取該訊息 ID 之後的歷史"
// @Success      200 {object} models.APIResponse{data=models.MessageHistoryResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id}/history [get]
func GetMessageHistory(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_SESSION_ID",
				Message: "缺少會話 ID",
			},
		})
		return
	}

	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	// 解析查詢參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := utils.ParseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := utils.ParseInt(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 模擬檢查會話是否存在
	validSessions := []string{"test_session", "session_001", "demo_session", "session_1", "session_2", "session_3"}
	sessionExists := false
	for _, validSession := range validSessions {
		if sessionID == validSession {
			sessionExists = true
			break
		}
	}

	if !sessionExists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 生成模擬歷史訊息
	totalMessages := 25
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	if endIndex > totalMessages {
		endIndex = totalMessages
	}

	var messages []models.MessageHistoryItem
	for i := startIndex; i < endIndex; i++ {
		// 交替生成用戶和助手訊息
		if i%2 == 0 {
			// 用戶訊息
			messages = append(messages, models.MessageHistoryItem{
				BaseModel: models.BaseModel{
					ID:        "msg_" + sessionID + "_" + utils.IntToString(i+1),
					CreatedAt: time.Now().Add(-time.Duration(totalMessages-i) * time.Hour),
					UpdatedAt: time.Now().Add(-time.Duration(totalMessages-i) * time.Hour),
				},
				SessionID:   sessionID,
				Role:        "user",
				Content:     "這是用戶訊息 " + utils.IntToString(i+1),
				NSFWLevel:   1,
			})
		} else {
			// 助手訊息
			messages = append(messages, models.MessageHistoryItem{
				BaseModel: models.BaseModel{
					ID:        "msg_" + sessionID + "_" + utils.IntToString(i+1),
					CreatedAt: time.Now().Add(-time.Duration(totalMessages-i) * time.Hour),
					UpdatedAt: time.Now().Add(-time.Duration(totalMessages-i) * time.Hour),
				},
				SessionID:        sessionID,
				Role:             "assistant",
				Content:          "你好，我是陸寒淵。很高興與你對話。",
				SceneDescription: "辦公室裡燈光微暖，陸寒淵放下手中的文件，深邃的眼眸望向你...",
				CharacterAction:  "他溫和地笑著，推了推鼻樑上的眼鏡",
				EmotionalState: map[string]interface{}{
					"affection":    50 + i,
					"mood":         "happy",
					"relationship": "friend",
				},
				NSFWLevel:    1,
				AIEngine:     "openai",
				ResponseTime: 1250 + i*50,
			})
		}
	}

	totalPages := (totalMessages + limit - 1) / limit
	
	response := models.MessageHistoryResponse{
		SessionID: sessionID,
		Messages:  messages,
		Pagination: models.PaginationInfo{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  totalMessages,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取對話歷史成功",
		Data:    response,
	})
}

// AddSessionTags godoc
// @Summary      為會話添加標籤
// @Description  為指定會話添加標籤
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話 ID"
// @Param        request body models.AddTagsRequest true "標籤列表"
// @Success      200 {object} models.APIResponse{data=models.ChatSession} "添加成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id}/tag [post]
func AddSessionTags(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_SESSION_ID",
				Message: "缺少會話 ID",
			},
		})
		return
	}

	var req models.AddTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "請求參數驗證失敗",
				Details: err.Error(),
			},
		})
		return
	}

	// 驗證標籤數量限制
	if len(req.Tags) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "EMPTY_TAGS",
				Message: "至少需要提供一個標籤",
			},
		})
		return
	}

	if len(req.Tags) > 10 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOO_MANY_TAGS",
				Message: "最多只能添加 10 個標籤",
			},
		})
		return
	}

	// 驗證標籤長度
	for _, tag := range req.Tags {
		if len(tag) > 20 {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "TAG_TOO_LONG",
					Message: "標籤長度不能超過 20 個字符: " + tag,
				},
			})
			return
		}
	}

	// 模擬檢查會話是否存在
	validSessions := []string{"test_session", "session_001", "demo_session", "session_1", "session_2", "session_3"}
	if !utils.StringInSlice(sessionID, validSessions) {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 模擬用戶ID獲取
	userID := "user_alice123_001"
	if authHeader == "Bearer demo_token" {
		userID = "user_demo_001"
	}

	// 模擬現有標籤
	existingTags := []string{"工作", "日常"}
	
	// 合併新舊標籤並去重
	allTagsMap := make(map[string]bool)
	for _, tag := range existingTags {
		allTagsMap[tag] = true
	}
	
	newTagsAdded := []string{}
	for _, tag := range req.Tags {
		if !allTagsMap[tag] {
			allTagsMap[tag] = true
			newTagsAdded = append(newTagsAdded, tag)
		}
	}

	// 轉換為切片
	updatedTags := make([]string, 0, len(allTagsMap))
	for tag := range allTagsMap {
		updatedTags = append(updatedTags, tag)
	}

	// 模擬更新後的會話
	updatedSession := models.ChatSession{
		BaseModel: models.BaseModel{
			ID:        sessionID,
			CreatedAt: time.Now().AddDate(0, 0, -1),
			UpdatedAt: time.Now(),
		},
		UserID:        userID,
		CharacterID:   "char_001",
		Title:         "與陸寒淵的對話",
		Mode:          "normal",
		Status:        "active",
		Tags:          updatedTags,
		MessageCount:  5,
		LastMessageAt: time.Now().Add(-2 * time.Hour),
	}

	response := map[string]interface{}{
		"session":         updatedSession,
		"new_tags_added":  newTagsAdded,
		"total_tags":      len(updatedTags),
		"all_tags":        updatedTags,
		"tags_added_at":   time.Now(),
	}

	utils.Logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"session_id":      sessionID,
		"new_tags":        newTagsAdded,
		"total_tag_count": len(updatedTags),
	}).Info("Session tags updated successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "會話標籤添加成功",
		Data:    response,
	})
}

// ExportChatHistory godoc
// @Summary      匯出對話記錄
// @Description  匯出指定會話的完整對話記錄
// @Tags         Chat
// @Accept       json
// @Produce      application/json
// @Security     BearerAuth
// @Param        session_id path string true "會話 ID"
// @Param        format query string false "匯出格式" default("json") Enums(json,txt,pdf)
// @Success      200 {object} models.APIResponse "匯出成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id}/export [get]
func ExportChatHistory(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_SESSION_ID",
				Message: "缺少會話 ID",
			},
		})
		return
	}

	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	// 匯出格式固定為 JSON
	format := "json"

	// 模擬檢查會話是否存在
	validSessions := []string{"test_session", "session_001", "demo_session", "session_1", "session_2", "session_3"}
	if !utils.StringInSlice(sessionID, validSessions) {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 生成匯出數據
	exportData := map[string]interface{}{
		"session_id":   sessionID,
		"export_time":  time.Now(),
		"format":       format,
		"total_messages": 25,
		"session_info": map[string]interface{}{
			"title":        "與陸寒淵的對話",
			"character_id": "char_001",
			"mode":         "normal",
			"created_at":   time.Now().AddDate(0, 0, -7),
		},
	}

	// JSON 格式 - 包含完整的結構化數據
	exportData["messages"] = []map[string]interface{}{
		{
			"id":        "msg_001",
			"role":      "user",
			"content":   "你好",
			"timestamp": time.Now().AddDate(0, 0, -7),
			"nsfw_level": 1,
		},
		{
			"id":               "msg_002",
			"role":             "assistant",
			"content":          "你好，我是陸寒淵。很高興見到你。",
			"scene_description": "辦公室裡燈光微暖，陸寒淵放下手中的文件，深邃的眼眸望向你...",
			"character_action":  "他溫和地笑著，推了推鼻樑上的眼鏡",
			"emotional_state": map[string]interface{}{
				"affection":    52,
				"mood":         "happy",
				"relationship": "friend",
			},
			"nsfw_level":    1,
			"ai_engine":     "openai",
			"response_time": 1250,
			"timestamp":     time.Now().AddDate(0, 0, -7).Add(2 * time.Minute),
		},
		{
			"id":        "msg_003",
			"role":      "user", 
			"content":   "今天工作累嗎？",
			"timestamp": time.Now().AddDate(0, 0, -7).Add(5 * time.Minute),
			"nsfw_level": 1,
		},
		{
			"id":               "msg_004",
			"role":             "assistant",
			"content":          "有點累，但看到你就不累了。",
			"scene_description": "他放下手中的筆，起身走向你，眼中帶著疲憊卻溫柔的光芒...",
			"character_action":  "他輕撫著太陽穴，然後對你露出寵溺的笑容",
			"emotional_state": map[string]interface{}{
				"affection":    55,
				"mood":         "tired_but_happy",
				"relationship": "friend",
			},
			"nsfw_level":    1,
			"ai_engine":     "openai", 
			"response_time": 1890,
			"timestamp":     time.Now().AddDate(0, 0, -7).Add(7 * time.Minute),
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "對話記錄匯出成功",
		Data:    exportData,
	})
}

// DeleteChatSession godoc
// @Summary      刪除對話會話
// @Description  刪除指定的對話會話及其所有訊息
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session_id path string true "會話 ID"
// @Success      200 {object} models.APIResponse "刪除成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "會話不存在"
// @Router       /chat/session/{session_id} [delete]
func DeleteChatSession(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || (len(authHeader) < 20 && authHeader != "Bearer demo_token") {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_SESSION_ID",
				Message: "缺少會話 ID",
			},
		})
		return
	}

	// 模擬檢查會話是否存在
	mockSessions := map[string]models.ChatSession{
		"test_session": {
			BaseModel: models.BaseModel{
				ID:        "test_session",
				CreatedAt: time.Now().AddDate(0, 0, -1),
				UpdatedAt: time.Now(),
			},
			UserID:        "user_alice123_001",
			CharacterID:   "char_001",
			Title:         "與陸寒淵的對話",
			Mode:          "normal",
			Status:        "active",
			Tags:          []string{"工作", "日常"},
			MessageCount:  5,
			LastMessageAt: time.Now().Add(-2 * time.Hour),
		},
		"session_001": {
			BaseModel: models.BaseModel{
				ID:        "session_001",
				CreatedAt: time.Now().AddDate(0, 0, -3),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			UserID:        "user_alice123_001",
			CharacterID:   "char_002",
			Title:         "與沈言墨的對話",
			Mode:          "normal",
			Status:        "active",
			Tags:          []string{"醫學", "溫柔"},
			MessageCount:  12,
			LastMessageAt: time.Now().Add(-1 * time.Hour),
		},
		"demo_session": {
			BaseModel: models.BaseModel{
				ID:        "demo_session",
				CreatedAt: time.Now().AddDate(0, 0, -7),
				UpdatedAt: time.Now().Add(-3 * time.Hour),
			},
			UserID:        "user_demo_001",
			CharacterID:   "char_001",
			Title:         "測試對話",
			Mode:          "nsfw",
			Status:        "paused",
			Tags:          []string{"測試", "NSFW"},
			MessageCount:  8,
			LastMessageAt: time.Now().Add(-3 * time.Hour),
		},
	}

	session, exists := mockSessions[sessionID]
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_NOT_FOUND",
				Message: "會話不存在",
			},
		})
		return
	}

	// 驗證會話所有權
	expectedUserID := "user_alice123_001"
	if authHeader == "Bearer demo_token" {
		expectedUserID = "user_demo_001"
	}

	if session.UserID != expectedUserID {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ACCESS_DENIED",
				Message: "無權限刪除此會話",
			},
		})
		return
	}

	// 檢查會話狀態，如果正在進行中給出警告信息
	var warningMessage string
	if session.Status == "active" {
		warningMessage = "此會話仍為活躍狀態，刪除後將無法恢復"
	}

	// 模擬刪除操作，記錄刪除的詳細信息
	deleteInfo := map[string]interface{}{
		"deleted_session": map[string]interface{}{
			"session_id":    sessionID,
			"title":         session.Title,
			"character_id":  session.CharacterID,
			"message_count": session.MessageCount,
			"created_at":    session.CreatedAt,
			"last_message_at": session.LastMessageAt,
		},
		"deletion_time": time.Now(),
		"user_id":       expectedUserID,
		"warning":       warningMessage,
		"deleted_messages_count": session.MessageCount,
	}

	utils.Logger.WithFields(logrus.Fields{
		"user_id":         expectedUserID,
		"session_id":      sessionID,
		"message_count":   session.MessageCount,
		"character_id":    session.CharacterID,
		"session_title":   session.Title,
	}).Info("Chat session deleted successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "會話刪除成功",
		Data:    deleteInfo,
	})
}

// Helper functions for chat session management

// calculateTotalMessages calculates the total message count across all sessions
func calculateTotalMessages(sessions []models.ChatSession) int {
	total := 0
	for _, session := range sessions {
		total += session.MessageCount
	}
	return total
}

// countSessionsByStatus counts sessions by their status
func countSessionsByStatus(sessions []models.ChatSession, status string) int {
	count := 0
	for _, session := range sessions {
		if session.Status == status {
			count++
		}
	}
	return count
}