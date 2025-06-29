package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ntx/internal/database"
)

// BackupService handles automatic database backups
// This service now works directly with the modern database.Manager using SQLC queries
// for improved type safety and performance over the previous repository pattern.
type BackupService struct {
	dbManager     *database.Manager
	backupDir     string
	logger        *slog.Logger
	config        *BackupConfig
	stopChan      chan struct{}
	backupHistory []BackupRecord
}

// BackupConfig holds backup service configuration
type BackupConfig struct {
	// Automatic backup interval (default: 24 hours)
	BackupInterval time.Duration

	// Maximum number of backup files to keep (default: 30)
	MaxBackupFiles int

	// Backup file retention period (default: 30 days)
	RetentionPeriod time.Duration

	// Enable compression for backup files (default: false)
	EnableCompression bool

	// Backup file name pattern (default: "ntx-backup-{timestamp}.db")
	FileNamePattern string

	// Enable automatic cleanup of old backups (default: true)
	AutoCleanup bool
}

// BackupRecord represents a backup operation record
type BackupRecord struct {
	FilePath  string        `json:"file_path"`
	Timestamp time.Time     `json:"timestamp"`
	Size      int64         `json:"size"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Automatic bool          `json:"automatic"`
}

// DefaultBackupConfig returns the default backup configuration
func DefaultBackupConfig() *BackupConfig {
	return &BackupConfig{
		BackupInterval:    24 * time.Hour,
		MaxBackupFiles:    30,
		RetentionPeriod:   30 * 24 * time.Hour,
		EnableCompression: false,
		FileNamePattern:   "ntx-backup-{timestamp}.db",
		AutoCleanup:       true,
	}
}

// NewBackupService creates a new backup service
// This constructor now accepts the modern database.Manager for direct integration
// with Goose migrations and SQLC queries instead of the repository pattern.
func NewBackupService(dbManager *database.Manager, backupDir string, logger *slog.Logger, config *BackupConfig) *BackupService {
	if config == nil {
		config = DefaultBackupConfig()
	}

	return &BackupService{
		dbManager:     dbManager,
		backupDir:     backupDir,
		logger:        logger,
		config:        config,
		stopChan:      make(chan struct{}),
		backupHistory: make([]BackupRecord, 0),
	}
}

// Start begins automatic backup scheduling
func (bs *BackupService) Start(ctx context.Context) error {
	// INFO: Ensure backup directory exists with proper permissions
	if err := os.MkdirAll(bs.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	bs.logger.Info("Starting backup service",
		"backup_dir", bs.backupDir,
		"interval", bs.config.BackupInterval,
		"max_files", bs.config.MaxBackupFiles)

	// Start automatic backup routine
	go bs.automaticBackupLoop(ctx)

	return nil
}

// Stop stops the backup service
func (bs *BackupService) Stop() error {
	bs.logger.Info("Stopping backup service")
	close(bs.stopChan)
	return nil
}

// CreateBackup creates a manual backup
func (bs *BackupService) CreateBackup(ctx context.Context) (*BackupRecord, error) {
	return bs.createBackup(ctx, false)
}

// createBackup performs the actual backup operation
func (bs *BackupService) createBackup(ctx context.Context, automatic bool) (*BackupRecord, error) {
	startTime := time.Now()
	timestamp := startTime.Format("20060102-150405")

	// Generate backup filename
	filename := strings.ReplaceAll(bs.config.FileNamePattern, "{timestamp}", timestamp)
	backupPath := filepath.Join(bs.backupDir, filename)

	record := &BackupRecord{
		FilePath:  backupPath,
		Timestamp: startTime,
		Automatic: automatic,
	}

	bs.logger.Info("Creating database backup",
		"path", backupPath,
		"automatic", automatic)

	// Create the backup
	err := bs.dbManager.Backup(ctx, backupPath)
	record.Duration = time.Since(startTime)

	if err != nil {
		record.Success = false
		record.Error = err.Error()
		bs.logger.Error("Backup failed",
			"path", backupPath,
			"duration", record.Duration,
			"error", err)

		// Clean up failed backup file
		os.Remove(backupPath)

		bs.backupHistory = append(bs.backupHistory, *record)
		return record, fmt.Errorf("backup failed: %w", err)
	}

	// Get backup file size
	if info, err := os.Stat(backupPath); err == nil {
		record.Size = info.Size()
	}

	record.Success = true
	bs.logger.Info("Backup completed successfully",
		"path", backupPath,
		"size", formatSize(record.Size),
		"duration", record.Duration)

	bs.backupHistory = append(bs.backupHistory, *record)

	// Trigger cleanup if enabled
	if bs.config.AutoCleanup {
		if err := bs.cleanupOldBackups(); err != nil {
			bs.logger.Warn("Failed to cleanup old backups", "error", err)
		}
	}

	return record, nil
}

// RestoreBackup restores database from a backup file
func (bs *BackupService) RestoreBackup(ctx context.Context, backupPath string) error {
	bs.logger.Info("Starting database restore", "backup_path", backupPath)

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	// Create a backup of current database before restore
	preRestoreRecord, err := bs.CreateBackup(ctx)
	if err != nil {
		bs.logger.Warn("Failed to create pre-restore backup", "error", err)
	} else {
		bs.logger.Info("Created pre-restore backup", "path", preRestoreRecord.FilePath)
	}

	// Perform restore
	err = bs.dbManager.Restore(ctx, backupPath)
	if err != nil {
		bs.logger.Error("Database restore failed", "backup_path", backupPath, "error", err)
		return fmt.Errorf("restore failed: %w", err)
	}

	bs.logger.Info("Database restore completed successfully", "backup_path", backupPath)
	return nil
}

// ListBackups returns a list of available backup files
func (bs *BackupService) ListBackups() ([]BackupInfo, error) {
	files, err := os.ReadDir(bs.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// INFO: Only consider files that match backup pattern
		if !strings.HasPrefix(file.Name(), "ntx-backup-") || !strings.HasSuffix(file.Name(), ".db") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		backup := BackupInfo{
			FileName:  file.Name(),
			FilePath:  filepath.Join(bs.backupDir, file.Name()),
			Size:      info.Size(),
			Timestamp: info.ModTime(),
		}

		// Parse timestamp from filename if possible
		if ts, err := parseTimestampFromFilename(file.Name()); err == nil {
			backup.Timestamp = ts
		}

		backups = append(backups, backup)
	}

	// Sort by timestamp (newest first)
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[j].Timestamp.After(backups[i].Timestamp) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	return backups, nil
}

// GetBackupHistory returns the backup operation history
func (bs *BackupService) GetBackupHistory() []BackupRecord {
	return bs.backupHistory
}

// GetBackupStats returns backup service statistics
func (bs *BackupService) GetBackupStats() (*BackupStats, error) {
	backups, err := bs.ListBackups()
	if err != nil {
		return nil, err
	}

	stats := &BackupStats{
		TotalBackups:     len(backups),
		TotalSize:        0,
		LastBackupTime:   time.Time{},
		OldestBackupTime: time.Time{},
	}

	if len(backups) > 0 {
		stats.LastBackupTime = backups[0].Timestamp
		stats.OldestBackupTime = backups[len(backups)-1].Timestamp
	}

	for _, backup := range backups {
		stats.TotalSize += backup.Size
	}

	// Count successful/failed operations from history
	for _, record := range bs.backupHistory {
		if record.Success {
			stats.SuccessfulBackups++
		} else {
			stats.FailedBackups++
		}
	}

	return stats, nil
}

// automaticBackupLoop runs the automatic backup scheduling
func (bs *BackupService) automaticBackupLoop(ctx context.Context) {
	ticker := time.NewTicker(bs.config.BackupInterval)
	defer ticker.Stop()

	// TODO: Create initial backup if none exists
	backups, err := bs.ListBackups()
	if err == nil && len(backups) == 0 {
		bs.logger.Info("No existing backups found, creating initial backup")
		if _, err := bs.createBackup(ctx, true); err != nil {
			bs.logger.Error("Failed to create initial backup", "error", err)
		}
	}

	for {
		select {
		case <-ticker.C:
			if _, err := bs.createBackup(ctx, true); err != nil {
				bs.logger.Error("Automatic backup failed", "error", err)
			}

		case <-bs.stopChan:
			bs.logger.Info("Automatic backup loop stopped")
			return

		case <-ctx.Done():
			bs.logger.Info("Automatic backup loop cancelled")
			return
		}
	}
}

// cleanupOldBackups removes old backup files based on retention policy
func (bs *BackupService) cleanupOldBackups() error {
	backups, err := bs.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups for cleanup: %w", err)
	}

	if len(backups) <= bs.config.MaxBackupFiles {
		return nil // Nothing to clean up
	}

	// Remove excess backups (keep the newest ones)
	toRemove := backups[bs.config.MaxBackupFiles:]
	for _, backup := range toRemove {
		if err := os.Remove(backup.FilePath); err != nil {
			bs.logger.Warn("Failed to remove old backup",
				"file", backup.FileName,
				"error", err)
		} else {
			bs.logger.Info("Removed old backup file", "file", backup.FileName)
		}
	}

	// Also remove backups older than retention period
	cutoffTime := time.Now().Add(-bs.config.RetentionPeriod)
	for _, backup := range backups {
		if backup.Timestamp.Before(cutoffTime) {
			if err := os.Remove(backup.FilePath); err != nil {
				bs.logger.Warn("Failed to remove expired backup",
					"file", backup.FileName,
					"error", err)
			} else {
				bs.logger.Info("Removed expired backup file", "file", backup.FileName)
			}
		}
	}

	return nil
}

// BackupInfo represents information about a backup file
type BackupInfo struct {
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
}

// BackupStats provides statistics about the backup system
type BackupStats struct {
	TotalBackups      int       `json:"total_backups"`
	TotalSize         int64     `json:"total_size"`
	LastBackupTime    time.Time `json:"last_backup_time"`
	OldestBackupTime  time.Time `json:"oldest_backup_time"`
	SuccessfulBackups int       `json:"successful_backups"`
	FailedBackups     int       `json:"failed_backups"`
}

// Helper functions

// parseTimestampFromFilename extracts timestamp from backup filename
func parseTimestampFromFilename(filename string) (time.Time, error) {
	// Expected format: ntx-backup-20060102-150405.db
	name := strings.TrimSuffix(filename, ".db")
	parts := strings.Split(name, "-")
	if len(parts) < 4 {
		return time.Time{}, fmt.Errorf("invalid filename format")
	}

	timestampStr := parts[len(parts)-2] + "-" + parts[len(parts)-1]
	return time.Parse("20060102-150405", timestampStr)
}

// formatSize formats byte size in human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
