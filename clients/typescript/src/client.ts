import { Buffer } from 'buffer';
import { ClientConfig, KeyInfo, RetryConfig } from './types';
import {
  KMSError,
  AuthenticationError,
  NotFoundError,
  PermissionError,
  InvalidInputError,
} from './errors';

/**
 * 🔐 Client for interacting with Xyphos
 */
export class LambdaKMSClient {
  private config: Required<ClientConfig>;
  private token: string | null = null;
  private tokenExpiry: number | null = null;
  private privateKey: CryptoKey | null = null;
  private publicKey: CryptoKey | null = null;
  private keyCache: KeyCache;
  private encoder: TextEncoder;
  private decoder: TextDecoder;

  constructor(config: ClientConfig) {
    this.config = {
      timeout: 30000,
      retryConfig: {
        maxRetries: 3,
        backoffFactor: 0.1,
        statusForcelist: [500, 502, 503, 504, 429],
      },
      ...config,
    };
    this.keyCache = KeyCache.getInstance();
    this.encoder = new TextEncoder();
    this.decoder = new TextDecoder();
  }

  /**
   * 🔄 Perform a fetch request with retries
   */
  private async fetchWithRetry(
    url: string,
    options: RequestInit,
    attempt: number = 0
  ): Promise<Response> {
    try {
      const response = await fetch(url, {
        ...options,
        signal: AbortSignal.timeout(this.config.timeout),
      });

      if (// @ts-ignore
        attempt < this.config?.retryConfig?.maxRetries &&
        (response.status >= 500 || response.status === 429)
      ) {
        const backoffMs =
            // @ts-ignore
          Math.min(1000 * Math.pow(2, attempt) * this.config.retryConfig.backoffFactor, 10000);
        await new Promise(resolve => setTimeout(resolve, backoffMs));
        return this.fetchWithRetry(url, options, attempt + 1);
      }

      if (!response.ok) {
        await this.handleErrorResponse(response);
      }

      return response;
    } catch (error) {
      if (
          // @ts-ignore
        attempt < this.config?.retryConfig?.maxRetries &&
        error instanceof TypeError
      ) {
        const backoffMs =
            // @ts-ignore
          Math.min(1000 * Math.pow(2, attempt) * this.config.retryConfig.backoffFactor, 10000);
        await new Promise(resolve => setTimeout(resolve, backoffMs));
        return this.fetchWithRetry(url, options, attempt + 1);
      }
      throw error;
    }
  }

  /**
   * 🚨 Handle error responses from the API
   */
  private async handleErrorResponse(response: Response): Promise<never> {
    switch (response.status) {
      case 401:
        throw new AuthenticationError();
      case 403:
        throw new PermissionError();
      case 404:
        throw new NotFoundError();
      case 400:
        throw new InvalidInputError();
      default:
        throw new KMSError(`Request failed with status ${response.status}`);
    }
  }

