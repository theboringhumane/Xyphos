package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"xyphos/internal/models"
	"xyphos/internal/services"

	"encoding/base64"

	"github.com/gin-gonic/gin"
)

// 🔐 EncryptionMiddleware handles request/response encryption using RSA-OAEP.
// It provides end-to-end encryption for client-server communication by:
// 1. Decrypting incoming requests that have the X-Request-Encrypted header
// 2. Encrypting outgoing responses when X-Response-Encrypted header is set
// 3. Using client-specific RSA key pairs for encryption/decryption
type EncryptionMiddleware struct {
	securityService *services.ClientSecurityService
}

// 🆕 NewEncryptionMiddleware creates a new encryption middleware instance.
// Parameters:
//   - securityService: Service for handling client security operations
//
// Returns:
//   - *EncryptionMiddleware: New middleware instance
func NewEncryptionMiddleware(securityService *services.ClientSecurityService) *EncryptionMiddleware {
	return &EncryptionMiddleware{
		securityService: securityService,
	}
}

// 🔒 HandleEncryption returns a Gin middleware function that handles request/response encryption.
// The middleware:
// 1. Checks for client configuration in the context (set by auth middleware)
// 2. Decrypts incoming requests if X-Request-Encrypted header is present
// 3. Encrypts outgoing responses if X-Response-Encrypted header is set
//
// Headers:
//   - X-Request-Encrypted: true - Indicates request body is encrypted
//   - X-Response-Encrypted: true - Indicates response should be encrypted
//
// Encryption Flow:
//
//	Client -> Server:
//	  Plaintext -> ChaCha20 encrypt -> RSA encrypt key -> Base64 -> Send
//	Server Process:
//	  Receive -> Base64 decode -> RSA decrypt -> ChaCha20 decrypt -> HSM encrypt -> Process
//	Server Response:
//	  Response -> HSM decrypt -> ChaCha20 encrypt -> RSA encrypt key -> Base64 -> Send
//	Client Decrypt:
//	  Receive -> Base64 decode -> RSA decrypt -> ChaCha20 decrypt -> Plaintext
//
// Returns:
//   - gin.HandlerFunc: Middleware function for Gin
func (m *EncryptionMiddleware) HandleEncryption() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client config from context (set by auth middleware)
		configAny, exists := c.Get("client_config")
		if !exists {
			c.Next()
			return
		}

		clientConfig, ok := configAny.(*models.ClientConfig)
		if !ok {
			m.sendEncryptedError(c, clientConfig, http.StatusInternalServerError, "invalid client config type")
			return
		}

		// Handle request decryption
		if c.GetHeader("X-Request-Encrypted") == "true" {
			// Ensure request body exists
			if c.Request.Body == nil {
				m.sendEncryptedError(c, clientConfig, http.StatusBadRequest, "empty request body")
				return
			}

			// Read request body
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				m.sendEncryptedError(c, clientConfig, http.StatusBadRequest, "failed to read request body")
				return
			}
			defer c.Request.Body.Close()

			// Base64 decode request body
			decodedBody, err := base64.StdEncoding.DecodeString(string(body))
			if err != nil {
				log.Printf("Failed to base64 decode request body: %v", err)
				m.sendEncryptedError(c, clientConfig, http.StatusBadRequest, "failed to base64 decode request body")
				return
			}

			// Check for empty body
			if len(decodedBody) == 0 {
				m.sendEncryptedError(c, clientConfig, http.StatusBadRequest, "empty request body")
				return
			}

			// Decrypt request body using client's public key
			decryptedBody, err := m.securityService.DecryptFromClient(clientConfig.PrivateKey, decodedBody)
			if err != nil {
				log.Printf("Failed to decrypt request body: %v", err)
				m.sendEncryptedError(c, clientConfig, http.StatusBadRequest, fmt.Sprintf("failed to decrypt request: %v", err))
				return
			}

			// Try to parse as JSON if the decrypted data starts with { or [
			if len(decryptedBody) > 0 && (decryptedBody[0] == '{' || decryptedBody[0] == '[') {
				var jsonBody any
				if err := json.Unmarshal(decryptedBody, &jsonBody); err != nil {
					log.Printf("Failed to parse decrypted body as JSON: %v", err)
					m.sendEncryptedError(c, clientConfig, http.StatusBadRequest, fmt.Sprintf("failed to parse decrypted json body: %v", err))
					return
				}
				// Set the Content-Type to application/json for downstream handlers
				c.Request.Header.Set("Content-Type", "application/json")
			}

			// Replace request body with decrypted data
			c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))
		}

		// Create response writer wrapper to capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			status:         http.StatusOK, // Default status
		}
		c.Writer = writer

		// Process the request
		c.Next()

		// Handle response encryption
		if writer.body != nil {
			var responseData []byte
			var err error

			// First encrypt with HSM key if needed
			if c.GetHeader("X-HSM-Encrypt") == "true" {
				responseData, err = m.securityService.EncryptWithHSM(writer.body)
				if err != nil {
					m.sendEncryptedError(c, clientConfig, http.StatusInternalServerError, fmt.Sprintf("failed to encrypt with HSM: %v", err))
					return
				}
			} else {
				responseData = writer.body
			}

			// Then encrypt for client transport
			encryptedData, err := m.securityService.EncryptForClient(clientConfig.PublicKey, responseData)
			if err != nil {
				m.sendEncryptedError(c, clientConfig, http.StatusInternalServerError, fmt.Sprintf("failed to encrypt response: %v", err))
				return
			}

			// Base64 encode the encrypted data
			encodedData := base64.StdEncoding.EncodeToString(encryptedData)

			// Write encrypted response
			writer.written = true
			c.Header("Content-Type", "application/octet-stream")
			c.Header("X-Response-Encrypted", "true")
			c.String(writer.status, encodedData)
		}
	}
}

// 📝 responseWriter wraps gin.ResponseWriter to capture the response body.
// This allows us to encrypt the response body before sending it to the client.
type responseWriter struct {
	gin.ResponseWriter
	body    []byte
	status  int
	written bool
}

// 📝 Write implements io.Writer to capture response body data.
// Parameters:
//   - b: Bytes to write
//
// Returns:
//   - int: Number of bytes written
//   - error: Error if write failed
func (w *responseWriter) Write(b []byte) (int, error) {
	if w.written {
		return w.ResponseWriter.Write(b)
	}
	w.body = append(w.body, b...)
	return len(b), nil
}

// 📝 WriteHeader captures the response status code.
// Parameters:
//   - status: HTTP status code
func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// 🔒 sendEncryptedError sends an encrypted error response
func (m *EncryptionMiddleware) sendEncryptedError(c *gin.Context, clientConfig *models.ClientConfig, status int, message string) {
	// Create error response
	errorResponse := gin.H{"error": message}
	jsonData, err := json.Marshal(errorResponse)
	if err != nil {
		log.Printf("Failed to marshal error response: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Encrypt error response
	encryptedData, err := m.securityService.EncryptForClient(clientConfig.PublicKey, jsonData)
	if err != nil {
		log.Printf("Failed to encrypt error response: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Base64 encode the encrypted data
	encodedData := base64.StdEncoding.EncodeToString(encryptedData)

	// Send encrypted error response
	c.Header("Content-Type", "application/octet-stream")
	c.Header("X-Response-Encrypted", "true")
	c.String(status, encodedData)
	c.Abort()
}
