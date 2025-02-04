'use client'

import { useQuery } from '@tanstack/react-query'
import { KeyIcon, PencilIcon, TrashIcon } from '@heroicons/react/24/outline'
import Link from 'next/link'

// 🔑 Mock data for development
const mockKeyrings = [
  {
    id: '1',
    name: 'Production Keys',
    description: 'Keys for production environment',
    keyCount: 12,
    createdAt: '2024-03-15T10:00:00Z',
  },
  {
    id: '2',
    name: 'Staging Keys',
    description: 'Keys for staging environment',
    keyCount: 8,
    createdAt: '2024-03-14T15:30:00Z',
  },
]

export function KeyringList() {
  // 🔄 Replace with actual API call
  const { data: keyrings, isLoading } = useQuery({
    queryKey: ['keyrings'],
    queryFn: async () => mockKeyrings,
  })

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-4">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="h-20 bg-background-card rounded-lg border border-border-light"
          />
        ))}
      </div>
    )
  }

  return (
    <div className="overflow-hidden bg-background-card shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
      <table className="min-w-full divide-y divide-border-light">
        <thead className="bg-background-hover">
          <tr>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-text-primary sm:pl-6"
            >
              Name
            </th>
            <th
              scope="col"
              className="px-3 py-3.5 text-left text-sm font-semibold text-text-primary"
            >
              Description
            </th>
            <th
              scope="col"
              className="px-3 py-3.5 text-left text-sm font-semibold text-text-primary"
            >
              Keys
            </th>
            <th
              scope="col"
              className="px-3 py-3.5 text-left text-sm font-semibold text-text-primary"
            >
              Created
            </th>
            <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
              <span className="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border-light bg-background-light">
          {keyrings?.map((keyring) => (
            <tr key={keyring.id}>
              <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-6">
                <div className="flex items-center">
                  <KeyIcon className="h-5 w-5 text-text-secondary mr-2" />
                  <div className="font-medium text-text-primary">{keyring.name}</div>
                </div>
              </td>
              <td className="whitespace-nowrap px-3 py-4 text-sm text-text-secondary">
                {keyring.description}
              </td>
              <td className="whitespace-nowrap px-3 py-4 text-sm text-text-secondary">
                {keyring.keyCount}
              </td>
              <td className="whitespace-nowrap px-3 py-4 text-sm text-text-secondary">
                {new Date(keyring.createdAt).toLocaleDateString()}
              </td>
              <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                <div className="flex items-center justify-end space-x-2">
                  <Link
                    href={`/keyrings/${keyring.id}/edit`}
                    className="text-primary-600 hover:text-primary-900"
                  >
                    <PencilIcon className="h-5 w-5" />
                    <span className="sr-only">Edit</span>
                  </Link>
                  <button
                    onClick={() => {
                      // TODO: Implement delete functionality
                    }}
                    className="text-error-DEFAULT hover:text-error-dark"
                  >
                    <TrashIcon className="h-5 w-5" />
                    <span className="sr-only">Delete</span>
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
} 