  /**
   * 🔑 Ensure a valid token is available
   */
  private async ensureToken(): Promise<void> {
    if (this.token && this.tokenExpiry && Date.now() < this.tokenExpiry) {
      return;
    }

    const response = await this.fetchWithRetry(
      `${this.config.baseURL}/api/v1/oauth/token`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          client_id: this.config.clientConfigId,
          client_secret: this.config.clientConfigSecret,
          grant_type: 'client_credentials',
        }),
      }
    );

    const data = await response.json();
    this.token = data.access_token;
    this.tokenExpiry = Date.now() + data.expires_in * 1000;
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

  /**
   * 💍 Create a new keyring
   */
  async createKeyring(name: string): Promise<void> {
    await this.fetchWithRetry(
      `${this.config.baseURL}/api/v1/tenant/keyrings`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({ name }),
      }
    );
  }

  /**
   * 📋 List all keyrings
   */
  async listKeyrings(): Promise<string[]> {
    const response = await this.fetchWithRetry(
      `${this.config.baseURL}/api/v1/tenant/keyrings`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );

    const data = await response.json();
    return data.keyrings;
  }

  /**
   * 📋 List all keys in a keyring
   */
  async listKeys(keyringName: string): Promise<string[]> {
    const response = await this.fetchWithRetry(
      `${this.config.baseURL}/api/v1/tenant/keyrings/${keyringName}/keys`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );

    const data = await response.json();
    return data.keys;
  }

  /**
   * 🔍 Get information about a key
   */
  async getKeyInfo(keyringName: string, keyName: string): Promise<KeyInfo> {
    const response = await this.fetchWithRetry(
      `${this.config.baseURL}/api/v1/tenant/keyrings/${keyringName}/keys/${keyName}`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );

    const data = await response.json();
    return {
      ...data,
      createdAt: new Date(data.created_at),
      nextRotation: new Date(data.next_rotation),
    };
  }

  /**
   * 🔑 Create a new key in the specified keyring
   */
  async createKey(keyringName: string, algorithm: string, purpose: string, rotationPeriod: number): Promise<void> {
    await this.fetchWithRetry(
      `${this.config.baseURL}/api/v1/tenant/keyrings/${keyringName}/keys`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({
          algorithm,
          purpose,
          rotation_period: rotationPeriod,
        }),
      }
    );
  }

  /**
   * 🔒 Encrypt data using a key in the specified keyring
   */
  async encrypt(keyringName: string, purpose: string, plaintext: Uint8Array): Promise<[Uint8Array, number]> {
    const response = await this.fetchWithRetry(
      `${this.config.baseURL}/api/v1/tenant/keyrings/${keyringName}/encrypt`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({
          plaintext: new TextDecoder().decode(plaintext),
          purpose,
        }),
      }
    );

    const data = await response.json();
    return [new TextEncoder().encode(data.ciphertext), data.keyVersion];
  }

  /**
   * 🔓 Decrypt data using a key in the specified keyring
   */
  async decrypt(keyringName: string, purpose: string, ciphertext: Uint8Array, keyVersion: number): Promise<Uint8Array> {
    const response = await this.fetchWithRetry(
      `${this.config.baseURL}/api/v1/tenant/keyrings/${keyringName}/decrypt`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({
          ciphertext: new TextDecoder().decode(ciphertext),
          keyVersion,
          purpose,
        }),
      }
    );

    const data = await response.json();
    return new TextEncoder().encode(data.plaintext);
  }

  // 🔐 Initialize client keys
  private async initializeKeys(privateKeyPEM: string) {
    // Import private key
    const privateKeyDER = this.pemToDER(privateKeyPEM);
    this.privateKey = await crypto.subtle.importKey(
      'pkcs8',
      privateKeyDER,
      {
        name: 'RSA-OAEP',
        hash: 'SHA-256',
      },
      true,
      ['decrypt']
    );

    // Extract public key
    if (this.privateKey) {
      const publicKey = await crypto.subtle.exportKey('spki', this.privateKey) ;
        this.publicKey = await crypto.subtle.importKey(
            'spki',
            publicKey,
            {
            name: 'RSA-OAEP',
            hash: 'SHA-256',
            },
            true,
            ['encrypt']
        );
    }
  }

  // 🔄 Convert PEM to DER format
  private pemToDER(pem: string): ArrayBuffer {
    const base64 = pem
      .replace('-----BEGIN PRIVATE KEY-----', '')
      .replace('-----END PRIVATE KEY-----', '')
      .replace(/\s/g, '');
    return Uint8Array.from(atob(base64), c => c.charCodeAt(0)).buffer;
  }

  // 🔒 Encrypt request body
  private async encryptRequest(data: any): Promise<ArrayBuffer> {
    if (!this.publicKey) {
      throw new Error('Client keys not initialized');
    }

    const jsonString = JSON.stringify(data);
    const encoder = new TextEncoder();
    const plaintext = encoder.encode(jsonString);

    return crypto.subtle.encrypt(
      {
        name: 'RSA-OAEP'
      },
      this.publicKey,
      plaintext
    );
  }

  // 🔓 Decrypt response body
  private async decryptResponse(encryptedData: ArrayBuffer): Promise<any> {
    if (!this.privateKey) {
      throw new Error('Client keys not initialized');
    }

    const decrypted = await crypto.subtle.decrypt(
      {
        name: 'RSA-OAEP'
      },
      this.privateKey,
      encryptedData
    );

    const decoder = new TextDecoder();
    const jsonString = decoder.decode(decrypted);
    return JSON.parse(jsonString);
  }

  // 🌐 Make an encrypted request
  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    if (!this.token) {
      await this.authenticate();
    }

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${this.token}`,
      ...(options.headers as Record<string, string> || {}),
    };

    if (options.body && this.publicKey) {
      const encryptedBody = await this.encryptRequest(JSON.parse(options.body as string));
      options.body = JSON.stringify({
        encrypted: Array.from(new Uint8Array(encryptedBody))
      });
      headers['X-Request-Encrypted'] = 'true';
    }

    const response = await fetch(`${this.config.baseURL}${path}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      throw new Error(`Request failed: ${response.statusText}`);
    }

    const contentType = response.headers.get('content-type');
    if (contentType?.includes('application/json')) {
      if (response.headers.get('X-Response-Encrypted') === 'true') {
        const encryptedData = await response.arrayBuffer();
        return await this.decryptResponse(encryptedData);
      }
      return await response.json();
    }

    return response.text() as T;
  }

  // 🔑 Authenticate with the server
  private async authenticate(): Promise<void> {
    const response = await fetch(`${this.config.baseURL}/auth/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        client_id: this.config.clientConfigId,
        client_secret: this.config.clientConfigSecret,
      }),
    });

    if (!response.ok) {
      throw new Error('Authentication failed');
    }

    const data = await response.json();
    this.token = data.access_token;
    this.tokenExpiry = Date.now() + data.expires_in * 1000;
  }
}

// 🔑 Key cache for storing parsed keys
class KeyCache {
  private static instance: KeyCache;
  private cache: Map<string, CryptoKey>;

  private constructor() {
    this.cache = new Map();
  }

  static getInstance(): KeyCache {
    if (!KeyCache.instance) {
      KeyCache.instance = new KeyCache();
    }
    return KeyCache.instance;
  }

  async get(key: string): Promise<CryptoKey | undefined> {
    return this.cache.get(key);
  }

  set(key: string, value: CryptoKey): void {
    this.cache.set(key, value);
  }
} 