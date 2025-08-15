package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
)

// GetAllTags godoc
// @Summary      獲取所有標籤
// @Description  獲取系統中所有可用的標籤列表
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /tags [get]
func GetAllTags(c *gin.Context) {
	// 靜態數據回應
	tags := []map[string]interface{}{
		{
			"id":         "tag_001",
			"name":       "甜寵",
			"category":   "genre",
			"usage_count": 1523,
			"color":      "#FF69B4",
			"created_at": time.Now().AddDate(0, -3, 0),
		},
		{
			"id":         "tag_002",
			"name":       "腹黑",
			"category":   "personality",
			"usage_count": 892,
			"color":      "#8B008B",
			"created_at": time.Now().AddDate(0, -3, 0),
		},
		{
			"id":         "tag_003",
			"name":       "霸總",
			"category":   "role",
			"usage_count": 2156,
			"color":      "#4169E1",
			"created_at": time.Now().AddDate(0, -3, 0),
		},
		{
			"id":         "tag_004",
			"name":       "古風",
			"category":   "style",
			"usage_count": 1678,
			"color":      "#8B4513",
			"created_at": time.Now().AddDate(0, -3, 0),
		},
		{
			"id":         "tag_005",
			"name":       "現代",
			"category":   "style",
			"usage_count": 3421,
			"color":      "#00CED1",
			"created_at": time.Now().AddDate(0, -3, 0),
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取標籤列表成功",
		Data: gin.H{
			"tags":        tags,
			"total_count": len(tags),
			"categories": []string{"genre", "personality", "role", "style"},
		},
	})
}

// GetPopularTags godoc
// @Summary      獲取熱門標籤
// @Description  獲取使用次數最多的熱門標籤
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        limit query int false "數量限制" default(10)
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /tags/popular [get]
func GetPopularTags(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// 靜態數據回應 - 按使用次數排序
	popularTags := []map[string]interface{}{
		{
			"id":          "tag_005",
			"name":        "現代",
			"category":    "style",
			"usage_count": 3421,
			"trend":       "up", // up, down, stable
			"trend_percentage": 12.5,
		},
		{
			"id":          "tag_003",
			"name":        "霸總",
			"category":    "role",
			"usage_count": 2156,
			"trend":       "up",
			"trend_percentage": 8.3,
		},
		{
			"id":          "tag_004",
			"name":        "古風",
			"category":    "style",
			"usage_count": 1678,
			"trend":       "stable",
			"trend_percentage": 0.5,
		},
		{
			"id":          "tag_001",
			"name":        "甜寵",
			"category":    "genre",
			"usage_count": 1523,
			"trend":       "down",
			"trend_percentage": -3.2,
		},
		{
			"id":          "tag_002",
			"name":        "腹黑",
			"category":    "personality",
			"usage_count": 892,
			"trend":       "up",
			"trend_percentage": 15.7,
		},
	}

	// 根據 limit 返回相應數量
	if limit < len(popularTags) {
		popularTags = popularTags[:limit]
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取熱門標籤成功",
		Data: gin.H{
			"tags":        popularTags,
			"period":      "last_7_days",
			"updated_at":  time.Now(),
		},
	})
}