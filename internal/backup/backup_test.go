package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	interval := 5 * time.Second

	manager, err := New(tempDir, interval)
	if err != nil {
		t.Fatalf("failed to create backup manager: %v", err)
	}

	if manager.backupDir != tempDir {
		t.Errorf("expected backup directory %s, got %s", tempDir, manager.backupDir)
	}

	if manager.interval != interval {
		t.Errorf("expected interval %v, got %v", interval, manager.interval)
	}
}

func TestCreateBackup(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := New(tempDir, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create backup manager: %v", err)
	}

	data := map[string]string{"test": "data"}
	metadata := map[string]string{"type": "test"}

	backup, err := manager.CreateBackup(context.Background(), data, metadata)
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	if backup.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	if backup.Data.(map[string]string)["test"] != "data" {
		t.Error("expected data to be preserved")
	}

	if backup.Metadata["type"] != "test" {
		t.Error("expected metadata to be preserved")
	}

	// Verify file was created
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read backup directory: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 backup file, got %d", len(entries))
	}
}

func TestRestoreBackup(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := New(tempDir, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create backup manager: %v", err)
	}

	// Create a backup
	data := map[string]string{"test": "data"}
	metadata := map[string]string{"type": "test"}
	backup, err := manager.CreateBackup(context.Background(), data, metadata)
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// List backups to get the filename
	backups, err := manager.ListBackups(context.Background())
	if err != nil {
		t.Fatalf("failed to list backups: %v", err)
	}

	if len(backups) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(backups))
	}

	// Restore the backup
	restored, err := manager.RestoreBackup(context.Background(), backups[0])
	if err != nil {
		t.Fatalf("failed to restore backup: %v", err)
	}

	if restored.Timestamp != backup.Timestamp {
		t.Error("expected timestamps to match")
	}

	if restored.Data.(map[string]string)["test"] != "data" {
		t.Error("expected data to be restored")
	}

	if restored.Metadata["type"] != "test" {
		t.Error("expected metadata to be restored")
	}
}

func TestListBackups(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := New(tempDir, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create backup manager: %v", err)
	}

	// Create multiple backups
	for i := 0; i < 3; i++ {
		_, err := manager.CreateBackup(context.Background(), map[string]int{"index": i}, nil)
		if err != nil {
			t.Fatalf("failed to create backup: %v", err)
		}
	}

	backups, err := manager.ListBackups(context.Background())
	if err != nil {
		t.Fatalf("failed to list backups: %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("expected 3 backups, got %d", len(backups))
	}
}

func TestDeleteBackup(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := New(tempDir, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create backup manager: %v", err)
	}

	// Create a backup
	_, err = manager.CreateBackup(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// List backups to get the filename
	backups, err := manager.ListBackups(context.Background())
	if err != nil {
		t.Fatalf("failed to list backups: %v", err)
	}

	if len(backups) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(backups))
	}

	// Delete the backup
	if err := manager.DeleteBackup(context.Background(), backups[0]); err != nil {
		t.Fatalf("failed to delete backup: %v", err)
	}

	// Verify backup was deleted
	backups, err = manager.ListBackups(context.Background())
	if err != nil {
		t.Fatalf("failed to list backups: %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}
}

func TestCleanupOldBackups(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := New(tempDir, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create backup manager: %v", err)
	}

	// Create a backup
	_, err = manager.CreateBackup(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// Clean up backups older than 1 second
	time.Sleep(2 * time.Second)
	if err := manager.CleanupOldBackups(context.Background(), time.Second); err != nil {
		t.Fatalf("failed to cleanup old backups: %v", err)
	}

	// Verify backup was cleaned up
	backups, err := manager.ListBackups(context.Background())
	if err != nil {
		t.Fatalf("failed to list backups: %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}
} 