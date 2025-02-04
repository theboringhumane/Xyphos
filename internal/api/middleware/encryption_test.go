package middleware

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"xyphos/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 🧪 TestEncryptionMiddleware_NoEncryption tests the middleware with no encryption
func TestEncryptionMiddleware_NoEncryption(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Create test router
	r := gin.New()
	r.Use(middleware.HandleEncryption())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"test":"data"}`))
	req.Header.Set("Content-Type", "application/json")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

// 🧪 TestEncryptionMiddleware_WithEncryption tests the middleware with encryption
func TestEncryptionMiddleware_WithEncryption(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Generate test client config with keys
	clientConfig, err := securityService.GenerateClientConfig("test-user", "test-client", []string{"write:keys"}, 24*60*60)
	require.NoError(t, err)
	require.NotEmpty(t, clientConfig.PublicKey, "Public key should not be empty")
	require.NotEmpty(t, clientConfig.PrivateKey, "Private key should not be empty")

	// Create test router
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("client_config", clientConfig)
		c.Next()
	})
	r.Use(middleware.HandleEncryption())
	r.POST("/test", func(c *gin.Context) {
		var data map[string]string
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"received": data["test"]})
	})

	// Create test request payload
	testPayload := map[string]string{"test": "encrypted-data"}

	// Encrypt test payload using client's public key
	encryptedPayload, err := securityService.EncryptForClient(clientConfig.PublicKey, mustMarshal(t, testPayload))
	require.NoError(t, err)

	// Base64 encode encrypted payload
	encodedPayload := base64.StdEncoding.EncodeToString(encryptedPayload)

	// Create test request with base64 encoded encrypted payload
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(encodedPayload))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "true", w.Header().Get("X-Response-Encrypted"))

	// Base64 decode response
	decodedResponse, err := base64.StdEncoding.DecodeString(w.Body.String())
	require.NoError(t, err, "Failed to base64 decode response")

	// Decrypt and verify response using client's private key
	decryptedData, err := securityService.DecryptFromClient(clientConfig.PrivateKey, decodedResponse)
	require.NoError(t, err)

	var response map[string]string
	err = json.Unmarshal(decryptedData, &response)
	require.NoError(t, err)
	assert.Equal(t, testPayload["test"], response["received"])
}

// 🧪 TestEncryptionMiddleware_InvalidEncryption tests the middleware with invalid encryption
func TestEncryptionMiddleware_InvalidEncryption(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Generate test client config with keys
	clientConfig, err := securityService.GenerateClientConfig("test-user", "test-client", []string{"write:keys"}, 24*60*60)
	require.NoError(t, err)
	require.NotEmpty(t, clientConfig.PublicKey, "Public key should not be empty")
	require.NotEmpty(t, clientConfig.PrivateKey, "Private key should not be empty")

	// Create test router
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("client_config", clientConfig)
		c.Next()
	})
	r.Use(middleware.HandleEncryption())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create invalid payload (not encrypted)
	invalidPayload := base64.StdEncoding.EncodeToString([]byte("invalid data"))

	// Create test request with invalid payload
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(invalidPayload))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Base64 decode response
	decodedResponse, err := base64.StdEncoding.DecodeString(w.Body.String())
	require.NoError(t, err, "Failed to base64 decode response")

	// Decrypt and verify error response
	decryptedData, err := securityService.DecryptFromClient(clientConfig.PrivateKey, decodedResponse)
	require.NoError(t, err)

	var response map[string]string
	err = json.Unmarshal(decryptedData, &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "failed to decrypt request")
}

// 🧪 TestEncryptionMiddleware_MissingClientConfig tests the middleware without client config
func TestEncryptionMiddleware_MissingClientConfig(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Create test router
	r := gin.New()
	r.Use(middleware.HandleEncryption())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"test":"data"}`))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert - should pass through without encryption since no client config
	assert.Equal(t, http.StatusOK, w.Code)
}

