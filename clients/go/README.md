# 🔐 Xyphos Go Client

> Because Go developers love their keys type-safe and their errors handled! 🎯

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A powerful Go client for Xyphos - your open-source key management system. Built with ❤️ for Go developers who take security seriously (and love emojis 🎉).

## 📦 Installation

```bash
go get github.com/theboringhumane/xyphos/clients/go
```

## 🚀 Quick Start

Here's a complete example that shows how to use the Xyphos Go client:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    kms "github.com/theboringhumane/xyphos/clients/go"
)

func main() {
    // 🔧 Create a new client with custom retry config
    client := kms.NewClient(
        "https://your-kms-server.com",
        "your-client-id",
        "your-client-secret",
        kms.WithRetryConfig(kms.RetryConfig{
            MaxRetries:  3,
            InitialWait: 100 * time.Millisecond,
            MaxWait:     2 * time.Second,
        }),
    )
    
    ctx := context.Background()
    
    // 📦 Create a project
    project, err := client.CreateProject(ctx, "my-project", "My awesome project")
    if err != nil {
        panic(err) // Don't do this in production! 😅
    }
    
    // 📍 Create a location
    location, err := client.CreateLocation(ctx, project.ID, "us-west1")
    if err != nil {
        panic(err)
    }
    
    // 💍 Create a keyring
    keyring, err := client.CreateKeyRing(ctx, project.ID, location.ID, "my-keyring")
    if err != nil {
        panic(err)
    }
    
    // 🔑 Create a key
    key, err := client.CreateCryptoKey(ctx, project.ID, location.ID, keyring.ID,
        "AES256-GCM",           // algorithm
        "ENCRYPT_DECRYPT",      // purpose
        24 * 60 * 60,          // rotation period (24 hours)
    )
    if err != nil {
        panic(err)
    }
    
    // 🔒 Encrypt some data
    plaintext := []byte("super secret message")
    ciphertext, err := client.Encrypt(ctx, project.ID, location.ID, keyring.ID, key.ID, plaintext)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Encrypted: %x\n", ciphertext)
    
    // 🔓 Decrypt the data
    decrypted, err := client.Decrypt(ctx, project.ID, location.ID, keyring.ID, key.ID, ciphertext)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Decrypted: %s\n", decrypted)
}
```

## ✨ Features

- 🔒 Secure key management with hierarchical structure
- 🔄 Automatic retries with exponential backoff
- 🎯 Type-safe API with Go-idiomatic error handling
- 🚦 Context support for timeouts and cancellation
- 🧪 Testing utilities included

## 📚 API Reference

### Resource Types

- 📦 **Project**: Top-level resource for organizing keys
  ```go
  type Project struct {
      ID          string
      Name        string
      Description string
      CreatedAt   time.Time
  }
  ```

- 📍 **Location**: Geographic location for keys
  ```go
  type Location struct {
      ID        string
      Name      string
      CreatedAt time.Time
  }
  ```

- 💍 **KeyRing**: Container for cryptographic keys
  ```go
  type KeyRing struct {
      ID        string
      Name      string
      CreatedAt time.Time
  }
  ```

- 🔑 **CryptoKey**: The actual encryption key
  ```go
  type CryptoKey struct {
      ID             string
      Name           string
      Algorithm      string
      Purpose        string
      RotationPeriod int
      CreatedAt      time.Time
      NextRotation   time.Time
      Version        int
  }
  ```

### Client Configuration

```go
// Create a new client with options
client := kms.NewClient(
    baseURL,
    clientID,
    clientSecret,
    kms.WithRetryConfig(kms.RetryConfig{
        MaxRetries:  3,
        InitialWait: 100 * time.Millisecond,
        MaxWait:     2 * time.Second,
    }),
    kms.WithTimeout(30 * time.Second),
)
```

### Project Operations

```go
// Create a project
project, err := client.CreateProject(ctx, "my-project", "My awesome project")

// List projects
projects, err := client.ListProjects(ctx)
```

### Location Operations

```go
// Create a location
location, err := client.CreateLocation(ctx, projectID, "us-west1")

// List locations
locations, err := client.ListLocations(ctx, projectID)
```

### KeyRing Operations

```go
// Create a keyring
keyring, err := client.CreateKeyRing(ctx, projectID, locationID, "my-keyring")

// List keyrings
keyrings, err := client.ListKeyRings(ctx, projectID, locationID)
```

### CryptoKey Operations

```go
// Create a key
key, err := client.CreateCryptoKey(ctx, projectID, locationID, keyringID,
    "AES256-GCM",      // algorithm
    "ENCRYPT_DECRYPT", // purpose
    24 * 60 * 60,     // rotation period (24 hours)
)

// List keys
keys, err := client.ListCryptoKeys(ctx, projectID, locationID, keyringID)

// Rotate a key
err = client.RotateCryptoKey(ctx, projectID, locationID, keyringID, keyID)
```

### Cryptographic Operations

```go
// Encrypt data
ciphertext, err := client.Encrypt(ctx, projectID, locationID, keyringID, keyID, plaintext)

// Decrypt data
plaintext, err := client.Decrypt(ctx, projectID, locationID, keyringID, keyID, ciphertext)
```

## 🔒 Security Best Practices

1. 🔑 **API Key Management**
   - Store client credentials securely (e.g., environment variables)
   - Rotate client credentials regularly
   - Use separate credentials for development and production

2. 🌐 **TLS Configuration**
   - Always use HTTPS for production
   - Verify TLS certificates
   - Consider using mutual TLS for additional security

3. ⏱️ **Timeouts and Retries**
   - Set appropriate timeouts for your use case
   - Configure retries with exponential backoff
   - Handle rate limiting gracefully

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...
```

## 🤝 Contributing

We love contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🎭 Author

Created with ❤️ by [Your Name] - because someone had to make key management fun! 

> Q: Why do cryptographers make great comedians?
> A: Because they always have the perfect key-punchline! 🎯 