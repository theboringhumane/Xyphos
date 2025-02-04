package models

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a tenant in the KMS system
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	OwnerID     string    `json:"ownerId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewProject creates a new project with default values
func NewProject(name, displayName, ownerID string) *Project {
	now := time.Now()
	return &Project{
		ID:          uuid.New().String(),
		Name:        name,
		DisplayName: displayName,
		OwnerID:     ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Location represents a geographic location where resources can be stored
type Location struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// PredefinedLocations represents the available locations in the system
var PredefinedLocations = []Location{
	{ID: "us-west1", Name: "us-west1", DisplayName: "US West (Oregon)", Description: "US West region in Oregon"},
	{ID: "us-east1", Name: "us-east1", DisplayName: "US East (Virginia)", Description: "US East region in Virginia"},
	{ID: "eu-west1", Name: "eu-west1", DisplayName: "Europe West (Ireland)", Description: "Europe West region in Ireland"},
	{ID: "asia-east1", Name: "asia-east1", DisplayName: "Asia East (Taiwan)", Description: "Asia East region in Taiwan"},
}
