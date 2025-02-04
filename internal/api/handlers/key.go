package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"xyphos/internal/api/services"
	"xyphos/internal/auth"
	"xyphos/internal/crypto"
	"xyphos/internal/hsm"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
)

// 🔑 KeyHandler handles key operations
type KeyHandler struct {
	keyStore         store.Store
	hsmService       hsm.Service
	keyringHandler   *KeyringHandler
	masterKeyManager *crypto.LocationMasterKeyManager
}

// 🆕 NewKeyHandler creates a new key handler
func NewKeyHandler(keyStore store.Store, hsmService hsm.Service, keyringHandler *KeyringHandler, masterKeyManager *crypto.LocationMasterKeyManager) *KeyHandler {
	return &KeyHandler{
		keyStore:         keyStore,
		hsmService:       hsmService,
		keyringHandler:   keyringHandler,
		masterKeyManager: masterKeyManager,
	}
}

// 🔑 HandleCreate handles key creation
func (h *KeyHandler) HandleCreate(c *gin.Context) {
	keyringName := c.Param("name")
	if keyringName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing keyring name"})
		return
	}

	var req struct {
		Algorithm      string `json:"algorithm" binding:"required"`
		Purpose        string `json:"purpose" binding:"required"`
		RotationPeriod int    `json:"rotation_period" binding:"required"`
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

	// Generate key material
	keyMaterial, err := h.hsmService.GenerateKeyMaterial(req.Algorithm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key material"})
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

	// Wrap key material with master key
	wrappedKey, err := h.hsmService.WrapKey(masterKey, keyMaterial)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to wrap key"})
		return
	}

	// Create initial version
	initialVersion := store.KeyVersion{
		Version:      1,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(time.Duration(req.RotationPeriod) * time.Second),
		EncryptedKey: wrappedKey,
	}

	// Create key record
	key := &store.Key{
		ID:             services.GenerateID(),
		KeyRing:        keyring.ID,
		Tenant:         tenant,
		Algorithm:      req.Algorithm,
		Purpose:        req.Purpose,
		State:          "ENABLED",
		CreatedAt:      time.Now(),
		Versions:       []store.KeyVersion{initialVersion},
		CurrentVersion: 1,
	}

	if err := h.keyStore.CreateKey(context.Background(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"algorithm":      key.Algorithm,
		"purpose":        key.Purpose,
		"state":          key.State,
		"createdAt":      key.CreatedAt,
		"currentVersion": "1",
	})
}

// 📋 HandleList handles listing keys in a keyring
func (h *KeyHandler) HandleList(c *gin.Context) {
	keyringName := c.Param("name")
	if keyringName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing keyring name"})
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

	// List keys
	keys, err := h.keyStore.ListKeys(context.Background(), keyring.ID, tenant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keys"})
		return
	}

	// Convert to response format
	response := make([]gin.H, len(keys))
	for i, key := range keys {
		currentVersion := "NONE"
		if len(key.Versions) > 0 {
			currentVersion = fmt.Sprintf("%d", key.CurrentVersion)
		}
		response[i] = gin.H{
			"algorithm":      key.Algorithm,
			"purpose":        key.Purpose,
			"state":          key.State,
			"createdAt":      key.CreatedAt,
			"currentVersion": currentVersion,
		}
	}

	c.JSON(http.StatusOK, response)
}

// 🔄 HandleRotate handles key rotation
func (h *KeyHandler) HandleRotate(c *gin.Context) {
	keyID := c.Param("id")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing key id"})
		return
	}

	// Get tenant from claims
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	tenant := claims.(*auth.Claims).Tenant

	// Get existing key
	key, err := h.keyStore.GetKey(context.Background(), keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get key"})
		return
	}
	if key == nil || key.Tenant != tenant {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	// Generate new key material
	keyMaterial, err := h.hsmService.GenerateKeyMaterial(key.Algorithm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key material"})
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

	// Wrap new key material
	wrappedKey, err := h.hsmService.WrapKey(masterKey, keyMaterial)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to wrap key"})
		return
	}

	// Create new version
	newVersion := store.KeyVersion{
		Version:      key.CurrentVersion + 1,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(time.Duration(24) * time.Hour), // Default 24h rotation period
		EncryptedKey: wrappedKey,
	}

	// Update previous version state to DEPRECATED
	for i := range key.Versions {
		if key.Versions[i].Version == key.CurrentVersion {
			key.Versions[i].State = "DEPRECATED"
			break
		}
	}

	// Add new version and update current version
	key.Versions = append(key.Versions, newVersion)
	key.CurrentVersion = newVersion.Version

	// Save updated key
	if err := h.keyStore.UpdateKey(context.Background(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Key rotated successfully",
		"currentVersion": newVersion.Version,
	})
}

// Helper function to find active key for purpose
func (h *KeyHandler) findActiveKeyForPurpose(keyringID, tenant, purpose string) (*store.Key, error) {
	keys, err := h.keyStore.ListKeys(context.Background(), keyringID, tenant)
	if err != nil {
		return nil, err
	}

	// Find the most recently created active key for the purpose
	var mostRecent *store.Key
	for _, k := range keys {
		if k.Purpose == purpose && len(k.Versions) > 0 {
			lastVersion := k.Versions[len(k.Versions)-1]
			if lastVersion.State == "ENABLED" {
				if mostRecent == nil || k.CreatedAt.After(mostRecent.CreatedAt) {
					mostRecent = k
				}
			}
		}
	}
	return mostRecent, nil
}
