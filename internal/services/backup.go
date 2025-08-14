package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"ypeskov/budget-go/internal/config"
)

type BackupService struct {
	cfg *config.Config
}

func NewBackupService(cfg *config.Config) *BackupService {
	return &BackupService{cfg: cfg}
}

type BackupResult struct {
	Filename string
	FilePath string
}

func (s *BackupService) CreatePostgresBackup() (*BackupResult, error) {
	if err := s.ensureBackupDir(); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
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

func (s *BackupService) ensureBackupDir() error {
	if _, err := os.Stat(s.cfg.DBBackupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.cfg.DBBackupDir, 0755); err != nil {
			return err
		}
		log.Infof("Created backup directory: %s", s.cfg.DBBackupDir)
	}
	return nil
}