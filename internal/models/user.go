package models

import (
	"fmt"
	"time"
)

// 👤 User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	GithubID  int       `json:"github_id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 🔑 ClientConfig represents API client configuration
type ClientConfig struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Name         string     `json:"name"`
	ClientID     string     `json:"client_id"`     // Unique client identifier
	ClientSecret string     `json:"client_secret"` // Secret for client authentication
	PublicKey    string     `json:"public_key"`    // Client's public key for request encryption
	PrivateKey   string     `json:"-"`             // Client's private key (only sent once during creation)
	KeyAlgorithm string     `json:"key_algorithm"` // Algorithm used for key pair (e.g., "RSA-4096")
	Permissions  []string   `json:"permissions"`   // read:keys, write:keys, etc.
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	LastUsedAt   time.Time  `json:"last_used_at"`
	Status       string     `json:"status"`       // active, revoked, expired
	IPWhitelist  []string   `json:"ip_whitelist"` // Optional IP restrictions
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`

	// 🔐 Request Encryption Settings
	RequestEncryption bool   `json:"request_encryption"` // Whether to enforce request encryption
	EncryptionAlg     string `json:"encryption_alg"`     // RSA-OAEP-256
	SigningAlg        string `json:"signing_alg"`        // RSA-PSS-SHA256
}

// 🔐 ClientConfigStatus represents the possible states of a client configuration
const (
	ClientConfigStatusActive  = "active"
	ClientConfigStatusRevoked = "revoked"
	ClientConfigStatusExpired = "expired"
)

// 📝 ClientConfigPermission defines available permissions
const (
	PermissionReadKeys      = "read:keys"
	PermissionWriteKeys     = "write:keys"
	PermissionRotateKeys    = "rotate:keys"
	PermissionManageConfig  = "manage:config"
	PermissionViewAuditLogs = "view:audit_logs"
)

// 🔍 ValidatePermissions checks if all permissions are valid
func (c *ClientConfig) ValidatePermissions() error {
	validPerms := map[string]bool{
		PermissionReadKeys:      true,
		PermissionWriteKeys:     true,
		PermissionRotateKeys:    true,
		PermissionManageConfig:  true,
		PermissionViewAuditLogs: true,
	}

	for _, perm := range c.Permissions {
		if !validPerms[perm] {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}
	return nil
}

// ⏰ IsExpired checks if the client config has expired
func (c *ClientConfig) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// ❌ IsRevoked checks if the client config has been revoked
func (c *ClientConfig) IsRevoked() bool {
	return c.RevokedAt != nil
}

// ✅ IsActive checks if the client config is active
func (c *ClientConfig) IsActive() bool {
	return !c.IsExpired() && !c.IsRevoked() && c.Status == ClientConfigStatusActive
}

// 🔄 UpdateLastUsed updates the last used timestamp
func (c *ClientConfig) UpdateLastUsed() {
	c.LastUsedAt = time.Now()
}
