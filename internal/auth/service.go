package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	// 🚫 ErrInvalidToken indicates an invalid or expired token
	ErrInvalidToken = errors.New("invalid or expired token")

	// 🚫 ErrInvalidCredentials indicates invalid client credentials
	ErrInvalidCredentials = errors.New("invalid client credentials")
)

// 👤 Client represents an API client
type Client struct {
	ID        string
	Secret    string
	Tenant    string
	CreatedAt time.Time
}

// 🔑 Service handles authentication and authorization
type Service interface {
	CreateClient(tenant string) (*Client, error)
	ValidateCredentials(clientID, clientSecret string) (*Client, error)
	GenerateToken(client *Client) (string, error)
	ValidateToken(token string) (*Claims, error)
}

// 📜 Claims represents the JWT claims
type Claims struct {
	ClientID string `json:"client_id"`
	Tenant   string `json:"tenant"`
	jwt.RegisteredClaims
}

// 🏭 authService implements the Service interface
type authService struct {
	clients map[string]*Client
	secret  []byte
}

// 🎯 NewService creates a new auth service instance
func NewService() Service {
	// Generate a random signing key
	secret := make([]byte, 32)
	rand.Read(secret)

	service := &authService{
		clients: make(map[string]*Client),
		secret:  secret,
	}

	// Create a default test client
	testClient := &Client{
		ID:        "test-client",
		Secret:    "test-secret",
		Tenant:    "test-tenant",
		CreatedAt: time.Now(),
	}
	service.clients[testClient.ID] = testClient

	return service
}

// 👥 CreateClient creates a new API client
func (s *authService) CreateClient(tenant string) (*Client, error) {
	// Generate random client ID and secret
	clientID := make([]byte, 16)
	clientSecret := make([]byte, 32)

	if _, err := rand.Read(clientID); err != nil {
		return nil, err
	}
	if _, err := rand.Read(clientSecret); err != nil {
		return nil, err
	}

	client := &Client{
		ID:        base64.URLEncoding.EncodeToString(clientID),
		Secret:    base64.URLEncoding.EncodeToString(clientSecret),
		Tenant:    tenant,
		CreatedAt: time.Now(),
	}

	s.clients[client.ID] = client
	return client, nil
}

// 🔍 ValidateCredentials validates client credentials
func (s *authService) ValidateCredentials(clientID, clientSecret string) (*Client, error) {
	client, exists := s.clients[clientID]
	if !exists || client.Secret != clientSecret {
		return nil, ErrInvalidCredentials
	}
	return client, nil
}

// 🎫 GenerateToken generates a new JWT token for a client
func (s *authService) GenerateToken(client *Client) (string, error) {
	claims := Claims{
		ClientID: client.ID,
		Tenant:   client.Tenant,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ✅ ValidateToken validates a JWT token
func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
