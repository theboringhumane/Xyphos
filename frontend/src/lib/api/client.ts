'use client'

// 🔐 API client for Xyphos
export class APIError extends Error {
  constructor(
    public status: number,
    public statusText: string,
    public data: any
  ) {
    super(`API Error: ${status} ${statusText}`)
    this.name = 'APIError'
  }
}

// 🌐 API client configuration
const API_URL = process.env.NEXT_PUBLIC_API_URL

// 🔑 API client class
export class APIClient {
  private token?: string
  private apiKey?: string

  constructor(token?: string, apiKey?: string) {
    this.token = token
    this.apiKey = apiKey
  }

  // 🔒 Set authentication token
  setToken(token: string) {
    this.token = token
  }

  // 🔑 Set API key
  setApiKey(apiKey: string) {
    this.apiKey = apiKey
  }

  // 📝 Generic request method
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    if (this.apiKey) {
      headers['X-API-Key'] = this.apiKey
    }

    const response = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      headers: {
        ...headers,
        ...options.headers,
      },
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({}))
      throw new Error(error.error || 'An error occurred')
    }

    return response.json()
  }

  // 🔑 Keyring operations
  async listKeyrings() {
    return this.request<{ keyrings: Keyring[] }>('/api/keyrings')
  }

  async createKeyring(data: CreateKeyringRequest) {
    return this.request<Keyring>('/api/keyrings', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async getKeyring(id: string) {
    return this.request<Keyring>(`/api/keyrings/${id}`)
  }

  async deleteKeyring(id: string) {
    return this.request(`/api/keyrings/${id}`, {
      method: 'DELETE',
    })
  }

  // 🔐 Key operations
  async listKeys(keyringId: string) {
    return this.request<{ keys: Key[] }>(`/api/keyrings/${keyringId}/keys`)
  }

  async createKey(keyringId: string, data: CreateKeyRequest) {
    return this.request<Key>(`/api/keyrings/${keyringId}/keys`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async rotateKey(keyringId: string, keyId: string) {
    return this.request<Key>(`/api/keyrings/${keyringId}/keys/${keyId}/rotate`, {
      method: 'POST',
    })
  }

  // 🔒 Cryptographic operations
  async encrypt(keyringId: string, data: EncryptRequest) {
    return this.request<EncryptResponse>(`/api/keyrings/${keyringId}/encrypt`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async decrypt(keyringId: string, data: DecryptRequest) {
    return this.request<DecryptResponse>(`/api/keyrings/${keyringId}/decrypt`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  // 👤 User methods
  async getCurrentUser() {
    return this.request<User>('/auth/github/user')
  }

  // 🔑 Client configuration methods
  async createClientConfig(data: CreateClientConfigRequest) {
    return this.request<ClientConfig>('/api/clients', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async listClientConfigs() {
    return this.request<{ configs: ClientConfig[] }>('/api/clients')
  }

  async getClientConfig() {
    return this.request<ClientConfig>('/api/clients/current')
  }
}

// 📝 Types
export interface User {
  id: string
  githubId: number
  email: string
  name: string
  login: string
  avatarUrl: string
  role: string
  createdAt: string
  updatedAt: string
}

export interface Keyring {
  id: string
  name: string
  description?: string
  createdAt: string
  updatedAt: string
}

export interface Key {
  id: string
  version: string
  algorithm: string
  purpose: string
  status: 'active' | 'inactive' | 'scheduled_for_deletion'
  createdAt: string
  expiresAt?: string
}

export interface ClientConfig {
  id: string
  userId: string
  name: string
  apiKey: string
  secretKey: string
  permissions: string[]
  createdAt: string
  expiresAt: string
}

export interface CreateKeyringRequest {
  name: string
  description?: string
}

export interface CreateKeyRequest {
  algorithm: string
  purpose: string
}

export interface CreateClientConfigRequest {
  name: string
  permissions: string[]
  expiresIn: string
}

export interface EncryptRequest {
  plaintext: string
  purpose: string
}

export interface EncryptResponse {
  ciphertext: string
  keyVersion: string
}

export interface DecryptRequest {
  ciphertext: string
  keyVersion: string
  purpose: string
}

export interface DecryptResponse {
  plaintext: string
}

// 🏭 Create API client instance
export const apiClient = new APIClient()