package api

import (
	"fmt"
	"net/http"
	"time"

	"xyphos/internal/hsm"
	"xyphos/internal/models"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
)

// 🔐 KMSHandler handles KMS operations
type KMSHandler struct {
	store     store.KMSStore
	userStore store.UserStore
	hsm       hsm.Service
}

// 🆕 NewKMSHandler creates a new KMS handler
func NewKMSHandler(kmsStore store.KMSStore, userStore store.UserStore, hsm hsm.Service) *KMSHandler {
	return &KMSHandler{
		store:     kmsStore,
		userStore: userStore,
		hsm:       hsm,
	}
}

//	@Summary		List all projects
//	@Description	Get a list of all projects
//	@Tags			projects
//	@Produce		json
//	@Success		200	{array}		Project
//	@Failure		401	{object}	ErrorResponse
//	@Router			/projects [get]
func (h *KMSHandler) ListProjects(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	projects, err := h.store.ListProjects(c.Request.Context(), user.(*models.User).ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list projects: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

//	@Summary		Create a new project
//	@Description	Create a new project in Xyphos KMS
//	@Tags			projects
//	@Accept			json
//	@Produce		json
//	@Param			project	body		CreateProjectRequest	true	"Project details"
//	@Success		200		{object}	Project
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/projects [post]
func (h *KMSHandler) CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	project.OwnerID = user.(*models.User).ID
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	if err := h.store.CreateProject(c.Request.Context(), &project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create project: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, project)
}

//	@Summary		Get a project by ID
//	@Description	Get detailed information about a specific project
//	@Tags			projects
//	@Produce		json
//	@Param			project_id	path		string	true	"Project ID"
//	@Success		200			{object}	Project
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Router			/projects/{project_id} [get]
func (h *KMSHandler) GetProject(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	project, err := h.store.GetProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get project: %v", err)})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// 📍 ListLocations lists all locations
func (h *KMSHandler) ListLocations(c *gin.Context) {
	locations, err := h.store.ListLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list locations: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"locations": locations})
}

// 📍 GetLocation gets a location by ID
func (h *KMSHandler) GetLocation(c *gin.Context) {
	locationID := c.Param("locationId")
	if locationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	location, err := h.store.GetLocation(c.Request.Context(), locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get location: %v", err)})
		return
	}

	if location == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	c.JSON(http.StatusOK, location)
}

// 💍 CreateKeyring creates a new keyring
func (h *KMSHandler) CreateKeyring(c *gin.Context) {
	var keyring models.Keyring
	if err := c.ShouldBindJSON(&keyring); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Create store keyring
	storeKeyring := &store.KeyRing{
		ID:        keyring.ID,
		Name:      keyring.Name,
		Tenant:    user.(*models.User).ID,
		CreatedAt: time.Now(),
	}

	if err := h.store.CreateKeyRing(c.Request.Context(), storeKeyring); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create keyring: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, keyring)
}

// 💍 ListKeyrings lists all keyrings
func (h *KMSHandler) ListKeyrings(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	keyrings, err := h.store.ListKeyRings(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list keyrings"})
		return
	}

	c.JSON(http.StatusOK, keyrings)
}

// 💍 GetKeyring gets a keyring by ID
func (h *KMSHandler) GetKeyring(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	keyring, err := h.store.GetKeyRing(c.Request.Context(), c.Param("keyringId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get keyring"})
		return
	}

	if keyring == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Keyring not found"})
		return
	}

	if keyring.Tenant != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, keyring)
}

//	@Summary		Create a new crypto key
//	@Description	Create a new cryptographic key in a keyring
//	@Tags			crypto-keys
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		string					true	"Project ID"
//	@Param			location_id	path		string					true	"Location ID"
//	@Param			keyring_id	path		string					true	"KeyRing ID"
//	@Param			key			body		CreateCryptoKeyRequest	true	"CryptoKey details"
//	@Success		200			{object}	CryptoKey
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys [post]
func (h *KMSHandler) CreateCryptoKey(c *gin.Context) {
	var key models.CryptoKey
	if err := c.ShouldBindJSON(&key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate initial key material
	keyMaterial, err := h.hsm.GenerateKeyMaterial(string(key.Algorithm))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate key material: %v", err)})
		return
	}

	// Create initial version
	version := models.NewCryptoKeyVersion(key.ID, keyMaterial)

	key.CreatedAt = time.Now()
	key.UpdatedAt = time.Now()
	key.Status = models.StatusEnabled
	key.Versions = []models.CryptoKeyVersion{*version}

	if err := h.store.CreateKey(c.Request.Context(), &store.Key{
		ID:        key.ID,
		KeyRing:   key.KeyringID,
		Algorithm: string(key.Algorithm),
		Purpose:   string(key.Purpose),
		State:     string(key.Status),
		CreatedAt: key.CreatedAt,
		Versions: []store.KeyVersion{{
			Version:      1,
			State:        string(version.State),
			CreatedAt:    version.CreatedAt,
			EncryptedKey: keyMaterial,
		}},
		CurrentVersion: 1,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create crypto key: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, key)
}

// 🔑 ListCryptoKeys lists all crypto keys in a keyring
func (h *KMSHandler) ListCryptoKeys(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	keyring, err := h.store.GetKeyRing(c.Request.Context(), c.Param("keyringId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get keyring"})
		return
	}

	if keyring == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Keyring not found"})
		return
	}

	if keyring.Tenant != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	keys, err := h.store.ListKeys(c.Request.Context(), keyring.ID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list keys"})
		return
	}

	c.JSON(http.StatusOK, keys)
}

