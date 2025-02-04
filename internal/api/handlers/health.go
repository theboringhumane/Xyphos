package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 🏥 Health handles health check requests
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}
