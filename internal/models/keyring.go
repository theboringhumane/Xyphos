package models

import (
	"time"

	"github.com/google/uuid"
)

// Keyring represents a collection of cryptographic keys
type Keyring struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"projectId"`
	LocationID  string    `json:"locationId"`
	Owner       string    `json:"owner"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewKeyring creates a new keyring with default values
func NewKeyring(projectID, locationID, name, description, owner string) *Keyring {
	now := time.Now()
	return &Keyring{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		LocationID:  locationID,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Owner:       owner,
	}
}
