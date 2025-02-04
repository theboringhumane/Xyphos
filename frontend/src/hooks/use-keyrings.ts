// 💍 KeyRing Management Hook

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/contexts/api-context';
import type { KeyRing } from '@/lib/api';

interface KeyRingParams {
  projectId: string;
  locationId: string;
}

// 🎣 Hook for managing key rings
export function useKeyRings(params: KeyRingParams) {
  const { client, isLoading: isClientLoading } = useApi();
  const queryClient = useQueryClient();
  const { projectId, locationId } = params;

  // 📋 List key rings
  const {
    data: keyRings,
    isLoading: isLoadingKeyRings,
    error: keyRingsError,
  } = useQuery({
    queryKey: ['keyrings', projectId, locationId],
    queryFn: () => client?.listKeyRings(projectId, locationId) ?? Promise.resolve([]),
    enabled: !!client && !isClientLoading && !!projectId && !!locationId,
  });

  // ➕ Create key ring
  const createKeyRingMutation = useMutation({
    mutationFn: (name: string) =>
      client?.createKeyRing(projectId, locationId, name) ?? Promise.reject('No client available'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['keyrings', projectId, locationId] });
    },
  });

  // 🔍 Get single key ring
  const useKeyRing = (keyRingId: string) =>
    useQuery({
      queryKey: ['keyrings', projectId, locationId, keyRingId],
      queryFn: () =>
        client?.getKeyRing(projectId, locationId, keyRingId) ?? Promise.reject('No client available'),
      enabled: !!client && !isClientLoading && !!projectId && !!locationId && !!keyRingId,
    });

  return {
    // 📊 Data
    keyRings,
    isLoadingKeyRings,
    keyRingsError,

    // 🛠️ Operations
    createKeyRing: createKeyRingMutation.mutate,
    isCreatingKeyRing: createKeyRingMutation.isPending,
    createKeyRingError: createKeyRingMutation.error,

    // 🎣 Sub-hooks
    useKeyRing,
  };
} 