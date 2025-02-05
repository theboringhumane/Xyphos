package handlers

import (
	"context"
	"net/http"
	"time"

	"xyphos/internal/api/services"
	"xyphos/internal/models"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
)

// 💍 KeyringHandler handles keyring operations
type KeyringHandler struct {
	keyStore store.Store
}

// NewKeyringHandler creates a new keyring handler
func NewKeyringHandler(keyStore store.Store) *KeyringHandler {
	return &KeyringHandler{
		keyStore: keyStore,
	}
}

// 💍 HandleCreate handles keyring creation
func (h *KeyringHandler) HandleCreate(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get owner from claims
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	owner := user.(*models.User).ID

	// Check if keyring with this name already exists for tenant
	existingKeyrings, err := h.keyStore.ListKeyRings(context.Background(), owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing keyrings"})
		return
	}
	for _, kr := range existingKeyrings {
		if kr.Name == req.Name {
			c.JSON(http.StatusConflict, gin.H{"error": "keyring with this name already exists"})
			return
		}
	}

	// get projectID from context
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	locationID := c.Param("locationId")
	if locationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	// Create a new keyring
	keyring := &store.KeyRing{
		ID:         services.GenerateID(),
		Name:       req.Name,
		Owner:      owner,
		ProjectID:  projectID,
		LocationID: locationID,
		CreatedAt:  time.Now(),
	}

	if err := h.keyStore.CreateKeyRing(context.Background(), keyring); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create keyring"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"name":      keyring.Name,
		"createdAt": keyring.CreatedAt,
	})
}

// 🔍 HandleGet handles keyring retrieval by name
func (h *KeyringHandler) HandleGet(c *gin.Context) {
	name := c.Param("keyringId")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing keyring ID"})
		return
	}

	// Get tenant from claims
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	owner := user.(*models.User).ID

	// Find keyring by name
	keyring, err := h.findByName(owner, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find keyring"})
		return
	}
	if keyring == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "keyring not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        keyring.ID,
		"name":      keyring.Name,
		"createdAt": keyring.CreatedAt,
	})
}

// 📋 HandleList handles listing keyrings for a tenant
func (h *KeyringHandler) HandleList(c *gin.Context) {
	// Get tenant from claims
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	owner := user.(*models.User).ID

	// List keyrings
	keyrings, err := h.keyStore.ListKeyRings(context.Background(), owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keyrings"})
		return
	}

	// Convert to response format
	response := make([]gin.H, len(keyrings))
	for i, kr := range keyrings {
		response[i] = gin.H{
			"name":      kr.Name,
			"createdAt": kr.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to find keyring by name
func (h *KeyringHandler) findByName(owner, name string) (*store.KeyRing, error) {
	keyrings, err := h.keyStore.ListKeyRings(context.Background(), owner)
	if err != nil {
		return nil, err
	}

	for _, kr := range keyrings {
		if kr.Name == name {
			return kr, nil
		}
	}
	return nil, nil
}
