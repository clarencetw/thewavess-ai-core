package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
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
	_ = limit // Used for pagination (static implementation)

	// 靜態數據回應
	results := []gin.H{
		{
			"session_id":     "session_045",
			"character_id":   "char_001",
			"character_name": "陸燁銘",
			"message": gin.H{
				"id":        "msg_1234",
				"content":   "我一直在想，如果當初我們在不同的場合相遇，結果會不會不一樣？",
				"highlight": "如果當初我們在不同的<mark>場合相遇</mark>",
				"role":      "assistant",
				"timestamp": time.Now().AddDate(0, 0, -5),
			},
			"context": gin.H{
				"session_title": "深夜談心",
				"total_messages": 156,
				"session_date":   time.Now().AddDate(0, 0, -5),
			},
			"relevance": 0.92,
		},
		{
			"session_id":     "session_089",
			"character_id":   "char_002",
			"character_name": "沈言墨",
			"message": gin.H{
				"id":        "msg_5678",
				"content":   "第一次相遇的時候，我就覺得你很特別",
				"highlight": "第一次<mark>相遇</mark>的時候",
				"role":      "assistant",
				"timestamp": time.Now().AddDate(0, 0, -10),
			},
			"context": gin.H{
				"session_title": "回憶往事",
				"total_messages": 89,
				"session_date":   time.Now().AddDate(0, 0, -10),
			},
			"relevance": 0.85,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "搜尋成功",
		Data: gin.H{
			"query":        query,
			"results":      results,
			"total_found":  42,
			"current_page": page,
			"total_pages":  3,
			"facets": gin.H{
				"characters": []gin.H{
					{"id": "char_001", "name": "陸燁銘", "count": 25},
					{"id": "char_002", "name": "沈言墨", "count": 17},
				},
				"date_range": gin.H{
					"earliest": time.Now().AddDate(0, -1, 0),
					"latest":   time.Now(),
				},
				"message_types": gin.H{
					"user":      18,
					"assistant": 24,
				},
			},
			"search_time": "156ms",
		},
	})
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

	// 靜態數據回應
	results := gin.H{
		"query": query,
		"categories": gin.H{
			"chats": gin.H{
				"count": 15,
				"top_results": []gin.H{
					{
						"id":        "chat_001",
						"title":     "深夜談心",
						"excerpt":   "...我們聊到了很多關於未來的計劃...",
						"type":      "chat_session",
						"relevance": 0.95,
					},
				},
			},
			"characters": gin.H{
				"count": 2,
				"top_results": []gin.H{
					{
						"id":        "char_001",
						"name":      "陸燁銘",
						"excerpt":   "霸道總裁，外冷內熱",
						"type":      "character",
						"relevance": 0.88,
					},
				},
			},
			"memories": gin.H{
				"count": 8,
				"top_results": []gin.H{
					{
						"id":        "mem_001",
						"title":     "第一次見面",
						"excerpt":   "在咖啡廳的邂逅",
						"type":      "memory",
						"relevance": 0.82,
					},
				},
			},
			"novels": gin.H{
				"count": 3,
				"top_results": []gin.H{
					{
						"id":        "novel_001",
						"title":     "霸道總裁的溫柔",
						"excerpt":   "現代都市言情故事",
						"type":      "novel",
						"relevance": 0.75,
					},
				},
			},
		},
		"total_results": 28,
		"search_time":   "287ms",
		"suggestions": []string{
			"相遇的場景",
			"初次見面",
			"命運安排",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "全局搜尋成功",
		Data:    results,
	})
}