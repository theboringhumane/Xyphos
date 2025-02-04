'use client'

import { useState } from 'react'
import { Navigation } from '@/components/layout/Navigation'
import { KeyringList } from '@/components/keyrings/KeyringList'
import { KeyringForm } from '@/components/keyrings/KeyringForm'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@headlessui/react'
import { PlusIcon } from '@heroicons/react/24/outline'

// 🔑 Keyrings Page
export default function KeyringsPage() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)

  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <main className="py-10">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="sm:flex sm:items-center">
            <div className="sm:flex-auto">
              <h1 className="text-3xl font-semibold text-foreground">Keyrings</h1>
              <p className="mt-2 text-sm text-muted-foreground">
                A list of all keyrings in your organization.
              </p>
            </div>
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <Button
                onClick={() => setIsCreateModalOpen(true)}
                className="flex items-center"
                variant="default"
              >
                <PlusIcon className="h-5 w-5 mr-2" />
                New Keyring
              </Button>
            </div>
          </div>

          <div className="mt-8">
            <KeyringList />
          </div>
        </div>
      </main>

      {/* Create Keyring Modal */}
      <Dialog
        open={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        className="relative z-50"
      >
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm" aria-hidden="true" />

        <div className="fixed inset-0 flex items-center justify-center p-4">
          <Dialog.Panel className="mx-auto max-w-sm rounded-lg bg-background p-6 shadow-lg ring-1 ring-border">
            <Dialog.Title className="text-lg font-medium text-foreground">
              Create New Keyring
            </Dialog.Title>
            <Dialog.Description className="mt-2 text-sm text-muted-foreground">
              Create a new keyring to manage encryption keys.
            </Dialog.Description>

            <div className="mt-4">
              <KeyringForm
                onSuccess={() => {
                  setIsCreateModalOpen(false)
                }}
              />
            </div>
          </Dialog.Panel>
        </div>
      </Dialog>
    </div>
  )
}