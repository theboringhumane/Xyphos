# 🔐 Xyphos Clients

> Why do developers love our KMS clients? Because they keep their secrets... secret! 🤫

Official client libraries for Xyphos in multiple languages. Each client provides a type-safe, secure, and easy-to-use interface to interact with Xyphos.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python Package](https://img.shields.io/pypi/v/xyphos-client)](https://pypi.org/project/xyphos-client/)
[![npm](https://img.shields.io/npm/v/xyphos-client)](https://www.npmjs.com/package/xyphos-client)
[![Go Reference](https://pkg.go.dev/badge/github.com/theboringhumane/xyphos/clients/go.svg)](https://pkg.go.dev/github.com/theboringhumane/xyphos/clients/go)

## 📦 Available Clients

| Language   | Directory                | Package Manager | Status |
|------------|--------------------------|----------------|---------|
| Go         | [`/go`](go/)            | go get         | ✅ Ready |
| TypeScript | [`/typescript`](typescript/) | npm/yarn   | ✅ Ready |
| Python     | [`/python`](python/)     | pip           | ✅ Ready |

## 🌟 Features

- 🔐 **Secure by Default**
  - GitHub OAuth authentication
  - JWT-based API authentication
  - TLS certificate validation
  - Secure credential storage

- 🔄 **Resource Management**
  - Projects and locations
  - Key rings and crypto keys
  - Client configurations
  - Automatic key rotation

- 🎯 **Developer Friendly**
  - Type-safe APIs
  - Comprehensive error handling
  - Detailed logging
  - Promise/async support (TS/Python)

- 🔒 **Cryptographic Operations**
  - AES-256-GCM encryption/decryption
  - RSA-OAEP key wrapping
  - ECDSA signing
  - Automatic key versioning

## 🚀 Quick Start

### Go

```go
import (
    kms "github.com/theboringhumane/xyphos/clients/go"
)

func main() {
    client := kms.NewClient(kms.Config{
        BaseURL:    "http://localhost:8080",
        ProjectID:  "my-project",
        APIKey:     os.Getenv("KMS_API_KEY"),
    })

    // Create a project
    project, err := client.CreateProject(context.Background(), &kms.CreateProjectRequest{
        Name: "my-project",
        DisplayName: "My Project",
    })

    // Create a keyring
    keyring, err := client.CreateKeyRing(context.Background(), &kms.CreateKeyRingRequest{
        ProjectID:  project.ID,
        LocationID: "us-west1",
        KeyRingID:  "app-keys",
    })

    // Create a key
    key, err := client.CreateCryptoKey(context.Background(), &kms.CreateCryptoKeyRequest{
        ProjectID:  project.ID,
        LocationID: "us-west1",
        KeyRingID:  keyring.ID,
        KeyID:      "encryption-key",
        Algorithm:  kms.AlgorithmAES256GCM,
    })

    // 🔒 Encrypt data
    resp, err := client.Encrypt(context.Background(), &kms.EncryptRequest{
        ProjectID:  project.ID,
        LocationID: "us-west1",
        KeyRingID:  keyring.ID,
        KeyID:      key.ID,
        Plaintext:  []byte("secret message"),
    })
}
```

### TypeScript

```typescript
import { KMSClient } from '@xyphos/client'

const client = new KMSClient({
    baseUrl: 'http://localhost:8080',
    projectId: 'my-project',
    apiKey: process.env.KMS_API_KEY,
})

// Create resources
const project = await client.createProject({
    name: 'my-project',
    displayName: 'My Project',
})

const keyring = await client.createKeyRing({
    projectId: project.id,
    locationId: 'us-west1',
    keyRingId: 'app-keys',
})

const key = await client.createCryptoKey({
    projectId: project.id,
    locationId: 'us-west1',
    keyRingId: keyring.id,
    keyId: 'encryption-key',
    algorithm: 'AES256-GCM',
})

// 🔒 Encrypt data
const { ciphertext } = await client.encrypt({
    projectId: project.id,
    locationId: 'us-west1',
    keyRingId: keyring.id,
    keyId: key.id,
    plaintext: 'secret message',
})
```

### Python

```python
from lambda_kms import KMSClient

client = KMSClient(
    base_url="http://localhost:8080",
    project_id="my-project",
    api_key=os.getenv("KMS_API_KEY")
)

# Create resources
project = await client.create_project(
    name="my-project",
    display_name="My Project"
)

keyring = await client.create_keyring(
    project_id=project.id,
    location_id="us-west1",
    keyring_id="app-keys"
)

key = await client.create_crypto_key(
    project_id=project.id,
    location_id="us-west1",
    keyring_id=keyring.id,
    key_id="encryption-key",
    algorithm="AES256-GCM"
)

# 🔒 Encrypt data
response = await client.encrypt(
    project_id=project.id,
    location_id="us-west1",
    keyring_id=keyring.id,
    key_id=key.id,
    plaintext=b"secret message"
)
```

## 🔐 API Structure

```
/api
├── /projects
│   ├── GET /                    # List projects
│   ├── POST /                   # Create project
│   ├── GET /:projectId         # Get project
│   │
│   └── /locations
│       ├── GET /               # List locations
│       ├── GET /:locationId    # Get location
│       │
│       └── /keyrings
│           ├── POST /          # Create keyring
│           ├── GET /           # List keyrings
│           ├── GET /:keyringId # Get keyring
│           │
│           └── /keys
│               ├── POST /      # Create key
│               ├── GET /       # List keys
│               ├── GET /:keyId # Get key
│               ├── POST /:keyId:rotate  # Rotate key
│               ├── POST /:keyId:encrypt # Encrypt data
│               └── POST /:keyId:decrypt # Decrypt data
│
└── /client-configs
    ├── POST /                  # Create client config
    ├── GET /                   # List client configs
    ├── GET /:configId         # Get client config
    └── POST /:configId/revoke # Revoke client config
```

## 🔐 Security Best Practices

1. 🔑 **API Key Management**
   - Store API keys securely (use environment variables)
   - Rotate keys regularly
   - Use separate keys for development/production

2. 🌐 **Network Security**
   - Always use HTTPS in production
   - Validate TLS certificates
   - Set appropriate timeouts

3. 🔒 **Error Handling**
   - Never log sensitive data
   - Handle all crypto errors
   - Implement proper key rotation

## 🧪 Testing

Each client includes comprehensive tests:

```bash
# Go
cd go && go test ./...

# TypeScript
cd typescript && npm test

# Python
cd python && pytest
```

## 🤝 Contributing

> How many developers does it take to implement a crypto algorithm? None, they should use the standard library! 🔐

1. 🍴 Fork the repository
2. 🌿 Create your feature branch
3. 🔨 Make your changes
4. 🧪 Run the tests
5. 📬 Submit a pull request

## 📝 License

MIT License - see [LICENSE](LICENSE)

## 🧑‍💻 Author

Created with ❤️ by [Harsh VARDHAN GOSWAMI](https://github.com/theboringhumane)

> Why did the cryptographer get locked out? They used ROT13 on their password! 🔄

---
*Remember: Never roll your own crypto! 🎲* 