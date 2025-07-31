// src/components/ApolloProvider.tsx
import React, { ReactNode, useEffect, useState } from 'react';
import { ApolloProvider as BaseApolloProvider } from '@apollo/client';
import { apolloClient, restoreCache, persistCache } from '@/lib/graphql-client';
import { Spin } from 'antd';

interface ApolloProviderProps {
  children: ReactNode;
}

/**
 * Enhanced Apollo Provider with cache persistence and performance monitoring
 * Phase 2 optimization: Enterprise-grade initialization with offline support
 */
const ApolloProvider: React.FC<ApolloProviderProps> = ({ children }) => {
  const [clientReady, setClientReady] = useState(false);
  const [initializationError, setInitializationError] = useState<Error | null>(null);
  const [cacheRestored, setCacheRestored] = useState(false);

  useEffect(() => {
    let cleanupFn: (() => void) | undefined;

    const initializeClient = async () => {
      try {
        // Phase 2: Restore cache from localStorage for faster startup
        await restoreCache();
        setCacheRestored(true);
        
        // Wait for client to be ready
        await new Promise(resolve => setTimeout(resolve, 100));
        
        // Pre-warm the cache with essential data structure
        apolloClient.writeQuery({
          query: require('graphql-tag')`
            query PrewarmCache {
              __typename
            }
          `,
          data: {
            __typename: 'Query'
          }
        });
        
        setClientReady(true);
        
        // Phase 2: Setup cache persistence on window unload
        if (typeof window !== 'undefined') {
          const handleBeforeUnload = () => {
            persistCache();
          };
          
          window.addEventListener('beforeunload', handleBeforeUnload);
          
          // Periodic cache persistence for long-running sessions
          const persistInterval = setInterval(() => {
            persistCache();
          }, 5 * 60 * 1000); // Every 5 minutes
          
          // Store cleanup function
          cleanupFn = () => {
            window.removeEventListener('beforeunload', handleBeforeUnload);
            clearInterval(persistInterval);
            persistCache(); // Final persist
          };
        }
        
      } catch (error) {
        // Apollo Client initialization error - fallback mode enabled
        setClientReady(true);
        setInitializationError(error as Error);
      }
    };

    initializeClient();
    
    // Return cleanup function
    return () => {
      if (cleanupFn) {
        cleanupFn();
      }
    };
  }, []);

  // Enhanced loading UI with cache restoration status
  if (!clientReady) {
    return (
      <div style={{ 
        display: 'flex', 
        flexDirection: 'column',
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh',
        gap: '16px'
      }}>
        <Spin size="large" tip="正在初始化应用..." />
        {cacheRestored && (
          <div style={{ fontSize: '14px', color: '#666' }}>
            缓存已恢复，性能已优化
          </div>
        )}
      </div>
    );
  }

  return (
    <BaseApolloProvider client={apolloClient}>
      {children}
    </BaseApolloProvider>
  );
};

export default ApolloProvider;