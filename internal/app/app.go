package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"ntx/internal/app/services"
	"ntx/internal/database"
	"ntx/internal/security"
	"ntx/internal/ui/dashboard"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

// DynamicWriter allows switching the output destination dynamically
type DynamicWriter struct {
	mu     sync.RWMutex
	writer io.Writer
}

// NewDynamicWriter creates a new dynamic writer starting with stderr
func NewDynamicWriter() *DynamicWriter {
	return &DynamicWriter{
		writer: os.Stderr,
	}
}

// Write implements io.Writer interface
func (dw *DynamicWriter) Write(p []byte) (n int, err error) {
	dw.mu.RLock()
	defer dw.mu.RUnlock()
	return dw.writer.Write(p)
}

// SwitchTo changes the output destination
func (dw *DynamicWriter) SwitchTo(writer io.Writer) {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	dw.writer = writer
}

// Global dynamic writer instance
var globalLogWriter = NewDynamicWriter()

// App represents the main application
// This struct now uses the modern database.Manager with Goose migrations and SQLC queries
// instead of the previous repository pattern for better performance and type safety.
type App struct {
	config           *Config
	logger           *slog.Logger
	program          *tea.Program
	configDir        string
	credentials      *security.Credentials
	dbManager        *database.Manager
	backupService    *services.BackupService
	marketService    *services.MarketService
	portfolioService *services.PortfolioService
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

	// Redirect all logging to file IMMEDIATELY to prevent any stderr output
	// This ensures a completely clean terminal for the TUI experience
	logFile := filepath.Join(configDir, "logs", fmt.Sprintf("ntx-%s.log", time.Now().Format("2006-01-02")))
	if err := redirectLoggingToFileEarly(logFile); err != nil {
		// This is the only error we allow to stderr since logging setup failed
		return nil, fmt.Errorf("failed to redirect logging to file: %w", err)
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
	dbManager, err := initializeDatabase(context.Background(), config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize backup service
	backupService, err := initializeBackupService(dbManager, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize backup service: %w", err)
	}

	// Initialize market service
	marketService, err := initializeMarketService(dbManager, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize market service: %w", err)
	}

	// Initialize portfolio service
	portfolioService := initializePortfolioService(dbManager, logger)

	// Create the dashboard model
	model := dashboard.NewModel(marketService, logger)

	// Create Bubbletea program
	program := tea.NewProgram(model, tea.WithAltScreen())

	app := &App{
		config:           config,
		logger:           logger,
		program:          program,
		configDir:        configDir,
		credentials:      credentials,
		dbManager:        dbManager,
		backupService:    backupService,
		marketService:    marketService,
		portfolioService: portfolioService,
	}

	return app, nil
}

// Start starts the application
func (a *App) Start(ctx context.Context) error {
	// Logging is already redirected to file in New(), so all messages go to file
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

	// Start market service if enabled
	if a.marketService != nil {
		if err := a.marketService.Start(ctx); err != nil {
			a.logger.Error("Failed to start market service", "error", err)
		} else {
			a.logger.Info("Market service started")
		}
	}

	a.logger.Info("Services initialized, starting TUI interface...")

	// Start the TUI with clean terminal - when it exits (q pressed), the entire app should stop
	if _, err := a.program.Run(); err != nil {
		a.logger.Error("TUI program failed", "error", err)
		return err
	}

	a.logger.Info("TUI exited, stopping application...")

	// Restore stderr logging for final shutdown messages
	globalLogWriter.SwitchTo(os.Stderr)
	a.logger.Info("Application shut down complete")

	return nil
}

// redirectLoggingToFileEarly redirects logging to file during app initialization
func redirectLoggingToFileEarly(logPath string) error {
	// Create logs directory
	logsDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Open log file for appending
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Redirect all logging to the file
	globalLogWriter.SwitchTo(logFile)

	return nil
}

// redirectLoggingToFile redirects the global log writer to a file
func (a *App) redirectLoggingToFile(logPath string) error {
	// This is now just a wrapper around the early function
	return redirectLoggingToFileEarly(logPath)
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

	// Stop market service
	if a.marketService != nil {
		if err := a.marketService.Stop(); err != nil {
			a.logger.Error("Failed to stop market service", "error", err)
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
	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.SetDefault("RefreshInterval", "60s")
	viper.SetDefault("StartupView", "portfolio")
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("DataDir", configDir)
	viper.SetDefault("DatabasePath", filepath.Join(configDir, "data.db"))
	viper.SetDefault("BackupEnabled", true)
	viper.SetDefault("BackupInterval", "24h")

	vp := viper.GetViper()
	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := vp.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
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

	// Use the global dynamic writer instead of stderr directly
	handler := slog.NewTextHandler(globalLogWriter, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// initializeDatabase initializes the database manager and runs migrations
// This now uses the modern database.Manager with Goose migrations and SQLC queries
// replacing the previous custom migration system for industry-standard practices.
func initializeDatabase(ctx context.Context, config *Config, logger *slog.Logger) (*database.Manager, error) {
	// Create database manager
	dbManager := database.NewManager()

	// Connect to database
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

// initializeMarketService initializes the market service if enabled
// This service provides real-time NEPSE market data and integrates with the database.
func initializeMarketService(dbManager *database.Manager, config *Config, logger *slog.Logger) (*services.MarketService, error) {
	// Create market service configuration
	marketConfig := services.DefaultMarketConfig()

	// Customize configuration based on app config if needed
	// TODO: Add market-specific configuration to Config struct

	// Create market service
	marketService, err := services.NewMarketService(dbManager, logger, marketConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create market service: %w", err)
	}

	return marketService, nil
}

// initializePortfolioService initializes the portfolio service for real-time P&L calculations
// This service integrates with the database to provide accurate portfolio metrics.
func initializePortfolioService(dbManager *database.Manager, logger *slog.Logger) *services.PortfolioService {
	return services.NewPortfolioService(dbManager, logger)
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

// GetMarketService returns the market service (for external access)
func (a *App) GetMarketService() *services.MarketService {
	return a.marketService
}

// GetPortfolioService returns the portfolio service (for external access)
func (a *App) GetPortfolioService() *services.PortfolioService {
	return a.portfolioService
}
