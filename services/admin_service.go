package services

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

// AdminService 管理員業務邏輯服務
type AdminService struct {
	db *bun.DB
}

// NewAdminService 創建管理員服務
func NewAdminService() *AdminService {
	return &AdminService{
		db: GetDB(),
	}
}

// CreateAdmin 創建管理員
func (s *AdminService) CreateAdmin(ctx context.Context, req *models.AdminCreateRequest) (*models.Admin, error) {
	utils.Logger.WithFields(logrus.Fields{
		"username": req.Username,
		"role":     req.Role,
	}).Info("創建新管理員")

	// 驗證請求
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("管理員數據驗證失敗: %w", err)
	}

	// 檢查用戶名和郵箱是否已存在
	exists, err := s.checkAdminExists(ctx, req.Username, req.Email)
	if err != nil {
		return nil, fmt.Errorf("檢查管理員是否存在失敗: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("管理員用戶名或郵箱已存在")
	}

	// 加密密碼
	hashedPassword, err := req.HashPassword()
	if err != nil {
		return nil, fmt.Errorf("密碼加密失敗: %w", err)
	}

	// 轉換為領域模型
	admin, err := req.ToAdmin()
	if err != nil {
		return nil, err
	}

	// 設置預設權限
	if len(admin.Permissions) == 0 {
		admin.Permissions = models.GetRolePermissions(admin.Role)
	}

	// 轉換為資料庫模型
	adminDB := admin.ToDB()
	adminDB.Password = hashedPassword

	// 保存到資料庫
	_, err = s.db.NewInsert().
		Model(adminDB).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("創建管理員失敗")
		return nil, fmt.Errorf("保存管理員失敗: %w", err)
	}

	utils.Logger.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"username": admin.Username,
		"role":     admin.Role,
	}).Info("管理員創建成功")

	return admin, nil
}

// AuthenticateAdmin 管理員身份驗證
func (s *AdminService) AuthenticateAdmin(ctx context.Context, req *models.AdminLoginRequest) (*models.Admin, error) {
	utils.Logger.WithField("username", req.Username).Info("管理員登入嘗試")

	// 查詢管理員
	var adminDB db.AdminDB
	err := s.db.NewSelect().
		Model(&adminDB).
		Where("username = ? OR email = ?", req.Username, req.Username).
		Scan(ctx)

	if err != nil {
		// 記錄失敗嘗試但不暴露詳細信息
		utils.Logger.WithFields(logrus.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Warn("管理員登入失敗 - 用戶不存在")
		return nil, fmt.Errorf("用戶名或密碼錯誤")
	}

	// 檢查帳號狀態
	if !adminDB.IsActive() {
		// 記錄失敗嘗試
		s.recordFailedLogin(ctx, &adminDB, req.Username)
		return nil, fmt.Errorf("帳號已被停用或鎖定")
	}

	// 驗證密碼
	if err := models.ValidateAdminPassword(adminDB.Password, req.Password); err != nil {
		// 記錄失敗嘗試
		s.recordFailedLogin(ctx, &adminDB, req.Username)
		return nil, fmt.Errorf("用戶名或密碼錯誤")
	}

	// 登入成功 - 重置失敗計數並更新登入信息
	adminDB.ResetFailedAttempts()
	now := time.Now()
	adminDB.LastLoginAt = &now

	// 更新資料庫
	_, err = s.db.NewUpdate().
		Model(&adminDB).
		Where("id = ?", adminDB.ID).
		Column("failed_attempts", "locked_until", "status", "last_login_at", "updated_at").
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("更新管理員登入信息失敗")
	}

	utils.Logger.WithFields(logrus.Fields{
		"admin_id": adminDB.ID,
		"username": adminDB.Username,
		"role":     adminDB.Role,
	}).Info("管理員登入成功")

	return models.AdminFromDB(&adminDB), nil
}

// GetAdmin 獲取管理員詳情
func (s *AdminService) GetAdmin(ctx context.Context, adminID string) (*models.Admin, error) {
	var adminDB db.AdminDB
	err := s.db.NewSelect().
		Model(&adminDB).
		Where("id = ?", adminID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("管理員不存在: %w", err)
	}

	return models.AdminFromDB(&adminDB), nil
}

