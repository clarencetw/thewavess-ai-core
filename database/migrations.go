package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/uptrace/bun/migrate"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations 全局遷移器
var Migrations *migrate.Migrations

// InitMigrations 初始化遷移器
func InitMigrations() error {
	if DB == nil {
		return fmt.Errorf("DB not initialized")
	}

	// 從嵌入的文件系統創建遷移器
	migrationsFS, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migrations filesystem: %w", err)
	}

	Migrations = migrate.NewMigrations()
	if err := Migrations.Discover(migrationsFS); err != nil {
		return fmt.Errorf("failed to discover migrations: %w", err)
	}

	return nil
}

// MigrateUp 執行遷移（向上）
func MigrateUp(ctx context.Context) error {
	if Migrations == nil {
		return fmt.Errorf("migrations not initialized")
	}

	migrator := migrate.NewMigrator(DB, Migrations)
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	if group.IsZero() {
		fmt.Printf("there are no new migrations to run (database is up to date)\n")
		return nil
	}

	fmt.Printf("migrated to %s\n", group)
	return nil
}

// MigrateDown 回滾遷移（向下）
func MigrateDown(ctx context.Context) error {
	if Migrations == nil {
		return fmt.Errorf("migrations not initialized")
	}

	migrator := migrate.NewMigrator(DB, Migrations)
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	group, err := migrator.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("failed to rollback: %w", err)
	}

	if group.IsZero() {
		fmt.Printf("there are no migrations to rollback\n")
		return nil
	}

	fmt.Printf("rolled back %s\n", group)
	return nil
}

// MigrateStatus 檢查遷移狀態
func MigrateStatus(ctx context.Context) error {
	if Migrations == nil {
		return fmt.Errorf("migrations not initialized")
	}

	migrator := migrate.NewMigrator(DB, Migrations)
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	ms, err := migrator.MigrationsWithStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Printf("migrations: %s\n", ms)
	fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
	fmt.Printf("last migration group: %s\n", ms.LastGroup())

	return nil
}

// MigrateReset 重置所有遷移
func MigrateReset(ctx context.Context) error {
	if Migrations == nil {
		return fmt.Errorf("migrations not initialized")
	}

	migrator := migrate.NewMigrator(DB, Migrations)
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	// 回滾所有遷移
	for {
		group, err := migrator.Rollback(ctx)
		if err != nil {
			return fmt.Errorf("failed to rollback: %w", err)
		}
		if group.IsZero() {
			break
		}
		fmt.Printf("rolled back %s\n", group)
	}

	// 重新執行所有遷移
	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	fmt.Printf("reset and migrated to %s\n", group)
	return nil
}