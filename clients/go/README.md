# 🔒 Xyphos Go Client

Official Go client for Xyphos - the open-source key management system.

## ✨ Features

- 🔒 End-to-end encryption using RSA-OAEP
- 🔄 Automatic retry mechanism with exponential backoff
- 🎯 Strong type safety with Go 1.21+ support
- 🔐 JWT-based authentication
- 📦 Native crypto package integration
- ⚡ Context-aware operations

## 📦 Installation

```bash
go get github.com/xyphos/client
```

## 🚀 Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/xyphos/client"
)

func main() {
    // Initialize client
    cfg := client.Config{
        BaseURL:           "http://localhost:8080",
        ClientConfigID:    "your-client-id",
        ClientConfigSecret: "your-client-secret",
    }
    
    client, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Encrypt data
    plaintext := []byte("Hello, World!")
    ciphertext, keyVersion, err := client.Encrypt(ctx, "my-keyring", "ENCRYPT_DECRYPT", plaintext)
    if err != nil {
        log.Fatal(err)
    }
    
    // Decrypt data
    decrypted, err := client.Decrypt(ctx, "my-keyring", "ENCRYPT_DECRYPT", ciphertext, keyVersion)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Decrypted: %s", string(decrypted))
}
```

## 🔧 Configuration

```go
// Client configuration
type Config struct {
    // Required configuration
    BaseURL            string        // API base URL
    ClientConfigID     string        // Client ID from Xyphos
    ClientConfigSecret string        // Client secret from Xyphos

    // Optional configuration
    Timeout           time.Duration  // Request timeout (default: 30s)
    PrivateKey        string         // RSA private key for request encryption
    RetryConfig       *RetryConfig   // Retry configuration
}

// Retry configuration
type RetryConfig struct {
    MaxRetries      int           // Maximum retry attempts (default: 3)
    BackoffFactor   float64       // Exponential backoff factor (default: 0.1)
    StatusForcelist []int         // Status codes to retry (default: [500, 502, 503, 504, 429])
}
```

## 📚 API Reference

### Key Management

```go
// Create a keyring
CreateKeyring(ctx context.Context, name string) error

// List keyrings
ListKeyrings(ctx context.Context) ([]string, error)

// List keys in a keyring
ListKeys(ctx context.Context, keyringName string) ([]string, error)

// Get key information
GetKeyInfo(ctx context.Context, keyringName, keyName string) (*KeyInfo, error)

// Create a new key
CreateKey(ctx context.Context, keyringName, algorithm, purpose string, rotationPeriod int) error
```

### Encryption Operations

```go
// Encrypt data
Encrypt(
    ctx context.Context,
    keyringName string,
    purpose string,
    plaintext []byte,
) (ciphertext []byte, keyVersion int, err error)

// Decrypt data
Decrypt(
    ctx context.Context,
    keyringName string,
    purpose string,
    ciphertext []byte,
    keyVersion int,
) (plaintext []byte, err error)
```

## 🔐 Security Features

### End-to-End Encryption

```go
// Initialize client with encryption enabled
cfg := client.Config{
    BaseURL:            "http://localhost:8080",
    ClientConfigID:     "your-client-id",
    ClientConfigSecret: "your-client-secret",
    PrivateKey:        "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
}

client, err := client.New(cfg)
if err != nil {
    log.Fatal(err)
}

// All requests/responses will be automatically encrypted
```

### Authentication

The client automatically handles:
- JWT token acquisition
- Token refresh
- Secure token storage
- Request signing

## 🔄 Retry Mechanism

```go
// Custom retry configuration
cfg := client.Config{
    BaseURL:            "http://localhost:8080",
    ClientConfigID:     "your-client-id",
    ClientConfigSecret: "your-client-secret",
    RetryConfig: &client.RetryConfig{
        MaxRetries:      5,
        BackoffFactor:   0.2,
        StatusForcelist: []int{500, 502, 503, 504, 429},
    },
}
```

## 🚨 Error Handling

```go
import "github.com/xyphos/client/errors"

// Encrypt with error handling
ciphertext, keyVersion, err := client.Encrypt(ctx, "my-keyring", "ENCRYPT_DECRYPT", plaintext)
switch {
case errors.IsAuthentication(err):
    // Handle authentication failure
case errors.IsPermission(err):
    // Handle permission issues
case errors.IsNotFound(err):
    // Handle missing resources
case errors.IsInvalidInput(err):
    // Handle invalid input
case err != nil:
    // Handle other errors
}
```

## 🔍 Debugging

Enable debug logging:

```go
import "github.com/xyphos/client/log"

// Set log level
log.SetLevel(log.DebugLevel)

// Or use custom logger
client.SetLogger(yourCustomLogger)
```

## 🧪 Testing

```bash
# Run unit tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests
go test -tags=integration ./...
```

## 📝 Type Definitions

```go
import (
    "github.com/xyphos/client"
    "github.com/xyphos/client/types"
)

// Available types
type (
    Config     = client.Config
    KeyInfo    = types.KeyInfo
    RetryConfig = client.RetryConfig
)
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md).

## 📄 License

MIT License - see [LICENSE](../../LICENSE) 