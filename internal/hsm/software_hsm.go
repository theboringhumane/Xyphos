package hsm

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"sync"
)

// 🏭 softwareHSM implements Service using software-based crypto
type softwareHSM struct {
	mu        sync.RWMutex
	masterKey []byte
	aead      cipher.AEAD
	closed    bool
}

// 🆕 NewSoftwareHSM creates a new software HSM instance
func NewSoftwareHSM() (Service, error) {
	// Generate a random master key
	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &softwareHSM{
		masterKey: masterKey,
		aead:      aead,
		closed:    false,
	}, nil
}

// 🔒 Close implements Service.Close
func (h *softwareHSM) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return nil
	}

	// Zero out sensitive material
	for i := range h.masterKey {
		h.masterKey[i] = 0
	}
	h.closed = true
	return nil
}

// 🔑 GenerateKeyMaterial generates new key material for the specified algorithm
func (h *softwareHSM) GenerateKeyMaterial(algorithm string) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return nil, fmt.Errorf("hsm is closed")
	}

	var keySize int
	switch algorithm {
	case "AES-256-GCM":
		keySize = 32 // 256 bits
	case "RSA-OAEP-4096":
		keySize = 512 // 4096 bits
	case "ECDSA-P256":
		keySize = 32 // 256 bits
	default:
		return nil, ErrInvalidAlgorithm
	}

	keyMaterial := make([]byte, keySize)
	if _, err := rand.Read(keyMaterial); err != nil {
		return nil, fmt.Errorf("failed to generate key material: %w", err)
	}

	return keyMaterial, nil
}

// 🔒 Encrypt encrypts plaintext using the specified key version
func (h *softwareHSM) Encrypt(keyVersion []byte, plaintext []byte) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return nil, fmt.Errorf("hsm is closed")
	}

	// Create AES cipher for the key version
	block, err := aes.NewCipher(keyVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and seal
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// 🔓 Decrypt decrypts ciphertext using the specified key version
func (h *softwareHSM) Decrypt(keyVersion []byte, ciphertext []byte) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return nil, fmt.Errorf("hsm is closed")
	}

	// Create AES cipher for the key version
	block, err := aes.NewCipher(keyVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// 🎁 WrapKey wraps key material with the master key
func (h *softwareHSM) WrapKey(masterKey []byte, keyMaterial []byte) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return nil, fmt.Errorf("hsm is closed")
	}

	// Generate nonce
	nonce := make([]byte, h.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Wrap key using master key's AEAD
	return h.aead.Seal(nonce, nonce, keyMaterial, nil), nil
}

// 📦 UnwrapKey unwraps a wrapped key using the master key
func (h *softwareHSM) UnwrapKey(masterKey []byte, wrappedKey []byte) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return nil, fmt.Errorf("hsm is closed")
	}

	// Extract nonce from wrapped key
	if len(wrappedKey) < h.aead.NonceSize() {
		return nil, ErrInvalidKeySize
	}
	nonce := wrappedKey[:h.aead.NonceSize()]
	wrappedKey = wrappedKey[h.aead.NonceSize():]

	// Unwrap using the master key's AEAD
	keyMaterial, err := h.aead.Open(nil, nonce, wrappedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap key: %w", err)
	}

	return keyMaterial, nil
}

// 🔐 GenerateKey generates a new key of the specified algorithm
func (h *softwareHSM) GenerateKey(ctx context.Context, algorithm string) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return nil, fmt.Errorf("hsm is closed")
	}

	switch algorithm {
	case "AES-256-GCM":
		key := make([]byte, 32) // 256 bits
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate AES-256-GCM key: %w", err)
		}
		return key, nil

	case "RSA-OAEP-4096":
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA-OAEP-4096 key: %w", err)
		}
		return x509.MarshalPKCS1PrivateKey(privateKey), nil

	case "ECDSA-P256":
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA-P256 key: %w", err)
		}
		keyBytes, err := x509.MarshalECPrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ECDSA-P256 key: %w", err)
		}
		return keyBytes, nil

	default:
		return nil, ErrInvalidAlgorithm
	}
}

// 🔑 GenerateKeyPair generates a new public/private key pair using ECDSA-P256
func (h *softwareHSM) GenerateKeyPair() ([]byte, []byte, error) {
	// 🎯 Generate an ECDSA-P256 private key (the basis for the key pair)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ECDSA-P256 key pair: %w", err)
	}

	// 🛡 Marshal the private key into DER format
	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal ECDSA-P256 private key: %w", err)
	}

	// 🔓 Derive and marshal the public key from the generated private key
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal ECDSA-P256 public key: %w", err)
	}

	return pubBytes, privBytes, nil
}
