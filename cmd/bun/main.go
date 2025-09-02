package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/clarencetw/thewavess-ai-core/cmd/bun/migrations"
	"github.com/clarencetw/thewavess-ai-core/database"
	dbmodels "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/migrate"
)

func main() {
	app := &cli.App{
		Name:  "bun-cli",
		Usage: "Thewavess AI Core - CLI Tool for Database Management",
		Commands: []*cli.Command{
			{
				Name:  "db",
				Usage: "Database management commands",
				Subcommands: []*cli.Command{
					{
						Name:   "init",
						Usage:  "Initialize migration tables",
						Action: dbInit,
					},
					{
						Name:   "migrate",
						Usage:  "Run pending migrations",
						Action: dbMigrate,
					},
					{
						Name:   "rollback",
						Usage:  "Rollback last migration group",
						Action: dbRollback,
					},
					{
						Name:   "status",
						Usage:  "Show migration status",
						Action: dbStatus,
					},
					{
						Name:   "reset",
						Usage:  "Reset all migrations and data",
						Action: dbReset,
					},
					{
						Name:  "fixtures",
						Usage: "Load fixtures into database",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "recreate",
								Usage: "Recreate tables before loading fixtures",
							},
						},
						Action: dbFixtures,
					},
				},
			},
			{
				Name:      "create-migration",
				Usage:     "Create new migration files",
				ArgsUsage: "<migration_name>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "type",
						Value: "sql",
						Usage: "Migration type: sql or go",
					},
				},
				Action: createMigration,
			},
		},
		Before: func(c *cli.Context) error {
			// Initialize logger and environment
			utils.InitLogger()
			if err := utils.LoadEnv(); err != nil {
				utils.Logger.Warn("No .env file found or error loading .env file")
			}

			// Only initialize database for commands that need it
			cmdName := c.Args().First()
			if cmdName == "db" || cmdName == "create-migration" {
				app := database.InitDB()
				if app == nil {
					return fmt.Errorf("failed to initialize database")
				}
				// Store app in context for later use
				c.Context = context.WithValue(c.Context, "app", app)
			}

			return nil
		},
		After: func(c *cli.Context) error {
			// Close database connection if it was opened
			if app, ok := c.Context.Value("app").(*database.App); ok && app != nil {
				return app.Close()
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func dbInit(c *cli.Context) error {
	ctx := context.Background()
	app := c.Context.Value("app").(*database.App)
	migrator := migrate.NewMigrator(app.DB(), migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	fmt.Println("✅ Migration tables initialized successfully")
	return nil
}

func dbMigrate(c *cli.Context) error {
	ctx := context.Background()
	app := c.Context.Value("app").(*database.App)
	migrator := migrate.NewMigrator(app.DB(), migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if group.IsZero() {
		fmt.Println("✅ Database is up to date - no new migrations to run")
	} else {
		fmt.Printf("✅ Migration completed successfully: %s\n", group)
	}

	return nil
}

func dbRollback(c *cli.Context) error {
	ctx := context.Background()
	app := c.Context.Value("app").(*database.App)
	migrator := migrate.NewMigrator(app.DB(), migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	group, err := migrator.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if group.IsZero() {
		fmt.Println("✅ No migrations to rollback")
	} else {
		fmt.Printf("✅ Rollback completed successfully: %s\n", group)
	}

	return nil
}

func dbStatus(c *cli.Context) error {
	ctx := context.Background()
	app := c.Context.Value("app").(*database.App)
	migrator := migrate.NewMigrator(app.DB(), migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	ms, err := migrator.MigrationsWithStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Printf("📊 Migration Status:\n")
	fmt.Printf("  All migrations: %s\n", ms)
	fmt.Printf("  Unapplied: %s\n", ms.Unapplied())
	fmt.Printf("  Last group: %s\n", ms.LastGroup())

	return nil
}

func dbReset(c *cli.Context) error {
	fmt.Println("⚠️  WARNING: This will reset all migrations and data!")
	fmt.Print("Are you sure? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Operation cancelled")
		return nil
	}

	ctx := context.Background()
	app := c.Context.Value("app").(*database.App)
	migrator := migrate.NewMigrator(app.DB(), migrations.Migrations)

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
		fmt.Printf("🔄 Rolled back: %s\n", group)
	}

	// 重新執行所有遷移
	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	fmt.Printf("✅ Database reset and migrated to: %s\n", group)
	return nil
}

func dbFixtures(c *cli.Context) error {
	ctx := context.Background()
	recreate := c.Bool("recreate")

	app := c.Context.Value("app").(*database.App)
	db := app.DB()

	// Register all models
	registerModels(db)

	fmt.Println("🌱 Loading fixtures...")

	var options []dbfixture.FixtureOption
	if recreate {
		options = append(options, dbfixture.WithRecreateTables())
		fmt.Println("🔄 Recreating tables...")
	}

	fixture := dbfixture.New(db, options...)

	if err := fixture.Load(ctx, os.DirFS("cmd/bun/fixtures"), "fixtures.yml"); err != nil {
		return fmt.Errorf("failed to load fixtures: %w", err)
	}

	fmt.Println("✅ Fixtures loaded successfully")
	return nil
}

func registerModels(db *bun.DB) {
	// Register all database models (7 tables total)
	db.RegisterModel(
		// Core tables
		(*dbmodels.UserDB)(nil),
		(*dbmodels.CharacterDB)(nil),

		// Chat system
		(*dbmodels.ChatDB)(nil),
		(*dbmodels.MessageDB)(nil),

		// Optimized relationship system (renamed from emotion_states)
		(*dbmodels.RelationshipDB)(nil),

		// Admin system
		(*dbmodels.AdminDB)(nil),

		// Speech styles and scenes removed - functionality integrated into character.user_description
		// Memory system removed - functionality integrated into relationships.emotion_data

	)
}

func createMigration(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("migration name is required")
	}

	name := c.Args().First()
	migType := c.String("type")
	timestamp := time.Now().Format("20060102150405")

	switch migType {
	case "sql":
		upFile := fmt.Sprintf("database/migrations/%s_%s.up.sql", timestamp, name)
		downFile := fmt.Sprintf("database/migrations/%s_%s.down.sql", timestamp, name)

		// Create up migration file
		if err := os.WriteFile(upFile, []byte("-- SQL migration up\n"), 0644); err != nil {
			return fmt.Errorf("failed to create up migration: %w", err)
		}

		// Create down migration file
		if err := os.WriteFile(downFile, []byte("-- SQL migration down\n"), 0644); err != nil {
			return fmt.Errorf("failed to create down migration: %w", err)
		}

		fmt.Printf("✅ Created SQL migration files:\n")
		fmt.Printf("  - %s\n", upFile)
		fmt.Printf("  - %s\n", downFile)

	case "go":
		goFile := fmt.Sprintf("cmd/bun/migrations/%s_%s.go", timestamp, name)

		template := fmt.Sprintf(`package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// TODO: implement migration up
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// TODO: implement migration down
		return nil
	})
}
`)

		if err := os.WriteFile(goFile, []byte(template), 0644); err != nil {
			return fmt.Errorf("failed to create Go migration: %w", err)
		}

		fmt.Printf("✅ Created Go migration file:\n")
		fmt.Printf("  - %s\n", goFile)

	default:
		return fmt.Errorf("unsupported migration type: %s (supported: sql, go)", migType)
	}

	return nil
}

// (Fixtures 系統已取代所有 seed 相關功能)
