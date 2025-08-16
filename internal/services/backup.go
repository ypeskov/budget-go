package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"ypeskov/budget-go/internal/config"

	log "github.com/sirupsen/logrus"
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
		log.Debug("Creating BackupService instance")
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
			log.Warnf("Failed to clean backup directory: %v", err)
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

	log.Infof("Creating database backup: %s", backupPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to create database backup: %s", string(output))
		return nil, fmt.Errorf("pg_dump failed: %w - output: %s", err, string(output))
	}

	log.Infof("Database backup successfully created: %s", backupPath)

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
		log.Infof("Created backup directory: %s", s.cfg.DBBackupDir)
	}
	return nil
}

func (s *BackupServiceInstance) cleanBackupDir() error {
	log.Infof("Cleaning backup directory: %s (running in container)", s.cfg.DBBackupDir)

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
				log.Warnf("Failed to delete backup file %s: %v", file.Name(), err)
			} else {
				deletedCount++
				log.Debugf("Deleted old backup file: %s", file.Name())
			}
		}
	}

	log.Infof("Cleaned %d old backup files from directory", deletedCount)
	return nil
}
