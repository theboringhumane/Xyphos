package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

// 🔐 LocationMasterKeyManager manages location-specific master keys
type LocationMasterKeyManager struct {
	mu        sync.RWMutex
	keys      map[string][]byte // location -> master key
	storePath string
	wrapKey   []byte
}

// 🆕 NewLocationMasterKeyManager creates a new location master key manager
func NewLocationMasterKeyManager(storePath string, wrapKeyHex string) (*LocationMasterKeyManager, error) {
	wrapKey, err := hex.DecodeString(wrapKeyHex)
	if err != nil {
		return nil, fmt.Errorf("❌ invalid wrap key format: %w", err)
	}

	// Create store path if it doesn't exist
	if err := os.MkdirAll(storePath, 0700); err != nil {
		return nil, fmt.Errorf("❌ failed to create store path: %w", err)
	}

	return &LocationMasterKeyManager{
		keys:      make(map[string][]byte),
		storePath: storePath,
		wrapKey:   wrapKey,
	}, nil
}

// 🎲 GenerateLocationMasterKey generates a new master key for a location
func (m *LocationMasterKeyManager) GenerateLocationMasterKey(location string) ([]byte, error) {
	fmt.Printf("🔒 Starting master key generation for location: %s\n", location)
	// Since this is called from GetLocationMasterKey which already has the write lock,
	// we don't need to acquire locks here

	// 🎲 Generate random master key
	fmt.Println("🎲 Generating random master key")
	masterKey := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(masterKey); err != nil {
		fmt.Printf("❌ Failed to generate master key: %v\n", err)
		return nil, fmt.Errorf("❌ failed to generate master key: %w", err)
	}
	fmt.Println("✅ Generated master key successfully")

	// 🔒 Encrypt master key using XChaCha20-Poly1305
	fmt.Println("🔒 Creating AEAD cipher")
	aead, err := chacha20poly1305.NewX(m.wrapKey)
	if err != nil {
		fmt.Printf("❌ Failed to create AEAD: %v\n", err)
		return nil, fmt.Errorf("❌ failed to create AEAD: %w", err)
	}
	fmt.Println("✅ Created AEAD cipher successfully")

	// 🎲 Generate nonce
	fmt.Println("🎲 Generating nonce")
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		fmt.Printf("❌ Failed to generate nonce: %v\n", err)
		return nil, fmt.Errorf("❌ failed to generate nonce: %w", err)
	}
	fmt.Println("✅ Generated nonce successfully")

	// 🔐 Encrypt master key
	fmt.Println("🔐 Encrypting master key")
	encryptedKey := aead.Seal(nonce, nonce, masterKey, nil)
	fmt.Println("✅ Encrypted master key successfully")

	// 💾 Save encrypted key to location-specific file
	fmt.Println("💾 Saving encrypted key to file")
	keyPath := filepath.Join(m.storePath, fmt.Sprintf("%s.key", location))
	if err := os.WriteFile(keyPath, encryptedKey, 0600); err != nil {
		fmt.Printf("❌ Failed to save encrypted master key: %v\n", err)
		return nil, fmt.Errorf("❌ failed to save encrypted master key: %w", err)
	}
	fmt.Printf("✅ Saved encrypted key to: %s\n", keyPath)

	// 🗄️ Cache the master key
	fmt.Println("🗄️ Caching master key in memory")
	m.keys[location] = masterKey
	fmt.Println("✅ Cached master key successfully")

	fmt.Println("✅ Master key generation completed successfully")
	return masterKey, nil
}

