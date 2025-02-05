package tasks

import (
	"encoding/json"
	"time"
)

// 🏷️ Task types
const (
	TypeMasterKeyRotation = "master:rotate"
	TypeKeyRotation       = "key:rotate"
	TypeDBBackup          = "db:backup"
)

// 🔄 KeyRotationPayload represents the payload for key rotation task
type KeyRotationPayload struct {
	KeyID        string    `json:"key_id"`
	KeyRingID    string    `json:"keyring_id"`
	TenantID     string    `json:"tenant_id"`
	RotationTime time.Time `json:"rotation_time"`
}

// 💾 DBBackupPayload represents the payload for database backup task
type DBBackupPayload struct {
	BackupPath string    `json:"backup_path"`
	Timestamp  time.Time `json:"timestamp"`
}

// 🔄 MasterKeyRotationPayload represents the payload for master key rotation task
type MasterKeyRotationPayload struct {
	LocationID string    `json:"location_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// 🏭 NewKeyRotationTask creates a new key rotation task
func NewKeyRotationTask(keyID, keyRingID, tenantID string, rotationTime time.Time) ([]byte, error) {
	return json.Marshal(KeyRotationPayload{
		KeyID:        keyID,
		KeyRingID:    keyRingID,
		TenantID:     tenantID,
		RotationTime: rotationTime,
	})
}

// 🏭 NewDBBackupTask creates a new database backup task
func NewDBBackupTask(backupPath string) ([]byte, error) {
	return json.Marshal(DBBackupPayload{
		BackupPath: backupPath,
		Timestamp:  time.Now(),
	})
}

// 🏭 NewMasterKeyRotationTask creates a new master key rotation task
func NewMasterKeyRotationTask(locationID string) ([]byte, error) {
	return json.Marshal(MasterKeyRotationPayload{
		LocationID: locationID,
		Timestamp:  time.Now(),
	})
}