// 🔑 GetCryptoKey gets a crypto key by ID
func (h *KMSHandler) GetCryptoKey(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	key, err := h.store.GetKey(c.Request.Context(), c.Param("keyId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get key"})
		return
	}

	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}

	if key.Tenant != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, key)
}

// 🔄 RotateCryptoKey rotates a crypto key
func (h *KMSHandler) RotateCryptoKey(c *gin.Context) {
	keyID := c.Param("keyId")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key ID is required"})
		return
	}

	key, err := h.store.GetKey(c.Request.Context(), keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get crypto key: %v", err)})
		return
	}

	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Crypto key not found"})
		return
	}

	// Generate new key material
	newKeyMaterial, err := h.hsm.GenerateKeyMaterial(key.Algorithm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate key material: %v", err)})
		return
	}

	// Create new version
	newVersion := store.KeyVersion{
		Version:      key.CurrentVersion + 1,
		State:        "ENABLED",
		CreatedAt:    time.Now(),
		EncryptedKey: newKeyMaterial,
	}

	key.Versions = append(key.Versions, newVersion)
	key.CurrentVersion = newVersion.Version

	if err := h.store.UpdateKey(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update crypto key: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key rotated successfully", "version": newVersion.Version})
}

//	@Summary		Encrypt data
//	@Description	Encrypt data using a crypto key
//	@Tags			crypto-operations
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		string			true	"Project ID"
//	@Param			location_id	path		string			true	"Location ID"
//	@Param			keyring_id	path		string			true	"KeyRing ID"
//	@Param			key_id		path		string			true	"Key ID"
//	@Param			request		body		EncryptRequest	true	"Data to encrypt"
//	@Success		200			{object}	EncryptResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}:encrypt [post]
func (h *KMSHandler) Encrypt(c *gin.Context) {
	keyID := c.Param("keyId")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key ID is required"})
		return
	}

	var req struct {
		Plaintext []byte `json:"plaintext" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.store.GetKey(c.Request.Context(), keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get crypto key: %v", err)})
		return
	}

	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Crypto key not found"})
		return
	}

	// Get current version
	if len(key.Versions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key has no versions"})
		return
	}
	currentVersion := key.Versions[len(key.Versions)-1]

	// Encrypt the data
	ciphertext, err := h.hsm.Encrypt(currentVersion.EncryptedKey, req.Plaintext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to encrypt data: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ciphertext": ciphertext,
		"keyVersion": currentVersion.Version,
	})
}

//	@Summary		Decrypt data
//	@Description	Decrypt data using a crypto key
//	@Tags			crypto-operations
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		string			true	"Project ID"
//	@Param			location_id	path		string			true	"Location ID"
//	@Param			keyring_id	path		string			true	"KeyRing ID"
//	@Param			key_id		path		string			true	"Key ID"
//	@Param			request		body		DecryptRequest	true	"Data to decrypt"
//	@Success		200			{object}	DecryptResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}:decrypt [post]
func (h *KMSHandler) Decrypt(c *gin.Context) {
	keyID := c.Param("keyId")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key ID is required"})
		return
	}

	var req struct {
		Ciphertext []byte `json:"ciphertext" binding:"required"`
		KeyVersion int    `json:"keyVersion" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.store.GetKey(c.Request.Context(), keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get crypto key: %v", err)})
		return
	}

	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Crypto key not found"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key version not found"})
		return
	}

	// Decrypt the data
	plaintext, err := h.hsm.Decrypt(keyVersion.EncryptedKey, req.Ciphertext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to decrypt data: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plaintext":  plaintext,
		"keyVersion": keyVersion.Version,
	})
}

// 🔑 CreateClientConfig creates a new client configuration
func (h *KMSHandler) CreateClientConfig(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Permissions []string `json:"permissions" binding:"required"`
		ExpiresIn   string   `json:"expires_in" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Parse expiration duration
	duration, err := time.ParseDuration(req.ExpiresIn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expires_in format"})
		return
	}

	// Create client configuration
	config := &models.ClientConfig{
		UserID:      user.(*models.User).ID,
		Name:        req.Name,
		Permissions: req.Permissions,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(duration),
		Status:      models.ClientConfigStatusActive,
	}

	if err := h.userStore.CreateClientConfig(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create client configuration: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// 📋 ListClientConfigs lists all client configurations for a user
func (h *KMSHandler) ListClientConfigs(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	configs, err := h.userStore.ListClientConfigs(c.Request.Context(), user.(*models.User).ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list client configurations: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// 🔍 GetClientConfig gets a client configuration by ID
func (h *KMSHandler) GetClientConfig(c *gin.Context) {
	configID := c.Param("configId")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config ID is required"})
		return
	}

	config, err := h.userStore.GetClientConfig(c.Request.Context(), configID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get client configuration: %v", err)})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client configuration not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// 🔑 RevokeClientConfig revokes a client configuration
func (h *KMSHandler) RevokeClientConfig(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	config, err := h.userStore.GetClientConfig(c.Request.Context(), c.Param("configId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client configuration"})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client configuration not found"})
		return
	}

	if config.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	config.Status = "REVOKED"
	if err := h.userStore.UpdateClientConfig(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke client configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// 👤 getUserFromContext gets the user from the Gin context
func getUserFromContext(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
