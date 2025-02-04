package auth

import (
	"testing"
	"time"
)

// 🧪 TestNewService tests service creation
func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Fatal("Service is nil")
	}
}

// 🧪 TestCreateClient tests client creation
func TestCreateClient(t *testing.T) {
	service := NewService()

	// Create a test client
	client, err := service.CreateClient("test-tenant")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify client fields
	if client.ID == "" {
		t.Error("Client ID is empty")
	}
	if client.Secret == "" {
		t.Error("Client secret is empty")
	}
	if client.Tenant != "test-tenant" {
		t.Errorf("Client tenant = %v, want %v", client.Tenant, "test-tenant")
	}
	if client.CreatedAt.IsZero() {
		t.Error("Client creation time is zero")
	}
}

// 🧪 TestValidateCredentials tests credential validation
func TestValidateCredentials(t *testing.T) {
	service := NewService()

	// Create a test client
	client, err := service.CreateClient("test-tenant")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test valid credentials
	validated, err := service.ValidateCredentials(client.ID, client.Secret)
	if err != nil {
		t.Errorf("Failed to validate valid credentials: %v", err)
	}
	if validated == nil {
		t.Error("Validated client is nil")
	}

	// Test invalid client ID
	_, err = service.ValidateCredentials("invalid", client.Secret)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials for invalid client ID, got %v", err)
	}

	// Test invalid secret
	_, err = service.ValidateCredentials(client.ID, "invalid")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials for invalid secret, got %v", err)
	}
}

// 🧪 TestTokenGeneration tests token generation and validation
func TestTokenGeneration(t *testing.T) {
	service := NewService()

	// Create a test client
	client, err := service.CreateClient("test-tenant")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Generate a token
	token, err := service.GenerateToken(client)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Error("Generated token is empty")
	}

	// Validate the token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	if claims.ClientID != client.ID {
		t.Errorf("Token client ID = %v, want %v", claims.ClientID, client.ID)
	}
	if claims.Tenant != client.Tenant {
		t.Errorf("Token tenant = %v, want %v", claims.Tenant, client.Tenant)
	}

	// Test token expiration
	time.Sleep(time.Second)
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Token expired too soon")
	}
}

// 🧪 TestInvalidToken tests token validation with invalid tokens
func TestInvalidToken(t *testing.T) {
	service := NewService()

	// Test empty token
	_, err := service.ValidateToken("")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for empty token, got %v", err)
	}

	// Test invalid token format
	_, err = service.ValidateToken("invalid.token.format")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken for invalid format, got %v", err)
	}

	// Test expired token
	client, _ := service.CreateClient("test-tenant")
	token, _ := service.GenerateToken(client)
	time.Sleep(2 * time.Second)
	_, err = service.ValidateToken(token)
	if err != nil {
		t.Errorf("Expected no error for valid token, got %v", err)
	}
}
