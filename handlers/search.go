package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	dbmodel "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
)

// ChatSearchResult 描述單筆聊天搜索結果
type ChatSearchResult struct {
	ID        string              `json:"id"`
	ChatID    string              `json:"chat_id"`
	ChatTitle string              `json:"chat_title"`
	Role      string              `json:"role"`
	Dialogue  string              `json:"dialogue"`
	NSFWLevel int                 `json:"nsfw_level"`
	CreatedAt time.Time           `json:"created_at"`
	Character ChatSearchCharacter `json:"character"`
	Relevance float64             `json:"relevance"`
}

// ChatSearchCharacter 附帶的角色摘要
type ChatSearchCharacter struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// ChatSearchFilters 回傳用的篩選資訊
type ChatSearchFilters struct {
	CharacterID string `json:"character_id,omitempty"`
	DateFrom    string `json:"date_from,omitempty"`
	DateTo      string `json:"date_to,omitempty"`
}

// ChatSearchFacets 搜索分面統計
type ChatSearchFacets struct {
	Characters []CharacterFacet `json:"characters"`
	NSFWLevels []NSFWLevelFacet `json:"nsfw_levels"`
}

// CharacterFacet 角色分面
type CharacterFacet struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// NSFWLevelFacet NSFW 等級分面
type NSFWLevelFacet struct {
	Level int `json:"level"`
	Count int `json:"count"`
}

// ChatSearchResponse 聊天搜索回應
type ChatSearchResponse struct {
	Query      string             `json:"query"`
	Results    []ChatSearchResult `json:"results"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
	TookMs     int64              `json:"took_ms"`
	Filters    ChatSearchFilters  `json:"filters"`
	Facets     ChatSearchFacets   `json:"facets"`
}

// GlobalSearchResponse 全域搜索回應
type GlobalSearchResponse struct {
	Query        string              `json:"query"`
	TookMs       int64               `json:"took_ms"`
	TotalResults int                 `json:"total_results"`
	Suggestions  []string            `json:"suggestions"`
	Results      GlobalSearchResults `json:"results"`
}

// GlobalSearchResults 分類結果
type GlobalSearchResults struct {
	Chats      *GlobalChatCategory      `json:"chats,omitempty"`
	Characters *GlobalCharacterCategory `json:"characters,omitempty"`
}

// GlobalChatCategory 聊天搜索分類
type GlobalChatCategory struct {
	Count   int                `json:"count"`
	Results []GlobalChatResult `json:"results"`
}

// GlobalChatResult 全域搜索中的聊天摘要
type GlobalChatResult struct {
	ChatID    string                 `json:"chat_id"`
	Title     string                 `json:"title"`
	Excerpt   string                 `json:"excerpt"`
	Character GlobalCharacterSummary `json:"character"`
	CreatedAt time.Time              `json:"created_at"`
	Relevance float64                `json:"relevance"`
}

// GlobalCharacterSummary 聊天結果中的角色資訊
type GlobalCharacterSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GlobalCharacterCategory 角色搜索分類
type GlobalCharacterCategory struct {
	Count   int                     `json:"count"`
	Results []GlobalCharacterResult `json:"results"`
}

// GlobalCharacterResult 全域搜索中的角色摘要
type GlobalCharacterResult struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Excerpt   string  `json:"excerpt"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Relevance float64 `json:"relevance"`
}

// 使用全局 character service 實例
func getCharacterServiceForSearch() *services.CharacterService {
	return services.GetCharacterService()
}

