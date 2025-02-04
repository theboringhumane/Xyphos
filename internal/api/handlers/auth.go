package handlers

import (
	"net/http"
	"strings"

	"xyphos/internal/auth"

	"github.com/gin-gonic/gin"
)

// 🔐 AuthHandler handles authentication related operations
type AuthHandler struct {
	authService auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// 🎫 HandleToken handles OAuth token requests
func (h *AuthHandler) HandleToken(c *gin.Context) {
	var req struct {
		ClientID     string `json:"client_id" binding:"required"`
		ClientSecret string `json:"client_secret" binding:"required"`
		GrantType    string `json:"grant_type" binding:"required,eq=client_credentials"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate credentials
	client, err := h.authService.ValidateCredentials(req.ClientID, req.ClientSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

// 🔑 Middleware handles authentication middleware
func (h *AuthHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for token endpoint
		if c.Request.URL.Path == "/oauth/token" {
			c.Next()
			return
		}

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		// Validate token
		claims, err := h.authService.ValidateToken(tokenParts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Store claims in context
		c.Set("claims", claims)
		c.Next()
	}
}
