package services

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

var encoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

// 🆔 GenerateID generates a unique ID for resources
func GenerateID() string {
	// Generate 16 random bytes
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err) // This should never happen
	}

	// Encode using base32 for URL-safe IDs
	return strings.ToLower(encoding.EncodeToString(randomBytes))
}
