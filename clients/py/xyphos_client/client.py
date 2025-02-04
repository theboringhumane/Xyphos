import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
import time
import base64
from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional, List, Dict, Any

from .exceptions import (
    AuthenticationError,
    NotFoundError,
    PermissionError,
    InvalidInputError,
    RateLimitError,
    ServerError,
)


@dataclass
class RetryConfig:
    """🔄 Configuration for retry behavior"""
    max_retries: int = 3
    initial_wait: float = 0.1
    max_wait: float = 2.0


@dataclass
class ClientConfig:
    """🔧 Configuration for the Xyphos client"""
    base_url: str
    client_id: str
    client_secret: str
    timeout: int = 30
    retry_config: RetryConfig = field(default_factory=RetryConfig)


@dataclass
class Project:
    """📦 Project resource"""
    id: str
    name: str
    description: str
    created_at: datetime

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'Project':
        return cls(
            id=data['id'],
            name=data['name'],
            description=data['description'],
            created_at=datetime.fromisoformat(data['created_at'])
        )


@dataclass
class Location:
    """📍 Location resource"""
    id: str
    name: str
    created_at: datetime

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'Location':
        return cls(
            id=data['id'],
            name=data['name'],
            created_at=datetime.fromisoformat(data['created_at'])
        )


@dataclass
class KeyRing:
    """💍 KeyRing resource"""
    id: str
    name: str
    created_at: datetime

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'KeyRing':
        return cls(
            id=data['id'],
            name=data['name'],
            created_at=datetime.fromisoformat(data['created_at'])
        )


@dataclass
class CryptoKey:
    """🔑 CryptoKey resource"""
    id: str
    name: str
    algorithm: str
    purpose: str
    rotation_period: int
    created_at: datetime
    next_rotation: datetime
    version: int

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'CryptoKey':
        return cls(
            id=data['id'],
            name=data['name'],
            algorithm=data['algorithm'],
            purpose=data['purpose'],
            rotation_period=data['rotation_period'],
            created_at=datetime.fromisoformat(data['created_at']),
            next_rotation=datetime.fromisoformat(data['next_rotation']),
            version=data['version']
        )


