# 🔐 Xyphos TypeScript Client

> Why do TypeScript developers love our KMS client? Because it's strongly typed and weakly coupled! 🎯

[![npm version](https://img.shields.io/npm/v/lambda-kms-client.svg)](https://www.npmjs.com/package/lambda-kms-client)
[![TypeScript](https://img.shields.io/badge/%3C%2F%3E-TypeScript-%230074c1.svg)](http://www.typescriptlang.org/)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 📦 Installation

```bash
# npm
npm install @xyphos/client

# yarn
yarn add @xyphos/client

# pnpm
pnpm add @xyphos/client
```

## 🚀 Quick Start

```typescript
import { KMSClient } from '@xyphos/client'

// 🔐 Initialize client
const client = new KMSClient({
    baseUrl: 'http://localhost:8080',
    projectId: 'my-project',
    apiKey: process.env.KMS_API_KEY,
    retryConfig: {
        maxRetries: 3,
        backoffFactor: 0.1,
        statusForcelist: [500, 502, 503, 504, 429],
    },
})

async function example() {
    // 🔑 Create a keyring
    const keyring = await client.createKeyRing({
        locationId: 'us-west1',
        keyRingId: 'app-keys',
        description: 'Application encryption keys',
    })

    // 🗝️ Create a key
    const key = await client.createKey({
        locationId: 'us-west1',
        keyRingId: keyring.id,
        keyId: 'encryption-key',
        algorithm: 'AES256-GCM',
        purpose: 'ENCRYPT_DECRYPT',
    })

    // 🔒 Encrypt data
    const { ciphertext, keyVersion } = await client.encrypt({
        locationId: 'us-west1',
        keyRingId: keyring.id,
        keyId: key.id,
        plaintext: 'secret message',
    })

    console.log('Encrypted data:', ciphertext.toString('base64'))
}
```

## 🎯 Features

- 🔐 **Type Safety**
  - Full TypeScript support
  - Comprehensive type definitions
  - Compile-time validation

- 🔄 **Automatic Retries**
  - Configurable retry policy
  - Exponential backoff
  - Circuit breaker pattern

- 🌐 **Modern JavaScript**
  - ESM and CommonJS support
  - Promise-based API
  - WebCrypto integration

- 🧪 **Testing Utilities**
  - Mock client
  - Test fixtures
  - Jest matchers

## 📚 API Reference

### Client Configuration

```typescript
interface KMSConfig {
    baseUrl: string
    projectId: string
    apiKey: string
    retryConfig?: {
        maxRetries: number
        backoffFactor: number
        statusForcelist: number[]
    }
    httpClient?: AxiosInstance
    logger?: Logger
}
```

### Key Operations

```typescript
// Create a keyring
createKeyRing(params: CreateKeyRingParams): Promise<KeyRing>

// Create a key
createKey(params: CreateKeyParams): Promise<Key>

// Encrypt data
encrypt(params: EncryptParams): Promise<EncryptResponse>

// Decrypt data
decrypt(params: DecryptParams): Promise<DecryptResponse>

// Rotate key
rotateKey(params: RotateKeyParams): Promise<Key>
```

## 🔐 Security Best Practices

1. 🔑 **API Key Management**
```typescript
// ✅ Good: Use environment variables
const client = new KMSClient({
    apiKey: process.env.KMS_API_KEY,
})

// ❌ Bad: Hardcode credentials
const client = new KMSClient({
    apiKey: 'secret-key-1234',
})
```

2. 🌐 **HTTPS Enforcement**
```typescript
// ✅ Good: Custom axios config with HTTPS enforcement
const client = new KMSClient({
    httpClient: axios.create({
        httpsAgent: new https.Agent({
            minVersion: 'TLSv1.2',
            maxVersion: 'TLSv1.3',
        }),
    }),
})
```

3. ⏱️ **Timeouts & Retries**
```typescript
// ✅ Good: Configure timeouts and retries
const client = new KMSClient({
    retryConfig: {
        maxRetries: 3,
        backoffFactor: 0.1,
        statusForcelist: [500, 502, 503, 504, 429],
    },
    httpClient: axios.create({
        timeout: 5000,
    }),
})
```

## 🧪 Testing

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run type checking
npm run type-check
```

### Jest Matchers

```typescript
import { kmsMatchers } from '@xyphos/client/testing'

expect.extend(kmsMatchers)

test('encryption works', async () => {
    const response = await client.encrypt(params)
    expect(response).toBeValidEncryptResponse()
})
```

## 🤝 Contributing

> Why do TypeScript developers make great cryptographers? Because they keep their types secret! 🤫

1. Fork the repository
2. Create your feature branch
3. Run the tests
4. Submit a pull request

## 📝 License

MIT License - see [LICENSE](LICENSE)

## 🧑‍💻 Author

Created with ❤️ by [Harsh VARDHAN GOSWAMI](https://github.com/theboringhumane)

---
*Remember: Type safety is the best security! 🛡️* 