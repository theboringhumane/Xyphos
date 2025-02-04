package models

import (
	"time"

	"github.com/google/uuid"
)

// Purpose defines the allowed purposes for a CryptoKey
type Purpose string

const (
	PurposeEncryptDecrypt Purpose = "ENCRYPT_DECRYPT"
	PurposeSignVerify     Purpose = "SIGN_VERIFY"
)

// Algorithm defines the cryptographic algorithm for a CryptoKey
type Algorithm string

const (
	AlgorithmAES256GCM Algorithm = "AES-256-GCM"
	AlgorithmRSAOAEP   Algorithm = "RSA-OAEP-4096"
	AlgorithmECDSAP256 Algorithm = "ECDSA-P256"
)

// Status defines the status of a CryptoKey or CryptoKeyVersion
type Status string

const (
	StatusEnabled    Status = "ENABLED"
	StatusDisabled   Status = "DISABLED"
	StatusDestroying Status = "DESTROYING"
	StatusDestroyed  Status = "DESTROYED"
)

// CryptoKey represents a cryptographic key
type CryptoKey struct {
	ID             string             `json:"id"`
	KeyringID      string             `json:"keyringId"`
	Name           string             `json:"name"`
	Algorithm      Algorithm          `json:"algorithm"`
	Purpose        Purpose            `json:"purpose"`
	RotationPeriod string             `json:"rotationPeriod,omitempty"`
	NextRotation   time.Time          `json:"nextRotation,omitempty"`
	Status         Status             `json:"status"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	Versions       []CryptoKeyVersion `json:"versions,omitempty"`
}

// CryptoKeyVersion represents a version of a CryptoKey
type CryptoKeyVersion struct {
	ID           string    `json:"id"`
	CryptoKeyID  string    `json:"cryptoKeyId"`
	State        Status    `json:"state"`
	ProtectedKey []byte    `json:"-"` // Encrypted key material
	CreatedAt    time.Time `json:"createdAt"`
	DestroyAt    time.Time `json:"destroyAt,omitempty"`
}

// NewCryptoKey creates a new CryptoKey with default values
func NewCryptoKey(keyringID, name string, algorithm Algorithm, purpose Purpose, rotationPeriod string) *CryptoKey {
	now := time.Now()
	return &CryptoKey{
		ID:             uuid.New().String(),
		KeyringID:      keyringID,
		Name:           name,
		Algorithm:      algorithm,
		Purpose:        purpose,
		RotationPeriod: rotationPeriod,
		Status:         StatusEnabled,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewCryptoKeyVersion creates a new version of a CryptoKey
func NewCryptoKeyVersion(cryptoKeyID string, protectedKey []byte) *CryptoKeyVersion {
	return &CryptoKeyVersion{
		ID:           uuid.New().String(),
		CryptoKeyID:  cryptoKeyID,
		State:        StatusEnabled,
		ProtectedKey: protectedKey,
		CreatedAt:    time.Now(),
	}
}
