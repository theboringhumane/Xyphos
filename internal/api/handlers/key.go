package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"xyphos/internal/api/services"
	"xyphos/internal/crypto"
	"xyphos/internal/hsm"
	"xyphos/internal/models"
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
	// 📝 Log start of key creation
	fmt.Println("🔄 Starting key creation process")

	keyringName := c.Param("keyringId")

	if keyringName == "" {
		fmt.Println("❌ Missing keyring name")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing keyring name"})
		return
	}

	fmt.Printf("📋 Using keyring: %s\n", keyringName)

	var req struct {
		Name           string `json:"name" binding:"required"`
		Algorithm      string `json:"algorithm" binding:"required"`
		Purpose        string `json:"purpose" binding:"required"`
		RotationPeriod int    `json:"rotation_period" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ Invalid request body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("📋 Request params - Algorithm: %s, Purpose: %s, Rotation Period: %d\n",
		req.Algorithm, req.Purpose, req.RotationPeriod)

	owner, ok := c.Get("user")
	if !ok {
		fmt.Println("❌ Missing user")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user"})
		return
	}

	user := owner.(*models.User)

	// get tenant from request
	tenant := c.Param("tenantId")
	if tenant == "" {
		fmt.Println("❌ Missing tenant")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant"})
		return
	}

	fmt.Printf("👤 Using tenant: %s\n", tenant)

	// Find keyring by name
	fmt.Println("🔍 Finding keyring")

	keyring, err := h.keyringHandler.findByName(user.ID, keyringName)
	if err != nil {
		fmt.Printf("❌ Failed to find keyring: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find keyring"})
		return
	}

	if keyring == nil {
		fmt.Println("❌ Keyring not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "keyring not found"})
		return
	}

	fmt.Printf("✅ Found keyring with ID: %s\n", keyring.ID)

	// CHECK IF KEY EXISTS WITH SAME NAME
	key_exists, err := h.keyStore.GetKeyByName(context.Background(), req.Name)
	if err != nil {
		fmt.Println("✅ Key does not exist", err)
	}

	if key_exists != nil {
		fmt.Println("❌ Key already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "key already exists"})
		return
	}

	// Generate key material
	fmt.Println("🔐 Generating key material")
	keyMaterial, err := h.hsmService.GenerateKeyMaterial(req.Algorithm)
	if err != nil {
		fmt.Printf("❌ Failed to generate key material: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key material"})
		return
	}
	fmt.Println("✅ Key material generated successfully")

	// get location from request
	location := c.Param("locationId")
	if location == "" {
		fmt.Println("❌ Missing location")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing location"})
		return
	}
	fmt.Printf("📍 Using location: %s\n", location)

	// Get master key
	fmt.Println("🔑 Getting master key")
	masterKey, err := h.masterKeyManager.GetLocationMasterKey(location)
	if err != nil {
		fmt.Printf("❌ Failed to get master key: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get master key"})
		return
	}
	fmt.Println("✅ Master key retrieved successfully")

	// Wrap key material with master key
	fmt.Println("🔒 Wrapping key material")
	wrappedKey, err := h.hsmService.WrapKey(masterKey, keyMaterial)
	if err != nil {
		fmt.Printf("❌ Failed to wrap key: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to wrap key"})
		return
	}
	fmt.Println("✅ Key wrapped successfully")

	// Create initial version
	fmt.Println("📝 Creating initial key version")
	initialVersion := store.KeyVersion{
		Version:      1,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		RotationTime: time.Now().Add(time.Duration(req.RotationPeriod) * time.Second),
		EncryptedKey: wrappedKey,
	}

	// Create key record
	fmt.Println("📝 Creating key record")
	key := &store.Key{
		ID:             services.GenerateID(),
		Tenant:         tenant,
		Name:           req.Name,
		Algorithm:      req.Algorithm,
		Purpose:        req.Purpose,
		CreatedAt:      time.Now(),
		Versions:       []store.KeyVersion{initialVersion},
		CurrentVersion: 1,
		KeyRing:        keyringName,
		NextRotation:   initialVersion.RotationTime,
		RotationPeriod: time.Duration(req.RotationPeriod) * time.Second,
	}

	if err := h.keyStore.CreateKey(context.Background(), key); err != nil {
		fmt.Printf("❌ Failed to create key: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
		return
	}

	fmt.Printf("✅ Key created successfully with ID: %s\n", key.ID)

	c.JSON(http.StatusCreated, gin.H{
		"algorithm":      key.Algorithm,
		"purpose":        key.Purpose,
		"createdAt":      key.CreatedAt,
		"currentVersion": "1",
	})
	fmt.Println("✅ Key creation process completed")
}

// 📋 HandleList handles listing keys in a keyring
func (h *KeyHandler) HandleList(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user"})
		return
	}

	userID := user.(*models.User).ID

	keyringName := c.Param("keyringId")
	if keyringName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing keyring name"})
		return
	}

	// get tenant from request
	tenant := c.Param("tenantId")
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant"})
		return
	}

	// Find keyring by name
	keyring, err := h.keyringHandler.findByName(userID, keyringName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find keyring"})
		return
	}
	if keyring == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "keyring not found"})
		return
	}

	// List keys
	keys, err := h.keyStore.ListKeys(context.Background(), keyring.Name, tenant)
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

	// get tenant from request
	tenant := c.Param("tenantId")
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant"})
		return
	}

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
func (h *KeyHandler) findActiveKeyForPurpose(keyringID, purpose, tenant string) (*store.Key, error) {
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

// 🔑 findKey finds a key by ID
func (h *KeyHandler) findKey(keyID string) (*store.Key, error) {
	return h.keyStore.GetKey(context.Background(), keyID)
}
