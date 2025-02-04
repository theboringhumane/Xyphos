package handlers

import (
	"encoding/base64"
	"net/http"
	"xyphos/internal/auth"
	"xyphos/internal/hsm"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
)

// 🔐 CryptoHandler handles cryptographic operations
type CryptoHandler struct {
	keyStore       store.Store
	hsmService     hsm.Service
	keyHandler     *KeyHandler
	keyringHandler *KeyringHandler
	masterKey      []byte // Store master key during initialization
}

// 🆕 NewCryptoHandler creates a new crypto handler
func NewCryptoHandler(keyStore store.Store, hsmService hsm.Service, keyHandler *KeyHandler, keyringHandler *KeyringHandler, masterKey []byte) *CryptoHandler {
	return &CryptoHandler{
		keyStore:       keyStore,
		hsmService:     hsmService,
		keyHandler:     keyHandler,
		keyringHandler: keyringHandler,
		masterKey:      masterKey,
	}
}

// 🔒 HandleEncrypt handles encryption requests
func (h *CryptoHandler) HandleEncrypt(c *gin.Context) {
	keyringName := c.Param("name")
	if keyringName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing keyring name"})
		return
	}

	var req struct {
		Plaintext  string `json:"plaintext" binding:"required"`
		KeyPurpose string `json:"purpose" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant from claims
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	tenant := claims.(*auth.Claims).Tenant

	// Find keyring by name
	keyring, err := h.keyringHandler.findByName(tenant, keyringName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find keyring"})
		return
	}
	if keyring == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "keyring not found"})
		return
	}

	// Find active key for the purpose
	key, err := h.keyHandler.findActiveKeyForPurpose(keyring.ID, tenant, req.KeyPurpose)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find active key"})
		return
	}
	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active key found for purpose"})
		return
	}

	// Get current version
	if len(key.Versions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key has no versions"})
		return
	}
	currentVersion := key.Versions[len(key.Versions)-1]

	// Decode plaintext from base64
	plaintextBytes, err := base64.StdEncoding.DecodeString(req.Plaintext)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plaintext encoding"})
		return
	}

	// Unwrap the key using HSM's master key
	keyMaterial, err := h.hsmService.UnwrapKey(h.masterKey, currentVersion.EncryptedKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unwrap key"})
		return
	}

	// Encrypt the data
	ciphertext, err := h.hsmService.Encrypt(keyMaterial, plaintextBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt data"})
		return
	}

	// Encode result as base64
	result := base64.StdEncoding.EncodeToString(ciphertext)

	c.JSON(http.StatusOK, gin.H{
		"ciphertext": result,
		"keyVersion": currentVersion.Version,
		"keyring":    keyringName,
	})
}

// 🔓 HandleDecrypt handles decryption requests
func (h *CryptoHandler) HandleDecrypt(c *gin.Context) {
	keyringName := c.Param("name")
	if keyringName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing keyring name"})
		return
	}

	var req struct {
		Ciphertext string `json:"ciphertext" binding:"required"`
		KeyVersion int    `json:"keyVersion" binding:"required"`
		Purpose    string `json:"purpose" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant from claims
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	tenant := claims.(*auth.Claims).Tenant

	// Find keyring by name
	keyring, err := h.keyringHandler.findByName(tenant, keyringName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find keyring"})
		return
	}
	if keyring == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "keyring not found"})
		return
	}

	// Find key for purpose
	key, err := h.keyHandler.findActiveKeyForPurpose(keyring.ID, tenant, req.Purpose)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find key"})
		return
	}
	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no key found for purpose"})
		return
	}

	// Find the specified version
	var keyVersion *store.KeyVersion
	for i := range key.Versions {
		if key.Versions[i].Version == req.KeyVersion {
			keyVersion = &key.Versions[i]
			break
		}
	}

	if keyVersion == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key version not found"})
		return
	}

	// Allow decryption with deprecated keys
	if keyVersion.State != "ENABLED" && keyVersion.State != "DEPRECATED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key version is not valid for decryption"})
		return
	}

	// Decode ciphertext from base64
	ciphertextBytes, err := base64.StdEncoding.DecodeString(req.Ciphertext)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ciphertext encoding"})
		return
	}

	// Unwrap the key using HSM's master key
	keyMaterial, err := h.hsmService.UnwrapKey(h.masterKey, keyVersion.EncryptedKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unwrap key"})
		return
	}

	// Decrypt the data
	plaintext, err := h.hsmService.Decrypt(keyMaterial, ciphertextBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt data"})
		return
	}

	// Encode result as base64
	result := base64.StdEncoding.EncodeToString(plaintext)

	c.JSON(http.StatusOK, gin.H{
		"plaintext":  result,
		"keyVersion": keyVersion.Version,
	})
}
