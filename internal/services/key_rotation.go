package services

import (
	"context"
	"fmt"
	"log"
	"time"
	"xyphos/internal/hsm"
	"xyphos/internal/store"
	"xyphos/internal/tasks"
)

// 🔄 RotationConfig defines the configuration for key rotation
type RotationConfig struct {
	Period          time.Duration // How often to check for keys that need rotation
	RotationWindow  time.Duration // How long before expiry to rotate keys
	DefaultLifetime time.Duration // Default lifetime for new key versions
	RedisAddr       string        // Redis address for Asynq
}

// 🔄 KeyRotationService handles automatic key rotation
type KeyRotationService struct {
	store     store.KMSStore
	hsm       hsm.Service
	config    RotationConfig
	taskQueue *tasks.Client
}

// 🎯 NewKeyRotationService creates a new key rotation service
func NewKeyRotationService(store store.KMSStore, hsm hsm.Service, config RotationConfig) (*KeyRotationService, error) {
	taskQueue, err := tasks.NewClient(config.RedisAddr)
	if err != nil {
		return nil, fmt.Errorf("❌ failed to create task client: %w", err)
	}

	return &KeyRotationService{
		store:     store,
		hsm:       hsm,
		config:    config,
		taskQueue: taskQueue,
	}, nil
}

// 🚀 Start begins the key rotation service
func (s *KeyRotationService) Start(ctx context.Context) error {
	log.Println("🔄 Starting key rotation service...")
	return s.scheduleInitialRotations(ctx)
}

// 🛑 Stop stops the key rotation service
func (s *KeyRotationService) Stop() {
	log.Println("🛑 Stopping key rotation service...")
	if err := s.taskQueue.Close(); err != nil {
		log.Printf("❌ Error closing task queue: %v", err)
	}
}

// 🔍 scheduleInitialRotations schedules initial rotation tasks for all keys
func (s *KeyRotationService) scheduleInitialRotations(ctx context.Context) error {
	// Get all projects to iterate through their keyrings and keys
	projects, err := s.store.ListAllProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		// Get all keyrings for the project's tenant
		keyrings, err := s.store.ListKeyRings(ctx, project.OwnerID)
		if err != nil {
			log.Printf("❌ Error listing keyrings for project %s: %v", project.ID, err)
			continue
		}

		for _, keyring := range keyrings {
			// Get all keys in the keyring for the tenant
			keys, err := s.store.ListKeys(ctx, keyring.ID, keyring.Owner)
			if err != nil {
				log.Printf("❌ Error listing keys for keyring %s: %v", keyring.ID, err)
				continue
			}

			for _, key := range keys {
				if err := s.scheduleKeyRotation(key); err != nil {
					log.Printf("❌ Error scheduling rotation for key %s: %v", key.ID, err)
				}
			}
		}
	}

	return nil
}

// 🔄 scheduleKeyRotation schedules rotation for a single key
func (s *KeyRotationService) scheduleKeyRotation(key *store.Key) error {
	if key.Versions[key.CurrentVersion-1].State != "ENABLED" {
		return nil // Skip disabled keys
	}

	if len(key.Versions) == 0 {
		return fmt.Errorf("key %s has no versions", key.ID)
	}

	currentVersion := key.Versions[key.CurrentVersion-1]
	timeUntilRotation := time.Until(currentVersion.RotationTime)

	if timeUntilRotation <= 0 {
		return fmt.Errorf("key %s is already due for rotation", key.ID)
	}

	// Schedule rotation
	return s.taskQueue.SchedulePeriodicKeyRotation(
		key.ID,
		key.KeyRing,
		key.Tenant,
		timeUntilRotation,
	)
}

// 🔄 RotateKeyNow forces immediate rotation of a specific key
func (s *KeyRotationService) RotateKeyNow(ctx context.Context, keyID string) error {
	key, err := s.store.GetKey(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}
	if key == nil {
		return fmt.Errorf("key not found: %s", keyID)
	}

	return s.taskQueue.ScheduleKeyRotation(
		key.ID,
		key.KeyRing,
		key.Tenant,
		time.Now(),
	)
}

// 🔄 GetNextRotationTime returns the time when the key will next be rotated
func (s *KeyRotationService) GetNextRotationTime(key *store.Key) time.Time {
	if len(key.Versions) == 0 || key.Versions[key.CurrentVersion-1].State != "ENABLED" {
		return time.Time{} // Zero time for invalid keys
	}

	currentVersion := key.Versions[key.CurrentVersion-1]
	return currentVersion.RotationTime.Add(-s.config.RotationWindow)
}
