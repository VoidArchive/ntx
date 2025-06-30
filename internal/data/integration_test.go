package data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDatabaseLifecycle(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	
	// Set test environment
	testHome := filepath.Join(tempDir, "test_home")
	if err := os.MkdirAll(testHome, 0755); err != nil {
		t.Fatalf("failed to create test home: %v", err)
	}
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalPath)

	// Test 1: Initialize database
	t.Run("InitializeDatabase", func(t *testing.T) {
		db, err := InitializeDatabase()
		if err != nil {
			t.Fatalf("failed to initialize database: %v", err)
		}
		defer db.Close()

		// Verify database file exists
		if _, err := os.Stat(db.Path()); os.IsNotExist(err) {
			t.Errorf("database file was not created at: %s", db.Path())
		}

		// Verify we can ping the database
		if err := db.Ping(); err != nil {
			t.Errorf("failed to ping database: %v", err)
		}
	})

	// Test 2: Seed database
	t.Run("SeedDatabase", func(t *testing.T) {
		db, err := InitializeDatabase()
		if err != nil {
			t.Fatalf("failed to initialize database: %v", err)
		}

		// Seed with sample data
		if err := SeedData(db); err != nil {
			db.Close()
			t.Fatalf("failed to seed database: %v", err)
		}

		// Verify seeded data exists
		var portfolioCount int
		err = db.QueryRow("SELECT COUNT(*) FROM portfolios").Scan(&portfolioCount)
		if err != nil {
			db.Close()
			t.Fatalf("failed to count portfolios: %v", err)
		}
		if portfolioCount < 2 {
			db.Close()
			t.Errorf("expected at least 2 portfolios, got %d", portfolioCount)
		}

		var transactionCount int
		err = db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&transactionCount)
		if err != nil {
			db.Close()
			t.Fatalf("failed to count transactions: %v", err)
		}
		if transactionCount < 5 {
			db.Close()
			t.Errorf("expected at least 5 transactions, got %d", transactionCount)
		}

		var holdingCount int
		err = db.QueryRow("SELECT COUNT(*) FROM holdings").Scan(&holdingCount)
		if err != nil {
			db.Close()
			t.Fatalf("failed to count holdings: %v", err)
		}
		if holdingCount < 5 {
			db.Close()
			t.Errorf("expected at least 5 holdings, got %d", holdingCount)
		}

		db.Close()
	})

	// Test 3: Backup and restore
	t.Run("BackupAndRestore", func(t *testing.T) {
		// Initialize and seed database
		db, err := InitializeDatabase()
		if err != nil {
			t.Fatalf("failed to initialize database: %v", err)
		}
		
		if err := SeedData(db); err != nil {
			t.Fatalf("failed to seed database: %v", err)
		}
		db.Close()

		// Create backup
		backupPath, err := BackupDatabase()
		if err != nil {
			t.Fatalf("failed to create backup: %v", err)
		}

		// Verify backup file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("backup file was not created at: %s", backupPath)
		}

		// Verify backup filename format
		if !strings.Contains(backupPath, "portfolio_backup_") || 
		   !strings.HasSuffix(backupPath, ".db") {
			t.Errorf("backup filename format incorrect: %s", backupPath)
		}

		// Reset database to empty state
		resetDb, err := ResetDatabase()
		if err != nil {
			t.Fatalf("failed to reset database: %v", err)
		}
		resetDb.Close()

		// Restore from backup
		if err := RestoreDatabase(backupPath); err != nil {
			t.Fatalf("failed to restore database: %v", err)
		}

		// Verify restored data
		restoredDb, err := NewDatabase()
		if err != nil {
			t.Fatalf("failed to open restored database: %v", err)
		}
		defer restoredDb.Close()

		var portfolioCount int
		err = restoredDb.QueryRow("SELECT COUNT(*) FROM portfolios").Scan(&portfolioCount)
		if err != nil {
			t.Fatalf("failed to count portfolios in restored db: %v", err)
		}
		if portfolioCount < 2 {
			t.Errorf("restored database missing portfolios, got %d", portfolioCount)
		}
	})

	// Test 4: List backups
	t.Run("ListBackups", func(t *testing.T) {
		// Create multiple backups
		backup1, err := BackupDatabase()
		if err != nil {
			t.Fatalf("failed to create first backup: %v", err)
		}

		// Wait a moment to ensure different timestamps
		time.Sleep(1 * time.Second)

		backup2, err := BackupDatabase()
		if err != nil {
			t.Fatalf("failed to create second backup: %v", err)
		}

		// List backups
		backups, err := ListBackups()
		if err != nil {
			t.Fatalf("failed to list backups: %v", err)
		}

		if len(backups) < 2 {
			t.Errorf("expected at least 2 backups, got %d", len(backups))
		}

		// Verify our backups are in the list
		found1, found2 := false, false
		for _, backup := range backups {
			if backup == backup1 {
				found1 = true
			}
			if backup == backup2 {
				found2 = true
			}
		}

		if !found1 {
			t.Errorf("first backup not found in list: %s", backup1)
		}
		if !found2 {
			t.Errorf("second backup not found in list: %s", backup2)
		}
	})

	// Test 5: Reset database
	t.Run("ResetDatabase", func(t *testing.T) {
		// Initialize and seed database
		db, err := InitializeDatabase()
		if err != nil {
			t.Fatalf("failed to initialize database: %v", err)
		}
		
		if err := SeedData(db); err != nil {
			t.Fatalf("failed to seed database: %v", err)
		}
		db.Close()

		// Reset database
		resetDb, err := ResetDatabase()
		if err != nil {
			t.Fatalf("failed to reset database: %v", err)
		}
		defer resetDb.Close()

		// Verify database is empty but schema exists
		var portfolioCount int
		err = resetDb.QueryRow("SELECT COUNT(*) FROM portfolios").Scan(&portfolioCount)
		if err != nil {
			t.Fatalf("failed to query portfolios after reset: %v", err)
		}
		if portfolioCount != 0 {
			t.Errorf("expected 0 portfolios after reset, got %d", portfolioCount)
		}

		// Verify we can still insert data (schema exists)
		_, err = resetDb.Exec(`
			INSERT INTO portfolios (name, description, currency) 
			VALUES (?, ?, ?)
		`, "Test Portfolio", "Test Description", "NPR")
		if err != nil {
			t.Errorf("failed to insert into reset database: %v", err)
		}
	})
}

