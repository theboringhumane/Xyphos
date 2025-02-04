# 🔐 Xyphos Python Client

Official Python client for Xyphos - the open-source key management system.

## ✨ Features

- 🔒 End-to-end encryption using RSA-OAEP
- 🔄 Automatic retry mechanism with exponential backoff
- 🎯 Type hints with Python 3.8+ support
- 🔐 JWT-based authentication
- 📦 Modern cryptography library integration
- ⚡ Async/await support

## 📦 Installation

```bash
pip install xyphos-client
# or
poetry add xyphos-client
```

## 🚀 Quick Start

```python
from xyphos_client import XyphosClient

# Initialize client
client = XyphosClient(
    base_url="http://localhost:8080",
    client_config_id="your-client-id",
    client_config_secret="your-client-secret"
)

# Encrypt data
plaintext = b"Hello, World!"
ciphertext, key_version = await client.encrypt(
    keyring_name="my-keyring",
    purpose="ENCRYPT_DECRYPT",
    plaintext=plaintext
)

# Decrypt data
decrypted = await client.decrypt(
    keyring_name="my-keyring",
    purpose="ENCRYPT_DECRYPT",
    ciphertext=ciphertext,
    key_version=key_version
)
```

## 🔧 Configuration

```python
from dataclasses import dataclass
from typing import Optional, List

@dataclass
class RetryConfig:
    max_retries: int = 3
    backoff_factor: float = 0.1
    status_forcelist: List[int] = [500, 502, 503, 504, 429]

@dataclass
class ClientConfig:
    # Required configuration
    base_url: str                # API base URL
    client_config_id: str        # Client ID from Xyphos
    client_config_secret: str    # Client secret from Xyphos

    # Optional configuration
    timeout: float = 30.0        # Request timeout in seconds
    private_key: Optional[str] = None  # RSA private key for request encryption
    retry_config: Optional[RetryConfig] = None
```

## 📚 API Reference

### Key Management

```python
# Create a keyring
await client.create_keyring(name: str) -> None

# List keyrings
await client.list_keyrings() -> List[str]

# List keys in a keyring
await client.list_keys(keyring_name: str) -> List[str]

# Get key information
await client.get_key_info(
    keyring_name: str,
    key_name: str
) -> KeyInfo

# Create a new key
await client.create_key(
    keyring_name: str,
    algorithm: str,
    purpose: str,
    rotation_period: int
) -> None
```

### Encryption Operations

```python
# Encrypt data
await client.encrypt(
    keyring_name: str,
    purpose: str,
    plaintext: bytes
) -> Tuple[bytes, int]

# Decrypt data
await client.decrypt(
    keyring_name: str,
    purpose: str,
    ciphertext: bytes,
    key_version: int
) -> bytes
```

## 🔐 Security Features

### End-to-End Encryption

```python
# Initialize client with encryption enabled
client = XyphosClient(
    base_url="http://localhost:8080",
    client_config_id="your-client-id",
    client_config_secret="your-client-secret",
    private_key="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
)

# All requests/responses will be automatically encrypted
```

### Authentication

The client automatically handles:
- JWT token acquisition
- Token refresh
- Secure token storage
- Request signing

## 🔄 Retry Mechanism

```python
from xyphos_client import RetryConfig

# Custom retry configuration
client = XyphosClient(
    base_url="http://localhost:8080",
    client_config_id="your-client-id",
    client_config_secret="your-client-secret",
    retry_config=RetryConfig(
        max_retries=5,
        backoff_factor=0.2,
        status_forcelist=[500, 502, 503, 504, 429]
    )
)
```

## 🚨 Error Handling

```python
from xyphos_client.exceptions import (
    AuthenticationError,
    PermissionError,
    NotFoundError,
    InvalidInputError,
    KMSError
)

try:
    await client.encrypt(...)
except AuthenticationError:
    # Handle authentication failure
except PermissionError:
    # Handle permission issues
except NotFoundError:
    # Handle missing resources
except InvalidInputError:
    # Handle invalid input
except KMSError:
    # Handle general KMS errors
```

## 🔍 Debugging

Enable debug logging:

```python
import logging

# Set up logging
logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger("xyphos_client")
logger.setLevel(logging.DEBUG)
```

## 🧪 Testing

```bash
# Run unit tests
pytest

# Run with coverage
pytest --cov=xyphos_client

# Run integration tests
pytest tests/integration
```

## 📝 Type Hints

Full type hints are included:

```python
from xyphos_client.types import (
    ClientConfig,
    KeyInfo,
    EncryptResponse,
    DecryptResponse,
    XyphosError
)
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md).

## 📄 License

MIT License - see [LICENSE](../../LICENSE) 