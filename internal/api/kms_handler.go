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
	tenantHandler  *handlers.TenantHandler
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
	tenantHandler := handlers.NewTenantHandler(kmsStore)

	return &KMSHandler{
		cryptoHandler:  cryptoHandler,
		keyHandler:     keyHandler,
		keyringHandler: keyringHandler,
		tenantHandler:  tenantHandler,
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

// @Summary		List all locations
// @Description	Get a list of all locations
// @Tags			locations
// @Produce		json
// @Success		200	{array}		Location
// @Failure		401	{object}	ErrorResponse
// @Router			/locations [get]
// 📍 ListLocations lists all locations
func (h *KMSHandler) ListLocations(c *gin.Context) {
	locations, err := h.store.ListLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list locations: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"locations": locations})
}

// @Summary		Get a location by ID
// @Description	Get a location by ID
// @Tags			locations
// @Produce		json
// @Param			location_id	path		string	true	"Location ID"
// @Success		200			{object}	Location
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/locations/{location_id} [get]
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

// @Summary		Create a new keyring
// @Description	Create a new keyring in a location
// @Tags			keyrings
// @Accept			json
// @Produce		json
// @Param			location_id	path		string	true	"Location ID"
// @Param			keyring		body		CreateKeyringRequest	true	"Keyring details"
// @Success		200			{object}	KeyRing
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings [post]
// 💍 CreateKeyring creates a new keyring
func (h *KMSHandler) CreateKeyring(c *gin.Context) {
	h.keyringHandler.HandleCreate(c)
}

// @Summary		List all keyrings
// @Description	Get a list of all keyrings in a location
// @Tags			keyrings
// @Produce		json
// @Param			location_id	path		string	true	"Location ID"
// @Success		200			{array}		KeyRing
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings [get]
// 📋 ListKeyrings delegates to KeyringHandler
func (h *KMSHandler) ListKeyrings(c *gin.Context) {
	h.keyringHandler.HandleList(c)
}

// @Summary		Get a keyring by ID
// @Description	Get a keyring by ID
// @Tags			keyrings
// @Produce		json
// @Param			keyring_id	path		string	true	"KeyRing ID"
// @Success		200			{object}	KeyRing
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id} [get]
// 🔍 GetKeyring delegates to KeyringHandler
func (h *KMSHandler) GetKeyring(c *gin.Context) {
	h.keyringHandler.HandleGet(c)
}

// @Summary		Create a new tenant
// @Description	Create a new tenant in a keyring
// @Tags			tenants
// @Accept			json
// @Produce		json
// @Param			project_id	path		string	true	"Project ID"
// @Param			location_id	path		string	true	"Location ID"
// @Param			keyring_id	path		string	true	"KeyRing ID"
// @Param			tenant		body		CreateTenantRequest	true	"Tenant details"
// @Success		200			{object}	Tenant
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/tenants [post]
// 👥 CreateTenant creates a new tenant
func (h *KMSHandler) CreateTenant(c *gin.Context) {
	h.tenantHandler.HandleCreate(c)
}

// @Summary		List all tenants
// @Description	Get a list of all tenants in a keyring
// @Tags			tenants
// @Produce		json
// @Param			project_id	path		string	true	"Project ID"
// @Param			location_id	path		string	true	"Location ID"
// @Param			keyring_id	path		string	true	"KeyRing ID"
// @Success		200			{array}		Tenant
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/tenants [get]
// 📋 ListTenants lists all tenants
func (h *KMSHandler) ListTenants(c *gin.Context) {
	h.tenantHandler.HandleList(c)
}

// @Summary		Get a tenant by ID
// @Description	Get a tenant by ID
// @Tags			tenants
// @Produce		json
// @Param			project_id	path		string	true	"Project ID"
// @Param			location_id	path		string	true	"Location ID"
// @Param			keyring_id	path		string	true	"KeyRing ID"
// @Param			tenant_id	path		string	true	"Tenant ID"
// @Success		200			{object}	Tenant
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/tenants/{tenant_id} [get]
// 🔍 GetTenant gets a tenant by ID
func (h *KMSHandler) GetTenant(c *gin.Context) {
	h.tenantHandler.HandleGet(c)
}

