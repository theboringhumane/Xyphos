package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"xyphos/internal/backup"
	"xyphos/internal/crypto"
	"xyphos/internal/hsm"
	"xyphos/internal/store"

	"github.com/hibiken/asynq"
)

// 🎯 Processor handles processing of Asynq tasks
type Processor struct {
	store            store.KMSStore
	hsm              hsm.Service
	mux              *asynq.ServeMux
	server           *asynq.Server
	client           *asynq.Client
	backupService    *backup.BackupService
	masterKeyManager *crypto.LocationMasterKeyManager
}

// 🏭 NewProcessor creates a new task processor
func NewProcessor(redisAddr string, store store.KMSStore, hsm hsm.Service, masterKeyManager *crypto.LocationMasterKeyManager) (*Processor, error) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6, // For key rotations
				"default":  3, // For DB backups
				"low":      1,
			},
		},
	)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	processor := &Processor{
		store:            store,
		hsm:              hsm,
		server:           srv,
		client:           client,
		mux:              asynq.NewServeMux(),
		masterKeyManager: masterKeyManager,
	}

	// Register task handlers
	processor.mux.HandleFunc(TypeKeyRotation, processor.handleKeyRotation)
	processor.mux.HandleFunc(TypeDBBackup, processor.handleDBBackup)
	processor.mux.HandleFunc(TypeMasterKeyRotation, processor.handleMasterKeyRotation)

	return processor, nil
}

// 🚀 Start starts the task processor
func (p *Processor) Start() error {
	return p.server.Start(p.mux)
}

// 🛑 Shutdown gracefully shuts down the processor
func (p *Processor) Shutdown() {
	p.server.Shutdown()
	p.client.Close()
}

// 🔄 handleKeyRotation processes key rotation tasks
func (p *Processor) handleKeyRotation(ctx context.Context, t *asynq.Task) error {
	var payload KeyRotationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("❌ failed to unmarshal key rotation payload: %w", err)
	}

	// Get the key
	key, err := p.store.GetKey(ctx, payload.KeyID)
	if err != nil {
		return fmt.Errorf("❌ failed to get key: %w", err)
	}
	if key == nil {
		return fmt.Errorf("❌ key not found: %s", payload.KeyID)
	}

	// Generate new key material
	keyMaterial, err := p.hsm.GenerateKeyMaterial(key.Algorithm)
	if err != nil {
		return fmt.Errorf("❌ failed to generate key material: %w", err)
	}

	// Create new version
	newVersion := store.KeyVersion{
		Version:      key.CurrentVersion + 1,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(24 * time.Hour), // Default 24h rotation period
		EncryptedKey: keyMaterial,
	}

	// Update previous version state to DEPRECATED
	for i := range key.Versions {
		if key.Versions[i].Version == key.CurrentVersion {
			key.Versions[i].State = "DEPRECATED"
			key.Versions[i].DeprecatedAt = time.Now()
			break
		}
	}

	// Add new version and update current version
	key.Versions = append(key.Versions, newVersion)
	key.CurrentVersion = newVersion.Version
	key.NextRotation = newVersion.RotationTime

	// Save updated key
	if err := p.store.UpdateKey(ctx, key); err != nil {
		return fmt.Errorf("❌ failed to update key: %w", err)
	}

	log.Printf("✅ Successfully rotated key %s from version %d to %d",
		key.ID, key.CurrentVersion-1, newVersion.Version)

	// Schedule next rotation
	nextRotationTask := asynq.NewTask(
		TypeKeyRotation,
		mustMarshal(KeyRotationPayload{
			KeyID: key.ID,
		}),
		asynq.ProcessAt(key.NextRotation),
		asynq.Queue("critical"),
	)

	if _, err := p.client.EnqueueContext(ctx, nextRotationTask); err != nil {
		log.Printf("⚠️ Failed to schedule next rotation for key %s: %v", key.ID, err)
	}

	return nil
}

