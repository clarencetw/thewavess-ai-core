package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// SearchChats godoc
// @Summary      搜尋對話
// @Description  搜尋對話歷史記錄
// @Tags         Search
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        q query string true "搜尋關鍵詞"
// @Param        character_id query string false "角色ID過濾"
// @Param        date_from query string false "開始日期"
// @Param        date_to query string false "結束日期"
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Success      200 {object} models.APIResponse "搜尋成功"
// @Router       /search/chats [get]
func SearchChats(c *gin.Context) {
	startTime := time.Now()
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

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_QUERY",
				Message: "請提供搜尋關鍵詞",
			},
		})
		return
	}

	// 解析過濾參數
	characterID := c.Query("character_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	// 解析分頁參數
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

	// 執行搜尋
	results, totalCount, facets, err := searchChatMessages(ctx, userID.(string), query, characterID, dateFrom, dateTo, page, limit)
	if err != nil {
		utils.Logger.WithError(err).Error("搜尋對話失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SEARCH_ERROR",
				Message: "搜尋失敗",
			},
		})
		return
	}

	totalPages := (totalCount + limit - 1) / limit
	searchTime := time.Since(startTime)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "搜尋成功",
		Data: gin.H{
			"query":        query,
			"results":      results,
			"total_found":  totalCount,
			"current_page": page,
			"total_pages":  totalPages,
			"facets":       facets,
			"search_time":  fmt.Sprintf("%dms", searchTime.Milliseconds()),
			"filters": gin.H{
				"character_id": characterID,
				"date_from":    dateFrom,
				"date_to":      dateTo,
			},
		},
	})
}

// searchChatMessages performs the actual database search for chat messages
func searchChatMessages(ctx context.Context, userID, query, characterID, dateFrom, dateTo string, page, limit int) ([]gin.H, int, gin.H, error) {
	db := database.GetDB()
	if db == nil {
		return nil, 0, nil, fmt.Errorf("database connection unavailable")
	}

	// Base query for messages with session information
	baseQuery := db.NewSelect().
		Model((*models.Message)(nil)).
		Column("m.id", "m.session_id", "m.role", "m.content", "m.scene_description", "m.character_action", "m.nsfw_level", "m.created_at").
		Column("cs.title", "cs.character_id").
		Column("c.name", "c.avatar_url").
		Join("JOIN chat_sessions cs ON cs.id = m.session_id").
		Join("JOIN characters c ON c.id = cs.character_id").
		Where("cs.user_id = ?", userID).
		Where("cs.status != ?", "deleted")

	// Add full-text search on message content
	if query != "" {
		// Use PostgreSQL full-text search
		baseQuery = baseQuery.Where("to_tsvector('simple', m.content) @@ plainto_tsquery('simple', ?)", query)
	}

	// Apply filters
	if characterID != "" {
		baseQuery = baseQuery.Where("cs.character_id = ?", characterID)
	}

	if dateFrom != "" {
		baseQuery = baseQuery.Where("m.created_at >= ?", dateFrom)
	}

	if dateTo != "" {
		baseQuery = baseQuery.Where("m.created_at <= ?", dateTo)
	}

	// Get total count
	countQuery := baseQuery
	totalCount, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Apply pagination and ordering
	offset := (page - 1) * limit
	var results []struct {
		models.Message
		SessionTitle   string `bun:"title"`
		CharacterID    string `bun:"character_id"`
		CharacterName  string `bun:"name"`
		CharacterAvatar string `bun:"avatar_url"`
	}

	err = baseQuery.
		Order("m.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx, &results)

	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to execute search query: %w", err)
	}

	// Convert to response format
	searchResults := make([]gin.H, len(results))
	characterCounts := make(map[string]int)
	nsfw_level_counts := make(map[int]int)

	for i, result := range results {
		searchResults[i] = gin.H{
			"id":               result.ID,
			"session_id":       result.SessionID,
			"session_title":    result.SessionTitle,
			"role":             result.Role,
			"content":          result.Content,
			"scene_description": result.SceneDescription,
			"character_action": result.CharacterAction,
			"nsfw_level":       result.NSFWLevel,
			"created_at":       result.CreatedAt,
			"character": gin.H{
				"id":         result.CharacterID,
				"name":       result.CharacterName,
				"avatar_url": result.CharacterAvatar,
			},
			"relevance": calculateSearchRelevance(result.Content, query),
		}

		// Count facets
		characterCounts[result.CharacterName]++
		nsfw_level_counts[result.NSFWLevel]++
	}

	// Build facets for filtering
	facets := gin.H{
		"characters": characterCounts,
		"nsfw_levels": nsfw_level_counts,
	}

	return searchResults, totalCount, facets, nil
}

