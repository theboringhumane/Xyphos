package crypto

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// 🔐 MasterKeyConfig holds configuration for master key management
type MasterKeyConfig struct {
	StorePath    string // 📂 Base path for storing master keys
	WrapKeyEnv   string // 🔑 Environment variable name for wrap key
	KeySize      int    // 📏 Size of master keys in bytes
	RotationDays int    // 🔄 Number of days before key rotation
}

// 🆕 NewDefaultMasterKeyConfig creates a new config with secure defaults
func NewDefaultMasterKeyConfig() *MasterKeyConfig {
	return &MasterKeyConfig{
		StorePath:    "/var/lib/xyphos/master-keys",
		WrapKeyEnv:   "XYPHOS_WRAP_KEY",
		KeySize:      64, // 512-bit keys
		RotationDays: 90, // 90-day rotation period
	}
}

// ✅ Validate checks if the configuration is valid
func (c *MasterKeyConfig) Validate() error {
	// 🔍 Check store path
	if c.StorePath == "" {
		return fmt.Errorf("❌ store path cannot be empty")
	}

	// 📁 Create store directory if it doesn't exist
	if err := os.MkdirAll(c.StorePath, 0700); err != nil {
		return fmt.Errorf("❌ failed to create store directory: %w", err)
	}

	// 🔒 Check directory permissions
	info, err := os.Stat(c.StorePath)
	if err != nil {
		return fmt.Errorf("❌ failed to stat store directory: %w", err)
	}
	if mode := info.Mode().Perm(); mode != 0700 {
		return fmt.Errorf("❌ insecure store directory permissions: %v", mode)
	}

	// 🔑 Check wrap key environment variable
	if c.WrapKeyEnv == "" {
		return fmt.Errorf("❌ wrap key environment variable name cannot be empty")
	}
	if os.Getenv(c.WrapKeyEnv) == "" {
		return fmt.Errorf("❌ wrap key not found in environment: %s", c.WrapKeyEnv)
	}

	// 📏 Validate key size
	if c.KeySize < 32 {
		return fmt.Errorf("❌ key size must be at least 32 bytes (got %d)", c.KeySize)
	}

	// ⏰ Validate rotation period
	if c.RotationDays < 1 {
		return fmt.Errorf("❌ rotation period must be at least 1 day (got %d)", c.RotationDays)
	}

	return nil
}

// 🔧 WithStorePath sets the store path
func (c *MasterKeyConfig) WithStorePath(path string) *MasterKeyConfig {
	c.StorePath = path
	return c
}

// 🔑 WithWrapKeyEnv sets the wrap key environment variable name
func (c *MasterKeyConfig) WithWrapKeyEnv(env string) *MasterKeyConfig {
	c.WrapKeyEnv = env
	return c
}

// 📏 WithKeySize sets the key size in bytes
func (c *MasterKeyConfig) WithKeySize(size int) *MasterKeyConfig {
	c.KeySize = size
	return c
}

// 🔄 WithRotationDays sets the rotation period in days
func (c *MasterKeyConfig) WithRotationDays(days int) *MasterKeyConfig {
	c.RotationDays = days
	return c
}

// 📂 GetKeyPath returns the full path for a location's key file
func (c *MasterKeyConfig) GetKeyPath(location string) string {
	return filepath.Join(c.StorePath, fmt.Sprintf("%s.key", location))
}

// 🔐 GetWrapKey retrieves the wrap key from environment
func (c *MasterKeyConfig) GetWrapKey() ([]byte, error) {
	wrapKeyHex := os.Getenv(c.WrapKeyEnv)
	if wrapKeyHex == "" {
		return nil, fmt.Errorf("❌ wrap key not found in environment: %s", c.WrapKeyEnv)
	}

	wrapKey, err := hex.DecodeString(wrapKeyHex)
	if err != nil {
		return nil, fmt.Errorf("❌ invalid wrap key format: %w", err)
	}

	return wrapKey, nil
}
