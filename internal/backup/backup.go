package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Backup represents a system backup
type Backup struct {
	Timestamp time.Time
	Data      interface{}
	Metadata  map[string]string
}

// Manager handles backup operations
type Manager struct {
	backupDir string
	interval  time.Duration
}

// New creates a new backup manager
func New(backupDir string, interval time.Duration) (*Manager, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &Manager{
		backupDir: backupDir,
		interval:  interval,
	}, nil
}

// CreateBackup creates a new backup
func (m *Manager) CreateBackup(ctx context.Context, data interface{}, metadata map[string]string) (*Backup, error) {
	backup := &Backup{
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  metadata,
	}

	filename := fmt.Sprintf("backup_%s.json", backup.Timestamp.Format("20060102_150405"))
	path := filepath.Join(m.backupDir, filename)

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(backup); err != nil {
		return nil, fmt.Errorf("failed to encode backup: %w", err)
	}

	return backup, nil
}

// RestoreBackup restores data from a backup file
func (m *Manager) RestoreBackup(ctx context.Context, filename string) (*Backup, error) {
	path := filepath.Join(m.backupDir, filename)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var backup Backup
	if err := json.NewDecoder(file).Decode(&backup); err != nil {
		return nil, fmt.Errorf("failed to decode backup: %w", err)
	}

	return &backup, nil
}

// ListBackups returns a list of available backups
func (m *Manager) ListBackups(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			backups = append(backups, entry.Name())
		}
	}

	return backups, nil
}

// DeleteBackup deletes a backup file
func (m *Manager) DeleteBackup(ctx context.Context, filename string) error {
	path := filepath.Join(m.backupDir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}
	return nil
}

// CleanupOldBackups removes backups older than the specified duration
func (m *Manager) CleanupOldBackups(ctx context.Context, maxAge time.Duration) error {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(filepath.Join(m.backupDir, entry.Name())); err != nil {
				return fmt.Errorf("failed to delete old backup: %w", err)
			}
		}
	}

	return nil
}

// StartBackupScheduler starts the automatic backup scheduler
func (m *Manager) StartBackupScheduler(ctx context.Context, dataProvider func() interface{}) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			data := dataProvider()
			metadata := map[string]string{
				"type": "scheduled",
			}
			if _, err := m.CreateBackup(ctx, data, metadata); err != nil {
				// Log error but continue
				fmt.Printf("Failed to create scheduled backup: %v\n", err)
			}
		}
	}
} 