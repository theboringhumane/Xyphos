'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  apiClient,
  type Keyring,
  type Key,
  type CreateKeyringRequest,
  type CreateKeyRequest,
  type CreateClientConfigRequest,
  type User,
  type ClientConfig,
} from './client'

// 🔑 Keyring hooks
export function useKeyring(id: string) {
  return useQuery({
    queryKey: ['keyring', id],
    queryFn: () => apiClient.getKeyring(id),
  })
}

export function useKeyrings() {
  return useQuery({
    queryKey: ['keyrings'],
    queryFn: () => apiClient.listKeyrings(),
  })
}

export function useCreateKeyring() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateKeyringRequest) => apiClient.createKeyring(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['keyrings'] })
    },
  })
}

export function useDeleteKeyring() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => apiClient.deleteKeyring(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['keyrings'] })
    },
  })
}

// 🔐 Key hooks
export function useKeys(keyringId: string) {
  return useQuery({
    queryKey: ['keys', keyringId],
    queryFn: () => apiClient.listKeys(keyringId),
  })
}

export function useCreateKey() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ keyringId, ...data }: CreateKeyRequest & { keyringId: string }) =>
      apiClient.createKey(keyringId, data),
    onSuccess: (_, { keyringId }) => {
      queryClient.invalidateQueries({ queryKey: ['keys', keyringId] })
    },
  })
}

export function useRotateKey() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ keyringId, keyId }: { keyringId: string; keyId: string }) =>
      apiClient.rotateKey(keyringId, keyId),
    onSuccess: (_, { keyringId }) => {
      queryClient.invalidateQueries({ queryKey: ['keys', keyringId] })
    },
  })
}

// 👤 User hooks
export function useCurrentUser() {
  return useQuery<User>({
    queryKey: ['user'],
    queryFn: () => apiClient.getCurrentUser(),
  })
}

// 🔑 Client configuration hooks
export function useClientConfigs() {
  return useQuery<{ configs: ClientConfig[] }>({
    queryKey: ['client-configs'],
    queryFn: () => apiClient.listClientConfigs(),
  })
}

export function useCreateClientConfig() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateClientConfigRequest) => apiClient.createClientConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['client-configs'] })
    },
  })
}

export function useCurrentClientConfig() {
  return useQuery<ClientConfig>({
    queryKey: ['client-config', 'current'],
    queryFn: () => apiClient.getClientConfig(),
  })
}

// 🔒 Authentication hooks
export function useSetAuthToken() {
  return (token: string) => {
    apiClient.setToken(token)
  }
}

export function useSetApiKey() {
  return (apiKey: string) => {
    apiClient.setApiKey(apiKey)
  }
}