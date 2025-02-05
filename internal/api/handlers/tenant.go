package handlers

import (
	"context"
	"net/http"
	"time"

	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 👥 TenantHandler handles tenant operations
type TenantHandler struct {
	store store.Store
}

// 🆕 NewTenantHandler creates a new tenant handler
func NewTenantHandler(store store.Store) *TenantHandler {
	return &TenantHandler{
		store: store,
	}
}

// 📝 HandleCreate handles tenant creation
func (h *TenantHandler) HandleCreate(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keyringID := c.Param("keyringId")
	if keyringID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyring ID is required"})
		return
	}

	locationID := c.Param("locationId")
	if locationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location ID is required"})
		return
	}

	tenant := &store.Tenant{
		ID:        uuid.New().String(),
		Name:      req.Name,
		KeyRing:   keyringID,
		Location:  locationID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.store.CreateTenant(context.Background(), tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tenant"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// 📋 HandleList handles listing tenants
func (h *TenantHandler) HandleList(c *gin.Context) {
	keyringID := c.Param("keyringId")
	if keyringID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyring ID is required"})
		return
	}

	tenants, err := h.store.ListTenants(context.Background(), keyringID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tenants"})
		return
	}

	c.JSON(http.StatusOK, tenants)
}

// 🔍 HandleGet handles getting a tenant by ID
func (h *TenantHandler) HandleGet(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	tenant, err := h.store.GetTenant(context.Background(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// 🔑 HandleUpdate handles updating a tenant
func (h *TenantHandler) HandleUpdate(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := h.store.GetTenant(context.Background(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	tenant.Name = req.Name
	tenant.UpdatedAt = time.Now()

	if err := h.store.UpdateTenant(context.Background(), tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tenant"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}
