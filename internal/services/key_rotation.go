package services

import (
	"context"
	"fmt"
	"lambda-kms/internal/hsm"
	"lambda-kms/internal/store"
	"log"
	"sync"
	"time"
)

// 🔄 RotationConfig defines the configuration for key rotation
type RotationConfig struct {
	Period          time.Duration // How often to check for keys that need rotation
	RotationWindow  time.Duration // How long before expiry to rotate keys
	DefaultLifetime time.Duration // Default lifetime for new key versions
}

// 🔄 KeyRotationService handles automatic key rotation
type KeyRotationService struct {
	store    store.KMSStore
	hsm      hsm.Service
	config   RotationConfig
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// 🎯 NewKeyRotationService creates a new key rotation service
func NewKeyRotationService(store store.KMSStore, hsm hsm.Service, config RotationConfig) *KeyRotationService {
	return &KeyRotationService{
		store:    store,
		hsm:      hsm,
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// 🚀 Start begins the key rotation service
func (s *KeyRotationService) Start(ctx context.Context) error {
	log.Println("🔄 Starting key rotation service...")
	s.wg.Add(1)
	go s.rotationLoop(ctx)
	return nil
}

// 🛑 Stop stops the key rotation service
func (s *KeyRotationService) Stop() {
	log.Println("🛑 Stopping key rotation service...")
	close(s.stopChan)
	s.wg.Wait()
}

// 🔄 rotationLoop runs the main rotation check loop
func (s *KeyRotationService) rotationLoop(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.config.Period)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.checkAndRotateKeys(ctx); err != nil {
				log.Printf("❌ Error during key rotation: %v", err)
			}
		}
	}
}

// 🔍 checkAndRotateKeys checks for and rotates keys that need rotation
func (s *KeyRotationService) checkAndRotateKeys(ctx context.Context) error {
	// Get all projects to iterate through their keyrings and keys
	projects, err := s.store.ListAllProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		// Get all keyrings for the project
		keyrings, err := s.store.ListKeyRings(ctx, project.ID)
		if err != nil {
			log.Printf("❌ Error listing keyrings for project %s: %v", project.ID, err)
			continue
		}

		for _, keyring := range keyrings {
			// Get all keys in the keyring
			keys, err := s.store.ListKeys(ctx, keyring.ID, project.ID)
			if err != nil {
				log.Printf("❌ Error listing keys for keyring %s: %v", keyring.ID, err)
				continue
			}

			for _, key := range keys {
				if err := s.checkAndRotateKey(ctx, key); err != nil {
					log.Printf("❌ Error rotating key %s: %v", key.ID, err)
				}
			}
		}
	}

	return nil
}

// 🔄 checkAndRotateKey checks if a key needs rotation and rotates it if necessary
func (s *KeyRotationService) checkAndRotateKey(ctx context.Context, key *store.Key) error {
	if key.State != "ENABLED" {
		return nil // Skip disabled keys
	}

	if len(key.Versions) == 0 {
		return fmt.Errorf("key %s has no versions", key.ID)
	}

	currentVersion := key.Versions[key.CurrentVersion-1]
	timeUntilRotation := currentVersion.RotationTime.Sub(time.Now())

	// Check if it's time to rotate
	if timeUntilRotation > s.config.RotationWindow {
		return nil // Not time to rotate yet
	}

	log.Printf("🔄 Rotating key %s (current version: %d)", key.ID, key.CurrentVersion)

	// Generate new key material
	newKeyMaterial, err := s.hsm.GenerateKey(ctx, key.Algorithm)
	if err != nil {
		return fmt.Errorf("failed to generate new key material: %w", err)
	}

	// Create new version
	newVersion := store.KeyVersion{
		Version:      key.CurrentVersion + 1,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(s.config.DefaultLifetime),
		EncryptedKey: newKeyMaterial,
	}

	// Mark current version as deprecated
	key.Versions[key.CurrentVersion-1].State = "DEPRECATED"

	// Add new version and update current version
	key.Versions = append(key.Versions, newVersion)
	key.CurrentVersion = newVersion.Version

	// Update the key in the store
	if err := s.store.UpdateKey(ctx, key); err != nil {
		return fmt.Errorf("failed to update key with new version: %w", err)
	}

	log.Printf("✅ Successfully rotated key %s to version %d", key.ID, key.CurrentVersion)
	return nil
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

	return s.checkAndRotateKey(ctx, key)
}

// 🔄 GetNextRotationTime returns the time when the key will next be rotated
func (s *KeyRotationService) GetNextRotationTime(key *store.Key) time.Time {
	if len(key.Versions) == 0 || key.State != "ENABLED" {
		return time.Time{} // Zero time for invalid keys
	}

	currentVersion := key.Versions[key.CurrentVersion-1]
	return currentVersion.RotationTime.Add(-s.config.RotationWindow)
}
