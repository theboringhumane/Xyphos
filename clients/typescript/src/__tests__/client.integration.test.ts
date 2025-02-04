import { LambdaKMSClient } from '../client';
import {
  Project,
  Location,
  KeyRing,
  KMSKey,
  ClientConfig,
  ClientConfigInfo,
  CreateProjectRequest,
  CreateCryptoKeyRequest,
  CreateClientConfigRequest,
  GenerateKeyPairRequest,
  KeyPair,
} from '../types';

// 🧪 Jest configuration
import { describe, it, expect, beforeAll } from '@jest/globals';

// 🔧 Test configuration
const TEST_CONFIG: ClientConfig = {
  baseURL: process.env.XYPHOS_API_URL || 'http://localhost:8080',
  clientId: process.env.XYPHOS_CLIENT_ID || 'test-client',
  clientSecret: process.env.XYPHOS_CLIENT_SECRET || 'test-secret',
  timeout: 5000,
  retryConfig: {
    maxRetries: 3,
    backoffFactor: 0.1,
    statusForcelist: [500, 502, 503, 504, 429],
    initialWait: 100,
    maxWait: 1000,
  },
};

describe('🔐 LambdaKMSClient Integration Tests', () => {
  let client: LambdaKMSClient;
  let testProject: Project;
  let testLocation: Location;
  let testKeyRing: KeyRing;
  let testKey: KMSKey;
  let testClientConfig: ClientConfigInfo;
  let testKeyPair: KeyPair;

  // 🏗️ Setup
  beforeAll(async () => {
    client = new LambdaKMSClient(TEST_CONFIG);
  });

  // 📦 Project Tests
  describe('Project Operations', () => {
    it('should create a project', async () => {
      const request: CreateProjectRequest = {
        name: `test-project-${Date.now()}`,
        description: 'Test project for integration tests',
      };

      testProject = await client.createProject(request);

      expect(testProject).toBeDefined();
      expect(testProject.id).toBeDefined();
      expect(testProject.name).toBe(request.name);
      expect(testProject.description).toBe(request.description);
    });

    it('should list projects', async () => {
      const projects = await client.listProjects();

      expect(projects).toBeDefined();
      expect(Array.isArray(projects)).toBe(true);
      expect(projects.length).toBeGreaterThan(0);
      expect(projects.find(p => p.id === testProject.id)).toBeDefined();
    });

    it('should get a project', async () => {
      const project = await client.getProject(testProject.id);

      expect(project).toBeDefined();
      expect(project.id).toBe(testProject.id);
      expect(project.name).toBe(testProject.name);
      expect(project.description).toBe(testProject.description);
    });
  });

  // 📍 Location Tests
  describe('Location Operations', () => {
    it('should list locations', async () => {
      const locations = await client.listLocations(testProject.id);

      expect(locations).toBeDefined();
      expect(Array.isArray(locations)).toBe(true);
      expect(locations.length).toBeGreaterThan(0);

      // Save first location for further tests
      [testLocation] = locations;
    });

    it('should get a location', async () => {
      const location = await client.getLocation(testProject.id, testLocation.id);

      expect(location).toBeDefined();
      expect(location.id).toBe(testLocation.id);
      expect(location.name).toBe(testLocation.name);
    });
  });

  // 💍 KeyRing Tests
  describe('KeyRing Operations', () => {
    it('should create a keyring', async () => {
      const keyringName = `test-keyring-${Date.now()}`;

      testKeyRing = await client.createKeyRing(
        testProject.id,
        testLocation.id,
        keyringName
      );

      expect(testKeyRing).toBeDefined();
      expect(testKeyRing.id).toBeDefined();
      expect(testKeyRing.name).toBe(keyringName);
    });

    it('should list keyrings', async () => {
      const keyrings = await client.listKeyRings(testProject.id, testLocation.id);

      expect(keyrings).toBeDefined();
      expect(Array.isArray(keyrings)).toBe(true);
      expect(keyrings.length).toBeGreaterThan(0);
      expect(keyrings.find(k => k.id === testKeyRing.id)).toBeDefined();
    });

    it('should get a keyring', async () => {
      const keyring = await client.getKeyRing(
        testProject.id,
        testLocation.id,
        testKeyRing.id
      );

      expect(keyring).toBeDefined();
      expect(keyring.id).toBe(testKeyRing.id);
      expect(keyring.name).toBe(testKeyRing.name);
    });
  });

  // 🔑 CryptoKey Tests
  describe('CryptoKey Operations', () => {
    it('should create a crypto key', async () => {
      const request: CreateCryptoKeyRequest = {
        name: `test-key-${Date.now()}`,
        algorithm: 'AES256-GCM',
        purpose: 'ENCRYPT_DECRYPT',
        rotationPeriod: 86400, // 24 hours
      };

      testKey = await client.createCryptoKey(
        testProject.id,
        testLocation.id,
        testKeyRing.id,
        request
      );

      expect(testKey).toBeDefined();
      expect(testKey.id).toBeDefined();
      expect(testKey.name).toBe(request.name);
      expect(testKey.algorithm).toBe(request.algorithm);
      expect(testKey.purpose).toBe(request.purpose);
      expect(testKey.rotationPeriod).toBe(request.rotationPeriod);
    });

    it('should list crypto keys', async () => {
      const keys = await client.listCryptoKeys(
        testProject.id,
        testLocation.id,
        testKeyRing.id
      );

      expect(keys).toBeDefined();
      expect(Array.isArray(keys)).toBe(true);
      expect(keys.length).toBeGreaterThan(0);
      expect(keys.find(k => k.id === testKey.id)).toBeDefined();
    });

    it('should get a crypto key', async () => {
      const key = await client.getCryptoKey(
        testProject.id,
        testLocation.id,
        testKeyRing.id,
        testKey.id
      );

      expect(key).toBeDefined();
      expect(key.id).toBe(testKey.id);
      expect(key.name).toBe(testKey.name);
      expect(key.algorithm).toBe(testKey.algorithm);
      expect(key.purpose).toBe(testKey.purpose);
    });

    it('should rotate a crypto key', async () => {
      const rotatedKey = await client.rotateCryptoKey(
        testProject.id,
        testLocation.id,
        testKeyRing.id,
        testKey.id
      );

      expect(rotatedKey).toBeDefined();
      expect(rotatedKey.id).toBe(testKey.id);
      expect(rotatedKey.version).toBeGreaterThan(testKey.version);
    });
  });

  // 🔐 Cryptographic Operations
  describe('Cryptographic Operations', () => {
    it('should encrypt and decrypt data', async () => {
      const plaintext = 'Hello, World! 🌍';

      // Encrypt
      const encryptResult = await client.encrypt(
        testProject.id,
        testLocation.id,
        testKeyRing.id,
        testKey.id,
        plaintext
      );

      expect(encryptResult).toBeDefined();
      expect(encryptResult.ciphertext).toBeDefined();
      expect(encryptResult.keyVersion).toBeDefined();

      // Decrypt
      const decryptResult = await client.decrypt(
        testProject.id,
        testLocation.id,
        testKeyRing.id,
        testKey.id,
        encryptResult.ciphertext,
        encryptResult.keyVersion
      );

      expect(decryptResult).toBeDefined();
      expect(decryptResult.plaintext).toBe(plaintext);
    });

    it('should encrypt and decrypt binary data', async () => {
      const plaintext = new Uint8Array([1, 2, 3, 4, 5]);

      // Encrypt
      const encryptResult = await client.encrypt(
        testProject.id,
        testLocation.id,
        testKeyRing.id,
        testKey.id,
        plaintext
      );

      expect(encryptResult).toBeDefined();
      expect(encryptResult.ciphertext).toBeDefined();
      expect(encryptResult.keyVersion).toBeDefined();

      // Decrypt
      const decryptResult = await client.decrypt(
        testProject.id,
        testLocation.id,
        testKeyRing.id,
        testKey.id,
        encryptResult.ciphertext,
        encryptResult.keyVersion
      );

      expect(decryptResult).toBeDefined();
      const decryptedBytes = Buffer.from(decryptResult.plaintext, 'base64');
      expect(Array.from(decryptedBytes)).toEqual(Array.from(plaintext));
    });
  });

  // 🔑 Client Configuration Tests
  describe('Client Configuration Operations', () => {
    it('should create a client configuration', async () => {
      const request: CreateClientConfigRequest = {
        name: `test-client-${Date.now()}`,
        permissions: ['encrypt', 'decrypt'],
        expiresIn: '24h',
      };

      testClientConfig = await client.createClientConfig(request);

      expect(testClientConfig).toBeDefined();
      expect(testClientConfig.id).toBeDefined();
      expect(testClientConfig.name).toBe(request.name);
      expect(testClientConfig.permissions).toEqual(request.permissions);
      expect(testClientConfig.status).toBe('ACTIVE');
    });

    it('should list client configurations', async () => {
      const configs = await client.listClientConfigs();

      expect(configs).toBeDefined();
      expect(Array.isArray(configs)).toBe(true);
      expect(configs.length).toBeGreaterThan(0);
      expect(configs.find(c => c.id === testClientConfig.id)).toBeDefined();
    });

    it('should get a client configuration', async () => {
      const config = await client.getClientConfig(testClientConfig.id);

      expect(config).toBeDefined();
      expect(config.id).toBe(testClientConfig.id);
      expect(config.name).toBe(testClientConfig.name);
      expect(config.permissions).toEqual(testClientConfig.permissions);
    });

    it('should revoke a client configuration', async () => {
      const config = await client.revokeClientConfig(testClientConfig.id);

      expect(config).toBeDefined();
      expect(config.id).toBe(testClientConfig.id);
      expect(config.status).toBe('REVOKED');
    });
  });

  // 🧪 Error Handling Tests
  describe('Error Handling', () => {
    it('should handle 404 errors', async () => {
      await expect(
        client.getProject('non-existent-id')
      ).rejects.toThrow('Resource not found');
    });

    it('should handle invalid input errors', async () => {
      await expect(
        client.createProject({ name: '', description: '' })
      ).rejects.toThrow('Invalid input');
    });

    it('should handle authentication errors', async () => {
      const invalidClient = new LambdaKMSClient({
        ...TEST_CONFIG,
        clientSecret: 'invalid-secret',
      });

      await expect(
        invalidClient.listProjects()
      ).rejects.toThrow('Authentication failed');
    });
  });

  // 🔑 Key Pair Tests
  describe('Key Pair Operations', () => {
    it('should generate a new key pair', async () => {
      const request: GenerateKeyPairRequest = {
        algorithm: 'RSA-OAEP-4096',
        expiresIn: '24h',
      };

      testKeyPair = await client.generateKeyPair(request);

      expect(testKeyPair).toBeDefined();
      expect(testKeyPair.publicKey).toBeDefined();
      expect(testKeyPair.privateKey).toBeDefined();
      expect(testKeyPair.algorithm).toBe(request.algorithm);
      expect(testKeyPair.expiresAt).toBeDefined();

      // Verify the key pair is stored in the client
      expect(client.getKeyPair()).toEqual(testKeyPair);
    });

    it('should use key pair for end-to-end encryption', async () => {
      const request: CreateProjectRequest = {
        name: `test-project-e2e-${Date.now()}`,
        description: 'Test project with end-to-end encryption',
      };

      const project = await client.createProject(request);

      expect(project).toBeDefined();
      expect(project.id).toBeDefined();
      expect(project.name).toBe(request.name);
      expect(project.description).toBe(request.description);
    });

    it('should allow setting a new key pair', async () => {
      // Clear the key pair
      client.setKeyPair(null);
      expect(client.getKeyPair()).toBeNull();

      // Set the key pair back
      client.setKeyPair(testKeyPair);
      expect(client.getKeyPair()).toEqual(testKeyPair);
    });

    it('should handle encryption errors gracefully', async () => {
      // Set an invalid key pair
      client.setKeyPair({
        ...testKeyPair,
        privateKey: 'invalid-key',
      });

      await expect(
        client.listProjects()
      ).rejects.toThrow('Failed to decrypt response');

      // Reset the valid key pair
      client.setKeyPair(testKeyPair);
    });
  });
}); 