// @Summary		Update a tenant
// @Description	Update a tenant by ID
// @Tags			tenants
// @Accept			json
// @Produce		json
// @Param			project_id	path		string	true	"Project ID"
// @Param			location_id	path		string	true	"Location ID"
// @Param			keyring_id	path		string	true	"KeyRing ID"
// @Param			tenant_id	path		string	true	"Tenant ID"
// @Param			tenant		body		UpdateTenantRequest	true	"Tenant details"
// @Success		200			{object}	Tenant
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/tenants/{tenant_id} [put]
// 🔑 UpdateTenant updates a tenant
func (h *KMSHandler) UpdateTenant(c *gin.Context) {
	h.tenantHandler.HandleUpdate(c)
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

// @Summary		List crypto keys
// @Description	List all crypto keys in a keyring
// @Tags			crypto-keys
// @Produce		json
// @Param			project_id	path		string	true	"Project ID"
// @Param			location_id	path		string	true	"Location ID"
// @Param			keyring_id	path		string	true	"KeyRing ID"
// @Param			tenant_id	path		string	true	"Tenant ID"
// @Success		200			{array}	CryptoKey
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/{tenant_id}/keys [get]
// 📋 ListCryptoKeys delegates to KeyHandler
func (h *KMSHandler) ListCryptoKeys(c *gin.Context) {
	h.keyHandler.HandleList(c)
}

// @Summary		Get a crypto key by ID
// @Description	Get a crypto key by ID
// @Tags			crypto-keys
// @Produce		json
// @Param			key_id		path		string	true	"Key ID"
// @Success		200			{object}	CryptoKey
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/{tenant_id}/keys/{key_id} [get]
// 🔑 GetCryptoKey gets a crypto key by name
func (h *KMSHandler) GetCryptoKey(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	key, err := h.store.GetKeyByName(c.Request.Context(), c.Param("keyId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get key"})
		return
	}

	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}

	tenant := c.Param("tenantId")
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	if key.Tenant != tenant {
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

// @Summary		Create a new client configuration
// @Description	Create a new client configuration
// @Tags			clients
// @Accept			json
// @Produce		json
// @Param			request		body		CreateClientConfigRequest	true	"Client configuration details"
// @Success		200			{object}	ClientConfig
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/clients [post]
// 🔑 CreateClientConfig creates a new client configuration
func (h *KMSHandler) CreateClientConfig(c *gin.Context) {
	h.clientHandler.CreateClientConfig(c)
}

// @Summary		List client configurations
// @Description	List all client configurations for a user
// @Tags			clients
// @Produce		json
// @Success		200			{array}		ClientConfig
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Router			/clients [get]
// 📋 ListClientConfigs lists all client configurations for a user
func (h *KMSHandler) ListClientConfigs(c *gin.Context) {
	h.clientHandler.ListClientConfigs(c)
}

// @Summary		Get a client configuration by ID
// @Description	Get a client configuration by ID
// @Tags			clients
// @Produce		json
// @Param			client_id	path		string	true	"Client ID"
// @Success		200			{object}	ClientConfig
// @Failure		400			{object}	ErrorResponse
// @Failure		401			{object}	ErrorResponse
// @Failure		404			{object}	ErrorResponse
// @Router			/clients/{client_id} [get]
// 🔑 GetClientConfig gets a client configuration by ID
func (h *KMSHandler) GetClientConfig(c *gin.Context) {
	h.clientHandler.GetClientConfig(c)
}

// @Summary		Revoke a client configuration
// @Description	Revoke a client configuration
// @Tags			clients
// @Produce		json
// @Param			client_id	path		string	true	"Client ID"
// @Router			/clients/{client_id} [delete]
// 🔑 RevokeClientConfig revokes a client configuration
func (h *KMSHandler) RevokeClientConfig(c *gin.Context) {
	h.clientHandler.RevokeClientConfig(c)
}

// 👤 getUserFromContext gets the user from the Gin context
func GetUserFromContext(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
