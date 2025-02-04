/**
 * 🚨 Base error class for KMS errors
 */
export class KMSError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'KMSError';
  }
}

/**
 * 🚨 Authentication error
 */
export class AuthenticationError extends KMSError {
  constructor(message: string = 'Authentication failed') {
    super(message);
    this.name = 'AuthenticationError';
  }
}

/**
 * 🚨 Resource not found error
 */
export class NotFoundError extends KMSError {
  constructor(message: string = 'Resource not found') {
    super(message);
    this.name = 'NotFoundError';
  }
}

/**
 * 🚨 Permission denied error
 */
export class PermissionError extends KMSError {
  constructor(message: string = 'Permission denied') {
    super(message);
    this.name = 'PermissionError';
  }
}

/**
 * 🚨 Invalid input error
 */
export class InvalidInputError extends KMSError {
  constructor(message: string = 'Invalid input provided') {
    super(message);
    this.name = 'InvalidInputError';
  }
}

/**
 * 🔄 Configuration for retry behavior
 */
export interface RetryConfig {
  maxRetries: number;
  initialWait: number;
  maxWait: number;
}

/**
 * 🔧 Configuration for the Xyphos client
 */
export interface ClientConfig {
  baseUrl: string;
  clientId: string;
  clientSecret: string;
  timeout?: number;
  retryConfig?: RetryConfig;
}

/**
 * 📦 Project resource
 */
export interface Project {
  id: string;
  name: string;
  description: string;
  createdAt: Date;
}

/**
 * 📍 Location resource
 */
export interface Location {
  id: string;
  name: string;
  createdAt: Date;
}

/**
 * 💍 KeyRing resource
 */
export interface KeyRing {
  id: string;
  name: string;
  createdAt: Date;
}

/**
 * 🔑 CryptoKey resource
 */
export interface CryptoKey {
  id: string;
  name: string;
  algorithm: string;
  purpose: string;
  rotationPeriod: number;
  createdAt: Date;
  nextRotation: Date;
  version: number;
}

/**
 * 🔐 Client for interacting with Xyphos
 */
export class LambdaKMSClient {
  private config: Required<ClientConfig>;
  private token: string | null = null;
  private tokenExpiry: number | null = null;

  constructor(config: ClientConfig) {
    this.config = {
      timeout: 30000,
      retryConfig: {
        maxRetries: 3,
        initialWait: 100,
        maxWait: 2000,
      },
      ...config,
    };
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

      if (
        attempt < this.config.retryConfig.maxRetries &&
        (response.status >= 500 || response.status === 429)
      ) {
        const wait = Math.min(
          this.config.retryConfig.initialWait * Math.pow(2, attempt),
          this.config.retryConfig.maxWait
        );
        await new Promise(resolve => setTimeout(resolve, wait));
        return this.fetchWithRetry(url, options, attempt + 1);
      }

      if (!response.ok) {
        await this.handleErrorResponse(response);
      }

      return response;
    } catch (error) {
      if (
        attempt < this.config.retryConfig.maxRetries &&
        error instanceof TypeError
      ) {
        const wait = Math.min(
          this.config.retryConfig.initialWait * Math.pow(2, attempt),
          this.config.retryConfig.maxWait
        );
        await new Promise(resolve => setTimeout(resolve, wait));
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
      `${this.config.baseUrl}/api/oauth/token`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          client_id: this.config.clientId,
          client_secret: this.config.clientSecret,
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

  // 📦 Project operations

  /**
   * Create a new project
   */
  async createProject(name: string, description: string): Promise<Project> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({ name, description }),
      }
    );
    return response.json();
  }

  /**
   * List all projects
   */
  async listProjects(): Promise<Project[]> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  // 📍 Location operations

  /**
   * Create a new location
   */
  async createLocation(projectId: string, name: string): Promise<Location> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({ name }),
      }
    );
    return response.json();
  }

  /**
   * List all locations in a project
   */
  async listLocations(projectId: string): Promise<Location[]> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  // 💍 KeyRing operations

  /**
   * Create a new keyring
   */
  async createKeyRing(projectId: string, locationId: string, name: string): Promise<KeyRing> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations/${locationId}/keyrings`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({ name }),
      }
    );
    return response.json();
  }

  /**
   * List all keyrings in a location
   */
  async listKeyRings(projectId: string, locationId: string): Promise<KeyRing[]> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations/${locationId}/keyrings`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  // 🔑 CryptoKey operations

  /**
   * Create a new crypto key
   */
  async createCryptoKey(
    projectId: string,
    locationId: string,
    keyringId: string,
    algorithm: string,
    purpose: string,
    rotationPeriod: number
  ): Promise<CryptoKey> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys`,
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
    return response.json();
  }

  /**
   * List all crypto keys in a keyring
   */
  async listCryptoKeys(
    projectId: string,
    locationId: string,
    keyringId: string
  ): Promise<CryptoKey[]> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys`,
      {
        method: 'GET',
        headers: await this.getHeaders(),
      }
    );
    return response.json();
  }

  /**
   * Rotate a crypto key
   */
  async rotateCryptoKey(
    projectId: string,
    locationId: string,
    keyringId: string,
    keyId: string
  ): Promise<void> {
    await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys/${keyId}:rotate`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
      }
    );
  }

  /**
   * 🔒 Encrypt data using a crypto key
   */
  async encrypt(
    projectId: string,
    locationId: string,
    keyringId: string,
    keyId: string,
    plaintext: Uint8Array
  ): Promise<Uint8Array> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys/${keyId}:encrypt`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({
          plaintext: Buffer.from(plaintext).toString('base64'),
        }),
      }
    );
    const data = await response.json();
    return Buffer.from(data.ciphertext, 'base64');
  }

  /**
   * 🔓 Decrypt data using a crypto key
   */
  async decrypt(
    projectId: string,
    locationId: string,
    keyringId: string,
    keyId: string,
    ciphertext: Uint8Array
  ): Promise<Uint8Array> {
    const response = await this.fetchWithRetry(
      `${this.config.baseUrl}/api/projects/${projectId}/locations/${locationId}/keyrings/${keyringId}/keys/${keyId}:decrypt`,
      {
        method: 'POST',
        headers: await this.getHeaders(),
        body: JSON.stringify({
          ciphertext: Buffer.from(ciphertext).toString('base64'),
        }),
      }
    );
    const data = await response.json();
    return Buffer.from(data.plaintext, 'base64');
  }
} 