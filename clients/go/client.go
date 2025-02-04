package kms

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

// 🚨 Error types for better error handling
var (
	ErrAuthentication = fmt.Errorf("authentication failed")
	ErrNotFound       = fmt.Errorf("resource not found")
	ErrPermission     = fmt.Errorf("permission denied")
	ErrInvalidInput   = fmt.Errorf("invalid input")
)

// 🔄 RetryConfig configures the retry behavior
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// 🔧 ClientConfig represents the client configuration
type ClientConfig struct {
	BaseURL            string
	ClientConfigID     string // 🔑 Client config ID from the KMS service
	ClientConfigSecret string // 🔐 Client config secret from the KMS service
	Timeout            time.Duration
	RetryConfig        *RetryConfig
	PrivateKey         string // 🔐 RSA private key in PEM format
	PublicKey          string // 🔑 RSA public key in PEM format
}

// 🔐 ClientConfigInfo represents client configuration information
type ClientConfigInfo struct {
	ID          string    `json:"id"`
	Secret      string    `json:"secret"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Permissions []string  `json:"permissions"`
}

// 🌐 Client represents a Xyphos client
type Client struct {
	config      ClientConfig
	httpClient  *http.Client
	token       string
	tokenExpiry time.Time
	keyCache    sync.Map
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
}

// 📦 Project represents a KMS project
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// 📍 Location represents a KMS location
type Location struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// 💍 KeyRing represents a KMS keyring
type KeyRing struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// 🔑 CryptoKey represents a KMS key
type CryptoKey struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Algorithm      string    `json:"algorithm"`
	Purpose        string    `json:"purpose"`
	RotationPeriod int       `json:"rotation_period"`
	CreatedAt      time.Time `json:"created_at"`
	NextRotation   time.Time `json:"next_rotation"`
	Version        int       `json:"version"`
}

// 🎯 ClientOption allows configuring the client
type ClientOption func(*Client)

// 🔧 WithRetryConfig sets the retry configuration
func WithRetryConfig(config RetryConfig) ClientOption {
	return func(c *Client) {
		c.config.RetryConfig = &config
	}
}

// 🔧 WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.config.Timeout = timeout
	}
}

// 🎯 NewClient creates a new Xyphos client
func NewClient(config ClientConfig) (*Client, error) {
	if config.RetryConfig == nil {
		config.RetryConfig = &RetryConfig{
			MaxRetries:  3,
			InitialWait: time.Second,
			MaxWait:     time.Second * 10,
		}
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	if config.PrivateKey != "" {
		if err := client.initKeys(config.PrivateKey); err != nil {
			return nil, fmt.Errorf("failed to initialize keys: %w", err)
		}
	}

	return client, nil
}

// 🔐 initKeys initializes RSA keys from PEM
func (c *Client) initKeys(privateKeyPEM string) error {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not RSA")
	}

	c.privateKey = rsaKey
	c.publicKey = &rsaKey.PublicKey
	return nil
}

// 🔒 Encrypt encrypts data using a KMS key
func (c *Client) Encrypt(ctx context.Context, projectID, locationID, keyringID, keyID string, plaintext []byte) ([]byte, error) {
	// Generate ChaCha20-Poly1305 key
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Create cipher
	cipher, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := cipher.Seal(nil, nonce, plaintext, nil)

	// Get RSA public key
	publicKey, err := c.getPublicKey(c.config.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Encrypt symmetric key
	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		key,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt key: %w", err)
	}

	// Combine all parts
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.LittleEndian, uint32(len(encryptedKey)))
	buf.Write(encryptedKey)
	buf.Write(nonce)
	buf.Write(ciphertext)

	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/api/projects/%s/locations/%s/keyrings/%s/keys/%s:encrypt",
			c.config.BaseURL, projectID, locationID, keyringID, keyID),
		&buf,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Request-Encrypted", "true")
	if err := c.setAuthHeader(ctx, req); err != nil {
		return nil, err
	}

	// Send request
	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, nil
}

// 🔓 Decrypt decrypts data using a KMS key
func (c *Client) Decrypt(ctx context.Context, projectID, locationID, keyringID, keyID string, ciphertext []byte) ([]byte, error) {
	// Get RSA private key
	privateKey, err := c.getPrivateKey(c.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	// Extract parts
	var keyLength uint32
	if err := binary.Read(bytes.NewReader(ciphertext), binary.LittleEndian, &keyLength); err != nil {
		return nil, fmt.Errorf("failed to read key length: %w", err)
	}

	if len(ciphertext) < int(keyLength)+4+chacha20poly1305.NonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	encryptedKey := ciphertext[4 : 4+keyLength]
	nonce := ciphertext[4+keyLength : 4+keyLength+chacha20poly1305.NonceSize]
	encryptedData := ciphertext[4+keyLength+chacha20poly1305.NonceSize:]

	// Decrypt symmetric key
	key, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %w", err)
	}

	// Create cipher
	cipher, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt data
	plaintext, err := cipher.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// 🔑 Get cached public key
func (c *Client) getPublicKey(pemKey string) (*rsa.PublicKey, error) {
	if key, ok := c.keyCache.Load(pemKey); ok {
		return key.(*rsa.PublicKey), nil
	}

	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	c.keyCache.Store(pemKey, rsaPub)
	return rsaPub, nil
}

// 🔑 Get cached private key
func (c *Client) getPrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	if key, ok := c.keyCache.Load(pemKey); ok {
		return key.(*rsa.PrivateKey), nil
	}

	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	c.keyCache.Store(pemKey, rsaPriv)
	return rsaPriv, nil
}

// 🔄 Set authorization header
func (c *Client) setAuthHeader(ctx context.Context, req *http.Request) error {
	if err := c.ensureToken(ctx); err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	return nil
}

// 🔄 Ensure valid token
func (c *Client) ensureToken(ctx context.Context) error {
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	reqBody := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
	}{
		ClientID:     c.config.ClientConfigID,
		ClientSecret: c.config.ClientConfigSecret,
		GrantType:    "client_credentials",
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/api/v1/oauth/token", c.config.BaseURL),
		bytes.NewBuffer(reqJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get token: status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.token = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// 📦 Project operations

// CreateProject creates a new KMS project
func (c *Client) CreateProject(ctx context.Context, name, description string) (*Project, error) {
	reqBody := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		Name:        name,
		Description: description,
	}

	var project Project
	err := c.doRequest(ctx, "POST", "/projects", reqBody, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &project, nil
}

// ListProjects lists all KMS projects
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/projects", c.config.BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list projects request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list projects: status %d", resp.StatusCode)
	}

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode projects response: %w", err)
	}

	return projects, nil
}

// 📍 Location operations

// CreateLocation creates a new KMS location in a project
func (c *Client) CreateLocation(ctx context.Context, projectID, name string) (*Location, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	type createLocationRequest struct {
		Name string `json:"name"`
	}

	reqBody := createLocationRequest{
		Name: name,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal location request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/projects/%s/locations", c.config.BaseURL, projectID), bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create location request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create location: status %d", resp.StatusCode)
	}

	var location Location
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		return nil, fmt.Errorf("failed to decode location response: %w", err)
	}

	return &location, nil
}

// ListLocations lists all KMS locations in a project
func (c *Client) ListLocations(ctx context.Context, projectID string) ([]Location, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/projects/%s/locations", c.config.BaseURL, projectID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list locations request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list locations: status %d", resp.StatusCode)
	}

	var locations []Location
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, fmt.Errorf("failed to decode locations response: %w", err)
	}

	return locations, nil
}

// 💍 KeyRing operations

// CreateKeyRing creates a new KMS keyring in a location
func (c *Client) CreateKeyRing(ctx context.Context, projectID, locationID, name string) (*KeyRing, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	type createKeyRingRequest struct {
		Name string `json:"name"`
	}

	reqBody := createKeyRingRequest{
		Name: name,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal keyring request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/projects/%s/locations/%s/keyrings", c.config.BaseURL, projectID, locationID), bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create keyring: status %d", resp.StatusCode)
	}

	var keyring KeyRing
	if err := json.NewDecoder(resp.Body).Decode(&keyring); err != nil {
		return nil, fmt.Errorf("failed to decode keyring response: %w", err)
	}

	return &keyring, nil
}

// ListKeyRings lists all KMS keyrings in a location
func (c *Client) ListKeyRings(ctx context.Context, projectID, locationID string) ([]KeyRing, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/projects/%s/locations/%s/keyrings", c.config.BaseURL, projectID, locationID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list keyrings request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list keyrings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list keyrings: status %d", resp.StatusCode)
	}

	var keyrings []KeyRing
	if err := json.NewDecoder(resp.Body).Decode(&keyrings); err != nil {
		return nil, fmt.Errorf("failed to decode keyrings response: %w", err)
	}

	return keyrings, nil
}

// 🔑 CryptoKey operations

// CreateCryptoKey creates a new KMS key in a keyring
func (c *Client) CreateCryptoKey(ctx context.Context, projectID, locationID, keyringID string, algorithm, purpose string, rotationPeriod int) (*CryptoKey, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	type createKeyRequest struct {
		Algorithm      string `json:"algorithm"`
		Purpose        string `json:"purpose"`
		RotationPeriod int    `json:"rotation_period"`
	}

	reqBody := createKeyRequest{
		Algorithm:      algorithm,
		Purpose:        purpose,
		RotationPeriod: rotationPeriod,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/projects/%s/locations/%s/keyrings/%s/keys", c.config.BaseURL, projectID, locationID, keyringID), bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create key request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create key: status %d", resp.StatusCode)
	}

	var key CryptoKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, fmt.Errorf("failed to decode key response: %w", err)
	}

	return &key, nil
}

// ListCryptoKeys lists all KMS keys in a keyring
func (c *Client) ListCryptoKeys(ctx context.Context, projectID, locationID, keyringID string) ([]CryptoKey, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/projects/%s/locations/%s/keyrings/%s/keys", c.config.BaseURL, projectID, locationID, keyringID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list keys request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list keys: status %d", resp.StatusCode)
	}

	var keys []CryptoKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode keys response: %w", err)
	}

	return keys, nil
}

// RotateCryptoKey rotates a KMS key
func (c *Client) RotateCryptoKey(ctx context.Context, projectID, locationID, keyringID, keyID string) error {
	if err := c.ensureToken(ctx); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/projects/%s/locations/%s/keyrings/%s/keys/%s:rotate", c.config.BaseURL, projectID, locationID, keyringID, keyID), nil)
	if err != nil {
		return fmt.Errorf("failed to create rotate key request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return fmt.Errorf("failed to rotate key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to rotate key: status %d", resp.StatusCode)
	}

	return nil
}

// 🔄 Do request with retry
func (c *Client) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	wait := c.config.RetryConfig.InitialWait

	for i := 0; i <= c.config.RetryConfig.MaxRetries; i++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if i == c.config.RetryConfig.MaxRetries {
				break
			}
			time.Sleep(wait)
			wait = min(wait*2, c.config.RetryConfig.MaxWait)
			continue
		}

		if resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}

		resp.Body.Close()
		if i == c.config.RetryConfig.MaxRetries {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			break
		}

		time.Sleep(wait)
		wait = min(wait*2, c.config.RetryConfig.MaxWait)
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", c.config.RetryConfig.MaxRetries, lastErr)
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// 🌐 doRequest performs an HTTP request with JSON handling
func (c *Client) doRequest(ctx context.Context, method, path string, reqBody interface{}, result interface{}) error {
	if err := c.ensureToken(ctx); err != nil {
		return err
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s/api%s", c.config.BaseURL, path),
		bytes.NewBuffer(reqJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: status %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
