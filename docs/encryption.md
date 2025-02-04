# 🔐 Xyphos Encryption Documentation

## Overview

Xyphos provides end-to-end encryption for client-server communication using RSA-OAEP encryption. This ensures that sensitive data remains secure during transit.

## Key Features

- 🔑 4096-bit RSA key pairs for strong security
- 🔒 Request/response encryption using RSA-OAEP with SHA-256
- 📦 Base64 encoding for encrypted data transport
- 🔄 Automatic key rotation support
- 🛡️ Per-client encryption keys

## Client Configuration

Each client gets its own RSA key pair during client configuration creation:

```go
config, err := client.CreateClientConfig(ctx, "my-client", []string{"encrypt", "decrypt"})
// Store config.PrivateKey securely - it's only sent once during creation
```

## Request Flow

1. **Client-side Encryption**:
   ```typescript
   // TypeScript example
   const client = new XyphosClient({
     baseURL: "http://localhost:8080",
     clientConfigId: config.id,
     clientConfigSecret: config.secret,
     privateKey: storedPrivateKey  // From client config creation
   });

   // Request is automatically encrypted
   const result = await client.encrypt(projectId, locationId, keyringId, keyId, plaintext);
   ```

2. **Server-side Decryption**:
   - Middleware automatically detects encrypted requests via `X-Request-Encrypted` header
   - Decrypts using client's private key stored on server
   - Processes decrypted request
   - Encrypts response if needed

## Headers

- `X-Request-Encrypted: true` - Indicates request body is encrypted
- `X-Response-Encrypted: true` - Indicates response body is encrypted

## Encryption Process

1. **Request Encryption**:
   ```
   [Request Body] -> JSON -> RSA-OAEP -> Base64 -> [Transport]
   ```

2. **Response Decryption**:
   ```
   [Response Body] -> Base64 Decode -> RSA-OAEP Decrypt -> JSON -> [Data]
   ```

## Client Libraries

### Go Client

```go
client := kms.NewClient(kms.ClientConfig{
    BaseURL: "http://localhost:8080",
    ClientConfigID: config.ID,
    ClientConfigSecret: config.Secret,
    PrivateKey: storedPrivateKey,
})

// Encryption happens automatically
result, err := client.Encrypt(ctx, projectID, locationID, keyringID, keyID, plaintext)
```

### Python Client

```python
client = XyphosClient(ClientConfig(
    base_url="http://localhost:8080",
    client_config_id=config.id,
    client_config_secret=config.secret,
    private_key=stored_private_key
))

# Encryption happens automatically
result = client.encrypt(project_id, location_id, keyring_id, key_id, plaintext)
```

### TypeScript Client

```typescript
const client = new XyphosClient({
  baseURL: "http://localhost:8080",
  clientConfigId: config.id,
  clientConfigSecret: config.secret,
  privateKey: storedPrivateKey
});

// Encryption happens automatically
const result = await client.encrypt(projectId, locationId, keyringId, keyId, plaintext);
```

## Security Considerations

1. **Key Storage**:
   - Store private keys securely
   - Never log or expose private keys
   - Use environment variables or secure key storage solutions

2. **Key Rotation**:
   - Rotate client keys periodically
   - Implement key version tracking
   - Handle key rotation gracefully in applications

3. **Error Handling**:
   - Handle encryption/decryption errors gracefully
   - Don't expose sensitive information in error messages
   - Log encryption failures securely

## Best Practices

1. **Always Use HTTPS**:
   - Even with request encryption, always use HTTPS
   - Encryption provides additional security layer

2. **Key Management**:
   - Implement secure key storage
   - Regular key rotation
   - Key backup and recovery procedures

3. **Monitoring**:
   - Monitor encryption/decryption failures
   - Alert on suspicious patterns
   - Track key usage and rotation

## Testing

Run the encryption tests:

```bash
go test -v ./internal/api/middleware/encryption_test.go
```

The tests cover:
- Basic request/response flow
- Encryption/decryption process
- Error handling
- Invalid encryption scenarios 