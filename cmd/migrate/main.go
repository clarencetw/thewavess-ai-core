package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

func main() {
	cmd := flag.String("cmd", "up", "Migration command: up, down, status, reset")
	flag.Parse()

	// 初始化 logger
	utils.InitLogger()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		utils.Logger.Warn("No .env file found or error loading .env file")
	} else {
		utils.Logger.Info("Successfully loaded .env file")
	}

	// 初始化數據庫連接
	if err := database.InitDB(); err != nil {
		utils.Logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer database.CloseDB()

	// 初始化遷移器
	if err := database.InitMigrations(); err != nil {
		utils.Logger.WithError(err).Fatal("Failed to initialize migrations")
	}

	ctx := context.Background()

	switch *cmd {
	case "up":
		if err := database.MigrateUp(ctx); err != nil {
			utils.Logger.WithError(err).Fatal("Migration up failed")
		}
		fmt.Println("✅ Migration up completed successfully")

	case "down":
		if err := database.MigrateDown(ctx); err != nil {
			utils.Logger.WithError(err).Fatal("Migration down failed")
		}
		fmt.Println("✅ Migration down completed successfully")

	case "status":
		if err := database.MigrateStatus(ctx); err != nil {
			utils.Logger.WithError(err).Fatal("Failed to check migration status")
		}

	case "reset":
		fmt.Println("⚠️  WARNING: This will reset all migrations and re-run them!")
		fmt.Print("Are you sure? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Operation cancelled")
			os.Exit(0)
		}

		if err := database.MigrateReset(ctx); err != nil {
			utils.Logger.WithError(err).Fatal("Reset failed")
		}
		fmt.Println("✅ Database reset completed successfully")

	default:
		fmt.Printf("Unknown command: %s\n", *cmd)
		fmt.Println("Available commands: up, down, status, reset")
		os.Exit(1)
	}
}