/*
NTX Portfolio Management TUI - Main Entry Point

Bootstrap configuration prioritizes CLI flags > env vars > config file > defaults
to support both casual users and power users with complex setups.

Bubbletea's Model-View-Update pattern chosen for predictable state management
and excellent terminal event handling - critical for financial data accuracy.
*/

package main

import (
	"flag"
	"fmt"
	"ntx/internal/app"
	"ntx/internal/config"
	"ntx/internal/data"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

// Bootstrap handles configuration cascade and error recovery
// Early exit prevents corrupted financial data from invalid configs
func main() {
	// CLI flags override config file to support CI/CD and scripting
	var (
		themeFlag  = flag.String("theme", "", "Theme to use (tokyo_night, rose_pine, gruvbox, default)")
		configFlag = flag.String("config", "", "Path to config file")
		dbCommand  = flag.String("db", "", "Database command (init, migrate, status, reset, seed, backup, restore, list-backups)")
	)
	flag.Parse()

	// Handle database commands
	if *dbCommand != "" {
		handleDatabaseCommand(*dbCommand)
		return
	}

	// Viper cascade ensures consistent config precedence across deployments
	if *themeFlag != "" {
		viper.Set("ui.theme", *themeFlag)
	}
	if *configFlag != "" {
		viper.SetConfigFile(*configFlag)
	}

	// Config validation prevents runtime errors with financial calculations
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Dependency injection pattern enables testing and configuration flexibility
	model := app.NewModelWithConfig(cfg)

	// Alt screen preserves user's terminal history during portfolio sessions
	// Mouse support enables modern interaction patterns for data exploration
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Graceful error handling prevents data corruption during market sessions
	// NOTE: Application state persists across restarts for session continuity
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running NTX Portfolio Management TUI: %v\n", err)
		os.Exit(1)
	}
}

// handleDatabaseCommand processes database management commands
func handleDatabaseCommand(command string) {
	switch command {
	case "init":
		fmt.Println("Initializing database...")
		db, err := data.InitializeDatabase()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		fmt.Printf("Database initialized successfully at: %s\n", db.Path())

	case "migrate":
		fmt.Println("Running migrations...")
		db, err := data.NewDatabase()
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		
		if err := data.RunMigrations(db.DB); err != nil {
			fmt.Printf("Error running migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations completed successfully")

	case "status":
		fmt.Println("Migration status:")
		db, err := data.NewDatabase()
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		
		if err := data.MigrationStatus(db.DB); err != nil {
			fmt.Printf("Error getting migration status: %v\n", err)
			os.Exit(1)
		}

	case "reset":
		fmt.Println("Resetting database...")
		db, err := data.ResetDatabase()
		if err != nil {
			fmt.Printf("Error resetting database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		fmt.Println("Database reset successfully")

	case "seed":
		fmt.Println("Seeding database with sample data...")
		db, err := data.InitializeDatabase()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		
		if err := data.SeedData(db); err != nil {
			fmt.Printf("Error seeding database: %v\n", err)
			os.Exit(1)
		}

	case "backup":
		fmt.Println("Creating database backup...")
		backupPath, err := data.BackupDatabase()
		if err != nil {
			fmt.Printf("Error creating backup: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Database backup created: %s\n", backupPath)

	case "restore":
		if len(os.Args) < 4 {
			fmt.Println("Usage: ntx -db restore <backup-file-path>")
			os.Exit(1)
		}
		backupPath := os.Args[3]
		
		fmt.Printf("Restoring database from: %s\n", backupPath)
		if err := data.RestoreDatabase(backupPath); err != nil {
			fmt.Printf("Error restoring database: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Database restored successfully")

	case "list-backups":
		fmt.Println("Available backups:")
		backups, err := data.ListBackups()
		if err != nil {
			fmt.Printf("Error listing backups: %v\n", err)
			os.Exit(1)
		}
		
		if len(backups) == 0 {
			fmt.Println("  No backups found")
		} else {
			for i, backup := range backups {
				fmt.Printf("  %d. %s\n", i+1, backup)
			}
		}

	default:
		fmt.Printf("Unknown database command: %s\n", command)
		fmt.Println("Available commands: init, migrate, status, reset, seed, backup, restore, list-backups")
		os.Exit(1)
	}
}
