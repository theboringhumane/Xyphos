'use client'

import { KeyringForm } from '@/components/keyrings/KeyringForm'
import { Navigation } from '@/components/layout/Navigation'

export default function NewKeyringPage() {
  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <main className="py-10">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="space-y-6">
            <div>
              <h1 className="text-3xl font-bold text-foreground">
                Create New Keyring
              </h1>
              <p className="mt-2 text-muted-foreground">
                Create a new keyring to manage encryption keys
              </p>
            </div>
            <div className="bg-background shadow-lg ring-1 ring-border rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <KeyringForm />
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  )
} 