// calculateSearchRelevance calculates search relevance score based on query match
func calculateSearchRelevance(content, query string) float64 {
	if query == "" {
		return 1.0
	}

	content = strings.ToLower(content)
	query = strings.ToLower(query)

	// Simple relevance calculation
	if strings.Contains(content, query) {
		// Calculate position-based relevance (earlier matches score higher)
		pos := strings.Index(content, query)
		lengthFactor := float64(len(query)) / float64(len(content))
		positionFactor := 1.0 - (float64(pos) / float64(len(content)))
		return 0.5 + (lengthFactor * 0.3) + (positionFactor * 0.2)
	}

	// Check for partial matches
	queryWords := strings.Fields(query)
	matchCount := 0
	for _, word := range queryWords {
		if strings.Contains(content, word) {
			matchCount++
		}
	}

	if matchCount > 0 {
		return float64(matchCount) / float64(len(queryWords)) * 0.7
	}

	return 0.1
}

// GlobalSearch godoc  
// @Summary      全局搜尋
// @Description  在所有內容中搜尋（對話、角色、記憶等）
// @Tags         Search
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        q query string true "搜尋關鍵詞"
// @Param        type query string false "內容類型過濾"
// @Success      200 {object} models.APIResponse "搜尋成功"
// @Router       /search/global [get]
func GlobalSearch(c *gin.Context) {
	startTime := time.Now()
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

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_QUERY",
				Message: "請提供搜尋關鍵詞",
			},
		})
		return
	}

	typeFilter := c.Query("type")

	// 執行全局搜尋
	results, err := performGlobalSearch(ctx, userID.(string), query, typeFilter)
	if err != nil {
		utils.Logger.WithError(err).Error("全局搜尋失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SEARCH_ERROR",
				Message: "搜尋失敗",
			},
		})
		return
	}

	searchTime := time.Since(startTime)
	results["search_time"] = fmt.Sprintf("%dms", searchTime.Milliseconds())
	results["query"] = query

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "全局搜尋成功",
		Data:    results,
	})
}

// performGlobalSearch executes global search across multiple content types
func performGlobalSearch(ctx context.Context, userID, query, typeFilter string) (gin.H, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection unavailable")
	}

	categories := gin.H{}
	totalResults := 0

	// Search in chat messages if type filter allows
	if typeFilter == "" || typeFilter == "all" || typeFilter == "chats" {
		chatResults, count, err := searchInChats(ctx, db, userID, query)
		if err != nil {
			return nil, fmt.Errorf("chat search failed: %w", err)
		}
		categories["chats"] = gin.H{
			"count": count,
			"top_results": chatResults,
		}
		totalResults += count
	}

	// Search in characters if type filter allows
	if typeFilter == "" || typeFilter == "all" || typeFilter == "characters" {
		characterResults, count, err := searchInCharacters(ctx, db, query)
		if err != nil {
			return nil, fmt.Errorf("character search failed: %w", err)
		}
		categories["characters"] = gin.H{
			"count": count,
			"top_results": characterResults,
		}
		totalResults += count
	}

	// Search in memories if type filter allows
	if typeFilter == "" || typeFilter == "all" || typeFilter == "memories" {
		memoryResults, count, err := searchInMemories(ctx, db, userID, query)
		if err != nil {
			return nil, fmt.Errorf("memory search failed: %w", err)
		}
		categories["memories"] = gin.H{
			"count": count,
			"top_results": memoryResults,
		}
		totalResults += count
	}

	// Generate search suggestions based on query
	suggestions := generateSearchSuggestions(query)

	results := gin.H{
		"categories": categories,
		"total_results": totalResults,
		"suggestions": suggestions,
	}

	return results, nil
}

