package data

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// InitializeDatabase creates a new database connection and runs migrations
func InitializeDatabase() (*Database, error) {
	db, err := NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Run migrations
	if err := RunMigrations(db.DB); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// ResetDatabase drops all tables and re-runs migrations (for development)
func ResetDatabase() (*Database, error) {
	db, err := NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Drop all tables by rolling back all migrations
	for {
		err := MigrateDown(db.DB)
		if err != nil {
			// No more migrations to roll back
			break
		}
	}

	// Re-run all migrations
	if err := RunMigrations(db.DB); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// BackupDatabase creates a backup copy of the database file
func BackupDatabase() (string, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return "", fmt.Errorf("failed to get database path: %w", err)
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", fmt.Errorf("database file does not exist: %s", dbPath)
	}

	// Create backup directory if it doesn't exist
	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate timestamped backup filename
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("portfolio_backup_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	// Copy database file to backup location
	if err := copyFile(dbPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// RestoreDatabase restores the database from a backup file
func RestoreDatabase(backupPath string) error {
	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	dbPath, err := getDatabasePath()
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	// Create database directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Copy backup file to database location
	if err := copyFile(backupPath, dbPath); err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	return nil
}

// ListBackups returns a list of available backup files
func ListBackups() ([]string, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	
	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []string{}, nil // No backups exist
	}

	// Read backup directory
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			backups = append(backups, filepath.Join(backupDir, entry.Name()))
		}
	}

	return backups, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Sync to ensure data is written to disk
	return dstFile.Sync()
}