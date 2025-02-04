'use client'

import { useState } from 'react'
import { useClientConfigs, useCreateClientConfig } from '@/lib/api/hooks'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@headlessui/react'
import { PlusIcon, KeyIcon } from '@heroicons/react/24/outline'

// 🔑 Client Settings Page
export default function ClientSettingsPage() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const { data: clientConfigsData, isLoading } = useClientConfigs()
  const { mutate: createClientConfig, isPending: isCreating } = useCreateClientConfig()

  // 📝 Handle client creation
  const handleCreateClient = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    
    createClientConfig({
      name: formData.get('name') as string,
      permissions: ['read:keys', 'write:keys'], // Default permissions
      expiresIn: formData.get('expiresIn') as string,
    }, {
      onSuccess: () => {
        setIsCreateModalOpen(false)
      },
    })
  }

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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold text-text-primary">API Clients</h2>
          <p className="text-sm text-text-secondary">
            Manage your API clients and access keys
          </p>
        </div>
        <Button
          onClick={() => setIsCreateModalOpen(true)}
          className="flex items-center space-x-2"
        >
          <PlusIcon className="h-5 w-5" />
          <span>New Client</span>
        </Button>
      </div>

      <div className="bg-background-card shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="space-y-4">
            {clientConfigsData?.configs.map((config) => (
              <div
                key={config.id}
                className="flex items-center justify-between p-4 border border-border-light rounded-lg"
              >
                <div className="flex items-center space-x-4">
                  <KeyIcon className="h-6 w-6 text-text-secondary" />
                  <div>
                    <h3 className="text-sm font-medium text-text-primary">
                      {config.name}
                    </h3>
                    <p className="text-xs text-text-secondary">
                      Expires: {new Date(config.expiresAt).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <code className="text-xs bg-background-hover px-2 py-1 rounded">
                    {config.apiKey}
                  </code>
                </div>
              </div>
            ))}

            {clientConfigsData?.configs.length === 0 && (
              <div className="text-center py-8">
                <KeyIcon className="mx-auto h-12 w-12 text-text-secondary" />
                <h3 className="mt-2 text-sm font-medium text-text-primary">
                  No API clients
                </h3>
                <p className="mt-1 text-sm text-text-secondary">
                  Create an API client to get started
                </p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Create Client Modal */}
      <Dialog
        open={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        className="relative z-50"
      >
        <div className="fixed inset-0 bg-black/30" aria-hidden="true" />

        <div className="fixed inset-0 flex items-center justify-center p-4">
          <Dialog.Panel className="mx-auto max-w-sm rounded-lg bg-white p-6 shadow-xl">
            <Dialog.Title className="text-lg font-medium text-gray-900">
              Create New API Client
            </Dialog.Title>
            <Dialog.Description className="mt-2 text-sm text-gray-500">
              Create a new API client to access the KMS API
            </Dialog.Description>

            <form onSubmit={handleCreateClient} className="mt-4 space-y-4">
              <div>
                <label
                  htmlFor="name"
                  className="block text-sm font-medium text-text-primary"
                >
                  Client Name
                </label>
                <input
                  type="text"
                  name="name"
                  id="name"
                  required
                  className="mt-1 block w-full rounded-md border-border-light shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                />
              </div>

              <div>
                <label
                  htmlFor="expiresIn"
                  className="block text-sm font-medium text-text-primary"
                >
                  Expires In
                </label>
                <select
                  name="expiresIn"
                  id="expiresIn"
                  required
                  className="mt-1 block w-full rounded-md border-border-light shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                >
                  <option value="24h">24 hours</option>
                  <option value="7d">7 days</option>
                  <option value="30d">30 days</option>
                  <option value="90d">90 days</option>
                </select>
              </div>

              <div className="mt-6 flex justify-end space-x-3">
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => setIsCreateModalOpen(false)}
                >
                  Cancel
                </Button>
                <Button type="submit" isLoading={isCreating}>
                  {isCreating ? 'Creating...' : 'Create Client'}
                </Button>
              </div>
            </form>
          </Dialog.Panel>
        </div>
      </Dialog>
    </div>
  )
} 