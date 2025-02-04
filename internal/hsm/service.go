// Package hsm provides hardware security module (HSM) functionality
package hsm

import (
	"context"
	"errors"
)

var (
	// 🚫 ErrInvalidKeySize indicates an invalid key size
	ErrInvalidKeySize = errors.New("invalid key size")

	// 🚫 ErrInvalidAlgorithm indicates an unsupported algorithm
	ErrInvalidAlgorithm = errors.New("unsupported algorithm")
)

// 🔐 Service defines the interface for HSM operations
type Service interface {
	// Key generation
	GenerateKeyMaterial(algorithm string) ([]byte, error)

	// Encryption/Decryption
	Encrypt(keyVersion []byte, plaintext []byte) ([]byte, error)
	Decrypt(keyVersion []byte, ciphertext []byte) ([]byte, error)

	// Key wrapping
	WrapKey(masterKey []byte, keyMaterial []byte) ([]byte, error)
	UnwrapKey(masterKey []byte, wrappedKey []byte) ([]byte, error)

	// GenerateKey generates a new key of the specified algorithm
	GenerateKey(ctx context.Context, algorithm string) ([]byte, error)

	// GenerateKeyPair generates a new public/private key pair
	GenerateKeyPair() ([]byte, []byte, error)

	// Close closes the HSM service
	Close() error
}
