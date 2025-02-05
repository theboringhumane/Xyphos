package tasks

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// 🎯 Client handles scheduling of Asynq tasks
type Client struct {
	client *asynq.Client
}

// 🏭 NewClient creates a new task client
func NewClient(redisAddr string) (*Client, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &Client{client: client}, nil
}

// 🛑 Close closes the client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// 🔄 ScheduleKeyRotation schedules a key rotation task
func (c *Client) ScheduleKeyRotation(keyID, keyRingID, tenantID string, rotationTime time.Time) error {
	payload, err := NewKeyRotationTask(keyID, keyRingID, tenantID, rotationTime)
	if err != nil {
		return fmt.Errorf("❌ failed to create key rotation task: %w", err)
	}

	task := asynq.NewTask(TypeKeyRotation, payload)
	_, err = c.client.Enqueue(task, asynq.Queue("critical"))
	if err != nil {
		return fmt.Errorf("❌ failed to enqueue key rotation task: %w", err)
	}

	return nil
}

// 💾 ScheduleDBBackup schedules a database backup task
func (c *Client) ScheduleDBBackup(backupPath string) error {
	payload, err := NewDBBackupTask(backupPath)
	if err != nil {
		return fmt.Errorf("❌ failed to create db backup task: %w", err)
	}

	task := asynq.NewTask(TypeDBBackup, payload)
	_, err = c.client.Enqueue(task, asynq.Queue("default"))
	if err != nil {
		return fmt.Errorf("❌ failed to enqueue db backup task: %w", err)
	}

	return nil
}

// 🔄 SchedulePeriodicKeyRotation schedules periodic key rotation
func (c *Client) SchedulePeriodicKeyRotation(keyID, keyRingID, tenantID string, period time.Duration) error {
	payload, err := NewKeyRotationTask(keyID, keyRingID, tenantID, time.Now().Add(period))
	if err != nil {
		return fmt.Errorf("❌ failed to create periodic key rotation task: %w", err)
	}

	task := asynq.NewTask(TypeKeyRotation, payload)
	_, err = c.client.Enqueue(task,
		asynq.Queue("critical"),
		asynq.ProcessIn(period),
		asynq.Unique(24*time.Hour), // Prevent duplicate rotations
	)
	if err != nil {
		return fmt.Errorf("❌ failed to enqueue periodic key rotation task: %w", err)
	}

	return nil
}

// 💾 SchedulePeriodicDBBackup schedules periodic database backup
func (c *Client) SchedulePeriodicDBBackup(backupPath string, period time.Duration) error {
	payload, err := NewDBBackupTask(backupPath)
	if err != nil {
		return fmt.Errorf("❌ failed to create periodic db backup task: %w", err)
	}

	task := asynq.NewTask(TypeDBBackup, payload)
	_, err = c.client.Enqueue(task,
		asynq.Queue("default"),
		asynq.ProcessIn(period),
		asynq.Unique(24*time.Hour), // Prevent duplicate backups
	)
	if err != nil {
		return fmt.Errorf("❌ failed to enqueue periodic db backup task: %w", err)
	}

	return nil
}

// 🔄 ScheduleMasterKeyRotation schedules a master key rotation task
func (c *Client) ScheduleMasterKeyRotation(locationID string) error {
	payload, err := NewMasterKeyRotationTask(locationID)
	if err != nil {
		return fmt.Errorf("❌ failed to create master key rotation task: %w", err)
	}

	task := asynq.NewTask(TypeMasterKeyRotation, payload)
	_, err = c.client.Enqueue(task, asynq.Queue("critical"))
	if err != nil {
		return fmt.Errorf("❌ failed to enqueue master key rotation task: %w", err)
	}

	return nil
}
