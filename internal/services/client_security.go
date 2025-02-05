package services

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
	"xyphos/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/chacha20poly1305"
)

// 🔐 ClientSecurityService handles client security operations
type ClientSecurityService struct {
	keySize         int             // RSA key size in bits
	serverKeyPair   *rsa.PrivateKey // Server's key pair for testing
	serverPublicKey string          // Cached PEM-encoded public key
	keyCache        *sync.Map
	connections     chan *HSMConnection
	maxPoolSize     int
}

// 🆕 NewClientSecurityService creates a new client security service
func NewClientSecurityService() *ClientSecurityService {
	maxPoolSize := runtime.NumCPU() * 2
	service := &ClientSecurityService{
		keySize:     4096, // Using 4096-bit RSA keys for strong security
		keyCache:    &sync.Map{},
		maxPoolSize: maxPoolSize,
		connections: make(chan *HSMConnection, maxPoolSize),
	}

	// Generate server key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, service.keySize)
	if err != nil {
		panic(fmt.Sprintf("failed to generate server key pair: %v", err))
	}
	service.serverKeyPair = privateKey

	// Cache the PEM-encoded public key using PKIX format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal server public key: %v", err))
	}
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	service.serverPublicKey = string(pem.EncodeToMemory(publicKeyPEM))

	// Initialize connection pool
	for i := 0; i < maxPoolSize; i++ {
		conn := newHSMConnection()
		service.connections <- conn
	}

	return service
}

// 🎯 GenerateClientConfig creates a new client configuration with security credentials
func (s *ClientSecurityService) GenerateClientConfig(userID string, name string, permissions []string, expiresIn time.Duration) (*models.ClientConfig, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, s.keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// Export public key to PEM format using PKIX
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyStr := string(pem.EncodeToMemory(publicKeyPEM))

	// Export private key to PEM format using PKCS8
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privateKeyPEM := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyStr := string(pem.EncodeToMemory(privateKeyPEM))

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

// 🔒 EncryptWithHSM encrypts data using HSM with connection pooling
func (s *ClientSecurityService) EncryptWithHSM(data []byte) ([]byte, error) {
	// Get connection from pool
	conn := <-s.connections
	defer func() { s.connections <- conn }()

	// Check key cache first
	if cachedCipher, ok := s.keyCache.Load("hsm_key"); ok {
		return conn.Encrypt(data, cachedCipher.([]byte))
	}

	// Get key from HSM if not cached
	key, err := conn.GetKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get HSM key: %w", err)
	}

	// Cache the key
	s.keyCache.Store("hsm_key", key)

	// Encrypt data
	return conn.Encrypt(data, key)
}

// 🔓 DecryptWithHSM decrypts data using HSM with connection pooling
func (s *ClientSecurityService) DecryptWithHSM(data []byte) ([]byte, error) {
	// Get connection from pool
	conn := <-s.connections
	defer func() { s.connections <- conn }()

	// Check key cache first
	if cachedKey, ok := s.keyCache.Load("hsm_key"); ok {
		return conn.Decrypt(data, cachedKey.([]byte))
	}

	// Get key from HSM if not cached
	key, err := conn.GetKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get HSM key: %w", err)
	}

	// Cache the key
	s.keyCache.Store("hsm_key", key)

	// Decrypt data
	return conn.Decrypt(data, key)
}

// 🔒 EncryptForClient encrypts data for client transport with optimized buffering
func (s *ClientSecurityService) EncryptForClient(publicKeyPEM string, data []byte) ([]byte, error) {
	// Use buffer pool for better performance
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	// Parse public key (cached)
	publicKey, err := s.getPublicKey(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	// Use ChaCha20-Poly1305 for faster encryption of large data
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt data with ChaCha20-Poly1305
	ciphertext := aead.Seal(nil, nonce, data, nil)

	// Encrypt key with RSA
	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		key,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Write to buffer
	binary.Write(buf, binary.LittleEndian, uint32(len(encryptedKey)))
	buf.Write(encryptedKey)
	buf.Write(nonce)
	buf.Write(ciphertext)

	return buf.Bytes(), nil
}

// 🔓 DecryptFromClient decrypts data from client transport with optimized buffering
func (s *ClientSecurityService) DecryptFromClient(privateKeyPEM string, data []byte) ([]byte, error) {
	// Use buffer pool for better performance
	buf := bytes.NewBuffer(data)

	// Parse private key (cached)
	privateKey, err := s.getPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	// Read encrypted key length
	var keyLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &keyLen); err != nil {
		return nil, err
	}

	// Read encrypted key
	encryptedKey := make([]byte, keyLen)
	if _, err := io.ReadFull(buf, encryptedKey); err != nil {
		return nil, err
	}

	// Decrypt key with RSA
	key, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Create ChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Read nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(buf, nonce); err != nil {
		return nil, err
	}

	// Decrypt data
	ciphertext := buf.Bytes()
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Buffer pool for reusing buffers
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Key cache methods
func (s *ClientSecurityService) getPublicKey(pemStr string) (*rsa.PublicKey, error) {
	if cached, ok := s.keyCache.Load(pemStr); ok {
		return cached.(*rsa.PublicKey), nil
	}

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	s.keyCache.Store(pemStr, rsaPub)
	return rsaPub, nil
}

func (s *ClientSecurityService) getPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	if cached, ok := s.keyCache.Load(pemStr); ok {
		return cached.(*rsa.PrivateKey), nil
	}

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	s.keyCache.Store(pemStr, rsaPriv)
	return rsaPriv, nil
}

// 🔐 DecryptWithServerKey decrypts data using the server's private key
func (s *ClientSecurityService) DecryptWithServerKey(encryptedData []byte) ([]byte, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(s.serverKeyPair)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal server private key: %w", err)
	}
	privateKeyPEM := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	return s.DecryptFromClient(string(pem.EncodeToMemory(privateKeyPEM)), encryptedData)
}

// 🔐 GetServerPublicKey returns the server's public key for request encryption
func (s *ClientSecurityService) GetServerPublicKey() (string, error) {
	return s.serverPublicKey, nil
}

// 🔐 EncryptWithServerKey encrypts data using the server's public key
func (s *ClientSecurityService) EncryptWithServerKey(publicKeyPEM string, data []byte) ([]byte, error) {
	return s.EncryptForClient(publicKeyPEM, data)
}