// ListAdmins 獲取管理員列表
func (s *AdminService) ListAdmins(ctx context.Context, query *models.AdminListQuery) ([]*models.Admin, *models.PaginationResponse, error) {
	// 構建查詢
	dbQuery := s.db.NewSelect().Model((*db.AdminDB)(nil))

	// 應用過濾條件
	if query.Role != "" {
		dbQuery = dbQuery.Where("role = ?", query.Role)
	}
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("(username ILIKE ? OR email ILIKE ? OR display_name ILIKE ?)",
			searchPattern, searchPattern, searchPattern)
	}

	// 計算總數
	totalCount, err := dbQuery.Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("統計管理員數量失敗: %w", err)
	}

	// 應用排序和分頁
	dbQuery = dbQuery.Order("created_at DESC").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize)

	// 執行查詢
	var adminDBs []db.AdminDB
	err = dbQuery.Scan(ctx, &adminDBs)
	if err != nil {
		return nil, nil, fmt.Errorf("查詢管理員列表失敗: %w", err)
	}

	// 轉換為領域模型
	admins := make([]*models.Admin, len(adminDBs))
	for i, adminDB := range adminDBs {
		admins[i] = models.AdminFromDB(&adminDB)
	}

	// 構建分頁響應
	totalPages := (totalCount + query.PageSize - 1) / query.PageSize
	pagination := &models.PaginationResponse{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalCount: int64(totalCount),
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}

	return admins, pagination, nil
}

// UpdateAdmin 更新管理員信息
func (s *AdminService) UpdateAdmin(ctx context.Context, adminID string, req *models.AdminUpdateRequest) (*models.Admin, error) {
	// 獲取現有管理員
	admin, err := s.GetAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}

	// 應用更新
	adminDB := admin.ToDB()
	updateQuery := s.db.NewUpdate().Model(adminDB).Where("id = ?", adminID)
	hasUpdate := false

	if req.DisplayName != nil {
		adminDB.DisplayName = req.DisplayName
		updateQuery = updateQuery.Column("display_name")
		hasUpdate = true
	}
	if req.Role != nil {
		adminDB.Role = db.AdminRole(*req.Role)
		updateQuery = updateQuery.Column("role")
		hasUpdate = true
	}
	if req.Status != nil {
		adminDB.Status = db.AdminStatus(*req.Status)
		updateQuery = updateQuery.Column("status")
		hasUpdate = true
	}
	if req.Permissions != nil {
		adminDB.Permissions = req.Permissions
		updateQuery = updateQuery.Column("permissions")
		hasUpdate = true
	}
	if req.Notes != nil {
		adminDB.Notes = req.Notes
		updateQuery = updateQuery.Column("notes")
		hasUpdate = true
	}

	if hasUpdate {
		updateQuery = updateQuery.Column("updated_at")
		_, err := updateQuery.Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("更新管理員失敗: %w", err)
		}
	}

	return models.AdminFromDB(adminDB), nil
}

// DeleteAdmin 軟刪除管理員（設為inactive）
func (s *AdminService) DeleteAdmin(ctx context.Context, adminID string) error {
	_, err := s.db.NewUpdate().
		Model((*db.AdminDB)(nil)).
		Where("id = ?", adminID).
		Set("status = ?", db.AdminStatusInactive).
		Set("updated_at = ?", time.Now()).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("刪除管理員失敗: %w", err)
	}

	utils.Logger.WithField("admin_id", adminID).Info("管理員已刪除")
	return nil
}

// 內部輔助方法

// checkAdminExists 檢查管理員是否已存在
func (s *AdminService) checkAdminExists(ctx context.Context, username, email string) (bool, error) {
	count, err := s.db.NewSelect().
		Model((*db.AdminDB)(nil)).
		Where("username = ? OR email = ?", username, email).
		Count(ctx)
	return count > 0, err
}

// recordFailedLogin 記錄失敗的登入嘗試
func (s *AdminService) recordFailedLogin(ctx context.Context, adminDB *db.AdminDB, username string) {
	adminDB.IncrementFailedAttempts()

	// 更新資料庫
	_, err := s.db.NewUpdate().
		Model(adminDB).
		Where("id = ?", adminDB.ID).
		Column("failed_attempts", "locked_until", "status", "updated_at").
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("記錄管理員登入失敗嘗試失敗")
	}

	utils.Logger.WithFields(logrus.Fields{
		"admin_id":        adminDB.ID,
		"username":        username,
		"failed_attempts": adminDB.FailedAttempts,
	}).Warn("管理員登入失敗")
}

// GetAdminService 獲取管理員服務實例
func GetAdminService() *AdminService {
	return NewAdminService()
}
