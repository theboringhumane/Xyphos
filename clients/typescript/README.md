# 🔐 Xyphos TypeScript Client

Official TypeScript client for Xyphos - the open-source key management system.

## ✨ Features

- 🔒 End-to-end encryption using RSA-OAEP
- 🔄 Automatic retry mechanism with exponential backoff
- 🎯 Type-safe API with full TypeScript support
- 🔐 JWT-based authentication
- 📦 WebCrypto API integration
- ⚡ Promise-based async/await API

## 📦 Installation

```bash
npm install @xyphos/client
# or
yarn add @xyphos/client
```

## 🚀 Quick Start

```typescript
import { XyphosClient } from '@xyphos/client';

const client = new XyphosClient({
  baseURL: 'http://localhost:8080',
  clientConfigId: 'your-client-id',
  clientConfigSecret: 'your-client-secret'
});

// Encrypt data
const plaintext = new TextEncoder().encode('Hello, World!');
const [ciphertext, keyVersion] = await client.encrypt(
  'my-keyring',
  'ENCRYPT_DECRYPT',
  plaintext
);

// Decrypt data
const decrypted = await client.decrypt(
  'my-keyring',
  'ENCRYPT_DECRYPT',
  ciphertext,
  keyVersion
);
```

## 🔧 Configuration

```typescript
interface ClientConfig {
  // Required configuration
  baseURL: string;              // API base URL
  clientConfigId: string;       // Client ID from Xyphos
  clientConfigSecret: string;   // Client secret from Xyphos

  // Optional configuration
  timeout?: number;             // Request timeout in ms (default: 30000)
  privateKey?: string;          // RSA private key for request encryption
  retryConfig?: {
    maxRetries: number;         // Maximum retry attempts (default: 3)
    backoffFactor: number;      // Exponential backoff factor (default: 0.1)
    statusForcelist: number[];  // Status codes to retry (default: [500, 502, 503, 504, 429])
  };
}
```

## 📚 API Reference

### Key Management

```typescript
// Create a keyring
await client.createKeyring(name: string): Promise<void>

// List keyrings
await client.listKeyrings(): Promise<string[]>

// List keys in a keyring
await client.listKeys(keyringName: string): Promise<string[]>

// Get key information
await client.getKeyInfo(
  keyringName: string,
  keyName: string
): Promise<KeyInfo>

// Create a new key
await client.createKey(
  keyringName: string,
  algorithm: string,
  purpose: string,
  rotationPeriod: number
): Promise<void>
```

### Encryption Operations

```typescript
// Encrypt data
await client.encrypt(
  keyringName: string,
  purpose: string,
  plaintext: Uint8Array
): Promise<[Uint8Array, number]>

// Decrypt data
await client.decrypt(
  keyringName: string,
  purpose: string,
  ciphertext: Uint8Array,
  keyVersion: number
): Promise<Uint8Array>
```

## 🔐 Security Features

### End-to-End Encryption

```typescript
// Initialize client with encryption enabled
const client = new XyphosClient({
  baseURL: 'http://localhost:8080',
  clientConfigId: 'your-client-id',
  clientConfigSecret: 'your-client-secret',
  privateKey: '-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----'
});

// All requests/responses will be automatically encrypted
```

### Authentication

The client automatically handles:
- JWT token acquisition
- Token refresh
- Secure token storage
- Request signing

## 🔄 Retry Mechanism

```typescript
// Custom retry configuration
const client = new XyphosClient({
  baseURL: 'http://localhost:8080',
  clientConfigId: 'your-client-id',
  clientConfigSecret: 'your-client-secret',
  retryConfig: {
    maxRetries: 5,
    backoffFactor: 0.2,
    statusForcelist: [500, 502, 503, 504, 429]
  }
});
```

## 🚨 Error Handling

```typescript
try {
  await client.encrypt(/*...*/);
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

## 🔍 Debugging

Enable debug logging:

```typescript
// Set environment variable
process.env.DEBUG = 'xyphos:*';

// Or in browser
localStorage.debug = 'xyphos:*';
```

## 🧪 Testing

```bash
# Run unit tests
npm test

# Run with coverage
npm run test:coverage

# Run integration tests
npm run test:integration
```

## 📝 Type Definitions

Full TypeScript definitions are included:

```typescript
import type {
  ClientConfig,
  KeyInfo,
  EncryptResponse,
  DecryptResponse,
  XyphosError
} from '@xyphos/client';
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md).

## 📄 License

MIT License - see [LICENSE](../../LICENSE) 