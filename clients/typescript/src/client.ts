import { Buffer } from 'buffer';
import {
  ClientConfig,
  ClientConfigInfo,
  RetryConfig,
  Project,
  Location,
  KeyRing,
  KMSKey,
  KeyPair,
  CreateProjectRequest,
  CreateCryptoKeyRequest,
  CreateClientConfigRequest,
  GenerateKeyPairRequest,
  EncryptRequest,
  EncryptResponse,
  DecryptRequest,
  DecryptResponse,
  EncryptedEnvelope,
} from './types';
import {
  KMSError,
  AuthenticationError,
  NotFoundError,
  PermissionError,
  InvalidInputError,
  RateLimitError,
} from './errors';

/**
 * 🔐 Client for interacting with Xyphos KMS
 */
export class LambdaKMSClient {
  private config: Required<ClientConfig> & { retryConfig: Required<RetryConfig> };
  private token: string | null = null;
  private tokenExpiry: number | null = null;
  private keyCache: Map<string, KMSKey> = new Map();
  private encoder: TextEncoder;
  private decoder: TextDecoder;
  private keyPair: KeyPair | null = null;

  constructor(config: ClientConfig) {
    const defaultRetryConfig: Required<RetryConfig> = {
      maxRetries: 3,
      backoffFactor: 0.1,
      statusForcelist: [500, 502, 503, 504, 429],
      initialWait: 100,
      maxWait: 10000,
    };

    this.config = {
      timeout: 30000,
      retryConfig: {
        ...defaultRetryConfig,
        ...(config.retryConfig || {}),
      },
      ...config,
    } as Required<ClientConfig> & { retryConfig: Required<RetryConfig> };

    this.encoder = new TextEncoder();
    this.decoder = new TextDecoder();
    this.keyPair = config.keyPair || null;

    // Load stored key pair if available
    if (!this.keyPair) {
      this.loadStoredKeyPair();
    }
  }

  /**
   * 🔄 Perform a fetch request with retries and encryption
   */
  private async fetchWithRetry(
    path: string,
    init: RequestInit,
    retryCount = 0
  ): Promise<Response> {
    try {
      // Add encryption headers if key pair is available
      const headers: Record<string, string> = {
        ...init.headers as Record<string, string>,
        'User-Agent': 'XyphosKMS-TS-Client/1.0',
      };

      // Always use encryption if key pair is available
      if (this.keyPair) {
        headers['X-Request-Encrypted'] = 'true';
        headers['X-Response-Encrypted'] = 'true';
        headers['Content-Type'] = 'application/octet-stream';
        
        // Add HSM encryption header for sensitive operations
        if (path.includes('/encrypt') || path.includes('/decrypt')) {
          headers['X-HSM-Encrypt'] = 'true';
        }
      }

      const response = await fetch(`${this.config.baseURL}/api${path}`, {
        ...init,
        headers,
      });

      if (response.status === 429) {
        throw new RateLimitError('Rate limit exceeded');
      }

      if (
        !response.ok &&
        this.config.retryConfig.statusForcelist.includes(response.status) &&
        retryCount < this.config.retryConfig.maxRetries
      ) {
        const waitTime = Math.min(
          this.config.retryConfig.initialWait * 
          Math.pow(this.config.retryConfig.backoffFactor, retryCount),
          this.config.retryConfig.maxWait
        );
        await new Promise(resolve => setTimeout(resolve, waitTime));
        return this.fetchWithRetry(path, init, retryCount + 1);
      }

      if (response.status === 401) {
        this.token = null;
        throw new AuthenticationError('Authentication failed');
      }

      if (response.status === 403) {
        throw new PermissionError('Permission denied');
      }

      if (response.status === 404) {
        throw new NotFoundError('Resource not found');
      }

      if (response.status === 400) {
        throw new InvalidInputError('Invalid input');
      }

      return response;
    } catch (error) {
      if (error instanceof KMSError) {
        throw error;
      }
      throw new KMSError(`Request failed: ${(error as Error).message}`);
    }
  }

