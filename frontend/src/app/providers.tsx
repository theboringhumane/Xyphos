'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { SessionProvider, signIn, useSession } from 'next-auth/react'
import { useEffect } from 'react'
import { useSetAuthToken } from '@/lib/api/hooks'

// 🔄 Query client
const queryClient = new QueryClient()

// 🔒 Auth provider with API client integration
function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: session } = useSession()
  const setAuthToken = useSetAuthToken()

  useEffect(() => {
    if (session?.accessToken) {
      setAuthToken(session.accessToken as string)
    }
  }, [session, setAuthToken])

  return <>{children}</>
}

// 🛡️ Protected route wrapper
export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { status } = useSession({
    required: true,
    onUnauthenticated() {
      signIn('github')
    },
  })

  if (status === 'loading') {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-primary-500" />
      </div>
    )
  }

  return <>{children}</>
}

// 🏭 Root providers
export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <SessionProvider>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>{children}</AuthProvider>
      </QueryClientProvider>
    </SessionProvider>
  )
} 