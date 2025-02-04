package hsm

import (
	"bytes"
	"testing"
)

// 🧪 TestNewSoftwareHSM tests HSM creation
func TestNewSoftwareHSM(t *testing.T) {
	hsm, err := NewSoftwareHSM()
	if err != nil {
		t.Fatalf("Failed to create HSM: %v", err)
	}
	if hsm == nil {
		t.Fatal("HSM is nil")
	}
}

// 🧪 TestGenerateKeyMaterial tests key generation
func TestGenerateKeyMaterial(t *testing.T) {
	hsm, err := NewSoftwareHSM()
	if err != nil {
		t.Fatalf("Failed to create HSM: %v", err)
	}

	tests := []struct {
		name      string
		algorithm string
		wantSize  int
		wantErr   bool
	}{
		{
			name:      "AES-256-GCM",
			algorithm: "AES-256-GCM",
			wantSize:  32,
			wantErr:   false,
		},
		{
			name:      "RSA-OAEP-4096",
			algorithm: "RSA-OAEP-4096",
			wantSize:  512,
			wantErr:   false,
		},
		{
			name:      "ECDSA-P256",
			algorithm: "ECDSA-P256",
			wantSize:  32,
			wantErr:   false,
		},
		{
			name:      "Invalid algorithm",
			algorithm: "INVALID",
			wantSize:  0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := hsm.GenerateKeyMaterial(tt.algorithm)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateKeyMaterial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(key) != tt.wantSize {
				t.Errorf("GenerateKeyMaterial() key size = %v, want %v", len(key), tt.wantSize)
			}
		})
	}
}

// 🧪 TestEncryptDecrypt tests data encryption and decryption
func TestEncryptDecrypt(t *testing.T) {
	hsm, err := NewSoftwareHSM()
	if err != nil {
		t.Fatalf("Failed to create HSM: %v", err)
	}

	// Generate a test key
	keyVersion, err := hsm.GenerateKeyMaterial("AES-256-GCM")
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Test data
	plaintext := []byte("test data")

	// Encrypt the data
	ciphertext, err := hsm.Encrypt(keyVersion, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Decrypt the data
	decrypted, err := hsm.Decrypt(keyVersion, ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	// Verify the decrypted data matches the original
	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted data does not match original")
	}
}

// 🧪 TestDecryptInvalidCiphertext tests decryption with invalid input
func TestDecryptInvalidCiphertext(t *testing.T) {
	hsm, err := NewSoftwareHSM()
	if err != nil {
		t.Fatalf("Failed to create HSM: %v", err)
	}

	// Generate a test key
	keyVersion, err := hsm.GenerateKeyMaterial("AES-256-GCM")
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Try to decrypt invalid ciphertext
	_, err = hsm.Decrypt(keyVersion, []byte("invalid"))
	if err == nil {
		t.Error("Expected error when decrypting invalid ciphertext")
	}
}

// 🧪 TestKeyWrapping tests key wrapping and unwrapping
func TestKeyWrapping(t *testing.T) {
	hsm, err := NewSoftwareHSM()
	if err != nil {
		t.Fatalf("Failed to create HSM: %v", err)
	}

	// Generate a master key
	masterKey, err := hsm.GenerateKeyMaterial("AES-256-GCM")
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}

	// Generate key material to wrap
	keyMaterial, err := hsm.GenerateKeyMaterial("AES-256-GCM")
	if err != nil {
		t.Fatalf("Failed to generate key material: %v", err)
	}

	// Wrap the key
	wrappedKey, err := hsm.WrapKey(masterKey, keyMaterial)
	if err != nil {
		t.Fatalf("Failed to wrap key: %v", err)
	}

	// Unwrap the key
	unwrappedKey, err := hsm.UnwrapKey(masterKey, wrappedKey)
	if err != nil {
		t.Fatalf("Failed to unwrap key: %v", err)
	}

	// Verify the unwrapped key matches the original
	if !bytes.Equal(keyMaterial, unwrappedKey) {
		t.Error("Unwrapped key does not match original")
	}
}
