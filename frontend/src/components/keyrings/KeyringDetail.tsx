'use client'

import { useKeyring, useKeys } from '@/lib/api/hooks'
import { KeyIcon, PlusIcon } from '@heroicons/react/24/outline'
import Link from 'next/link'
import { Button } from '@/components/ui/Button'

interface KeyringDetailProps {
  id: string
}

export function KeyringDetail({ id }: KeyringDetailProps) {
  const { data: keyring, isLoading: isLoadingKeyring, error: keyringError } = useKeyring(id)
  const { data: keysData, isLoading: isLoadingKeys, error: keysError } = useKeys(id)

  if (keyringError || keysError) {
    const error = keyringError || keysError
    return (
      <div className="text-center py-12">
        <div className="rounded-full bg-error-50 p-3 w-12 h-12 mx-auto mb-4">
          <KeyIcon className="h-6 w-6 text-error-600" />
        </div>
        <h3 className="text-sm font-semibold text-gray-900">Error loading keyring</h3>
        <p className="mt-1 text-sm text-gray-500">
          {error instanceof Error ? error.message : 'An unexpected error occurred'}
        </p>
        <Button
          variant="outline"
          size="sm"
          className="mt-4"
          onClick={() => window.location.reload()}
        >
          Try Again
        </Button>
      </div>
    )
  }

  if (isLoadingKeyring || isLoadingKeys) {
    return (
      <div className="animate-pulse space-y-8">
        <div className="h-20 bg-white rounded-lg" />
        <div className="h-64 bg-white rounded-lg" />
      </div>
    )
  }

  if (!keyring) {
    return (
      <div className="text-center py-12">
        <KeyIcon className="mx-auto h-12 w-12 text-gray-400" />
        <h3 className="mt-2 text-sm font-semibold text-gray-900">Keyring not found</h3>
        <p className="mt-1 text-sm text-gray-500">
          This keyring doesn't exist or you don't have access to it.
        </p>
        <Link href="/keyrings">
          <Button variant="outline" className="mt-4">
            Back to Keyrings
          </Button>
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{keyring.name}</h1>
          <p className="mt-2 text-gray-500">
            Created on {new Date(keyring.createdAt).toLocaleDateString()}
          </p>
        </div>
        <Link href={`/keyrings/${id}/keys/new`}>
          <Button>
            <PlusIcon className="h-5 w-5 mr-2" />
            New Key
          </Button>
        </Link>
      </div>

      {/* Keys */}
      <div className="bg-white shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h2 className="text-lg font-medium text-gray-900">Keys</h2>
          {keysData?.keys?.length ? (
            <div className="mt-4">
              <table className="min-w-full divide-y divide-gray-200">
                <thead>
                  <tr>
                    <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900">
                      Version
                    </th>
                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Algorithm
                    </th>
                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Purpose
                    </th>
                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Status
                    </th>
                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Created
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {keysData.keys.map((key) => (
                    <tr key={key.version}>
                      <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900">
                        {key.version}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        {key.algorithm}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        {key.purpose}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm">
                        <span
                          className={`inline-flex rounded-full px-2 text-xs font-semibold leading-5 ${
                            key.status === 'active'
                              ? 'bg-success-50 text-success-700'
                              : 'bg-error-50 text-error-700'
                          }`}
                        >
                          {key.status}
                        </span>
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        {new Date(key.createdAt).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-12">
              <KeyIcon className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-semibold text-gray-900">No keys</h3>
              <p className="mt-1 text-sm text-gray-500">
                Get started by creating a new key.
              </p>
              <Link href={`/keyrings/${id}/keys/new`}>
                <Button className="mt-4">Create Key</Button>
              </Link>
            </div>
          )}
        </div>
      </div>
    </div>
  )
} 