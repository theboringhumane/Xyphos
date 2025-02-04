package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"xyphos/internal/auth"
	"xyphos/internal/hsm"
	"xyphos/internal/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// 🌐 Server represents the API server
type Server struct {
	router      *gin.Engine
	authHandler *auth.Handler
	kmsHandler  *KMSHandler
}

// 🎯 NewServer creates a new API server instance
func NewServer(store store.KMSStore, userStore store.UserStore, hsm hsm.Service) *Server {
	// Create handlers
	authHandler := auth.NewAuthHandler(userStore)
	kmsHandler := NewKMSHandler(store, userStore, hsm)

	// Create router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	server := &Server{
		router:      router,
		authHandler: authHandler,
		kmsHandler:  kmsHandler,
	}

	// Set up routes
	server.setupRoutes()

	return server
}

// 🚀 Start starts the API server
func (s *Server) Start(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed to start server: %v", err)
		}
	}()

	return nil
}

// 🛑 Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	server := &http.Server{
		Addr:    s.router.BasePath(),
		Handler: s.router,
	}

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Disable keep-alives for new connections
	server.SetKeepAlivesEnabled(false)

	// Shut down the server
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}

// 🛣️ setupRoutes sets up the API routes
func (s *Server) setupRoutes() {
	// 📚 Swagger documentation route
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 🏥 Health check route
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// 🔐 GitHub OAuth routes
	s.authHandler.InitGitHubAuthRoutes(s.router)

	// 🔒 Protected API routes
	api := s.router.Group("/api")
	api.Use(s.authHandler.AuthMiddleware())
	{
		// 📦 Project routes
		projects := api.Group("/projects")
		{
			projects.GET("", s.kmsHandler.ListProjects)
			projects.POST("", s.kmsHandler.CreateProject)
			projects.GET("/:projectId", s.kmsHandler.GetProject)

			// 📍 Location routes
			locations := projects.Group("/:projectId/locations")
			{
				locations.GET("", s.kmsHandler.ListLocations)
				locations.GET("/:locationId", s.kmsHandler.GetLocation)

				// 💍 KeyRing routes
				keyrings := locations.Group("/:locationId/keyrings")
				{
					keyrings.POST("", s.kmsHandler.CreateKeyring)
					keyrings.GET("", s.kmsHandler.ListKeyrings)
					keyrings.GET("/:keyringId", s.kmsHandler.GetKeyring)

					// 🔑 CryptoKey routes
					keys := keyrings.Group("/:keyringId/keys")
					{
						keys.POST("", s.kmsHandler.CreateCryptoKey)
						keys.GET("", s.kmsHandler.ListCryptoKeys)
						keys.GET("/:keyId", s.kmsHandler.GetCryptoKey)
						keys.POST("/:keyId:rotate", s.kmsHandler.RotateCryptoKey)
						keys.POST("/:keyId:encrypt", s.kmsHandler.Encrypt)
						keys.POST("/:keyId:decrypt", s.kmsHandler.Decrypt)
					}
				}
			}
		}

		// 🔑 Client configuration routes
		clients := api.Group("/client-configs")
		{
			clients.POST("", s.kmsHandler.CreateClientConfig)
			clients.GET("", s.kmsHandler.ListClientConfigs)
			clients.GET("/:configId", s.kmsHandler.GetClientConfig)
			clients.POST("/:configId/revoke", s.kmsHandler.RevokeClientConfig)
		}
	}
}
