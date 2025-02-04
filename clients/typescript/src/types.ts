/**
 * 🔄 Configuration for retry behavior
 */
export interface RetryConfig {
  maxRetries: number;
  backoffFactor: number;
  statusForcelist: number[];
}

/**
 * 🔧 Configuration for the Xyphos client
 */
export interface ClientConfig {
  baseURL: string;
  clientConfigId: string;  // 🔑 Client config ID from the KMS service
  clientConfigSecret: string;  // 🔐 Client config secret from the KMS service
  timeout?: number;
  retryConfig?: Partial<RetryConfig>;
}

/**
 * 📊 Information about a key
 */
export interface KeyInfo {
  id: string;
  name: string;
  algorithm: string;
  purpose: string;
  rotationPeriod: number;
  version: number;
  nextRotation: string;
  createdAt: string;
}

export interface ClientConfigResponse {
  id: string;
  secret: string;
  expiresAt: string;
  permissions: string[];
} 