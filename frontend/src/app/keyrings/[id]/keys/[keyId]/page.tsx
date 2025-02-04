'use client'

import { Layout } from '@/components/layout/Layout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { KeyDetail } from '@/components/keys/KeyDetail'
import { useParams } from 'next/navigation'

export default function KeyDetailPage() {
  const params = useParams()
  const keyringId = params.id as string
  const keyId = params.keyId as string

  return (
    <ProtectedRoute>
      <Layout>
        <KeyDetail keyringId={keyringId} keyId={keyId} />
      </Layout>
    </ProtectedRoute>
  )
} 