class XyphosClient:
    """🔐 Client for interacting with Xyphos KMS"""

    def __init__(self, base_url: str, client_id: str, client_secret: str,
                 timeout: int = 30, retry_config: Optional[RetryConfig] = None):
        """Initialize the client with the given configuration"""
        self.config = ClientConfig(
            base_url=base_url,
            client_id=client_id,
            client_secret=client_secret,
            timeout=timeout,
            retry_config=retry_config or RetryConfig()
        )

        self._session = requests.Session()
        retry_strategy = Retry(
            total=self.config.retry_config.max_retries,
            backoff_factor=self.config.retry_config.initial_wait,
            status_forcelist=[500, 502, 503, 504, 429]
        )
        adapter = HTTPAdapter(max_retries=retry_strategy)
        self._session.mount("http://", adapter)
        self._session.mount("https://", adapter)

        self._token: Optional[str] = None
        self._token_expiry: Optional[float] = None

    def _ensure_token(self) -> None:
        """🔑 Ensure a valid token is available"""
        if self._token and self._token_expiry and time.time() < self._token_expiry:
            return

        response = self._session.post(
            f"{self.config.base_url}/api/oauth/token",
            json={
                "client_id": self.config.client_id,
                "client_secret": self.config.client_secret,
                "grant_type": "client_credentials"
            },
            timeout=self.config.timeout
        )

        if response.status_code == 401:
            raise AuthenticationError("Invalid credentials")
        elif response.status_code != 200:
            raise ServerError(f"Failed to obtain token: {response.status_code}")

        data = response.json()
        self._token = data["access_token"]
        self._token_expiry = time.time() + data["expires_in"]

    def _request(self, method: str, path: str, **kwargs) -> Dict[str, Any]:
        """🌐 Make an authenticated request to the API"""
        self._ensure_token()

        headers = {
            "Authorization": f"Bearer {self._token}",
            "Content-Type": "application/json",
        }

        if "headers" in kwargs:
            headers.update(kwargs.pop("headers"))

        response = self._session.request(
            method,
            f"{self.config.base_url}/api{path}",
            headers=headers,
            timeout=self.config.timeout,
            **kwargs
        )

        if response.status_code == 401:
            self._token = None
            raise AuthenticationError("Authentication failed")
        elif response.status_code == 403:
            raise PermissionError("Permission denied")
        elif response.status_code == 404:
            raise NotFoundError("Resource not found")
        elif response.status_code == 400:
            raise InvalidInputError(response.json().get("message", "Invalid input"))
        elif response.status_code == 429:
            raise RateLimitError("Rate limit exceeded")
        elif response.status_code >= 500:
            raise ServerError(f"Server error: {response.status_code}")

        return response.json()

    # 📦 Project Operations
    def create_project(self, name: str, description: str) -> Project:
        """Create a new project"""
        data = self._request("POST", "/projects", json={
            "name": name,
            "description": description
        })
        return Project.from_dict(data)

    def list_projects(self) -> List[Project]:
        """List all projects"""
        data = self._request("GET", "/projects")
        return [Project.from_dict(item) for item in data]

    def get_project(self, project_id: str) -> Project:
        """Get a project by ID"""
        data = self._request("GET", f"/projects/{project_id}")
        return Project.from_dict(data)

    # 💍 KeyRing Operations
    def create_keyring(self, project_id: str, location_id: str, name: str) -> KeyRing:
        """Create a new keyring"""
        data = self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings",
            json={"name": name}
        )
        return KeyRing.from_dict(data)

    def list_keyrings(self, project_id: str, location_id: str) -> List[KeyRing]:
        """List all keyrings in a location"""
        data = self._request("GET", f"/projects/{project_id}/locations/{location_id}/keyrings")
        return [KeyRing.from_dict(item) for item in data]

    def get_keyring(self, project_id: str, location_id: str, keyring_id: str) -> KeyRing:
        """Get a keyring by ID"""
        data = self._request(
            "GET",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}"
        )
        return KeyRing.from_dict(data)

    # 🔑 CryptoKey Operations
    def create_crypto_key(
            self, project_id: str, location_id: str, keyring_id: str,
            name: str, algorithm: str, purpose: str, rotation_period: int
    ) -> CryptoKey:
        """Create a new crypto key"""
        data = self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys",
            json={
                "name": name,
                "algorithm": algorithm,
                "purpose": purpose,
                "rotation_period": rotation_period
            }
        )
        return CryptoKey.from_dict(data)

    def list_crypto_keys(
            self, project_id: str, location_id: str, keyring_id: str
    ) -> List[CryptoKey]:
        """List all crypto keys in a keyring"""
        data = self._request(
            "GET",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys"
        )
        return [CryptoKey.from_dict(item) for item in data]

    def get_crypto_key(
            self, project_id: str, location_id: str, keyring_id: str, key_id: str
    ) -> CryptoKey:
        """Get a crypto key by ID"""
        data = self._request(
            "GET",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}"
        )
        return CryptoKey.from_dict(data)

    def rotate_crypto_key(
            self, project_id: str, location_id: str, keyring_id: str, key_id: str
    ) -> CryptoKey:
        """Rotate a crypto key"""
        data = self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}:rotate"
        )
        return CryptoKey.from_dict(data)

    # 🔒 Cryptographic Operations
    def encrypt(
            self, project_id: str, location_id: str, keyring_id: str,
            key_id: str, plaintext: str
    ) -> Dict[str, str]:
        """Encrypt data using a crypto key"""
        data = self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}:encrypt",
            json={"plaintext": base64.b64encode(plaintext.encode()).decode()}
        )
        return {
            "ciphertext": data["ciphertext"],
            "key_version": data["key_version"]
        }

    def decrypt(
            self, project_id: str, location_id: str, keyring_id: str,
            key_id: str, ciphertext: str
    ) -> Dict[str, str]:
        """Decrypt data using a crypto key"""
        data = self._request(
            "POST",
            f"/projects/{project_id}/locations/{location_id}/keyrings/{keyring_id}/keys/{key_id}:decrypt",
            json={"ciphertext": ciphertext}
        )
        return {
            "plaintext": base64.b64decode(data["plaintext"]).decode()
        }