// SearchChats godoc
// @Summary      搜尋對話
// @Description  用戶搜尋自己的對話歷史記錄
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
// @Param        offset query int false "結果偏移量（可選，優先於 page）"
// @Success      200 {object} models.APIResponse "搜尋成功"
// @Router       /search/chats [get]
func SearchChats(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	if ctx == nil {
		ctx = context.Background()
	}

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

	query := strings.TrimSpace(c.Query("q"))
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

	characterID := c.Query("character_id")
	dateFromRaw := c.Query("date_from")
	dateToRaw := c.Query("date_to")

	page := parsePositiveInt(c.Query("page"), 1)
	limit := parseBoundedPositiveInt(c.Query("limit"), 20, 1, 50)

	offset := (page - 1) * limit
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
			page = (offset / limit) + 1
		}
	}

	var dateFromPtr *time.Time
	if dateFromRaw != "" {
		parsed, err := parseDateParam(dateFromRaw)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_DATE",
					Message: fmt.Sprintf("date_from 無效: %v", err),
				},
			})
			return
		}
		dateFromPtr = &parsed
	}

	var dateToPtr *time.Time
	if dateToRaw != "" {
		parsed, err := parseDateParam(dateToRaw)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_DATE",
					Message: fmt.Sprintf("date_to 無效: %v", err),
				},
			})
			return
		}
		dateToPtr = &parsed
	}

	results, totalCount, facets, err := searchChatMessages(ctx, userID.(string), query, characterID, dateFromPtr, dateToPtr, limit, offset)
	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("搜尋對話失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SEARCH_ERROR",
				Message: "搜尋失敗",
			},
		})
		return
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}

	searchTime := time.Since(startTime)

	utils.Logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"query":       query,
		"total_found": totalCount,
		"search_time": searchTime.Milliseconds(),
	}).Info("用戶執行聊天搜尋")

	response := ChatSearchResponse{
		Query:      query,
		Results:    results,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		TookMs:     searchTime.Milliseconds(),
		Filters: ChatSearchFilters{
			CharacterID: characterID,
			DateFrom:    dateFromRaw,
			DateTo:      dateToRaw,
		},
		Facets: facets,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "搜尋成功",
		Data:    response,
	})
}

func searchChatMessages(ctx context.Context, userID, query, characterID string, dateFrom, dateTo *time.Time, limit, offset int) ([]ChatSearchResult, int, ChatSearchFacets, error) {
	db := GetDB()
	if db == nil {
		return nil, 0, ChatSearchFacets{}, fmt.Errorf("database connection unavailable")
	}

	baseQuery := db.NewSelect().
		Model((*dbmodel.MessageDB)(nil)).
		Column("m.id", "m.chat_id", "m.role", "m.dialogue", "m.nsfw_level", "m.created_at").
		ColumnExpr("cs.title AS chat_title").
		ColumnExpr("cs.character_id AS character_id").
		ColumnExpr("c.name AS character_name").
		ColumnExpr("c.avatar_url AS character_avatar").
		Join("JOIN chats cs ON cs.id = m.chat_id").
		Join("JOIN characters c ON c.id = cs.character_id").
		Where("cs.status != ?", "deleted").
		Where("cs.user_id = ?", userID).
		Where("to_tsvector('simple', coalesce(m.dialogue, '')) @@ plainto_tsquery('simple', ?)", query)

	if characterID != "" {
		baseQuery = baseQuery.Where("cs.character_id = ?", characterID)
	}

	if dateFrom != nil {
		baseQuery = baseQuery.Where("m.created_at >= ?", *dateFrom)
	}

	if dateTo != nil {
		baseQuery = baseQuery.Where("m.created_at <= ?", *dateTo)
	}

	totalCount, err := baseQuery.Clone().Count(ctx)
	if err != nil {
		return nil, 0, ChatSearchFacets{}, fmt.Errorf("failed to count search results: %w", err)
	}

	if totalCount == 0 {
		return []ChatSearchResult{}, 0, ChatSearchFacets{}, nil
	}

	var rows []struct {
		ID              string    `bun:"id"`
		ChatID          string    `bun:"chat_id"`
		Role            string    `bun:"role"`
		Dialogue        string    `bun:"dialogue"`
		NSFWLevel       int       `bun:"nsfw_level"`
		CreatedAt       time.Time `bun:"created_at"`
		ChatTitle       string    `bun:"chat_title"`
		CharacterID     string    `bun:"character_id"`
		CharacterName   string    `bun:"character_name"`
		CharacterAvatar *string   `bun:"character_avatar"`
	}

	queryWithPaging := baseQuery.Clone().
		Order("m.created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := queryWithPaging.Scan(ctx, &rows); err != nil {
		return nil, 0, ChatSearchFacets{}, fmt.Errorf("failed to execute search query: %w", err)
	}

	results := make([]ChatSearchResult, len(rows))
	characterFacetMap := make(map[string]*CharacterFacet)
	nsfwFacetMap := make(map[int]*NSFWLevelFacet)

	for i, row := range rows {
		character := ChatSearchCharacter{
			ID:        row.CharacterID,
			Name:      row.CharacterName,
			AvatarURL: row.CharacterAvatar,
		}

		results[i] = ChatSearchResult{
			ID:        row.ID,
			ChatID:    row.ChatID,
			ChatTitle: row.ChatTitle,
			Role:      row.Role,
			Dialogue:  row.Dialogue,
			NSFWLevel: row.NSFWLevel,
			CreatedAt: row.CreatedAt,
			Character: character,
			Relevance: calculateSearchRelevance(row.Dialogue, query),
		}

		if facet, exists := characterFacetMap[row.CharacterID]; exists {
			facet.Count++
		} else {
			characterFacetMap[row.CharacterID] = &CharacterFacet{
				ID:    row.CharacterID,
				Name:  row.CharacterName,
				Count: 1,
			}
		}

		if facet, exists := nsfwFacetMap[row.NSFWLevel]; exists {
			facet.Count++
		} else {
			nsfwFacetMap[row.NSFWLevel] = &NSFWLevelFacet{
				Level: row.NSFWLevel,
				Count: 1,
			}
		}
	}

	characterFacets := make([]CharacterFacet, 0, len(characterFacetMap))
	for _, facet := range characterFacetMap {
		characterFacets = append(characterFacets, *facet)
	}
	sort.Slice(characterFacets, func(i, j int) bool {
		if characterFacets[i].Count == characterFacets[j].Count {
			return characterFacets[i].Name < characterFacets[j].Name
		}
		return characterFacets[i].Count > characterFacets[j].Count
	})

	nsfwFacets := make([]NSFWLevelFacet, 0, len(nsfwFacetMap))
	for _, facet := range nsfwFacetMap {
		nsfwFacets = append(nsfwFacets, *facet)
	}
	sort.Slice(nsfwFacets, func(i, j int) bool {
		return nsfwFacets[i].Level < nsfwFacets[j].Level
	})

	facets := ChatSearchFacets{
		Characters: characterFacets,
		NSFWLevels: nsfwFacets,
	}

	return results, totalCount, facets, nil
}

