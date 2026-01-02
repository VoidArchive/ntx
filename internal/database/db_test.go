package database

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func TestAutoMigrate(t *testing.T) {
	db, err := OpenTestDB()
	require.NoError(t, err)
	defer db.Close()

	err = AutoMigrate(db)
	require.NoError(t, err)

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='stocks'").Scan(&tableName)
	require.NoError(t, err)
	require.Equal(t, "stocks", tableName)
}

func TestQueryGeneration(t *testing.T) {
	db, err := OpenTestDB()
	require.NoError(t, err)
	defer db.Close()

	queries := sqlc.New(db)
	require.NotNil(t, queries)
}