// 🔓 GetLocationMasterKey gets or loads the master key for a location
func (m *LocationMasterKeyManager) GetLocationMasterKey(location string) ([]byte, error) {
	fmt.Printf("🔍 Attempting to get master key for location: %s\n", location)

	m.mu.RLock()
	if key, exists := m.keys[location]; exists {
		fmt.Println("✅ Found cached master key")
		m.mu.RUnlock()
		return key, nil
	}
	m.mu.RUnlock()

	fmt.Println("🔄 Key not in cache, upgrading to write lock")
	// 🔄 Upgrade to write lock for loading/generating
	m.mu.Lock()
	defer m.mu.Unlock()

	// 📖 Check again in case another goroutine loaded it
	if key, exists := m.keys[location]; exists {
		fmt.Println("✅ Found cached master key after lock upgrade")
		return key, nil
	}

	fmt.Println("📂 Attempting to load existing key from file")
	// 📂 Try to load existing key
	keyPath := filepath.Join(m.storePath, fmt.Sprintf("%s.key", location))
	encryptedKey, err := os.ReadFile(keyPath)
	if err != nil {
		// 🆕 Generate new key if file doesn't exist
		if os.IsNotExist(err) {
			fmt.Println("🆕 Key file not found, generating new master key")
			return m.GenerateLocationMasterKey(location)
		}
		fmt.Printf("❌ Failed to read master key file: %v\n", err)
		return nil, fmt.Errorf("❌ failed to read master key: %w", err)
	}

	fmt.Println("🔓 Decrypting master key")
	// 🔓 Decrypt the key
	aead, err := chacha20poly1305.NewX(m.wrapKey)
	if err != nil {
		fmt.Printf("❌ Failed to create AEAD: %v\n", err)
		return nil, fmt.Errorf("❌ failed to create AEAD: %w", err)
	}

	if len(encryptedKey) < aead.NonceSize() {
		fmt.Println("❌ Encrypted key data is too short")
		return nil, fmt.Errorf("❌ encrypted key too short")
	}

	nonce := encryptedKey[:aead.NonceSize()]
	ciphertext := encryptedKey[aead.NonceSize():]

	masterKey, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Printf("❌ Failed to decrypt master key: %v\n", err)
		return nil, fmt.Errorf("❌ failed to decrypt master key: %w", err)
	}

	fmt.Println("🗄️ Caching decrypted master key")
	// 🗄️ Cache the decrypted key
	m.keys[location] = masterKey

	fmt.Println("✅ Successfully retrieved master key")
	return masterKey, nil
}

// 🔄 RotateLocationMasterKey rotates the master key for a location
func (m *LocationMasterKeyManager) RotateLocationMasterKey(location string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 🔍 Get the old master key first
	oldMasterKey, exists := m.keys[location]
	if !exists {
		// If no old key exists, just generate a new one
		return m.GenerateLocationMasterKey(location)
	}

	// 🎲 Generate new master key
	newMasterKey := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(newMasterKey); err != nil {
		return nil, fmt.Errorf("❌ failed to generate new master key: %w", err)
	}

	// 🔒 Encrypt new master key using XChaCha20-Poly1305
	aead, err := chacha20poly1305.NewX(m.wrapKey)
	if err != nil {
		return nil, fmt.Errorf("❌ failed to create AEAD: %w", err)
	}

	// 🎲 Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("❌ failed to generate nonce: %w", err)
	}

	// 🔐 Encrypt new master key
	encryptedKey := aead.Seal(nonce, nonce, newMasterKey, nil)

	// Create backup of old key file
	oldKeyPath := filepath.Join(m.storePath, fmt.Sprintf("%s.key", location))
	backupPath := filepath.Join(m.storePath, fmt.Sprintf("%s.key.bak", location))
	if err := os.Rename(oldKeyPath, backupPath); err != nil {
		return nil, fmt.Errorf("❌ failed to backup old master key: %w", err)
	}

	// 💾 Save new encrypted key to location-specific file
	if err := os.WriteFile(oldKeyPath, encryptedKey, 0600); err != nil {
		// Restore backup if saving new key fails
		os.Rename(backupPath, oldKeyPath)
		return nil, fmt.Errorf("❌ failed to save new master key: %w", err)
	}

	// Store both old and new keys in memory temporarily
	m.keys[location+"_old"] = oldMasterKey
	m.keys[location] = newMasterKey

	// After successful rotation, schedule cleanup of old key
	go func() {
		time.Sleep(24 * time.Hour) // Keep old key for 24 hours
		m.mu.Lock()
		delete(m.keys, location+"_old")
		os.Remove(backupPath)
		m.mu.Unlock()
	}()

	return newMasterKey, nil
}

// 🔍 GetOldLocationMasterKey gets the previous master key for a location during rotation
func (m *LocationMasterKeyManager) GetOldLocationMasterKey(location string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key, exists := m.keys[location+"_old"]
	return key, exists
}