// calculateSearchRelevance calculates search relevance score based on query match
func calculateSearchRelevance(content, query string) float64 {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return 1.0
	}

	content = strings.ToLower(content)
	contentLength := len(content)
	if contentLength == 0 {
		return 0.1
	}

	if strings.Contains(content, query) {
		pos := strings.Index(content, query)
		lengthFactor := float64(len(query)) / float64(contentLength)
		positionFactor := 1.0 - (float64(pos) / float64(contentLength))
		return 0.5 + (lengthFactor * 0.3) + (positionFactor * 0.2)
	}

	queryWords := strings.Fields(query)
	if len(queryWords) == 0 {
		return 0.2
	}

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
	ctx := c.Request.Context()
	if ctx == nil {
		ctx = context.Background()
	}

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

	query := strings.TrimSpace(c.Query("q"))
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

	typeFilter := strings.ToLower(strings.TrimSpace(c.Query("type")))

	var results GlobalSearchResults
	totalResults := 0

	if typeFilter == "" || typeFilter == "all" || typeFilter == "chats" {
		chatResults, count, err := searchInChats(ctx, userID.(string), query)
		if err != nil {
			utils.Logger.WithError(err).Error("全局搜尋聊天失敗")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "SEARCH_ERROR",
					Message: "聊天搜尋失敗",
				},
			})
			return
		}
		results.Chats = &GlobalChatCategory{
			Count:   count,
			Results: chatResults,
		}
		totalResults += count
	}

	if typeFilter == "" || typeFilter == "all" || typeFilter == "characters" {
		characterResults, count, err := searchInCharacters(ctx, query)
		if err != nil {
			utils.Logger.WithError(err).Error("全局搜尋角色失敗")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "SEARCH_ERROR",
					Message: "角色搜尋失敗",
				},
			})
			return
		}
		results.Characters = &GlobalCharacterCategory{
			Count:   count,
			Results: characterResults,
		}
		totalResults += count
	}

	searchTime := time.Since(startTime)
	suggestions := generateSearchSuggestions(query)

	response := GlobalSearchResponse{
		Query:        query,
		TookMs:       searchTime.Milliseconds(),
		TotalResults: totalResults,
		Suggestions:  suggestions,
		Results:      results,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "全局搜尋成功",
		Data:    response,
	})
}

