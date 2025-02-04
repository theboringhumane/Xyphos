// Package main Xyphos Server
//
//	@title						Xyphos API
//	@version					1.0
//	@description				A secure multi-tenant key management system that provides cryptographic operations and key management capabilities.
//	@termsOfService				https://xyphos.io/terms
//
//	@contact.name				API Support
//	@contact.url				https://xyphos.io/support
//	@contact.email				support@xyphos.io
//
//	@license.name				MIT
//	@license.url				https://github.com/yourusername/xyphos/blob/main/LICENSE
//
//	@host						localhost:8080
//	@BasePath					/api
//	@schemes					http https
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"xyphos/internal/api"
	"xyphos/internal/backup"
	"xyphos/internal/hsm"
	"xyphos/internal/store"
)

// 🚀 Main entry point for the Xyphos server
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 🔐 Initialize the HSM service
	hsmService, err := hsm.NewSoftwareHSM()
	if err != nil {
		log.Fatalf("Failed to initialize HSM service: %v", err)
	}

	// 📦 Initialize the store
	dataStore, err := store.NewBadgerStore("./data")
	if err != nil {
		log.Fatalf("Failed to initialize data store: %v", err)
	}
	defer dataStore.Close()

	// 💾 Initialize backup service
	backupService, err := backup.NewBackupService(backup.BackupConfig{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("R2_ACCESS_KEY_SECRET"),
		BucketName:      os.Getenv("R2_BUCKET_NAME"),
		BackupInterval:  os.Getenv("BACKUP_INTERVAL"),
		DataDir:         "./data",
	}, dataStore.DB())
	if err != nil {
		log.Printf("⚠️ Warning: Failed to initialize backup service: %v", err)
	} else {
		// Start backup scheduler in a goroutine
		go backupService.StartBackupScheduler(ctx)
	}

	// 🌐 Initialize and start the API server
	server := api.NewServer(dataStore, dataStore, hsmService)
	go func() {
		if err := server.Start(":8080"); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// 🛑 Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
