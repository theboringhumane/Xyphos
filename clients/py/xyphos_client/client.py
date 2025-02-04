import base64
from dataclasses import dataclass, field
import json
import os
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Union, Any

import httpx
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding, rsa
from cryptography.hazmat.primitives.ciphers.aead import ChaCha20Poly1305
from cryptography.hazmat.primitives.serialization import load_pem_private_key, load_pem_public_key

from .models import (
    ClientConfig,
    RetryConfig,
    Project,
    Location,
    KeyRing,
    KMSKey,
    ClientConfigInfo,
    CreateProjectRequest,
    CreateCryptoKeyRequest,
    CreateClientConfigRequest,
    EncryptResponse,
    DecryptResponse,
)
from .exceptions import (
    KMSError,
    AuthenticationError,
    NotFoundError,
    PermissionError,
    InvalidInputError,
    RateLimitError,
    ServerError,
)

# 🔐 Client configuration
@dataclass
class ClientConfig:
    base_url: str
    client_id: str
    client_secret: str
    private_key: str
    public_key: str
    timeout: float = 30.0
    retry_config: Optional[RetryConfig] = None

# 🔄 Retry configuration
@dataclass
class RetryConfig:
    max_retries: int = 3
    initial_wait: float = 1.0
    max_wait: float = 10.0
    backoff_factor: float = 2.0
    status_forcelist: List[int] = field(default_factory=lambda: [429])

