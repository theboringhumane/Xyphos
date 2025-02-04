// 🔐 CryptoKey Management Hook

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/contexts/api-context';
import type { CryptoKey } from '@/lib/api';

interface CryptoKeyParams {
  projectId: string;
  locationId: string;
  keyRingId: string;
}

interface CreateCryptoKeyParams {
  name: string;
  algorithm: string;
  purpose: string;
  rotationPeriod: number;
}

// 🎣 Hook for managing crypto keys
export function useCryptoKeys(params: CryptoKeyParams) {
  const { client, isLoading: isClientLoading } = useApi();
  const queryClient = useQueryClient();
  const { projectId, locationId, keyRingId } = params;

  // 📋 List crypto keys
  const {
    data: cryptoKeys,
    isLoading: isLoadingCryptoKeys,
    error: cryptoKeysError,
  } = useQuery({
    queryKey: ['cryptokeys', projectId, locationId, keyRingId],
    queryFn: () => client?.listCryptoKeys(projectId, locationId, keyRingId) ?? Promise.resolve([]),
    enabled: !!client && !isClientLoading && !!projectId && !!locationId && !!keyRingId,
  });

  // ➕ Create crypto key
  const createCryptoKeyMutation = useMutation({
    mutationFn: (params: CreateCryptoKeyParams) =>
      client?.createCryptoKey(projectId, locationId, keyRingId, params) ??
      Promise.reject('No client available'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cryptokeys', projectId, locationId, keyRingId] });
    },
  });

  // 🔄 Rotate crypto key
  const rotateCryptoKeyMutation = useMutation({
    mutationFn: (keyId: string) =>
      client?.rotateCryptoKey(projectId, locationId, keyRingId, keyId) ??
      Promise.reject('No client available'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cryptokeys', projectId, locationId, keyRingId] });
    },
  });

  // 🔒 Encrypt data
  const encryptMutation = useMutation({
    mutationFn: ({
      keyId,
      plaintext,
    }: {
      keyId: string;
      plaintext: string;
    }) =>
      client?.encrypt(projectId, locationId, keyRingId, keyId, plaintext) ??
      Promise.reject('No client available'),
  });

  // 🔓 Decrypt data
  const decryptMutation = useMutation({
    mutationFn: ({
      keyId,
      ciphertext,
    }: {
      keyId: string;
      ciphertext: string;
    }) =>
      client?.decrypt(projectId, locationId, keyRingId, keyId, ciphertext) ??
      Promise.reject('No client available'),
  });

  // 🔍 Get single crypto key
  const useCryptoKey = (keyId: string) =>
    useQuery({
      queryKey: ['cryptokeys', projectId, locationId, keyRingId, keyId],
      queryFn: () =>
        client?.getCryptoKey(projectId, locationId, keyRingId, keyId) ??
        Promise.reject('No client available'),
      enabled: !!client && !isClientLoading && !!projectId && !!locationId && !!keyRingId && !!keyId,
    });

  return {
    // 📊 Data
    cryptoKeys,
    isLoadingCryptoKeys,
    cryptoKeysError,

    // 🛠️ Operations
    createCryptoKey: createCryptoKeyMutation.mutate,
    isCreatingCryptoKey: createCryptoKeyMutation.isPending,
    createCryptoKeyError: createCryptoKeyMutation.error,

    rotateCryptoKey: rotateCryptoKeyMutation.mutate,
    isRotatingCryptoKey: rotateCryptoKeyMutation.isPending,
    rotateCryptoKeyError: rotateCryptoKeyMutation.error,

    encrypt: encryptMutation.mutate,
    isEncrypting: encryptMutation.isPending,
    encryptError: encryptMutation.error,

    decrypt: decryptMutation.mutate,
    isDecrypting: decryptMutation.isPending,
    decryptError: decryptMutation.error,

    // 🎣 Sub-hooks
    useCryptoKey,
  };
} 