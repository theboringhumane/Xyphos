package store

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// 🗄️ Store interface defines operations for managing KMS resources
type Store interface {
	KMSStore
	UserStore
}

// 💍 KeyRing represents a collection of cryptographic keys
type KeyRing struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tenant    string    `json:"tenant"`
	CreatedAt time.Time `json:"created_at"`
}

// 🔑 Key represents a cryptographic key
type Key struct {
	ID             string       `json:"id"`
	KeyRing        string       `json:"keyring"`
	Tenant         string       `json:"tenant"`
	Algorithm      string       `json:"algorithm"`
	Purpose        string       `json:"purpose"`
	State          string       `json:"state"`
	CreatedAt      time.Time    `json:"created_at"`
	Versions       []KeyVersion `json:"versions"`
	CurrentVersion int          `json:"current_version"`
}

// 📦 KeyVersion represents a specific version of a key
type KeyVersion struct {
	Version      int       `json:"version"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
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
