package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"lambda-kms/internal/models"
	"time"

	"github.com/google/uuid"
)

// 🔐 ClientSecurityService handles client security operations
type ClientSecurityService struct {
	keySize int // RSA key size in bits
}

// 🆕 NewClientSecurityService creates a new client security service
func NewClientSecurityService() *ClientSecurityService {
	return &ClientSecurityService{
		keySize: 4096, // Using 4096-bit RSA keys for strong security
	}
}

// 🎯 GenerateClientConfig creates a new client configuration with security credentials
func (s *ClientSecurityService) GenerateClientConfig(userID string, name string, permissions []string, expiresIn time.Duration) (*models.ClientConfig, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, s.keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// Export public key to PEM format
	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	}
	publicKeyStr := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(publicKeyPEM))

	// Export private key to PEM format (temporary, will be sent to client)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyStr := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(privateKeyPEM))

	// Generate client credentials
	clientID := uuid.New().String()
	clientSecret := generateSecureSecret()

	// Create client configuration
	config := &models.ClientConfig{
		ID:           uuid.New().String(),
		UserID:       userID,
		Name:         name,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		PublicKey:    publicKeyStr,
		PrivateKey:   privateKeyStr, // This will be removed after sending to client
		KeyAlgorithm: "RSA-4096",
		Permissions:  permissions,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(expiresIn),
		LastUsedAt:   time.Now(),
		Status:       models.ClientConfigStatusActive,
	}

	// Validate permissions
	if err := config.ValidatePermissions(); err != nil {
		return nil, err
	}

	return config, nil
}

// 🔒 generateSecureSecret generates a cryptographically secure secret
func generateSecureSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) // This should never happen
	}
	return base64.URLEncoding.EncodeToString(b)
}

// 🔍 ValidateClientCredentials validates client credentials
func (s *ClientSecurityService) ValidateClientCredentials(clientID, clientSecret string) bool {
	// TODO: Implement secure constant-time comparison
	// This is a placeholder - actual implementation should use the database
	return false
}

// 🔐 EncryptForClient encrypts data for a specific client using their public key
func (s *ClientSecurityService) EncryptForClient(publicKeyPEM string, data []byte) ([]byte, error) {
	// Decode PEM
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse public key
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Encrypt data
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return ciphertext, nil
}

// 🔓 DecryptFromClient decrypts data from a client using their private key (server-side only)
func (s *ClientSecurityService) DecryptFromClient(privateKeyPEM string, ciphertext []byte) ([]byte, error) {
	// Decode PEM
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Decrypt data
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
