import base64
import json
import os
import time
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Dict, Optional, Union

import httpx
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding, rsa
from cryptography.hazmat.primitives.ciphers.aead import ChaCha20Poly1305
from cryptography.hazmat.primitives.serialization import load_pem_private_key, load_pem_public_key

# 🔐 Client configuration
@dataclass
class ClientConfig:
    base_url: str
    client_id: str
    client_secret: str
    private_key: str
    public_key: str
    timeout: float = 30.0

# 🔄 Retry configuration
@dataclass
class RetryConfig:
    max_retries: int = 3
    initial_wait: float = 1.0
    max_wait: float = 10.0

# 🌐 Main client class
class Client:
    def __init__(self, config: ClientConfig):
        self.config = config
        self.token = None
        self.token_expiry = None
        self._key_cache = {}
        self._http_client = httpx.Client(timeout=config.timeout)

    # 🔒 Encrypt data
    async def encrypt(
        self,
        project_id: str,
        location_id: str,
        keyring_id: str,
        key_id: str,
        plaintext: Union[str, bytes],
    ) -> bytes:
        # Ensure we have bytes
        if isinstance(plaintext, str):
            data = plaintext.encode()
        else:
            data = plaintext

        # Generate ChaCha20Poly1305 key
        key = os.urandom(32)
        cipher = ChaCha20Poly1305(key)
        nonce = os.urandom(12)

        # Encrypt data
        ciphertext = cipher.encrypt(nonce, data, None)

        # Get RSA public key
        public_key = self._get_public_key(self.config.public_key)

        # Encrypt symmetric key
        encrypted_key = public_key.encrypt(
            key,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None,
            ),
        )

        # Combine all parts
        key_length = len(encrypted_key).to_bytes(4, byteorder='little')
        combined = key_length + encrypted_key + nonce + ciphertext

        # Base64 encode for transport
        return combined

    # 🔓 Decrypt data
    async def decrypt(
        self,
        project_id: str,
        location_id: str,
        keyring_id: str,
        key_id: str,
        ciphertext: bytes,
    ) -> bytes:
        # Get RSA private key
        private_key = self._get_private_key(self.config.private_key)

        # Extract parts
        key_length = int.from_bytes(ciphertext[:4], byteorder='little')
        encrypted_key = ciphertext[4:4 + key_length]
        nonce = ciphertext[4 + key_length:4 + key_length + 12]
        encrypted_data = ciphertext[4 + key_length + 12:]

        # Decrypt symmetric key
        key = private_key.decrypt(
            encrypted_key,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None,
            ),
        )

        # Decrypt data
        cipher = ChaCha20Poly1305(key)
        decrypted = cipher.decrypt(nonce, encrypted_data, None)

        return decrypted

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

    # 🔄 Ensure valid token
    async def _ensure_token(self):
        if self.token and self.token_expiry and datetime.now() < self.token_expiry:
            return

        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{self.config.base_url}/api/v1/oauth/token",
                json={
                    "client_id": self.config.client_id,
                    "client_secret": self.config.client_secret,
                    "grant_type": "client_credentials",
                },
            )

            if response.status_code != 200:
                raise Exception(f"Failed to get token: {response.text}")

            data = response.json()
            self.token = data["access_token"]
            self.token_expiry = datetime.now() + timedelta(seconds=data["expires_in"])

    # 🔄 Make authenticated request
    async def _request(
        self,
        method: str,
        path: str,
        data: Optional[Union[str, bytes]] = None,
        headers: Optional[Dict[str, str]] = None,
    ) -> httpx.Response:
        await self._ensure_token()

        if headers is None:
            headers = {}

        headers["Authorization"] = f"Bearer {self.token}"

        async with httpx.AsyncClient() as client:
            response = await client.request(
                method,
                f"{self.config.base_url}/api{path}",
                content=data,
                headers=headers,
            )

            if response.status_code == 401:
                self.token = None
                await self._ensure_token()
                headers["Authorization"] = f"Bearer {self.token}"
                response = await client.request(
                    method,
                    f"{self.config.base_url}/api{path}",
                    content=data,
                    headers=headers,
                )

            return response 