// searchInChats searches for messages in chat sessions
func searchInChats(ctx context.Context, db *bun.DB, userID, query string) ([]gin.H, int, error) {
	var results []struct {
		models.Message
		SessionTitle   string `bun:"title"`
		CharacterName  string `bun:"name"`
	}

	// Search in message content
	err := db.NewSelect().
		Model((*models.Message)(nil)).
		Column("m.id", "m.session_id", "m.content", "m.created_at").
		Column("cs.title").
		Column("c.name").
		Join("JOIN chat_sessions cs ON cs.id = m.session_id").
		Join("JOIN characters c ON c.id = cs.character_id").
		Where("cs.user_id = ?", userID).
		Where("cs.status != ?", "deleted").
		Where("to_tsvector('simple', m.content) @@ plainto_tsquery('simple', ?)", query).
		Order("m.created_at DESC").
		Limit(5).
		Scan(ctx, &results)

	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	chatResults := make([]gin.H, len(results))
	for i, result := range results {
		excerpt := result.Content
		if len(excerpt) > 100 {
			excerpt = excerpt[:100] + "..."
		}

		chatResults[i] = gin.H{
			"id":        result.SessionID,
			"title":     result.SessionTitle,
			"excerpt":   excerpt,
			"type":      "chat_session",
			"relevance": calculateSearchRelevance(result.Content, query),
			"character": result.CharacterName,
			"created_at": result.CreatedAt,
		}
	}

	return chatResults, len(results), nil
}

// searchInCharacters searches for characters by name and description
func searchInCharacters(ctx context.Context, db *bun.DB, query string) ([]gin.H, int, error) {
	var characters []models.Character

	// Search in character name and description
	err := db.NewSelect().
		Model(&characters).
		Where("to_tsvector('simple', name || ' ' || description) @@ plainto_tsquery('simple', ?)", query).
		Where("is_active = ?", true).
		Order("name").
		Limit(5).
		Scan(ctx)

	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	characterResults := make([]gin.H, len(characters))
	for i, char := range characters {
		excerpt := char.Description
		if len(excerpt) > 100 {
			excerpt = excerpt[:100] + "..."
		}

		characterResults[i] = gin.H{
			"id":        char.ID,
			"name":      char.Name,
			"excerpt":   excerpt,
			"type":      "character",
			"relevance": calculateSearchRelevance(char.Name+" "+char.Description, query),
			"avatar_url": char.AvatarURL,
		}
	}

	return characterResults, len(characters), nil
}

// searchInMemories searches in memory content
func searchInMemories(ctx context.Context, db *bun.DB, userID, query string) ([]gin.H, int, error) {
	// For now, return empty results as memory search is complex
	// This would require implementing memory storage tables first
	return []gin.H{}, 0, nil
}

// generateSearchSuggestions creates related search suggestions
func generateSearchSuggestions(query string) []string {
	// Simple suggestion generation based on query
	suggestions := []string{}

	queryLower := strings.ToLower(query)
	commonSuggestions := map[string][]string{
		"愛":   {"愛情", "戀愛", "愛好"},
		"工作":  {"職場", "事業", "工作環境"},
		"生活":  {"日常生活", "生活方式", "生活習慣"},
		"音樂":  {"喜歡的音樂", "音樂偏好", "歌曲"},
		"電影":  {"電影類型", "最愛電影", "觀影習慣"},
		"旅行":  {"旅遊地點", "旅行經驗", "度假"},
		"食物":  {"美食", "料理", "餐廳"},
		"運動":  {"健身", "體育活動", "運動習慣"},
	}

	for keyword, related := range commonSuggestions {
		if strings.Contains(queryLower, keyword) {
			suggestions = append(suggestions, related...)
			break
		}
	}

	// If no specific suggestions found, provide general ones
	if len(suggestions) == 0 {
		suggestions = []string{
			"最近的對話",
			"重要回憶",
			"角色互動",
		}
	}

	// Limit to 3 suggestions
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return suggestions
}