package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"xyphos/internal/api/handlers"
	"xyphos/internal/crypto"
	"xyphos/internal/hsm"
	"xyphos/internal/models"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 🔐 KMSHandler coordinates all KMS operations through specialized handlers
type KMSHandler struct {
	cryptoHandler  *handlers.CryptoHandler
	keyHandler     *handlers.KeyHandler
	keyringHandler *handlers.KeyringHandler
	clientHandler  *handlers.ClientHandler
	store          store.Store
}

// 🆕 NewKMSHandler creates a new KMS handler with all required sub-handlers
func NewKMSHandler(
	kmsStore store.Store,
	userStore store.UserStore,
	hsmService hsm.Service,
	masterKeyConfig crypto.MasterKeyConfig,
) *KMSHandler {

	// 🔐 Initialize master key manager
	masterKeyManager, err := crypto.NewLocationMasterKeyManager(
		masterKeyConfig.StorePath,
		os.Getenv(masterKeyConfig.WrapKeyEnv),
	)
	if err != nil {
		log.Fatalf("❌ Failed to initialize master key manager: %v", err)
	}

	// Initialize handlers in correct order due to dependencies
	keyringHandler := handlers.NewKeyringHandler(kmsStore)
	keyHandler := handlers.NewKeyHandler(kmsStore, hsmService, keyringHandler, masterKeyManager)
	cryptoHandler := handlers.NewCryptoHandler(kmsStore, hsmService, keyHandler, keyringHandler, masterKeyManager)
	clientHandler := handlers.NewClientHandler(userStore)

	return &KMSHandler{
		cryptoHandler:  cryptoHandler,
		keyHandler:     keyHandler,
		keyringHandler: keyringHandler,
		clientHandler:  clientHandler,
		store:          kmsStore,
	}
}

// @Summary		List all projects
// @Description	Get a list of all projects
// @Tags			projects
// @Produce		json
// @Success		200	{array}		Project
// @Failure		401	{object}	ErrorResponse
// @Router			/projects [get]
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

// @Summary		Create a new project
// @Description	Create a new project in Xyphos KMS
// @Tags			projects
// @Accept			json
// @Produce		json
// @Param			project	body		CreateProjectRequest	true	"Project details"
// @Success		200		{object}	Project
// @Failure		400		{object}	ErrorResponse
// @Failure		401		{object}	ErrorResponse
// @Router			/projects [post]
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

	project.ID = uuid.New().String()
	project.OwnerID = user.(*models.User).ID
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	if err := h.store.CreateProject(c.Request.Context(), &project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create project: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// @Summary		Get a project by ID
// @Description	Get detailed information about a specific project
// @Tags			projects
// @Produce		json
// @Param			project_id	path		string	true	"Project ID"
// @Success		200			{object}	Project
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id} [get]
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
	h.keyringHandler.HandleCreate(c)
}

// 📋 ListKeyrings delegates to KeyringHandler
func (h *KMSHandler) ListKeyrings(c *gin.Context) {
	h.keyringHandler.HandleList(c)
}

// 🔍 GetKeyring delegates to KeyringHandler
func (h *KMSHandler) GetKeyring(c *gin.Context) {
	h.keyringHandler.HandleGet(c)
}

// @Summary		Create a new crypto key
// @Description	Create a new cryptographic key in a keyring
// @Tags			crypto-keys
// @Accept			json
// @Produce		json
// @Param			project_id	path		string					true	"Project ID"
// @Param			location_id	path		string					true	"Location ID"
// @Param			keyring_id	path		string					true	"KeyRing ID"
// @Param			key			body		CreateCryptoKeyRequest	true	"CryptoKey details"
// @Success		200			{object}	CryptoKey
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/{tenant_id}/keys [post]

// 🔑 CreateCryptoKey delegates to KeyHandler
func (h *KMSHandler) CreateCryptoKey(c *gin.Context) {
	h.keyHandler.HandleCreate(c)
}

// 📋 ListCryptoKeys delegates to KeyHandler
func (h *KMSHandler) ListCryptoKeys(c *gin.Context) {
	h.keyHandler.HandleList(c)
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

	if key.Owner != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, key)
}

// 🔄 RotateCryptoKey rotates a crypto key
func (h *KMSHandler) RotateCryptoKey(c *gin.Context) {
	h.keyHandler.HandleRotate(c)
}

// @Summary		Encrypt data
// @Description	Encrypt data using a crypto key
// @Tags			crypto-operations
// @Accept			json
// @Produce		json
// @Param			project_id	path		string			true	"Project ID"
// @Param			location_id	path		string			true	"Location ID"
// @Param			keyring_id	path		string			true	"KeyRing ID"
// @Param			key_id		path		string			true	"Key ID"
// @Param			request		body		EncryptRequest	true	"Data to encrypt"
// @Success		200			{object}	EncryptResponse
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/{tenant_id}/keys/{key_id}:encrypt [post]
// 🔐 Encrypt delegates to CryptoHandler
func (h *KMSHandler) Encrypt(c *gin.Context) {
	h.cryptoHandler.HandleEncrypt(c)
}

// @Summary		Decrypt data
// @Description	Decrypt data using a crypto key
// @Tags			crypto-operations
// @Accept			json
// @Produce		json
// @Param			project_id	path		string			true	"Project ID"
// @Param			location_id	path		string			true	"Location ID"
// @Param			keyring_id	path		string			true	"KeyRing ID"
// @Param			key_id		path		string			true	"Key ID"
// @Param			request		body		DecryptRequest	true	"Data to decrypt"
// @Success		200			{object}	DecryptResponse
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/{tenant_id}/keys/{key_id}:decrypt [post]
// 🔓 Decrypt delegates to CryptoHandler
func (h *KMSHandler) Decrypt(c *gin.Context) {
	h.cryptoHandler.HandleDecrypt(c)
}

// 🔑 CreateClientConfig creates a new client configuration
func (h *KMSHandler) CreateClientConfig(c *gin.Context) {
	h.clientHandler.CreateClientConfig(c)
}

// 📋 ListClientConfigs lists all client configurations for a user
func (h *KMSHandler) ListClientConfigs(c *gin.Context) {
	h.clientHandler.ListClientConfigs(c)
}

// 🔍 GetClientConfig gets a client configuration by ID
func (h *KMSHandler) GetClientConfig(c *gin.Context) {
	h.clientHandler.GetClientConfig(c)
}

// 🔑 RevokeClientConfig revokes a client configuration
func (h *KMSHandler) RevokeClientConfig(c *gin.Context) {
	h.clientHandler.RevokeClientConfig(c)
}

// 👤 getUserFromContext gets the user from the Gin context
func getUserFromContext(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
