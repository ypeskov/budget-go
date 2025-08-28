package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/logger"
)

type BackupService interface {
	CreatePostgresBackup() (*BackupResult, error)
}

type BackupServiceInstance struct {
	cfg *config.Config
}

var (
	backupInstance *BackupServiceInstance
	backupOnce     sync.Once
)

func NewBackupService(cfg *config.Config) BackupService {
	backupOnce.Do(func() {
		logger.Debug("Creating BackupService instance")
		backupInstance = &BackupServiceInstance{cfg: cfg}
	})

	return backupInstance
}

type BackupResult struct {
	Filename string
	FilePath string
}

func (s *BackupServiceInstance) CreatePostgresBackup() (*BackupResult, error) {
	if err := s.ensureBackupDir(); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Clean backup directory if running in container to prevent disk space issues
	if s.cfg.RunningInContainer {
		if err := s.cleanBackupDir(); err != nil {
			logger.Warn("Failed to clean backup directory", "error", err)
			// Don't fail the backup, just log the warning
		}
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("%s-%s_%s.sql", s.cfg.Environment, s.cfg.DbName, timestamp)
	backupPath := filepath.Join(s.cfg.DBBackupDir, filename)

	command := []string{
		"pg_dump",
		"-h", s.cfg.DbHost,
		"-p", s.cfg.DbPort,
		"-U", s.cfg.DbUser,
		"-d", s.cfg.DbName,
		"-F", "p",
		"--no-sync",
		"--verbose",
		"-f", backupPath,
	}

	cmd := exec.Command(command[0], command[1:]...)

	env := os.Environ()
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", s.cfg.DbPassword))
	cmd.Env = env

	logger.Info("Creating database backup", "path", backupPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to create database backup", "output", string(output))
		return nil, fmt.Errorf("pg_dump failed: %w - output: %s", err, string(output))
	}

	logger.Info("Database backup successfully created", "path", backupPath)

	return &BackupResult{
		Filename: filename,
		FilePath: backupPath,
	}, nil
}

func (s *BackupServiceInstance) ensureBackupDir() error {
	if _, err := os.Stat(s.cfg.DBBackupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.cfg.DBBackupDir, 0755); err != nil {
			return err
		}
		logger.Info("Created backup directory", "dir", s.cfg.DBBackupDir)
	}
	return nil
}

func (s *BackupServiceInstance) cleanBackupDir() error {
	logger.Info("Cleaning backup directory (running in container)", "dir", s.cfg.DBBackupDir)

	// Read all files in backup directory
	files, err := os.ReadDir(s.cfg.DBBackupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	deletedCount := 0
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(s.cfg.DBBackupDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				logger.Warn("Failed to delete backup file", "file", file.Name(), "error", err)
			} else {
				deletedCount++
				logger.Debug("Deleted old backup file", "file", file.Name())
			}
		}
	}

	logger.Info("Cleaned old backup files from directory", "count", deletedCount)
	return nil
}
