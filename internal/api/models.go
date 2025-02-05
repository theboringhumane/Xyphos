package api

import "time"

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status" example:"healthy"`
}

// ErrorResponse represents an API error response
// @Description API error response
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Invalid request parameters"`
}

// Project represents a Xyphos project
// @Description A Xyphos project that can contain multiple keyrings
type Project struct {
	ID          string    `json:"id" example:"proj-123"`
	Name        string    `json:"name" example:"My Project"`
	Description string    `json:"description" example:"A project for managing encryption keys"`
	CreatedAt   time.Time `json:"created_at" example:"2025-02-04T12:00:00Z"`
}

// CreateProjectRequest represents the request to create a new project
// @Description Request body for creating a new project
type CreateProjectRequest struct {
	Name        string `json:"name" example:"My Project" binding:"required"`
	Description string `json:"description" example:"A project for managing encryption keys"`
}

// Location represents a KMS location
type Location struct {
	ID        string    `json:"id" example:"us-west1"`
	Name      string    `json:"name" example:"us-west1"`
	CreatedAt time.Time `json:"created_at" example:"2024-02-04T12:00:00Z"`
}

// KeyRing represents a collection of cryptographic keys
// @Description A collection of cryptographic keys
type KeyRing struct {
	ID        string    `json:"id" example:"kr-123"`
	Name      string    `json:"name" example:"Production Keys"`
	CreatedAt time.Time `json:"created_at" example:"2025-02-04T12:00:00Z"`
}

// Tenant represents a tenant in a keyring
type Tenant struct {
	ID        string    `json:"id" example:"tenant-123"`
	Name      string    `json:"name" example:"Production"`
	CreatedAt time.Time `json:"created_at" example:"2025-02-04T12:00:00Z"`
}

// CreateTenantRequest represents the request to create a new tenant
// @Description Request body for creating a new tenant
type CreateTenantRequest struct {
	Name string `json:"name" example:"Production" binding:"required"`
}

// CreateKeyRingRequest represents the request to create a new keyring
// @Description Request body for creating a new keyring
type CreateKeyRingRequest struct {
	Name string `json:"name" example:"Production Keys" binding:"required"`
}

// CryptoKey represents a cryptographic key
// @Description A cryptographic key used for encryption/decryption
type CryptoKey struct {
	ID             string    `json:"id" example:"key-123"`
	Name           string    `json:"name" example:"AES-256 Key"`
	Algorithm      string    `json:"algorithm" example:"AES-256-GCM"`
	Purpose        string    `json:"purpose" example:"ENCRYPT_DECRYPT"`
	RotationPeriod int       `json:"rotation_period" example:"86400"`
	CreatedAt      time.Time `json:"created_at" example:"2025-02-04T12:00:00Z"`
	NextRotation   time.Time `json:"next_rotation" example:"2025-02-05T12:00:00Z"`
	Version        int       `json:"version" example:"1"`
}

// CreateCryptoKeyRequest represents the request to create a new crypto key
// @Description Request body for creating a new cryptographic key
type CreateCryptoKeyRequest struct {
	Name           string `json:"name" example:"AES-256 Key" binding:"required"`
	Algorithm      string `json:"algorithm" example:"AES-256-GCM" binding:"required"`
	Purpose        string `json:"purpose" example:"ENCRYPT_DECRYPT" binding:"required"`
	RotationPeriod int    `json:"rotation_period" example:"86400"`
}

// EncryptRequest represents the request to encrypt data
// @Description Request body for encrypting data
type EncryptRequest struct {
	Plaintext string `json:"plaintext" example:"Hello, World!" binding:"required"`
}

// EncryptResponse represents the response from encrypting data
// @Description Response body after encrypting data
type EncryptResponse struct {
	Ciphertext string `json:"ciphertext" example:"base64-encoded-encrypted-data"`
	KeyVersion int    `json:"key_version" example:"1"`
}

// DecryptRequest represents the request to decrypt data
// @Description Request body for decrypting data
type DecryptRequest struct {
	Ciphertext string `json:"ciphertext" example:"base64-encoded-encrypted-data" binding:"required"`
}

// DecryptResponse represents the response from decrypting data
// @Description Response body after decrypting data
type DecryptResponse struct {
	Plaintext string `json:"plaintext" example:"Hello, World!"`
}

// ClientConfig represents a client configuration
type ClientConfig struct {
	ID          string    `json:"id" example:"config-123"`
	Name        string    `json:"name" example:"my-client"`
	ClientID    string    `json:"client_id" example:"client-123"`
	CreatedAt   time.Time `json:"created_at" example:"2024-02-04T12:00:00Z"`
	RevokedAt   time.Time `json:"revoked_at,omitempty"`
	Permissions []string  `json:"permissions" example:"['encrypt', 'decrypt']"`
}

// CreateClientConfigRequest represents the request to create a client configuration
type CreateClientConfigRequest struct {
	Name        string   `json:"name" example:"my-client" binding:"required"`
	Permissions []string `json:"permissions" example:"['encrypt', 'decrypt']" binding:"required"`
}

// RevokeResponse represents the response from revoking a client configuration
type RevokeResponse struct {
	RevokedAt time.Time `json:"revoked_at" example:"2024-02-04T12:00:00Z"`
}