// 💾 handleDBBackup processes database backup tasks
func (p *Processor) handleDBBackup(ctx context.Context, t *asynq.Task) error {
	var payload DBBackupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("❌ failed to unmarshal db backup payload: %w", err)
	}

	// Perform backup using BackupService from service.go 📦🔒
	if err := p.backupService.PerformBackup(ctx); err != nil {
		return fmt.Errorf("❌ failed to backup database: %w", err)
	}

	log.Printf("✅ Successfully backed up database to %s", payload.BackupPath)
	return nil
}

// 🔄 handleMasterKeyRotation processes master key rotation tasks
func (p *Processor) handleMasterKeyRotation(ctx context.Context, t *asynq.Task) error {
	var payload MasterKeyRotationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("❌ failed to unmarshal master key rotation payload: %w", err)
	}

	// Get the location
	location, err := p.store.GetLocation(ctx, payload.LocationID)
	if err != nil {
		return fmt.Errorf("❌ failed to get location: %w", err)
	}
	if location == nil {
		return fmt.Errorf("❌ location not found: %s", payload.LocationID)
	}

	// First rotate the HSM's master key
	if err := p.hsm.RotateMasterKey(ctx, payload.LocationID); err != nil {
		return fmt.Errorf("❌ failed to rotate HSM master key: %w", err)
	}

	// Then rotate the location's master key
	newMasterKey, err := p.masterKeyManager.RotateLocationMasterKey(payload.LocationID)
	if err != nil {
		return fmt.Errorf("❌ failed to rotate location master key: %w", err)
	}

	// Get all keys in the location that need re-encryption
	keyrings, err := p.store.ListKeyRings(ctx, location.ID)
	if err != nil {
		return fmt.Errorf("❌ failed to list keyrings: %w", err)
	}

	// Re-encrypt all keys with the new master key
	for _, keyring := range keyrings {
		keys, err := p.store.ListKeys(ctx, keyring.ID, keyring.Owner)
		if err != nil {
			log.Printf("❌ Error listing keys for keyring %s: %v", keyring.ID, err)
			continue
		}

		for _, key := range keys {
			// Get the old master key for decryption
			oldMasterKey, exists := p.masterKeyManager.GetOldLocationMasterKey(payload.LocationID)
			if !exists {
				log.Printf("⚠️ Old master key not found for key %s, skipping re-encryption", key.ID)
				continue
			}

			// Get current version's key material
			currentVersion := key.Versions[key.CurrentVersion-1]
			keyMaterial, err := p.hsm.UnwrapKey(oldMasterKey, currentVersion.EncryptedKey)
			if err != nil {
				log.Printf("❌ Error unwrapping key %s version %d: %v", key.ID, currentVersion.Version, err)
				continue
			}

			// Create new version with the same key material but wrapped with new master key
			newWrappedKey, err := p.hsm.WrapKey(newMasterKey, keyMaterial)
			if err != nil {
				log.Printf("❌ Error wrapping key %s with new master key: %v", key.ID, err)
				continue
			}

			// Mark all existing versions as DEPRECATED
			for i := range key.Versions {
				if key.Versions[i].State == "ENABLED" {
					key.Versions[i].State = "DEPRECATED"
					key.Versions[i].DeprecatedAt = time.Now()
				}
			}

			// Create new version
			newVersion := store.KeyVersion{
				Version:      key.CurrentVersion + 1,
				State:        "ENABLED",
				CreatedAt:    time.Now(),
				RotationTime: time.Now().Add(24 * time.Hour), // Default 24h rotation period
				EncryptedKey: newWrappedKey,
			}

			// Add new version and update current version
			key.Versions = append(key.Versions, newVersion)
			key.CurrentVersion = newVersion.Version
			key.NextRotation = newVersion.RotationTime

			// Save the updated key
			if err := p.store.UpdateKey(ctx, key); err != nil {
				log.Printf("❌ Error updating key %s: %v", key.ID, err)
				continue
			}

			log.Printf("✅ Successfully created new version %d for key %s with new master key", newVersion.Version, key.ID)
		}
	}

	log.Printf("✅ Successfully rotated master key for location %s", payload.LocationID)
	return nil
}

// 🔨 mustMarshal marshals data to JSON and panics on error
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal data: %v", err))
	}
	return data
}
