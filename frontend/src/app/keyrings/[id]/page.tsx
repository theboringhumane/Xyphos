'use client'

import { Navigation } from '@/components/layout/Navigation'
import { KeyringDetail } from '@/components/keyrings/KeyringDetail'
import React from 'react'

// 🔑 Keyring Detail Page
export default function KeyringDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = React.use(params)

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />
      <main className="py-10">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <KeyringDetail id={id} />
        </div>
      </main>
    </div>
  )
} 