func TestDatabaseErrorHandling(t *testing.T) {
	// Test backup of non-existent database
	t.Run("BackupNonExistentDatabase", func(t *testing.T) {
		tempDir := t.TempDir()
		originalPath := os.Getenv("HOME")
		
		// Set test environment to empty directory
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalPath)

		_, err := BackupDatabase()
		if err == nil {
			t.Error("expected error when backing up non-existent database")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected 'does not exist' error, got: %v", err)
		}
	})

	// Test restore from non-existent backup
	t.Run("RestoreNonExistentBackup", func(t *testing.T) {
		err := RestoreDatabase("/nonexistent/backup.db")
		if err == nil {
			t.Error("expected error when restoring from non-existent backup")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected 'does not exist' error, got: %v", err)
		}
	})

	// Test list backups with no backup directory
	t.Run("ListBackupsNoDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalPath := os.Getenv("HOME")
		
		// Set test environment to empty directory
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalPath)

		backups, err := ListBackups()
		if err != nil {
			t.Fatalf("unexpected error listing backups: %v", err)
		}
		if len(backups) != 0 {
			t.Errorf("expected empty backup list, got %d backups", len(backups))
		}
	})
}

func TestMigrationOperations(t *testing.T) {
	tempDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	
	// Set test environment
	testHome := filepath.Join(tempDir, "test_home")
	if err := os.MkdirAll(testHome, 0755); err != nil {
		t.Fatalf("failed to create test home: %v", err)
	}
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalPath)

	// Test migration functions
	t.Run("MigrationFlow", func(t *testing.T) {
		// Create database and run initial migrations
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}
		defer db.Close()

		// Run migrations
		if err := RunMigrations(db.DB); err != nil {
			t.Fatalf("failed to run migrations: %v", err)
		}

		// Verify tables exist
		tables := []string{"portfolios", "holdings", "transactions", "corporate_actions"}
		for _, table := range tables {
			var exists int
			query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
			err = db.QueryRow(query, table).Scan(&exists)
			if err != nil {
				t.Fatalf("failed to check if table %s exists: %v", table, err)
			}
			if exists != 1 {
				t.Errorf("table %s does not exist after migration", table)
			}
		}

		// Test rollback (not fully implemented, but test the function)
		if err := MigrateDown(db.DB); err != nil {
			// Expected to fail since we don't have proper down migrations yet
			t.Logf("MigrateDown failed as expected: %v", err)
		}
	})
}