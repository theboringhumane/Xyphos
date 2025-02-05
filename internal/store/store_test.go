package store

import (
	"context"
	"os"
	"testing"
	"time"
)

// 🏗️ setupTestStore creates a temporary store for testing
func setupTestStore(t *testing.T) (*BadgerStore, func()) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "lambda-kms-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create store
	store, err := NewBadgerStore(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create store: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

// 🧪 TestKeyRingOperations tests keyring CRUD operations
func TestKeyRingOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create test keyring
	keyring := &KeyRing{
		ID:        "test-keyring",
		Name:      "Test Keyring",
		Owner:     "test-tenant",
		CreatedAt: time.Now(),
	}

	// Test creation
	if err := store.CreateKeyRing(ctx, keyring); err != nil {
		t.Fatalf("Failed to create keyring: %v", err)
	}

	// Test retrieval
	retrieved, err := store.GetKeyRing(ctx, keyring.ID)
	if err != nil {
		t.Fatalf("Failed to get keyring: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved keyring is nil")
	}
	if retrieved.ID != keyring.ID || retrieved.Name != keyring.Name {
		t.Error("Retrieved keyring does not match original")
	}

	// Test listing
	keyrings, err := store.ListKeyRings(ctx, keyring.Owner)
	if err != nil {
		t.Fatalf("Failed to list keyrings: %v", err)
	}
	if len(keyrings) != 1 {
		t.Errorf("Expected 1 keyring, got %d", len(keyrings))
	}

	// Test deletion
	if err := store.DeleteKeyRing(ctx, keyring.ID); err != nil {
		t.Fatalf("Failed to delete keyring: %v", err)
	}

	// Verify deletion
	retrieved, err = store.GetKeyRing(ctx, keyring.ID)
	if err != nil {
		t.Fatalf("Failed to check keyring after deletion: %v", err)
	}
	if retrieved != nil {
		t.Error("Keyring still exists after deletion")
	}
}

// 🧪 TestKeyOperations tests key CRUD operations
func TestKeyOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create test key with versions
	initialVersion := KeyVersion{
		Version:      1,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(24 * time.Hour),
		EncryptedKey: []byte("test-key-material"),
	}

	key := &Key{
		ID:             "test-key",
		Name:           "Test Key",
		Algorithm:      "AES-256",
		Purpose:        "ENCRYPT_DECRYPT",
		Tenant:         "test-tenant",
		CurrentVersion: 1,
		Versions:       []KeyVersion{initialVersion},
		CreatedAt:      time.Now(),
		RotationPeriod: 24 * time.Hour,
		NextRotation:   time.Now().Add(24 * time.Hour),
	}

	// Test creation
	if err := store.CreateKey(ctx, key); err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	// Test retrieval
	retrieved, err := store.GetKey(ctx, key.ID)
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Retrieved key is nil")
	}
	if retrieved.ID != key.ID || retrieved.Algorithm != key.Algorithm {
		t.Error("Retrieved key does not match original")
	}
	if len(retrieved.Versions) != 1 {
		t.Error("Key version not properly stored")
	}

	// Test key version update
	newVersion := KeyVersion{
		Version:      2,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(24 * time.Hour),
		EncryptedKey: []byte("new-key-material"),
	}
	retrieved.Versions = append(retrieved.Versions, newVersion)
	retrieved.CurrentVersion = 2

	if err := store.UpdateKey(ctx, retrieved); err != nil {
		t.Fatalf("Failed to update key: %v", err)
	}

	// Verify update
	updated, err := store.GetKey(ctx, key.ID)
	if err != nil {
		t.Fatalf("Failed to get updated key: %v", err)
	}
	if updated.CurrentVersion != 2 {
		t.Error("Key version not updated")
	}
	if len(updated.Versions) != 2 {
		t.Error("Key versions not properly updated")
	}

	// Test listing
	keys, err := store.ListKeys(ctx, key.ID, key.Tenant)
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}

	// Test deletion
	if err := store.DeleteKey(ctx, key.ID); err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	// Verify deletion
	retrieved, err = store.GetKey(ctx, key.ID)
	if err != nil {
		t.Fatalf("Failed to check key after deletion: %v", err)
	}
	if retrieved != nil {
		t.Error("Key still exists after deletion")
	}
}

// 🧪 TestKeyVersionManagement tests key version management
func TestKeyVersionManagement(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create initial key with one version
	initialVersion := KeyVersion{
		Version:      1,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(24 * time.Hour),
		EncryptedKey: []byte("initial-key-material"),
	}

	key := &Key{
		ID:             "test-key",
		Name:           "Test Key",
		Algorithm:      "AES-256",
		Purpose:        "ENCRYPT_DECRYPT",
		Tenant:         "test-tenant",
		CurrentVersion: 1,
		Versions:       []KeyVersion{initialVersion},
		CreatedAt:      time.Now(),
		RotationPeriod: 24 * time.Hour,
		NextRotation:   time.Now().Add(24 * time.Hour),
	}

	// Create key
	if err := store.CreateKey(ctx, key); err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	// Add new version
	retrieved, err := store.GetKey(ctx, key.ID)
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	newVersion := KeyVersion{
		Version:      2,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(24 * time.Hour),
		EncryptedKey: []byte("new-key-material"),
	}

	// Mark previous version as deprecated
	retrieved.Versions[0].State = "DEPRECATED"
	// Add new version
	retrieved.Versions = append(retrieved.Versions, newVersion)
	retrieved.CurrentVersion = 2

	// Update key
	if err := store.UpdateKey(ctx, retrieved); err != nil {
		t.Fatalf("Failed to update key: %v", err)
	}

	// Verify versions
	updated, err := store.GetKey(ctx, key.ID)
	if err != nil {
		t.Fatalf("Failed to get updated key: %v", err)
	}

	if len(updated.Versions) != 2 {
		t.Fatalf("Expected 2 versions, got %d", len(updated.Versions))
	}

	if updated.Versions[0].State != "DEPRECATED" {
		t.Error("Old version not marked as deprecated")
	}

	if updated.Versions[1].State != "ENABLED" {
		t.Error("New version not marked as enabled")
	}

	if updated.CurrentVersion != 2 {
		t.Error("Current version not updated")
	}
}