// searchInChats 搜尋聊天消息摘要
func searchInChats(ctx context.Context, userID, query string) ([]GlobalChatResult, int, error) {
	db := GetDB()
	if db == nil {
		return nil, 0, fmt.Errorf("database connection unavailable")
	}

	var rows []struct {
		ChatID        string    `bun:"chat_id"`
		Dialogue      string    `bun:"dialogue"`
		CreatedAt     time.Time `bun:"created_at"`
		ChatTitle     string    `bun:"title"`
		CharacterID   string    `bun:"character_id"`
		CharacterName string    `bun:"name"`
	}

	err := db.NewSelect().
		Model((*dbmodel.MessageDB)(nil)).
		Column("m.chat_id", "m.dialogue", "m.created_at").
		ColumnExpr("cs.title AS title").
		ColumnExpr("c.id AS character_id").
		ColumnExpr("c.name AS name").
		Join("JOIN chats cs ON cs.id = m.chat_id").
		Join("JOIN characters c ON c.id = cs.character_id").
		Where("cs.user_id = ?", userID).
		Where("cs.status != ?", "deleted").
		Where("to_tsvector('simple', coalesce(m.dialogue, '')) @@ plainto_tsquery('simple', ?)", query).
		Order("m.created_at DESC").
		Limit(5).
		Scan(ctx, &rows)

	if err != nil {
		return nil, 0, err
	}

	results := make([]GlobalChatResult, len(rows))
	for i, row := range rows {
		results[i] = GlobalChatResult{
			ChatID:  row.ChatID,
			Title:   row.ChatTitle,
			Excerpt: makeExcerpt(row.Dialogue, 120),
			Character: GlobalCharacterSummary{
				ID:   row.CharacterID,
				Name: row.CharacterName,
			},
			CreatedAt: row.CreatedAt,
			Relevance: calculateSearchRelevance(row.Dialogue, query),
		}
	}

	return results, len(results), nil
}

// searchInCharacters searches for characters by name and description using memory-based system
func searchInCharacters(ctx context.Context, query string) ([]GlobalCharacterResult, int, error) {
	service := getCharacterServiceForSearch()
	if service == nil {
		return nil, 0, fmt.Errorf("character service not available")
	}

	characters, err := service.SearchCharacters(ctx, query, 5)
	if err != nil {
		return nil, 0, err
	}

	results := make([]GlobalCharacterResult, len(characters))
	for i, char := range characters {
		description := ""
		if char.UserDescription != nil && *char.UserDescription != "" {
			description = *char.UserDescription
		} else {
			description = string(char.Type)
		}

		results[i] = GlobalCharacterResult{
			ID:        char.ID,
			Name:      char.Name,
			Excerpt:   makeExcerpt(description, 120),
			AvatarURL: char.Metadata.AvatarURL,
			Relevance: calculateSearchRelevance(char.Name+" "+description, query),
		}
	}

	return results, len(results), nil
}

// generateSearchSuggestions creates related search suggestions
func generateSearchSuggestions(query string) []string {
	suggestions := []string{}

	queryLower := strings.ToLower(query)
	commonSuggestions := map[string][]string{
		"愛":  {"愛情", "戀愛", "愛好"},
		"工作": {"職場", "事業", "工作環境"},
		"生活": {"日常生活", "生活方式", "生活習慣"},
		"音樂": {"喜歡的音樂", "音樂偏好", "歌曲"},
		"電影": {"電影類型", "最愛電影", "觀影習慣"},
		"旅行": {"旅遊地點", "旅行經驗", "度假"},
		"食物": {"美食", "料理", "餐廳"},
		"運動": {"健身", "體育活動", "運動習慣"},
	}

	for keyword, related := range commonSuggestions {
		if strings.Contains(queryLower, keyword) {
			suggestions = append(suggestions, related...)
			break
		}
	}

	if len(suggestions) == 0 {
		suggestions = []string{
			"最近的對話",
			"重要回憶",
			"角色互動",
		}
	}

	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return suggestions
}

func parseDateParam(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("無法解析日期格式: %s", raw)
}

func parsePositiveInt(raw string, defaultVal int) int {
	if raw == "" {
		return defaultVal
	}

	if v, err := strconv.Atoi(raw); err == nil && v > 0 {
		return v
	}

	return defaultVal
}

func parseBoundedPositiveInt(raw string, defaultVal, minVal, maxVal int) int {
	v := parsePositiveInt(raw, defaultVal)
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func makeExcerpt(text string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}

	trimmed := strings.TrimSpace(text)
	runes := []rune(trimmed)
	if len(runes) <= maxLength {
		return trimmed
	}

	return string(runes[:maxLength]) + "..."
}