// 🧪 TestEncryptionMiddleware_EmptyRequest tests the middleware with empty request body
func TestEncryptionMiddleware_EmptyRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Generate test client config
	clientConfig, err := securityService.GenerateClientConfig("test-user", "test-client", []string{"write:keys"}, 24*60*60)
	require.NoError(t, err)

	// Create test router
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("client_config", clientConfig)
		c.Next()
	})
	r.Use(middleware.HandleEncryption())
	r.POST("/test", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}
		if len(body) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty request body"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create test request with empty body
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, "true", w.Header().Get("X-Response-Encrypted"))

	// Base64 decode response
	decodedResponse, err := base64.StdEncoding.DecodeString(w.Body.String())
	require.NoError(t, err, "Failed to base64 decode response")

	// Decrypt and verify error response
	decryptedData, err := securityService.DecryptFromClient(clientConfig.PrivateKey, decodedResponse)
	require.NoError(t, err)

	var response map[string]string
	err = json.Unmarshal(decryptedData, &response)
	require.NoError(t, err)
	assert.Equal(t, "empty request body", response["error"])
}

// 🧪 TestEncryptionMiddleware_LargePayload tests the middleware with a large payload
func TestEncryptionMiddleware_LargePayload(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Generate test client config
	clientConfig, err := securityService.GenerateClientConfig("test-user", "test-client", []string{"write:keys"}, 24*60*60)
	require.NoError(t, err)

	// Create large test payload (1MB)
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	// Create test router
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("client_config", clientConfig)
		c.Next()
	})
	r.Use(middleware.HandleEncryption())
	r.POST("/test", func(c *gin.Context) {
		// Read raw request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Echo back request body in response
		c.Data(http.StatusOK, "application/octet-stream", body)
	})

	// Encrypt large payload using client's public key
	encryptedPayload, err := securityService.EncryptForClient(clientConfig.PublicKey, largePayload)
	require.NoError(t, err)

	// Base64 encode encrypted large payload
	encodedPayload := base64.StdEncoding.EncodeToString(encryptedPayload)

	// Create test request with base64 encoded encrypted payload
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(encodedPayload))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "true", w.Header().Get("X-Response-Encrypted"))

	// Base64 decode response
	decodedResponse, err := base64.StdEncoding.DecodeString(w.Body.String())
	require.NoError(t, err, "Failed to base64 decode response")

	// Decrypt and verify response using client's private key
	decryptedData, err := securityService.DecryptFromClient(clientConfig.PrivateKey, decodedResponse)
	require.NoError(t, err)

	// Verify the decrypted data matches the original payload
	assert.Equal(t, largePayload, decryptedData)
}

// 🏃 BenchmarkEncryptionMiddleware_SmallPayload benchmarks encryption with small payload
func BenchmarkEncryptionMiddleware_SmallPayload(b *testing.B) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Generate test client config
	clientConfig, _ := securityService.GenerateClientConfig("test-user", "test-client", []string{"encrypt"}, 24*60*60)

	// Create test router
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("client_config", clientConfig)
		c.Next()
	})
	r.Use(middleware.HandleEncryption())
	r.POST("/test", func(c *gin.Context) {
		c.Header("X-Response-Encrypted", "true")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create and encrypt test data
	testData := map[string]string{"test": "data"}
	encryptedData, _ := securityService.EncryptForClient(clientConfig.PublicKey, mustMarshal(b, testData))
	encodedData := base64.StdEncoding.EncodeToString(encryptedData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(encodedData))
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("X-Request-Encrypted", "true")
		r.ServeHTTP(w, req)
	}
}

// 🏃 BenchmarkEncryptionMiddleware_LargePayload benchmarks encryption with large payload
func BenchmarkEncryptionMiddleware_LargePayload(b *testing.B) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Generate test client config
	clientConfig, _ := securityService.GenerateClientConfig("test-user", "test-client", []string{"encrypt"}, 24*60*60)

	// Create large test data (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// Create test router
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("client_config", clientConfig)
		c.Next()
	})
	r.Use(middleware.HandleEncryption())
	r.POST("/test", func(c *gin.Context) {
		c.Header("X-Response-Encrypted", "true")
		c.JSON(http.StatusOK, gin.H{"size": len(largeData)})
	})

	// Encrypt large data
	encryptedData, _ := securityService.EncryptForClient(clientConfig.PublicKey, largeData)
	encodedData := base64.StdEncoding.EncodeToString(encryptedData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(encodedData))
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("X-Request-Encrypted", "true")
		r.ServeHTTP(w, req)
	}
}

// 🔧 Helper function for marshaling in tests
func mustMarshal(t testing.TB, v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}
	return data
}
