package store

import (
	"context"
	"xyphos/internal/models"
)

// 👥 UserStore defines operations for managing users and client configurations
type UserStore interface {
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByGithubID(ctx context.Context, githubID int) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error

	// Client configuration operations
	CreateClientConfig(ctx context.Context, config *models.ClientConfig) error
	GetClientConfig(ctx context.Context, id string) (*models.ClientConfig, error)
	GetClientConfigByClientID(ctx context.Context, clientID string) (*models.ClientConfig, error)
	ListClientConfigs(ctx context.Context, userID string) ([]*models.ClientConfig, error)
	UpdateClientConfig(ctx context.Context, config *models.ClientConfig) error

	// Close closes the store
	Close() error
}