  /**
   * 🔑 Ensure a valid token is available
   */
  private async ensureToken(): Promise<void> {
    if (this.token && this.tokenExpiry && Date.now() < this.tokenExpiry) {
      return;
    }

    const response = await this.fetchWithRetry('/oauth/token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret,
        grant_type: 'client_credentials',
      }),
    });

    const data = await response.json();
    this.token = data.access_token;
    this.tokenExpiry = Date.now() + (data.expires_in * 1000);
  }

  /**
   * 📝 Get headers with authorization token
   */
  private async getHeaders(): Promise<HeadersInit> {
    await this.ensureToken();
    return {
      'Authorization': `Bearer ${this.token}`,
      'Content-Type': 'application/json',
    };
  }

  // 📦 Project operations

  async createProject(request: CreateProjectRequest): Promise<Project> {
    const body = await this.prepareRequestBody(request);
    const response = await this.fetchWithRetry('/projects', {
      method: 'POST',
      headers: await this.getHeaders(),
      body,
    });

    return this.processEncryptedResponse<Project>(response);
  }

  async listProjects(): Promise<Project[]> {
    const response = await this.fetchWithRetry('/projects', {
      method: 'GET',
      headers: await this.getHeaders(),
    });
    return this.processEncryptedResponse<Project[]>(response);
  }

  async getProject(projectId: string): Promise<Project> {
    const response = await this.fetchWithRetry(`/projects/${projectId}`, {
      method: 'GET',
      headers: await this.getHeaders(),
    });
    return this.processEncryptedResponse<Project>(response);
  }

  // 📍 Location operations

  async listLocations(projectId: string): Promise<Location[]> {
    const response = await this.fetchWithRetry(`/projects/${projectId}/locations`, {
      method: 'GET',
      headers: await this.getHeaders(),
    });
    return response.json();
  }

  async getLocation(projectId: string, locationId: string): Promise<Location> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  // 💍 KeyRing operations

  async createKeyRing(
    projectId: string,
    locationId: string,
    name: string
  ): Promise<KeyRing> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({ name }),
      }
    );
    return response.json();
  }

  async listKeyRings(
    projectId: string,
    locationId: string
  ): Promise<KeyRing[]> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  async getKeyRing(
    projectId: string,
    locationId: string,
    keyringId: string
  ): Promise<KeyRing> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  // 🔑 KMS Key operations

  async createCryptoKey(
    projectId: string,
    locationId: string,
    keyringId: string,
    request: CreateCryptoKeyRequest
  ): Promise<KMSKey> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify(request),
      }
    );
    return response.json();
  }

  async listCryptoKeys(
    projectId: string,
    locationId: string,
    keyringId: string
  ): Promise<KMSKey[]> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  async getCryptoKey(
    projectId: string,
    locationId: string,
    keyringId: string,
    keyId: string
  ): Promise<KMSKey> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys/${keyId}`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  async rotateCryptoKey(
    projectId: string,
    locationId: string,
    keyringId: string,
    keyId: string
  ): Promise<KMSKey> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys/${keyId}/rotate`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  // 🔐 Cryptographic operations

  async encrypt(
    projectId: string,
    locationId: string,
    keyringId: string,
    keyId: string,
    plaintext: string | Uint8Array
  ): Promise<EncryptResponse> {
    const plaintextBase64 = typeof plaintext === 'string'
      ? Buffer.from(plaintext).toString('base64')
      : Buffer.from(plaintext).toString('base64');

    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys/${keyId}/encrypt`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({ plaintext: plaintextBase64 }),
      }
    );
    return response.json();
  }

  async decrypt(
    projectId: string,
    locationId: string,
    keyringId: string,
    keyId: string,
    ciphertext: string,
    keyVersion: number
  ): Promise<DecryptResponse> {
    const response = await this.fetchWithRetry(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys/${keyId}/decrypt`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({ ciphertext, keyVersion }),
      }
    );
    return response.json();
  }

  // 🔑 Client Configuration operations

  async createClientConfig(request: CreateClientConfigRequest): Promise<ClientConfigInfo> {
    const response = await this.fetchWithRetry('/client-configs', {
      method: 'POST',
      headers: await this.getHeaders(),
      body: JSON.stringify(request),
    });
    return response.json();
  }

  async listClientConfigs(): Promise<ClientConfigInfo[]> {
    const response = await this.fetchWithRetry('/client-configs', {
      method: 'GET',
      headers: await this.getHeaders(),
    });
    return response.json();
  }

  async getClientConfig(configId: string): Promise<ClientConfigInfo> {
    const response = await this.fetchWithRetry(`/client-configs/${configId}`, {
      method: 'GET',
      headers: await this.getHeaders(),
    });
    return response.json();
  }

  async revokeClientConfig(configId: string): Promise<ClientConfigInfo> {
    const response = await this.fetchWithRetry(`/client-configs/${configId}/revoke`, {
      method: 'POST',
      headers: await this.getHeaders(),
    });
    return response.json();
  }

  /**
   * 🔑 Generate a new key pair for end-to-end encryption
   */
  async generateKeyPair(request: GenerateKeyPairRequest): Promise<KeyPair> {
    const response = await this.fetchWithRetry('/key-pairs/generate', {
      method: 'POST',
      headers: await this.getHeaders(),
      body: JSON.stringify(request),
    });

    const keyPair = await response.json();
    this.keyPair = keyPair;

    // Store the key pair in localStorage for persistence if available
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('xyphos_key_pair', JSON.stringify(keyPair));
    }

    return keyPair;
  }

  /**
   * 🔐 Encrypt data with ChaCha20 and RSA
   */
  private async encryptWithServerKey(data: string | Uint8Array): Promise<EncryptedEnvelope> {
    if (!this.keyPair) {
      throw new Error('No key pair available. Call generateKeyPair first.');
    }

    // Generate random key for ChaCha20
    const chacha20Key = crypto.getRandomValues(new Uint8Array(32));
    const iv = crypto.getRandomValues(new Uint8Array(12));

    // Convert data to Uint8Array
    const dataBytes = typeof data === 'string'
      ? this.encoder.encode(data)
      : data;

    // Encrypt data with ChaCha20
    const encryptedData = await crypto.subtle.encrypt(
      {
        name: 'CHACHA20-POLY1305',
        iv,
      },
      await crypto.subtle.importKey(
        'raw',
        chacha20Key,
        'CHACHA20-POLY1305',
        false,
        ['encrypt']
      ),
      dataBytes
    );

    // Encrypt ChaCha20 key with RSA
    const response = await this.fetchWithRetry('/crypto/encrypt-envelope', {
      method: 'POST',
      headers: {
        ...(await this.getHeaders()),
        'X-Request-Encrypted': 'true',
        'X-Response-Encrypted': 'true',
        'Content-Type': 'application/octet-stream',
      },
      body: JSON.stringify({
        data: Buffer.from(chacha20Key).toString('base64'),
        publicKey: this.keyPair.publicKey,
        algorithm: this.keyPair.algorithm,
        iv: Buffer.from(iv).toString('base64'),
      }),
    });

    const envelope = await response.json();
    return {
      ...envelope,
      encryptedData: Buffer.from(encryptedData).toString('base64'),
    };
  }

  /**
   * 🔓 Decrypt data with ChaCha20 and RSA
   */
  private async decryptWithClientKey(envelope: EncryptedEnvelope): Promise<string> {
    if (!this.keyPair) {
      throw new Error('No key pair available. Call generateKeyPair first.');
    }

    // Decrypt the ChaCha20 key with RSA
    const response = await this.fetchWithRetry('/crypto/decrypt-envelope', {
      method: 'POST',
      headers: {
        ...(await this.getHeaders()),
        'X-Request-Encrypted': 'true',
        'X-Response-Encrypted': 'true',
        'Content-Type': 'application/octet-stream',
      },
      body: JSON.stringify({
        envelope,
        privateKey: this.keyPair.privateKey,
        algorithm: this.keyPair.algorithm,
      }),
    });

    const { key, iv } = await response.json();
    const chacha20Key = Buffer.from(key, 'base64');
    const ivBytes = Buffer.from(iv, 'base64');
    const encryptedData = Buffer.from(envelope.encryptedData, 'base64');

    // Decrypt data with ChaCha20
    const decryptedData = await crypto.subtle.decrypt(
      {
        name: 'CHACHA20-POLY1305',
        iv: ivBytes,
      },
      await crypto.subtle.importKey(
        'raw',
        chacha20Key,
        'CHACHA20-POLY1305',
        false,
        ['decrypt']
      ),
      encryptedData
    );

    return this.decoder.decode(decryptedData);
  }

  /**
   * 🔄 Process response with end-to-end encryption
   */
  private async processEncryptedResponse<T>(response: Response): Promise<T> {
    // Check if response is encrypted
    if (response.headers.get('X-Response-Encrypted') !== 'true') {
      return response.json();
    }

    // Get response as text since it's base64 encoded
    const encodedData = await response.text();
    
    try {
      // Base64 decode and decrypt
      const envelope: EncryptedEnvelope = JSON.parse(
        Buffer.from(encodedData, 'base64').toString()
      );
      
      const decryptedData = await this.decryptWithClientKey(envelope);
      return JSON.parse(decryptedData) as T;
    } catch (error) {
      throw new KMSError(`Failed to decrypt response: ${(error as Error).message}`);
    }
  }

  /**
   * 🔄 Prepare request body with encryption if needed
   */
  private async prepareRequestBody(data: any): Promise<string> {
    const jsonData = JSON.stringify(data);
    
    if (!this.keyPair) {
      return jsonData;
    }

    const envelope = await this.encryptWithServerKey(jsonData);
    return Buffer.from(JSON.stringify(envelope)).toString('base64');
  }

  /**
   * 🔑 Get the current key pair
   */
  getKeyPair(): KeyPair | null {
    return this.keyPair;
  }

  /**
   * 🔄 Set a new key pair
   */
  setKeyPair(keyPair: KeyPair | null): void {
    this.keyPair = keyPair;

    // Update localStorage if available
    if (typeof localStorage !== 'undefined') {
      if (keyPair) {
        localStorage.setItem('xyphos_key_pair', JSON.stringify(keyPair));
      } else {
        localStorage.removeItem('xyphos_key_pair');
      }
    }
  }

  /**
   * 🔄 Load key pair from storage if available
   */
  private loadStoredKeyPair(): void {
    if (typeof localStorage !== 'undefined') {
      const storedKeyPair = localStorage.getItem('xyphos_key_pair');
      if (storedKeyPair) {
        try {
          this.keyPair = JSON.parse(storedKeyPair);
        } catch (error) {
          console.warn('Failed to parse stored key pair:', error);
          localStorage.removeItem('xyphos_key_pair');
        }
      }
    }
  }
}

// 🔑 Key cache for storing parsed keys
class KeyCache {
  private static instance: KeyCache;
  private cache: Map<string, KMSKey>;

  private constructor() {
    this.cache = new Map();
  }

  static getInstance(): KeyCache {
    if (!KeyCache.instance) {
      KeyCache.instance = new KeyCache();
    }
    return KeyCache.instance;
  }

  async get(key: string): Promise<KMSKey | undefined> {
    return this.cache.get(key);
  }

  set(key: string, value: KMSKey): void {
    this.cache.set(key, value);
  }
} 