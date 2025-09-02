package pages

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	dbmodel "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/gin-gonic/gin"
)

// AdminLoginPageHandler 管理員登入頁面
// AJAX架構：純HTML頁面，認證檢查由前端JavaScript完成
func AdminLoginPageHandler(c *gin.Context) {
	// 純登入頁面，無需後端認證檢查
	// 前端JavaScript會檢查localStorage中的JWT並處理重導向
	data := gin.H{
		"Title":     "管理員登入",
		"BodyClass": "bg-gray-50 min-h-screen flex items-center justify-center",
	}
	c.HTML(http.StatusOK, "admin-login.html", data)
}

// AdminDashboardPageHandler 管理員儀表板頁面
func AdminDashboardPageHandler(c *gin.Context) {
	ctx := context.Background()

	// 獲取系統統計數據
	stats, err := getSystemStats(ctx)
	if err != nil {
		stats = getDefaultStats()
	}

	data := gin.H{
		"Title":       "管理員儀表板",
		"NavTitle":    "Thewavess AI 管理後台",
		"NavIcon":     "fas fa-shield-alt",
		"CurrentPage": "dashboard",
		"Stats":       stats,
	}
	c.HTML(http.StatusOK, "admin-dashboard.html", data)
}

// AdminUsersPageHandler 用戶管理頁面
func AdminUsersPageHandler(c *gin.Context) {
	ctx := context.Background()

	// 解析分頁參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// 獲取用戶列表和統計
	users, totalCount, userStats, err := getUsersWithStats(ctx, page, limit)
	if err != nil {
		users = []dbmodel.UserDB{}
		totalCount = 0
		userStats = getUserStatsDefault()
	}

	totalPages := (totalCount + limit - 1) / limit

	// 序列化用戶數據為 JSON 供 JavaScript 使用
	usersJSON, _ := json.Marshal(users)

	data := gin.H{
		"Title":       "用戶管理",
		"CurrentPage": "users",
		"Users":       users,
		"UsersJSON":   string(usersJSON),
		"UserStats":   userStats,
		"Pagination": gin.H{
			"CurrentPage": page,
			"TotalPages":  totalPages,
			"TotalCount":  totalCount,
			"Limit":       limit,
		},
	}
	c.HTML(http.StatusOK, "admin-users.html", data)
}

// AdminChatHistoryPageHandler 聊天記錄管理頁面
func AdminChatHistoryPageHandler(c *gin.Context) {
	ctx := context.Background()

	// 解析搜尋和分頁參數
	query := c.Query("q")
	userIDFilter := c.Query("user_id")
	characterID := c.Query("character_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

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

	// 獲取聊天會話列表
	chats, totalCount, err := getChatSessions(ctx, query, userIDFilter, characterID, dateFrom, dateTo, page, limit)
	if err != nil {
		chats = []gin.H{}
		totalCount = 0
	}

	totalPages := (totalCount + limit - 1) / limit

	data := gin.H{
		"Title":       "聊天記錄管理",
		"CurrentPage": "chats",
		"Chats":       chats,
		"SearchQuery": query,
		"Filters": gin.H{
			"UserID":      userIDFilter,
			"CharacterID": characterID,
			"DateFrom":    dateFrom,
			"DateTo":      dateTo,
		},
		"Pagination": gin.H{
			"CurrentPage": page,
			"TotalPages":  totalPages,
			"TotalCount":  totalCount,
			"Limit":       limit,
		},
	}
	c.HTML(http.StatusOK, "admin-chats.html", data)
}

// AdminCharactersPageHandler 角色管理頁面
func AdminCharactersPageHandler(c *gin.Context) {
	ctx := context.Background()

	// 解析搜尋和分頁參數
	query := c.Query("q")
	characterType := c.Query("type")
	locale := c.Query("locale")
	isActiveFilter := c.Query("is_active")

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

	// 獲取角色統計
	characterStats, err := getCharacterStats(ctx)
	if err != nil {
		characterStats = getCharacterStatsDefault()
	}

	// 獲取角色列表
	characters, totalCount, err := getCharacterList(ctx, query, characterType, locale, isActiveFilter, page, limit)
	if err != nil {
		characters = []gin.H{}
		totalCount = 0
	}

	totalPages := (totalCount + limit - 1) / limit

	data := gin.H{
		"Title":          "角色管理",
		"CurrentPage":    "characters",
		"Characters":     characters,
		"CharacterStats": characterStats,
		"SearchQuery":    query,
		"Filters": gin.H{
			"Type":     characterType,
			"Locale":   locale,
			"IsActive": isActiveFilter,
		},
		"Pagination": gin.H{
			"CurrentPage": page,
			"TotalPages":  totalPages,
			"TotalCount":  totalCount,
			"Limit":       limit,
		},
	}
	c.HTML(http.StatusOK, "admin-characters.html", data)
}

