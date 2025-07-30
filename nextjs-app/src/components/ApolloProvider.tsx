// src/components/ApolloProvider.tsx
import React, { ReactNode, useEffect, useState } from 'react';
import { ApolloProvider as BaseApolloProvider } from '@apollo/client';
import { apolloClient } from '@/lib/graphql-client';
import { Spin } from 'antd';

interface ApolloProviderProps {
  children: ReactNode;
}

/**
 * Resilient Apollo Provider that handles initialization errors gracefully
 * This wrapper prevents Apollo client invariant violations during app startup
 */
const ApolloProvider: React.FC<ApolloProviderProps> = ({ children }) => {
  const [clientReady, setClientReady] = useState(false);
  const [initializationError, setInitializationError] = useState<Error | null>(null);

  useEffect(() => {
    const initializeClient = async () => {
      try {
        // Wait for client to be ready
        await new Promise(resolve => setTimeout(resolve, 100));
        
        // Pre-warm the cache with empty data to prevent invariant violations
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
      } catch (error) {
        console.warn('Apollo Client initialization warning:', error);
        // Still mark as ready to allow fallback to REST API
        setClientReady(true);
        setInitializationError(error as Error);
      }
    };

    initializeClient();
  }, []);

  // Show loading spinner during initialization
  if (!clientReady) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh' 
      }}>
        <Spin size="large" tip="正在初始化应用..." />
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