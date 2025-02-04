'use client'

import { useQuery } from '@tanstack/react-query'
import { UserGroupIcon, PencilIcon, TrashIcon } from '@heroicons/react/24/outline'
import Link from 'next/link'

// 👥 Mock data for development
const mockTenants = [
  {
    id: '1',
    name: 'Production Team',
    description: 'Production environment access',
    keyringCount: 3,
    createdAt: '2024-03-15T10:00:00Z',
  },
  {
    id: '2',
    name: 'Development Team',
    description: 'Development environment access',
    keyringCount: 2,
    createdAt: '2024-03-14T15:30:00Z',
  },
]

export function TenantList() {
  // 🔄 Replace with actual API call
  const { data: tenants, isLoading } = useQuery({
    queryKey: ['tenants'],
    queryFn: async () => mockTenants,
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
              Keyrings
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
          {tenants?.map((tenant) => (
            <tr key={tenant.id}>
              <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-6">
                <div className="flex items-center">
                  <UserGroupIcon className="h-5 w-5 text-text-secondary mr-2" />
                  <div className="font-medium text-text-primary">{tenant.name}</div>
                </div>
              </td>
              <td className="whitespace-nowrap px-3 py-4 text-sm text-text-secondary">
                {tenant.description}
              </td>
              <td className="whitespace-nowrap px-3 py-4 text-sm text-text-secondary">
                {tenant.keyringCount}
              </td>
              <td className="whitespace-nowrap px-3 py-4 text-sm text-text-secondary">
                {new Date(tenant.createdAt).toLocaleDateString()}
              </td>
              <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                <div className="flex items-center justify-end space-x-2">
                  <Link
                    href={`/tenants/${tenant.id}/edit`}
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