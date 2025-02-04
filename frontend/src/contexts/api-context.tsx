// 🌐 API Context for Xyphos

'use client';

import React, { createContext, useContext, useEffect, useState } from 'react';
import { useSession } from 'next-auth/react';
import { XyphosClient } from '@/lib/api';

// 🔧 Context type
interface ApiContextType {
  client: XyphosClient | null;
  isLoading: boolean;
}

// 📝 Create context
const ApiContext = createContext<ApiContextType>({
  client: null,
  isLoading: true,
});

// 🎣 Hook for using the API context
export const useApi = () => {
  const context = useContext(ApiContext);
  if (!context) {
    throw new Error('useApi must be used within an ApiProvider');
  }
  return context;
};

// 🏗️ Provider component
export function ApiProvider({ children }: { children: React.ReactNode }) {
  const { data: session, status } = useSession();
  const [client, setClient] = useState<XyphosClient | null>(null);

  useEffect(() => {
    if (status === 'loading') return;

    // Create new client when session changes
    const newClient = new XyphosClient({
      baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
      token: session?.accessToken as string,
    });

    setClient(newClient);
  }, [session, status]);

  return (
    <ApiContext.Provider value={{ client, isLoading: status === 'loading' }}>
      {children}
    </ApiContext.Provider>
  );
} 