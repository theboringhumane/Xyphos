from dataclasses import dataclass
from datetime import datetime
from typing import List, Optional

# 🔧 Configuration models
@dataclass
class RetryConfig:
    """🔄 Configuration for retry behavior"""
    max_retries: int = 3
    backoff_factor: float = 0.1
    status_forcelist: List[int] = None
    initial_wait: float = 0.1
    max_wait: float = 10.0

    def __post_init__(self):
        if self.status_forcelist is None:
            self.status_forcelist = [500, 502, 503, 504, 429]

@dataclass
class ClientConfig:
    """🔧 Configuration for the Xyphos client"""
    base_url: str
    client_id: str  # 🔑 Client ID from the KMS service
    client_secret: str  # 🔐 Client secret from the KMS service
    timeout: float = 30.0
    retry_config: Optional[RetryConfig] = None

# 📦 Resource models
@dataclass
class Project:
    """📦 Project information"""
    id: str
    name: str
    description: str
    created_at: datetime

@dataclass
class Location:
    """📍 Location information"""
    id: str
    name: str
    created_at: datetime

@dataclass
class KeyRing:
    """💍 KeyRing information"""
    id: str
    name: str
    created_at: datetime

@dataclass
class KMSKey:
    """🔑 KMS Key information"""
    id: str
    name: str
    algorithm: str
    purpose: str
    rotation_period: int
    state: str
    version: int
    next_rotation: datetime
    created_at: datetime

@dataclass
class ClientConfigInfo:
    """🔐 Client configuration information"""
    id: str
    name: str
    permissions: List[str]
    expires_at: datetime
    status: str
    last_used_at: datetime

# 🔐 Request/Response models
@dataclass
class CreateProjectRequest:
    """📝 Create project request"""
    name: str
    description: str

@dataclass
class CreateCryptoKeyRequest:
    """🔑 Create crypto key request"""
    name: str
    algorithm: str
    purpose: str
    rotation_period: int

@dataclass
class CreateClientConfigRequest:
    """🎫 Create client config request"""
    name: str
    permissions: List[str]
    expires_in: str

@dataclass
class EncryptResponse:
    """🔒 Encryption response"""
    ciphertext: str
    key_version: int

@dataclass
class DecryptResponse:
    """🔓 Decryption response"""
    plaintext: str 