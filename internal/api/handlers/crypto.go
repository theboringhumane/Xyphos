package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"xyphos/internal/crypto"
	"xyphos/internal/hsm"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
)

// 🔐 CryptoHandler handles cryptographic operations
type CryptoHandler struct {
	keyStore         store.Store
	hsmService       hsm.Service
	keyHandler       *KeyHandler
	keyringHandler   *KeyringHandler
	masterKeyManager *crypto.LocationMasterKeyManager
}

// 🆕 NewCryptoHandler creates a new crypto handler
func NewCryptoHandler(keyStore store.Store, hsmService hsm.Service, keyHandler *KeyHandler, keyringHandler *KeyringHandler, masterKeyManager *crypto.LocationMasterKeyManager) *CryptoHandler {
	return &CryptoHandler{
		keyStore:         keyStore,
		hsmService:       hsmService,
		keyHandler:       keyHandler,
		keyringHandler:   keyringHandler,
		masterKeyManager: masterKeyManager,
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

	// get tenant from request
	tenant := c.Param("tenantId")
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant"})
		return
	}

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
	if currentVersion.State != "ENABLED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current key version is not enabled"})
		return
	}

	// Decode plaintext from base64
	plaintextBytes, err := base64.StdEncoding.DecodeString(req.Plaintext)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plaintext encoding"})
		return
	}

	// get location from request
	location := c.Param("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing location"})
		return
	}

	// Get master key
	masterKey, err := h.masterKeyManager.GetLocationMasterKey(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get master key"})
		return
	}

	// Unwrap the key using HSM's master key
	keyMaterial, err := h.hsmService.UnwrapKey(masterKey, currentVersion.EncryptedKey)
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
		"keyId":      key.ID,
		"algorithm":  key.Algorithm,
		"state":      currentVersion.State,
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
		KeyID      string `json:"keyId" binding:"required"`
		KeyVersion int    `json:"keyVersion" binding:"required"`
		Purpose    string `json:"purpose" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get tenant from request
	tenant := c.Param("tenantId")
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant"})
		return
	}

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

	// Find key by ID
	key, err := h.keyHandler.findKey(req.KeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find key"})
		return
	}
	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no key found for ID"})
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

	// Allow decryption with both ENABLED and DEPRECATED versions
	if keyVersion.State != "ENABLED" && keyVersion.State != "DEPRECATED" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("key version %d is in state %s and cannot be used for decryption",
				req.KeyVersion, keyVersion.State),
		})
		return
	}

	// Decode ciphertext from base64
	ciphertextBytes, err := base64.StdEncoding.DecodeString(req.Ciphertext)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ciphertext encoding"})
		return
	}

	// get location from request
	location := c.Param("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing location"})
		return
	}

	// Get master key
	masterKey, err := h.masterKeyManager.GetLocationMasterKey(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get master key"})
		return
	}

	// Unwrap the key using HSM's master key
	keyMaterial, err := h.hsmService.UnwrapKey(masterKey, keyVersion.EncryptedKey)
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
		"state":      keyVersion.State,
		"keyId":      key.ID,
	})
}
