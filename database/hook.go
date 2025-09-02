package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"golang.org/x/sync/errgroup"
)

// Hook 表示應用程式生命週期鉤子
type Hook func(ctx context.Context, app *App) error

// HookManager 管理應用程式鉤子
type HookManager struct {
	mu    sync.RWMutex
	hooks map[string]Hook
}

// NewHookManager 創建新的鉤子管理器
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[string]Hook),
	}
}

// OnStart 註冊啟動鉤子
func (hm *HookManager) OnStart(name string, hook Hook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.hooks == nil {
		hm.hooks = make(map[string]Hook)
	}

	hm.hooks[name] = hook
	utils.Logger.WithField("hook", name).Info("Hook registered")
}

// RunHooks 並發執行所有鉤子
func (hm *HookManager) RunHooks(ctx context.Context, app *App) error {
	hm.mu.RLock()
	hooks := make(map[string]Hook, len(hm.hooks))
	for name, hook := range hm.hooks {
		hooks[name] = hook
	}
	hm.mu.RUnlock()

	if len(hooks) == 0 {
		utils.Logger.Info("No hooks to execute")
		return nil
	}

	utils.Logger.WithField("count", len(hooks)).Info("Starting to execute hooks")

	// 使用 errgroup 並發執行鉤子
	g, ctx := errgroup.WithContext(ctx)

	for name, hook := range hooks {
		name, hook := name, hook // 捕獲循環變數
		g.Go(func() error {
			return hm.runSingleHook(ctx, app, name, hook)
		})
	}

	// 等待所有鉤子完成
	if err := g.Wait(); err != nil {
		utils.Logger.WithError(err).Error("One or more hooks failed")
		return fmt.Errorf("hook execution failed: %w", err)
	}

	utils.Logger.Info("All hooks executed successfully")
	return nil
}

// runSingleHook 執行單個鉤子並記錄性能
func (hm *HookManager) runSingleHook(ctx context.Context, app *App, name string, hook Hook) error {
	start := time.Now()

	// 為每個鉤子設置 30 秒超時
	hookCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	logger := utils.Logger.WithField("hook", name)
	logger.Info("Starting hook execution")

	// 執行鉤子
	err := hook(hookCtx, app)
	duration := time.Since(start)

	if err != nil {
		logger.WithError(err).WithField("duration", duration).Error("Hook execution failed")
		return fmt.Errorf("hook %s failed: %w", name, err)
	}

	// 記錄執行時間（超過 1 秒的會特別標注）
	if duration > time.Second {
		logger.WithField("duration", duration).Warn("Hook execution took longer than 1 second")
	} else {
		logger.WithField("duration", duration).Info("Hook execution completed")
	}

	return nil
}

// GetRegisteredHooks 獲取已註冊的鉤子名稱列表
func (hm *HookManager) GetRegisteredHooks() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	names := make([]string, 0, len(hm.hooks))
	for name := range hm.hooks {
		names = append(names, name)
	}
	return names
}

// RemoveHook 移除指定的鉤子
func (hm *HookManager) RemoveHook(name string) bool {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if _, exists := hm.hooks[name]; exists {
		delete(hm.hooks, name)
		utils.Logger.WithField("hook", name).Info("Hook removed")
		return true
	}
	return false
}

// ClearHooks 清除所有鉤子
func (hm *HookManager) ClearHooks() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	count := len(hm.hooks)
	hm.hooks = make(map[string]Hook)
	utils.Logger.WithField("count", count).Info("All hooks cleared")
}
