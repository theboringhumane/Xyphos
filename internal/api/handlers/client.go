package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"xyphos/internal/models"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// 🔑 ClientHandler handles client configuration operations
type ClientHandler struct {
	userStore store.UserStore
	limiter   *rate.Limiter
}

// 🆕 Create new client handler
func NewClientHandler(userStore store.UserStore) *ClientHandler {
	return &ClientHandler{
		userStore: userStore,
		limiter:   rate.NewLimiter(rate.Every(time.Second), 10), // 10 requests per second
	}
}

// 📝 Create client configuration request
type CreateClientConfigRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
	ExpiresIn   string   `json:"expires_in" binding:"required"` // e.g., "24h", "7d", "30d"
}

// 🔒 ValidatePermissions validates the requested permissions
func (h *ClientHandler) ValidatePermissions(permissions []string) error {
	validPermissions := map[string]bool{
		"encrypt":     true,
		"decrypt":     true,
		"create_key":  true,
		"list_keys":   true,
		"rotate_key":  true,
		"delete_key":  true,
		"create_ring": true,
		"list_rings":  true,
		"delete_ring": true,
	}

	for _, p := range permissions {
		if !validPermissions[p] {
			return fmt.Errorf("invalid permission: %s", p)
		}
	}
	return nil
}

// 📝 Create client configuration
func (h *ClientHandler) CreateClientConfig(c *gin.Context) {
	// Apply rate limiting
	if !h.limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		return
	}

	var req CreateClientConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate permissions
	if err := h.ValidatePermissions(req.Permissions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Parse expiration duration
	duration, err := time.ParseDuration(req.ExpiresIn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expires_in format"})
		return
	}

	// Enforce maximum expiration time (30 days)
	maxDuration := 30 * 24 * time.Hour
	if duration > maxDuration {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum expiration time is 30 days"})
		return
	}

	// Create client configuration
	config := &models.ClientConfig{
		UserID:      user.(*models.User).ID,
		Name:        req.Name,
		Permissions: req.Permissions,
		ExpiresAt:   time.Now().Add(duration),
		Status:      models.ClientConfigStatusActive,
		LastUsedAt:  time.Now(),
	}

	if err := h.userStore.CreateClientConfig(context.Background(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client configuration"})
		return
	}

	// Remove sensitive data from response
	response := gin.H{
		"name":        config.Name,
		"permissions": config.Permissions,
		"expiresAt":   config.ExpiresAt,
		"clientId":    config.ClientID,
		"status":      config.Status,
	}

	c.JSON(http.StatusCreated, response)
}

// 📋 List client configurations
func (h *ClientHandler) ListClientConfigs(c *gin.Context) {
	// Apply rate limiting
	if !h.limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	configs, err := h.userStore.ListClientConfigs(context.Background(), user.(*models.User).ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list client configurations"})
		return
	}

	// Convert to response format, excluding sensitive data
	response := make([]gin.H, len(configs))
	for i, config := range configs {
		response[i] = gin.H{
			"name":        config.Name,
			"permissions": config.Permissions,
			"expiresAt":   config.ExpiresAt,
			"clientId":    config.ClientID,
			"status":      config.Status,
			"lastUsedAt":  config.LastUsedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"configs": response})
}

// 🔍 Get client configuration by client ID
func (h *ClientHandler) GetClientConfig(c *gin.Context) {
	// Apply rate limiting
	if !h.limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}

	config, err := h.userStore.GetClientConfigByClientID(context.Background(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client configuration"})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client configuration not found"})
		return
	}

	// Check if the configuration has expired
	if time.Now().After(config.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Client configuration has expired"})
		return
	}

	// Check if the configuration is active
	if config.Status != models.ClientConfigStatusActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Client configuration is not active"})
		return
	}

	// Update last used time
	config.LastUsedAt = time.Now()
	if err := h.userStore.UpdateClientConfig(context.Background(), config); err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to update last used time: %v", err)
	}

	// Remove sensitive data from response
	response := gin.H{
		"name":        config.Name,
		"permissions": config.Permissions,
		"expiresAt":   config.ExpiresAt,
		"clientId":    config.ClientID,
		"status":      config.Status,
		"lastUsedAt":  config.LastUsedAt,
	}

	c.JSON(http.StatusOK, response)
}
