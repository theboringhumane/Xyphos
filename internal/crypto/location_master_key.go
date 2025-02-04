package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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
	m.mu.Lock()
	defer m.mu.Unlock()

	// 🎲 Generate random master key
	masterKey := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(masterKey); err != nil {
		return nil, fmt.Errorf("❌ failed to generate master key: %w", err)
	}

	// 🔒 Encrypt master key using XChaCha20-Poly1305
	aead, err := chacha20poly1305.NewX(m.wrapKey)
	if err != nil {
		return nil, fmt.Errorf("❌ failed to create AEAD: %w", err)
	}

	// 🎲 Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("❌ failed to generate nonce: %w", err)
	}

	// 🔐 Encrypt master key
	encryptedKey := aead.Seal(nonce, nonce, masterKey, nil)

	// 💾 Save encrypted key to location-specific file
	keyPath := filepath.Join(m.storePath, fmt.Sprintf("%s.key", location))
	if err := os.WriteFile(keyPath, encryptedKey, 0600); err != nil {
		return nil, fmt.Errorf("❌ failed to save encrypted master key: %w", err)
	}

	// 🗄️ Cache the master key
	m.keys[location] = masterKey

	return masterKey, nil
}

// 🔓 GetLocationMasterKey gets or loads the master key for a location
func (m *LocationMasterKeyManager) GetLocationMasterKey(location string) ([]byte, error) {
	m.mu.RLock()
	if key, exists := m.keys[location]; exists {
		m.mu.RUnlock()
		return key, nil
	}
	m.mu.RUnlock()

	// 🔄 Upgrade to write lock for loading/generating
	m.mu.Lock()
	defer m.mu.Unlock()

	// 📖 Check again in case another goroutine loaded it
	if key, exists := m.keys[location]; exists {
		return key, nil
	}

	// 📂 Try to load existing key
	keyPath := filepath.Join(m.storePath, fmt.Sprintf("%s.key", location))
	encryptedKey, err := os.ReadFile(keyPath)
	if err != nil {
		// 🆕 Generate new key if file doesn't exist
		if os.IsNotExist(err) {
			return m.GenerateLocationMasterKey(location)
		}
		return nil, fmt.Errorf("❌ failed to read master key: %w", err)
	}

	// 🔓 Decrypt the key
	aead, err := chacha20poly1305.NewX(m.wrapKey)
	if err != nil {
		return nil, fmt.Errorf("❌ failed to create AEAD: %w", err)
	}

	if len(encryptedKey) < aead.NonceSize() {
		return nil, fmt.Errorf("❌ encrypted key too short")
	}

	nonce := encryptedKey[:aead.NonceSize()]
	ciphertext := encryptedKey[aead.NonceSize():]

	masterKey, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("❌ failed to decrypt master key: %w", err)
	}

	// 🗄️ Cache the decrypted key
	m.keys[location] = masterKey

	return masterKey, nil
}

// 🔄 RotateLocationMasterKey rotates the master key for a location
func (m *LocationMasterKeyManager) RotateLocationMasterKey(location string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 🗑️ Remove old key from cache
	delete(m.keys, location)

	// 🆕 Generate new key
	return m.GenerateLocationMasterKey(location)
}
