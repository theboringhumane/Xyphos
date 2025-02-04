// 🔐 Xyphos API Client

import axios, { AxiosInstance, AxiosResponse, AxiosError } from 'axios';

// 📝 API Types
export interface Project {
  id: string;
  name: string;
  description: string;
  createdAt: string;
}

export interface Location {
  id: string;
  name: string;
  createdAt: string;
}

export interface KeyRing {
  id: string;
  name: string;
  createdAt: string;
}

export interface CryptoKey {
  id: string;
  name: string;
  algorithm: string;
  purpose: string;
  rotationPeriod: number;
  createdAt: string;
  nextRotation: string;
  version: number;
}

// 🔧 API Client Configuration
interface ApiConfig {
  baseURL: string;
  token?: string;
}

// 🌐 Xyphos API Client
export class XyphosClient {
  private client: AxiosInstance;

  constructor(config: ApiConfig) {
    this.client = axios.create({
      baseURL: config.baseURL,
      headers: {
        'Content-Type': 'application/json',
        ...(config.token && { Authorization: `Bearer ${config.token}` }),
      },
    });

    // Add response interceptor for error handling
    this.client.interceptors.response.use(
      (response: AxiosResponse) => response,
      (error: AxiosError) => {
        if (error.response?.status === 401) {
          // Handle unauthorized access
          window.location.href = '/auth/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // 🔑 Authentication
  async login(githubCode: string) {
    const { data } = await this.client.post('/auth/github/callback', { code: githubCode });
    return data;
  }

  // 📦 Project Operations
  async listProjects(): Promise<Project[]> {
    const { data } = await this.client.get('/projects');
    return data;
  }

  async createProject(name: string, description: string): Promise<Project> {
    const { data } = await this.client.post('/projects', { name, description });
    return data;
  }

  async getProject(id: string): Promise<Project> {
    const { data } = await this.client.get(`/projects/${id}`);
    return data;
  }

  // 📍 Location Operations
  async listLocations(projectId: string): Promise<Location[]> {
    const { data } = await this.client.get(`/projects/${projectId}/locations`);
    return data;
  }

  async getLocation(projectId: string, locationId: string): Promise<Location> {
    const { data } = await this.client.get(`/projects/${projectId}/locations/${locationId}`);
    return data;
  }

  // 💍 KeyRing Operations
  async listKeyRings(projectId: string, locationId: string): Promise<KeyRing[]> {
    const { data } = await this.client.get(`/projects/${projectId}/locations/${locationId}/keyrings`);
    return data;
  }

  async createKeyRing(projectId: string, locationId: string, name: string): Promise<KeyRing> {
    const { data } = await this.client.post(`/projects/${projectId}/locations/${locationId}/keyrings`, { name });
    return data;
  }

  async getKeyRing(projectId: string, locationId: string, keyRingId: string): Promise<KeyRing> {
    const { data } = await this.client.get(`/projects/${projectId}/locations/${locationId}/keyrings/${keyRingId}`);
    return data;
  }

  // 🔐 CryptoKey Operations
  async listCryptoKeys(projectId: string, locationId: string, keyRingId: string): Promise<CryptoKey[]> {
    const { data } = await this.client.get(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyRingId}/keys`
    );
    return data;
  }

  async createCryptoKey(
    projectId: string,
    locationId: string,
    keyRingId: string,
    params: {
      name: string;
      algorithm: string;
      purpose: string;
      rotationPeriod: number;
    }
  ): Promise<CryptoKey> {
    const { data } = await this.client.post(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyRingId}/keys`,
      params
    );
    return data;
  }

  async getCryptoKey(
    projectId: string,
    locationId: string,
    keyRingId: string,
    keyId: string
  ): Promise<CryptoKey> {
    const { data } = await this.client.get(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyRingId}/keys/${keyId}`
    );
    return data;
  }

  // 🔄 Key Rotation
  async rotateCryptoKey(
    projectId: string,
    locationId: string,
    keyRingId: string,
    keyId: string
  ): Promise<CryptoKey> {
    const { data } = await this.client.post(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyRingId}/keys/${keyId}/rotate`
    );
    return data;
  }

  // 🔒 Encryption
  async encrypt(
    projectId: string,
    locationId: string,
    keyRingId: string,
    keyId: string,
    plaintext: string
  ): Promise<{ ciphertext: string }> {
    const { data } = await this.client.post(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyRingId}/keys/${keyId}/encrypt`,
      { plaintext }
    );
    return data;
  }

  // 🔓 Decryption
  async decrypt(
    projectId: string,
    locationId: string,
    keyRingId: string,
    keyId: string,
    ciphertext: string
  ): Promise<{ plaintext: string }> {
    const { data } = await this.client.post(
      `/projects/${projectId}/locations/${locationId}/keyrings/${keyRingId}/keys/${keyId}/decrypt`,
      { ciphertext }
    );
    return data;
  }
} 