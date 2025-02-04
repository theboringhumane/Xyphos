package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dgraph-io/badger/v4"
)

// 🔧 BackupConfig holds the configuration for the backup service
type BackupConfig struct {
	AccountID       string // R2 account ID
	AccessKeyID     string // R2 access key ID
	AccessKeySecret string // R2 access key secret
	BucketName      string // R2 bucket name
	BackupInterval  string // Duration string (e.g., "24h", "1h30m")
	DataDir         string // BadgerDB data directory
}

// 🔄 BackupService handles automated backups to R2
type BackupService struct {
	config   BackupConfig
	db       *badger.DB
	s3Client *s3.Client
	logger   *log.Logger
}

// 🆕 NewBackupService creates a new backup service instance
func NewBackupService(config BackupConfig, db *badger.DB) (*BackupService, error) {
	// 🔐 Create R2 credentials
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", config.AccountID),
		}, nil
	})

	// ⚙️ Configure AWS SDK
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithEndpointResolverWithOptions(r2Resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AccessKeyID,
			config.AccessKeySecret,
			"",
		)),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("❌ unable to configure R2 client: %v", err)
	}

	// 📝 Create logger
	logger := log.New(os.Stdout, "🔄 [BACKUP] ", log.LstdFlags|log.Lmsgprefix)

	return &BackupService{
		config:   config,
		db:       db,
		s3Client: s3.NewFromConfig(cfg),
		logger:   logger,
	}, nil
}

// 💾 createBackup creates a backup file of BadgerDB
func (s *BackupService) createBackup() (string, error) {
	// Create temporary backup file
	backupFile := filepath.Join(os.TempDir(), fmt.Sprintf("xyphos-backup-%s.bak", time.Now().Format("2006-01-02-15-04-05")))
	file, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("❌ unable to create backup file: %v", err)
	}
	defer file.Close()

	// Perform backup
	_, err = s.db.Backup(file, 0)
	if err != nil {
		os.Remove(backupFile)
		return "", fmt.Errorf("❌ unable to backup database: %v", err)
	}

	return backupFile, nil
}

// 📤 uploadToR2 uploads the backup file to R2
func (s *BackupService) uploadToR2(ctx context.Context, backupFile string) error {
	file, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("❌ unable to open backup file: %v", err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("❌ unable to get file info: %v", err)
	}

	// Upload to R2
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(filepath.Base(backupFile)),
		Body:   file,
		Metadata: map[string]string{
			"CreatedAt": time.Now().UTC().Format(time.RFC3339),
			"Size":      fmt.Sprintf("%d", fileInfo.Size()),
		},
	})
	if err != nil {
		return fmt.Errorf("❌ unable to upload to R2: %v", err)
	}

	return nil
}

// 🧹 cleanup removes the temporary backup file
func (s *BackupService) cleanup(backupFile string) {
	if err := os.Remove(backupFile); err != nil {
		s.logger.Printf("⚠️ Warning: unable to remove temporary backup file %s: %v", backupFile, err)
	}
}

// 🔄 PerformBackup executes a complete backup operation
func (s *BackupService) PerformBackup(ctx context.Context) error {
	s.logger.Println("📦 Starting database backup...")

	// Create backup file
	backupFile, err := s.createBackup()
	if err != nil {
		return err
	}
	defer s.cleanup(backupFile)

	s.logger.Printf("💾 Created backup file: %s", backupFile)

	// Upload to R2
	if err := s.uploadToR2(ctx, backupFile); err != nil {
		return err
	}

	s.logger.Println("✅ Backup completed successfully!")
	return nil
}

// 🔄 StartBackupScheduler starts the backup scheduler
func (s *BackupService) StartBackupScheduler(ctx context.Context) {
	s.logger.Printf("🕒 Starting backup scheduler with interval: %s", s.config.BackupInterval)

	ticker := time.NewTicker(parseDuration(s.config.BackupInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Println("🛑 Backup scheduler stopped")
			return
		case <-ticker.C:
			if err := s.PerformBackup(ctx); err != nil {
				s.logger.Printf("❌ Backup failed: %v", err)
			}
		}
	}
}

// ⏰ parseDuration converts backup interval string to duration
func parseDuration(interval string) time.Duration {
	// Default to 24 hours if invalid
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return 24 * time.Hour
	}
	return duration
}
