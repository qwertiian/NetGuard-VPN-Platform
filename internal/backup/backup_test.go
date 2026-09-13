package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackupRestoreCycle(t *testing.T) {
	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "db.sqlite")
	backupDir := t.TempDir()
	
	// Create dummy files
	os.WriteFile(dbPath, []byte("dummy db"), 0600)
	os.WriteFile(filepath.Join(dataDir, "server.key"), []byte("privkey"), 0600)
	
	m := NewManager(dataDir, dbPath)
	
	backupPath := filepath.Join(backupDir, "backup.tar.gz")
	err := m.CreateBackup(backupPath)
	assert.NoError(t, err)
	
	info, err := os.Stat(backupPath)
	assert.NoError(t, err)
	assert.True(t, info.Size() > 0)
	
	backups, err := m.ListBackups(backupDir)
	assert.NoError(t, err)
	assert.Len(t, backups, 1)
	assert.Equal(t, backupPath, backups[0].Path)
	
	// Test restore
	restoreDir := t.TempDir()
	m2 := NewManager(restoreDir, filepath.Join(restoreDir, "db.sqlite"))
	
	err = m2.RestoreBackup(backupPath)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(restoreDir, "db.sqlite"))
	assert.NoError(t, err)
	assert.Equal(t, "dummy db", string(content))
	
	content, err = os.ReadFile(filepath.Join(restoreDir, "server.key"))
	assert.NoError(t, err)
	assert.Equal(t, "privkey", string(content))
}
