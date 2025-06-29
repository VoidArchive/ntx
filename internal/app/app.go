package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"ntx/internal/app/services"
	"ntx/internal/database"
	"ntx/internal/security"
	"ntx/internal/ui/dashboard"

	tea "github.com/charmbracelet/bubbletea"
)

// App represents the main application
// This struct now uses the modern database.Manager with Goose migrations and SQLC queries
// instead of the previous repository pattern for better performance and type safety.
type App struct {
	config        *Config
	logger        *slog.Logger
	program       *tea.Program
	configDir     string
	credentials   *security.Credentials
	dbManager     *database.Manager
	backupService *services.BackupService
}

// Config holds application configuration
type Config struct {
	RefreshInterval string
	StartupView     string
	LogLevel        string
	DataDir         string
	DatabasePath    string
	BackupEnabled   bool
	BackupInterval  string
}

// New creates a new application instance
// This initialization now uses the modern database.Manager with Goose + SQLC
// for improved type safety and performance over the previous repository pattern.
func New() (*App, error) {
	// Determine config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".ntx")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize security (credentials management)
	if err := security.InitializeCredentials(configDir); err != nil {
		return nil, fmt.Errorf("failed to initialize credentials: %w", err)
	}

	// Load credentials
	credentials, err := security.LoadCredentials(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	// Load configuration
	config, err := loadConfig(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Setup logger
	logger := setupLogger(config.LogLevel)

	// Initialize database with new Manager
	dbManager, err := initializeDatabase(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize backup service
	backupService, err := initializeBackupService(dbManager, config, logger)
	if err != nil {
		logger.Warn("Failed to initialize backup service", "error", err)
		// Continue without backup service - not critical for basic functionality
	}

	// Create the dashboard model
	model := dashboard.NewModel()

	// Create Bubbletea program
	program := tea.NewProgram(model, tea.WithAltScreen())

	app := &App{
		config:        config,
		logger:        logger,
		program:       program,
		configDir:     configDir,
		credentials:   credentials,
		dbManager:     dbManager,
		backupService: backupService,
	}

	return app, nil
}

// Start starts the application
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting NTX - NEPSE Power Terminal",
		"version", "0.1.0",
		"config_dir", a.configDir)

	// Start backup service if enabled
	if a.backupService != nil {
		if err := a.backupService.Start(ctx); err != nil {
			a.logger.Error("Failed to start backup service", "error", err)
		} else {
			a.logger.Info("Backup service started")
		}
	}

	// Start the TUI in a goroutine so we can handle context cancellation
	go func() {
		if _, err := a.program.Run(); err != nil {
			a.logger.Error("TUI program failed", "error", err)
		}
	}()

	return nil
}

// Stop stops the application gracefully
func (a *App) Stop() error {
	a.logger.Info("Stopping application...")

	// Stop backup service
	if a.backupService != nil {
		if err := a.backupService.Stop(); err != nil {
			a.logger.Error("Failed to stop backup service", "error", err)
		}
	}

	// Stop TUI program
	if a.program != nil {
		a.program.Quit()
	}

	// Close database connection
	if a.dbManager != nil {
		if err := a.dbManager.Close(); err != nil {
			a.logger.Error("Failed to close database", "error", err)
		}
	}

	a.logger.Info("Application stopped successfully")
	return nil
}

// loadConfig loads the application configuration
func loadConfig(configDir string) (*Config, error) {
	// TODO: Implement proper config loading from TOML file using Viper
	config := &Config{
		RefreshInterval: "60s",
		StartupView:     "portfolio",
		LogLevel:        "info",
		DataDir:         configDir,
		DatabasePath:    filepath.Join(configDir, "data.db"),
		BackupEnabled:   true,
		BackupInterval:  "24h",
	}

	return config, nil
}

// setupLogger configures structured logging
func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// initializeDatabase initializes the database manager and runs migrations
// This now uses the modern database.Manager with Goose migrations and SQLC queries
// replacing the previous custom migration system for industry-standard practices.
func initializeDatabase(config *Config, logger *slog.Logger) (*database.Manager, error) {
	// Create database manager
	dbManager := database.NewManager()

	// Connect to database
	ctx := context.Background()
	if err := dbManager.Connect(ctx, config.DatabasePath); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("Database connected successfully", "path", config.DatabasePath)

	// Run database migrations using Goose
	if err := dbManager.RunMigrations(ctx); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Get schema version for logging
	if version, err := dbManager.GetSchemaVersion(ctx); err == nil {
		logger.Info("Database migrations completed", "schema_version", version)
	}

	return dbManager, nil
}

// initializeBackupService initializes the backup service if enabled
// This service now works with the modern database.Manager for consistent backup operations.
func initializeBackupService(dbManager *database.Manager, config *Config, logger *slog.Logger) (*services.BackupService, error) {
	if !config.BackupEnabled {
		return nil, nil
	}

	backupDir := filepath.Join(config.DataDir, "backups")

	// Parse backup interval
	backupConfig := services.DefaultBackupConfig()
	if config.BackupInterval != "" {
		if interval, err := time.ParseDuration(config.BackupInterval); err == nil {
			backupConfig.BackupInterval = interval
		} else {
			logger.Warn("Invalid backup interval, using default", "interval", config.BackupInterval)
		}
	}

	backupService := services.NewBackupService(dbManager, backupDir, logger, backupConfig)
	return backupService, nil
}

// GetDatabaseManager returns the database manager (for external access)
func (a *App) GetDatabaseManager() *database.Manager {
	return a.dbManager
}

// GetBackupService returns the backup service (for external access)
func (a *App) GetBackupService() *services.BackupService {
	return a.backupService
}

// GetCredentials returns the loaded credentials (for external access)
func (a *App) GetCredentials() *security.Credentials {
	return a.credentials
}
