# 🔐 Xyphos

> Why did the cryptographer bring a ladder to work? Because they heard the encryption keys were stored at a higher level! 🪜

A secure, multi-tenant key management system built in Go. Think Google Cloud KMS, but open source and running in your own infrastructure! 

[![GitHub stars](https://img.shields.io/github/stars/theboringhumane/xyphos?style=social)](https://github.com/theboringhumane/xyphos/stargazers)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/theboringhumane/xyphos)](https://goreportcard.com/report/github.com/theboringhumane/xyphos)

## 🌟 Overview

Xyphos is a complete key management solution with two main components:

1. 🔒 **Server** ([Documentation](docs/server/README.md))
   - Go-based KMS service with BadgerDB storage
   - FIPS 140-2 compliant cryptographic operations
   - Secure key storage and management
   
2. 🌐 **Frontend Dashboard**
   - Modern React/Next.js interface
   - WebCrypto API integration
   - Real-time key management

## 🏗️ System Architecture

```mermaid
graph TD
    A[Frontend - Next.js] -->|HTTPS| B[API Gateway - Go/Gin]
    B --> C[Auth Service]
    B --> D[KMS Service]
    D --> E[HSM Service]
    D --> F[BadgerDB]
    D --> G[Audit Service]
    
    style A fill:#f9f,stroke:#333,stroke-width:2px
    style B fill:#bbf,stroke:#333,stroke-width:2px
    style C fill:#dfd,stroke:#333,stroke-width:2px
    style D fill:#fdd,stroke:#333,stroke-width:2px
```

## 🚀 Quick Start

> Why do developers prefer dark mode? Because light attracts bugs! 🪲

### Prerequisites

- 🐹 Go 1.21+
- 📦 Node.js 18+
- 🐳 Docker (optional)
- ☕ Coffee (lots of it!)

### Setup

1. Clone the repo:
```bash
git clone https://github.com/theboringhumane/xyphos.git
cd xyphos
```

2. Set up the server:
```bash
cp .env.example .env
# Generate JWT secret
openssl rand -base64 32  # Copy to JWT_SECRET in .env
go run cmd/main.go
```

3. Set up the frontend:
```bash
cd frontend
cp .env.example .env
# Generate NextAuth secret
openssl rand -base64 32  # Copy to NEXTAUTH_SECRET in .env
npm install && npm run dev
```

For detailed server documentation, including API endpoints and security architecture, see the [Server Documentation](docs/server/README.md).

## 🎮 Client SDK

### 📦 Installation

```bash
npm install @xyphos/client
# or
yarn add @xyphos/client
```

### 🔧 Configuration

```typescript
interface ClientConfig {
  baseURL: string;
  clientConfigId: string;
  clientConfigSecret: string;
  timeout?: number;        // Default: 30000ms
  retryConfig?: {
    maxRetries: number;    // Default: 3
    backoffFactor: number; // Default: 0.1
    statusForcelist: number[]; // Default: [500, 502, 503, 504, 429]
  };
  privateKey?: string;     // RSA private key for request encryption
}
```

### 🚀 Basic Usage

```typescript
import { APIClient } from '@xyphos/client'

// Initialize with basic config
const client = new APIClient({
  baseURL: 'http://localhost:8080',
  clientConfigId: 'your-client-id',
  clientConfigSecret: 'your-client-secret'
})

// Or with full configuration including encryption
const secureClient = new APIClient({
  baseURL: 'http://localhost:8080',
  clientConfigId: 'your-client-id',
  clientConfigSecret: 'your-client-secret',
  timeout: 30000,
  retryConfig: {
    maxRetries: 3,
    backoffFactor: 0.1,
    statusForcelist: [500, 502, 503, 504, 429]
  },
  privateKey: '-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----'
})

// 🔐 Encrypt data
const { ciphertext, keyVersion } = await client.encrypt({
  projectId: 'my-project',
  locationId: 'us-west1',
  keyringId: 'app-keys',
  keyId: 'encryption-key',
  plaintext: 'super secret message'
})

// 🔓 Decrypt data
const { plaintext } = await client.decrypt({
  projectId: 'my-project',
  locationId: 'us-west1',
  keyringId: 'app-keys',
  keyId: 'encryption-key',
  keyVersion: keyVersion,
  ciphertext: ciphertext
})
```

### 🔄 Automatic Retries

The client includes built-in retry logic for handling transient failures:

```typescript
// Custom retry configuration
const client = new APIClient({
  // ... other config
  retryConfig: {
    maxRetries: 5,                    // Maximum number of retry attempts
    backoffFactor: 0.2,              // Exponential backoff multiplier
    statusForcelist: [500, 502, 503]  // Status codes that trigger retries
  }
})
```

### 🔐 End-to-End Encryption

When initialized with a private key, the client automatically handles request/response encryption:

```typescript
// The client will automatically:
// 1. Encrypt request bodies using RSA-OAEP
// 2. Decrypt encrypted responses
// 3. Handle key rotation and management
const encryptedClient = new APIClient({
  // ... other config
  privateKey: process.env.CLIENT_PRIVATE_KEY
})

// All requests/responses will be automatically encrypted
const result = await encryptedClient.encrypt({
  projectId: 'my-project',
  keyringId: 'app-keys',
  keyId: 'encryption-key',
  plaintext: 'This request will be encrypted'
})
```

### 🚨 Error Handling

```typescript
try {
  await client.encrypt(/*...*/)
} catch (error) {
  if (error instanceof AuthenticationError) {
    // Handle authentication failure
  } else if (error instanceof PermissionError) {
    // Handle permission issues
  } else if (error instanceof NotFoundError) {
    // Handle missing resources
  } else if (error instanceof InvalidInputError) {
    // Handle invalid input
  } else if (error instanceof KMSError) {
    // Handle general KMS errors
  }
}
```

## 🤝 Contributing

> How many developers does it take to change a light bulb? None, that's a hardware problem! 💡

Contributions are welcome! Check out our [Contributing Guide](CONTRIBUTING.md).

## 📝 License

MIT License - see [LICENSE](LICENSE)

## 🧑‍💻 Author

Created with ❤️ by [Harsh VARDHAN GOSWAMI](https://github.com/theboringhumane)

> Why did the developer go broke? Because they used up all their cache! 💰

---
*Remember: In cryptography we trust, but we still verify! 🔍*