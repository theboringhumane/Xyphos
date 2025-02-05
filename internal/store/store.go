package store

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// 🗄️ Store interface defines operations for managing KMS resources
type Store interface {
	KMSStore
	// User operations
	UserStore
}

// 🏢 Tenant represents a tenant in a keyring
type Tenant struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	KeyRing     string            `json:"key_ring"`
	Location    string            `json:"location"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Labels      map[string]string `json:"labels"`
}

// 💍 KeyRing represents a collection of cryptographic keys
type KeyRing struct {
	ID          string            `json:"id"`
	ProjectID   string            `json:"project_id"`
	LocationID  string            `json:"location_id"`
	Name        string            `json:"name"`
	Owner       string            `json:"owner"`
	CreatedAt   time.Time         `json:"created_at"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

// 🔑 Key represents a cryptographic key
type Key struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Algorithm      string            `json:"algorithm"`
	Purpose        string            `json:"purpose"`
	CurrentVersion int               `json:"current_version"`
	Versions       []KeyVersion      `json:"versions"`
	CreatedAt      time.Time         `json:"created_at"`
	KeyRing        string            `json:"key_ring"`
	Tenant         string            `json:"tenant"`
	RotationPeriod time.Duration     `json:"rotation_period"`
	NextRotation   time.Time         `json:"next_rotation"`
	Labels         map[string]string `json:"labels"`
}

// 🔐 KeyVersion represents a version of a cryptographic key
type KeyVersion struct {
	Version      int       `json:"version"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	DeprecatedAt time.Time `json:"deprecated_at,omitempty"`
	RotationTime time.Time `json:"rotation_time"`
	EncryptedKey []byte    `json:"encrypted_key"`
}

// 🏪 BadgerStore implements Store using BadgerDB
type BadgerStore struct {
	db *badger.DB
}

// 🎯 NewBadgerStore creates a new BadgerDB-backed store
func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Disable Badger's default logger

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	return &BadgerStore{db: db}, nil
}
