package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"sync"
)

// 🔐 HSMClient interface for hardware security module operations
type HSMClient interface {
	GetKey() ([]byte, error)
	Encrypt(data []byte, key []byte) ([]byte, error)
	Decrypt(data []byte, key []byte) ([]byte, error)
}

// 🔌 HSMConnection represents a connection to the HSM
type HSMConnection struct {
	client HSMClient
	aead   cipher.AEAD
	mutex  sync.Mutex
}

// 🆕 newHSMConnection creates a new HSM connection
func newHSMConnection() *HSMConnection {
	return &HSMConnection{
		client: &softwareHSM{}, // Using software HSM for development
		mutex:  sync.Mutex{},
	}
}

// 🔒 Encrypt encrypts data using the HSM connection
func (c *HSMConnection) Encrypt(data []byte, key []byte) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.initAEAD(key); err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt data
	ciphertext := c.aead.Seal(nil, nonce, data, nil)

	// Combine nonce and ciphertext
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result, nonce)
	copy(result[len(nonce):], ciphertext)

	return result, nil
}

// 🔓 Decrypt decrypts data using the HSM connection
func (c *HSMConnection) Decrypt(data []byte, key []byte) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.initAEAD(key); err != nil {
		return nil, err
	}

	if len(data) < c.aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Split nonce and ciphertext
	nonce := data[:c.aead.NonceSize()]
	ciphertext := data[c.aead.NonceSize():]

	// Decrypt data
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// 🔧 initAEAD initializes the AEAD cipher
func (c *HSMConnection) initAEAD(key []byte) error {
	if c.aead != nil {
		return nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	c.aead, err = cipher.NewGCM(block)
	return err
}

// 🔐 GetKey gets the encryption key from HSM
func (c *HSMConnection) GetKey() ([]byte, error) {
	return c.client.GetKey()
}

// 💻 softwareHSM implements HSMClient interface for development/testing
type softwareHSM struct{}

// 🔑 GetKey generates a new key (simulated HSM)
func (h *softwareHSM) GetKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// 🔒 Encrypt encrypts data using software HSM
func (h *softwareHSM) Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// 🔓 Decrypt decrypts data using software HSM
func (h *softwareHSM) Decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
