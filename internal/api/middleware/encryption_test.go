package middleware

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"xyphos/internal/models"
	"xyphos/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
	clientConfig, err := securityService.GenerateClientConfig("test-user", "test-client", []string{"encrypt"}, 24*60*60)
	assert.NoError(t, err)

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
		c.Header("X-Response-Encrypted", "true")
		c.JSON(http.StatusOK, gin.H{"received": data["test"]})
	})

	// Create and encrypt test request
	testData := map[string]string{"test": "encrypted-data"}
	encryptedData, err := securityService.EncryptForClient(clientConfig.PublicKey, mustMarshal(t, testData))
	assert.NoError(t, err)

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(base64.StdEncoding.EncodeToString(encryptedData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "true", w.Header().Get("X-Response-Encrypted"))

	// Decrypt and verify response
	encryptedResponse, err := base64.StdEncoding.DecodeString(w.Body.String())
	assert.NoError(t, err)

	decryptedData, err := securityService.DecryptFromClient(clientConfig.PrivateKey, encryptedResponse)
	assert.NoError(t, err)

	var response map[string]string
	err = json.Unmarshal(decryptedData, &response)
	assert.NoError(t, err)
	assert.Equal(t, "encrypted-data", response["received"])
}

// 🧪 TestEncryptionMiddleware_InvalidEncryption tests the middleware with invalid encryption
func TestEncryptionMiddleware_InvalidEncryption(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Create test client config
	clientConfig := &models.ClientConfig{
		ID:         "test-id",
		PublicKey:  "invalid-key",
		PrivateKey: "invalid-key",
	}

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

	// Create test request with invalid encryption
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString("invalid-encrypted-data"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid encrypted data")
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
	req.Header.Set("Content-Type", "application/json")
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
	clientConfig, err := securityService.GenerateClientConfig("test-user", "test-client", []string{"encrypt"}, 24*60*60)
	assert.NoError(t, err)

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

	// Create test request with empty body
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// 🧪 TestEncryptionMiddleware_LargePayload tests the middleware with a large payload
func TestEncryptionMiddleware_LargePayload(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	securityService := services.NewClientSecurityService()
	middleware := NewEncryptionMiddleware(securityService)

	// Generate test client config
	clientConfig, err := securityService.GenerateClientConfig("test-user", "test-client", []string{"encrypt"}, 24*60*60)
	assert.NoError(t, err)

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
	encryptedData, err := securityService.EncryptForClient(clientConfig.PublicKey, largeData)
	assert.NoError(t, err)

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(base64.StdEncoding.EncodeToString(encryptedData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Encrypted", "true")

	// Test
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "true", w.Header().Get("X-Response-Encrypted"))
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
		req.Header.Set("Content-Type", "application/json")
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
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-Encrypted", "true")
		r.ServeHTTP(w, req)
	}
}

// Helper function for benchmarks
func mustMarshal(b *testing.B, v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		b.Fatal(err)
	}
	return data
}
