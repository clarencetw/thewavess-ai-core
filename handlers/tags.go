package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GetAllTags godoc
// @Summary      獲取所有標籤
// @Description  獲取系統中所有可用的標籤列表，支援分類篩選
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        category query string false "標籤分類篩選"
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /tags [get]
func GetAllTags(c *gin.Context) {
	ctx := context.Background()
	category := c.Query("category")
	
	// 查詢所有標籤
	query := database.GetApp().DB().NewSelect().Model((*db.TagDB)(nil))
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	var tagDBs []db.TagDB
	err := query.Order("category", "name").Scan(ctx, &tagDBs)
	if err != nil {
		utils.Logger.WithError(err).Error("查詢標籤失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "查詢標籤失敗",
			},
		})
		return
	}
	
	// 計算每個標籤的使用次數
	tagUsageCounts := make(map[string]int)
	for _, tagDB := range tagDBs {
		count, _ := database.GetApp().DB().NewSelect().
			Model((*db.CharacterTagDB)(nil)).
			Where("tag_id = ?", tagDB.ID).
			Count(ctx)
		tagUsageCounts[tagDB.ID] = count
	}
	
	// 轉換為響應格式
	tags := make([]models.Tag, len(tagDBs))
	categoryMap := make(map[string]int)
	
	for i, tagDB := range tagDBs {
		tags[i] = models.Tag{
			ID:          tagDB.ID,
			Name:        tagDB.Name,
			Category:    tagDB.Category,
			Color:       tagDB.Color,
			Description: tagDB.Description,
			UsageCount:  tagUsageCounts[tagDB.ID],
			CreatedAt:   tagDB.CreatedAt,
		}
		categoryMap[tagDB.Category]++
	}
	
	// 生成分類統計
	categories := make([]models.TagCategory, 0)
	categoryNames := map[string]string{
		"genre":       "類型",
		"personality": "性格",
		"role":        "職業",
		"style":       "風格",
	}
	
	for cat, count := range categoryMap {
		displayName := categoryNames[cat]
		if displayName == "" {
			displayName = cat
		}
		categories = append(categories, models.TagCategory{
			Name:        cat,
			DisplayName: displayName,
			Count:       count,
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取標籤列表成功",
		Data: models.TagsResponse{
			Tags:       tags,
			Categories: categories,
			TotalCount: len(tags),
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
	ctx := context.Background()
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// 查詢標籤及其使用次數
	var result []struct {
		db.TagDB
		UsageCount int `bun:"usage_count"`
	}
	
	err := database.GetApp().DB().NewSelect().
		Model((*db.TagDB)(nil)).
		Column("t.*").
		ColumnExpr("COUNT(ct.tag_id) AS usage_count").
		Join("LEFT JOIN character_tags ct ON ct.tag_id = t.id").
		Group("t.id", "t.name", "t.category", "t.color", "t.description", "t.created_at").
		Order("usage_count DESC").
		Limit(limit).
		Scan(ctx, &result)
		
	if err != nil {
		utils.Logger.WithError(err).Error("查詢熱門標籤失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "查詢熱門標籤失敗",
			},
		})
		return
	}
	
	// 轉換為響應格式，添加趨勢數據
	popularTags := make([]models.TagWithStats, len(result))
	for i, r := range result {
		// 簡單的趨勢模擬（基於使用次數）
		trend := "stable"
		trendPercentage := 0.0
		
		if r.UsageCount >= 3 {
			trend = "up"
			trendPercentage = float64(r.UsageCount*2) + 5.0
		} else if r.UsageCount <= 1 {
			trend = "down"
			trendPercentage = -2.0
		} else {
			trendPercentage = 1.0
		}
		
		popularTags[i] = models.TagWithStats{
			Tag: models.Tag{
				ID:          r.ID,
				Name:        r.Name,
				Category:    r.Category,
				Color:       r.Color,
				Description: r.Description,
				UsageCount:  r.UsageCount,
				CreatedAt:   r.CreatedAt,
			},
			Trend:           trend,
			TrendPercentage: trendPercentage,
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取熱門標籤成功",
		Data: models.PopularTagsResponse{
			Tags:      popularTags,
			Period:    "last_7_days",
			UpdatedAt: time.Now(),
		},
	})
}