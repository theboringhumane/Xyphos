'use client'

import { Layout } from '@/components/layout/Layout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { KeyForm } from '@/components/keys/KeyForm'
import { useParams } from 'next/navigation'

export default function NewKeyPage() {
  const params = useParams()
  const keyringId = params.id as string

  return (
    <ProtectedRoute>
      <Layout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold text-text-primary">Create New Key</h1>
            <p className="mt-2 text-text-secondary">
              Create a new cryptographic key in this keyring
            </p>
          </div>
          <div className="bg-background-card shadow sm:rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <KeyForm keyringId={keyringId} />
            </div>
          </div>
        </div>
      </Layout>
    </ProtectedRoute>
  )
} 