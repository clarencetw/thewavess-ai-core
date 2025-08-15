package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

func main() {
	cmd := flag.String("cmd", "up", "Migration command: up, down, status, reset")
	flag.Parse()

	// 初始化 logger
	utils.InitLogger()

	// 初始化數據庫連接
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// 初始化遷移器
	if err := database.InitMigrations(); err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}

	ctx := context.Background()

	switch *cmd {
	case "up":
		if err := database.MigrateUp(ctx); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("✅ Migration up completed successfully")

	case "down":
		if err := database.MigrateDown(ctx); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("✅ Migration down completed successfully")

	case "status":
		if err := database.MigrateStatus(ctx); err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
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
			log.Fatalf("Reset failed: %v", err)
		}
		fmt.Println("✅ Database reset completed successfully")

	default:
		fmt.Printf("Unknown command: %s\n", *cmd)
		fmt.Println("Available commands: up, down, status, reset")
		os.Exit(1)
	}
}