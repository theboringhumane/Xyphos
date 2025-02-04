/**
 * 🔄 Configuration for retry behavior
 */
export interface RetryConfig {
  maxRetries: number;
  backoffFactor: number;
  statusForcelist: number[];
  initialWait: number;
  maxWait: number;
}

/**
 * 🔑 Key pair information
 */
export interface KeyPair {
  publicKey: string;  // 🔐 Base64 encoded public key
  privateKey: string; // 🔐 Base64 encoded private key
  algorithm: string;  // 🔧 Key algorithm (e.g., 'RSA-OAEP-4096')
  expiresAt: string; // ⏰ Key expiration timestamp
}

/**
 * 🔧 Configuration for initializing the Xyphos client
 */
export interface ClientInit {
  baseURL: string;
  clientId: string;  // 🔑 Client ID for authentication
  clientSecret: string;  // 🔐 Client secret for authentication
  timeout?: number;
  retryConfig?: Partial<RetryConfig>;
}

/**
 * 🔧 Configuration for the Xyphos client
 */
export interface ClientConfig {
  baseURL: string;
  clientId: string;  // 🔑 Client ID for authentication
  clientSecret: string;  // 🔐 Client secret for authentication
  timeout?: number;
  retryConfig?: Partial<RetryConfig>;
  keyPair?: KeyPair;  // 🔑 Client key pair for end-to-end encryption
}

/**
 * 📦 Project information
 */
export interface Project {
  id: string;
  name: string;
  description: string;
  createdAt: string;
}

/**
 * 📍 Location information
 */
export interface Location {
  id: string;
  name: string;
  createdAt: string;
}

/**
 * 💍 KeyRing information
 */
export interface KeyRing {
  id: string;
  name: string;
  createdAt: string;
}

/**
 * 🔑 KMS Key information
 */
export interface KMSKey {
  id: string;
  name: string;
  algorithm: string;
  purpose: string;
  rotationPeriod: number;
  state: string;
  version: number;
  nextRotation: string;
  createdAt: string;
}

/**
 * 🔐 Client configuration information from the KMS service
 */
export interface ClientConfigInfo {
  id: string;
  name: string;
  permissions: string[];
  expiresAt: string;
  status: string;
  lastUsedAt: string;
}

/**
 * 🔒 Encryption request
 */
export interface EncryptRequest {
  plaintext: string;
}

/**
 * 🔓 Encryption response
 */
export interface EncryptResponse {
  ciphertext: string;
  keyVersion: number;
}

/**
 * 🔐 Decryption request
 */
export interface DecryptRequest {
  ciphertext: string;
  keyVersion: number;
}

/**
 * 🔓 Decryption response
 */
export interface DecryptResponse {
  plaintext: string;
}

/**
 * 📝 Create project request
 */
export interface CreateProjectRequest {
  name: string;
  description: string;
}

/**
 * 🔑 Create crypto key request
 */
export interface CreateCryptoKeyRequest {
  name: string;
  algorithm: string;
  purpose: string;
  rotationPeriod: number;
}

/**
 * 🎫 Create client config request
 */
export interface CreateClientConfigRequest {
  name: string;
  permissions: string[];
  expiresIn: string;
}

/**
 * 🔑 Generate key pair request
 */
export interface GenerateKeyPairRequest {
  algorithm: string;  // 🔧 Key algorithm (e.g., 'RSA-OAEP-4096')
  expiresIn: string; // ⏰ Key pair validity duration (e.g., '24h')
}

/**
 * 🔐 Encrypted response envelope
 */
export interface EncryptedEnvelope {
  encryptedData: string;  // 🔒 Base64 encoded encrypted data
  keyId: string;         // 🔑 ID of the key used for encryption
  algorithm: string;     // 🔧 Encryption algorithm used
  iv: string;           // 🎲 Base64 encoded initialization vector
  tag: string;          // ✅ Base64 encoded authentication tag
} 