# 🌐 Main client class
class Client:
    """🔐 Client for interacting with Xyphos KMS"""

    def __init__(self, config: ClientConfig):
        """Initialize the client with configuration"""
        self.config = config
        self.token: Optional[str] = None
        self.token_expiry: Optional[datetime] = None
        self._key_cache: Dict[str, Any] = {}

        if self.config.retry_config is None:
            self.config.retry_config = RetryConfig()

        self._http_client = httpx.AsyncClient(
            timeout=self.config.timeout,
            headers={"User-Agent": "XyphosKMS-Python-Client/1.0"},
        )

    async def close(self):
        """Close the HTTP client"""
        await self._http_client.aclose()

    async def __aenter__(self):
        """Async context manager entry"""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        await self.close()

    # 🔒 Authentication and request handling

    async def _ensure_token(self):
        """Ensure a valid token is available"""
        if self.token and self.token_expiry and datetime.now() < self.token_expiry:
            return

        response = await self._http_client.post(
            f"{self.config.base_url}/api/oauth/token",
            json={
                "client_id": self.config.client_id,
                "client_secret": self.config.client_secret,
                "grant_type": "client_credentials",
            },
        )

        if response.status_code != 200:
            raise AuthenticationError(f"Failed to get token: {response.text}")

        data = response.json()
        self.token = data["access_token"]
        self.token_expiry = datetime.now() + timedelta(seconds=data["expires_in"])

    async def _request(
        self,
        method: str,
        path: str,
        json_data: Optional[Dict] = None,
        retry_count: int = 0,
    ) -> httpx.Response:
        """Make an authenticated request with retry logic"""
        await self._ensure_token()

        headers = {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json",
        }

        try:
            response = await self._http_client.request(
                method,
                f"{self.config.base_url}/api{path}",
                json=json_data,
                headers=headers,
            )

            if response.status_code == 429:
                raise RateLimitError()

            if (
                response.status_code in self.config.retry_config.status_forcelist
                and retry_count < self.config.retry_config.max_retries
            ):
                wait_time = min(
                    self.config.retry_config.initial_wait
                    * (self.config.retry_config.backoff_factor ** retry_count),
                    self.config.retry_config.max_wait,
                )
                await httpx.AsyncClient.sleep(wait_time)
                return await self._request(method, path, json_data, retry_count + 1)

            if response.status_code == 401:
                self.token = None
                raise AuthenticationError()

            if response.status_code == 403:
                raise PermissionError()

            if response.status_code == 404:
                raise NotFoundError()

            if response.status_code == 400:
                raise InvalidInputError(response.text)

            if response.status_code >= 500:
                raise ServerError(f"Server error: {response.status_code}")

            return response

        except httpx.RequestError as e:
            raise KMSError(f"Request failed: {str(e)}")

    # 📦 Project operations

    async def create_project(self, request: CreateProjectRequest) -> Project:
        """Create a new project"""
        response = await self._request("POST", "/projects", json_data=request.__dict__)
        data = response.json()
        return Project(**data)

    async def list_projects(self) -> List[Project]:
        """List all projects"""
        response = await self._request("GET", "/projects")
        data = response.json()
        return [Project(**item) for item in data["projects"]]

    async def get_project(self, project_id: str) -> Project:
        """Get a project by ID"""
        response = await self._request("GET", f"/projects/{project_id}")
        data = response.json()
        return Project(**data)

    # 📍 Location operations

    async def list_locations(self, project_id: str) -> List[Location]:
        """List all locations in a project"""
        response = await self._request("GET", f"/projects/{project_id}/locations")
        data = response.json()
        return [Location(**item) for item in data["locations"]]

    async def get_location(self, project_id: str, location_id: str) -> Location:
        """Get a location by ID"""
        response = await self._request(
            "GET", f"/projects/{project_id}/locations/{location_id}"
        )
        data = response.json()
        return Location(**data)

    # 💍 KeyRing operations

    async def create_keyring(
        self, project_id: str, location_id: str, name: str
    ) -> KeyRing:
        """Create a new keyring"""
        response = await self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings",
            json_data={"name": name},
        )
        data = response.json()
        return KeyRing(**data)

    async def list_keyrings(
        self, project_id: str, location_id: str
    ) -> List[KeyRing]:
        """List all keyrings in a location"""
        response = await self._request(
            "GET", f"/projects/{project_id}/locations/{location_id}/keyrings"
        )
        data = response.json()
        return [KeyRing(**item) for item in data["keyrings"]]

    async def get_keyring(
        self, project_id: str, location_id: str, keyring_id: str
    ) -> KeyRing:
        """Get a keyring by ID"""
        response = await self._request(
            "GET",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}",
        )
        data = response.json()
        return KeyRing(**data)

    # 🔑 KMS Key operations

    async def create_crypto_key(
        self,
        project_id: str,
        location_id: str,
        keyring_id: str,
        request: CreateCryptoKeyRequest,
    ) -> KMSKey:
        """Create a new crypto key"""
        response = await self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys",
            json_data=request.__dict__,
        )
        data = response.json()
        return KMSKey(**data)

    async def list_crypto_keys(
        self, project_id: str, location_id: str, keyring_id: str
    ) -> List[KMSKey]:
        """List all crypto keys in a keyring"""
        response = await self._request(
            "GET",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys",
        )
        data = response.json()
        return [KMSKey(**item) for item in data["keys"]]

    async def get_crypto_key(
        self, project_id: str, location_id: str, keyring_id: str, key_id: str
    ) -> KMSKey:
        """Get a crypto key by ID"""
        response = await self._request(
            "GET",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}",
        )
        data = response.json()
        return KMSKey(**data)

    async def rotate_crypto_key(
        self, project_id: str, location_id: str, keyring_id: str, key_id: str
    ) -> KMSKey:
        """Rotate a crypto key"""
        response = await self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}/rotate",
        )
        data = response.json()
        return KMSKey(**data)

    # 🔐 Cryptographic operations

    async def encrypt(
        self,
        project_id: str,
        location_id: str,
        keyring_id: str,
        key_id: str,
        plaintext: Union[str, bytes],
    ) -> EncryptResponse:
        """Encrypt data using a crypto key"""
        if isinstance(plaintext, str):
            plaintext_bytes = plaintext.encode()
        else:
            plaintext_bytes = plaintext

        plaintext_base64 = base64.b64encode(plaintext_bytes).decode()

        response = await self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}/encrypt",
            json_data={"plaintext": plaintext_base64},
        )
        data = response.json()
        return EncryptResponse(**data)

    async def decrypt(
        self,
        project_id: str,
        location_id: str,
        keyring_id: str,
        key_id: str,
        ciphertext: str,
        key_version: int,
    ) -> DecryptResponse:
        """Decrypt data using a crypto key"""
        response = await self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}/decrypt",
            json_data={"ciphertext": ciphertext, "key_version": key_version},
        )
        data = response.json()
        return DecryptResponse(**data)

    # 🔑 Client configuration operations

    async def create_client_config(
        self, request: CreateClientConfigRequest
    ) -> ClientConfigInfo:
        """Create a new client configuration"""
        response = await self._request(
            "POST", "/client-configs", json_data=request.__dict__
        )
        data = response.json()
        return ClientConfigInfo(**data)

    async def list_client_configs(self) -> List[ClientConfigInfo]:
        """List all client configurations"""
        response = await self._request("GET", "/client-configs")
        data = response.json()
        return [ClientConfigInfo(**item) for item in data["configs"]]

    async def get_client_config(self, config_id: str) -> ClientConfigInfo:
        """Get a client configuration by ID"""
        response = await self._request("GET", f"/client-configs/{config_id}")
        data = response.json()
        return ClientConfigInfo(**data)

    async def revoke_client_config(self, config_id: str) -> ClientConfigInfo:
        """Revoke a client configuration"""
        response = await self._request("POST", f"/client-configs/{config_id}/revoke")
        data = response.json()
        return ClientConfigInfo(**data)

    # 🔑 Get cached public key
    def _get_public_key(self, pem_key: str) -> rsa.RSAPublicKey:
        if pem_key in self._key_cache:
            return self._key_cache[pem_key]

        key = load_pem_public_key(pem_key.encode())
        self._key_cache[pem_key] = key
        return key

    # 🔑 Get cached private key
    def _get_private_key(self, pem_key: str) -> rsa.RSAPrivateKey:
        if pem_key in self._key_cache:
            return self._key_cache[pem_key]

        key = load_pem_private_key(pem_key.encode(), password=None)
        self._key_cache[pem_key] = key
        return key 