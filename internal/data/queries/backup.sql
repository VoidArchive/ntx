-- Backup History Operations

-- name: CreateBackupRecord :execresult
INSERT INTO backup_history (backup_path, backup_size, created_at, notes)
VALUES (?, ?, ?, ?);

-- name: GetBackupRecord :one
SELECT id, backup_path, backup_size, created_at, restored_at, notes
FROM backup_history
WHERE id = ?;

-- name: GetAllBackupRecords :many
SELECT id, backup_path, backup_size, created_at, restored_at, notes
FROM backup_history
ORDER BY created_at DESC;

-- name: GetRecentBackups :many
SELECT id, backup_path, backup_size, created_at, restored_at, notes
FROM backup_history
ORDER BY created_at DESC
LIMIT ?;

-- name: UpdateBackupRestore :exec
UPDATE backup_history
SET restored_at = ?
WHERE id = ?;

-- name: DeleteBackupRecord :exec
DELETE FROM backup_history WHERE id = ?;

-- name: GetBackupStats :one
SELECT 
    COUNT(*) as total_backups,
    COALESCE(SUM(backup_size), 0) as total_size,
    MAX(created_at) as latest_backup
FROM backup_history;