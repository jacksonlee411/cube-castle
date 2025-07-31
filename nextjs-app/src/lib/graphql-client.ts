// src/lib/graphql-client.ts
import { ApolloClient, InMemoryCache, HttpLink, from, split } from '@apollo/client';
import { getMainDefinition } from '@apollo/client/utilities';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';

// HTTP link for queries and mutations
const httpLink = new HttpLink({
  uri: process.env.NEXT_PUBLIC_GRAPHQL_ENDPOINT || 'http://localhost:8080/graphql',
});

// WebSocket link for subscriptions
const wsLink = typeof window !== 'undefined' ? new GraphQLWsLink(
  createClient({
    url: process.env.NEXT_PUBLIC_GRAPHQL_WS_ENDPOINT || 'ws://localhost:8080/graphql',
    connectionParams: () => ({
      authorization: typeof window !== 'undefined' ? localStorage.getItem('token') : null,
    }),
  })
) : null;

// Auth link to add authorization header
const authLink = setContext((_, { headers }) => {
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  const tenantId = typeof window !== 'undefined' ? localStorage.getItem('tenantId') : null;
  
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : '',
      'x-tenant-id': tenantId || '',
    }
  };
});

// Error link to handle GraphQL errors
const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
  if (graphQLErrors) {
    graphQLErrors.forEach(({ message, locations, path }) => {
      // GraphQL error logged - handled by error boundary
      
      // Handle specific errors
      if (message.includes('UNAUTHORIZED')) {
        // Redirect to login or refresh token
        if (typeof window !== 'undefined') {
          localStorage.removeItem('token');
          window.location.href = '/login';
        }
      }
    });
  }

  if (networkError) {
    // Network error - will fallback to REST API
    
    // Handle network errors gracefully
    if (networkError.message.includes('fetch') || 
        networkError.message.includes('404') ||
        networkError.message.includes('Failed to fetch')) {
      // GraphQL endpoint not available, will fallback to REST API
      
      // Don't throw the error, let the component handle fallback
      return;
    }
  }
});

// Split link to route queries/mutations to HTTP and subscriptions to WebSocket
const splitLink = typeof window !== 'undefined' && wsLink
  ? split(
      ({ query }) => {
        const definition = getMainDefinition(query);
        return (
          definition.kind === 'OperationDefinition' &&
          definition.operation === 'subscription'
        );
      },
      wsLink,
      from([errorLink, authLink, httpLink])
    )
  : from([errorLink, authLink, httpLink]);

// Enhanced cache configuration for enterprise-grade performance
const createOptimizedCache = () => {
  return new InMemoryCache({
    // Global cache configuration
    addTypename: true,
    resultCaching: true,
    
    // Enhanced type policies for core business entities
    typePolicies: {
      Query: {
        fields: {
          employees: {
            keyArgs: ['filters', 'pagination'],
            merge(existing, incoming, { args }) {
              // Smart merge strategy for employee lists
              if (!existing || args?.pagination?.offset === 0) {
                return incoming;
              }
              return {
                ...incoming,
                edges: [...(existing.edges || []), ...(incoming.edges || [])],
                totalCount: incoming.totalCount,
              };
            },
          },
          organizations: {
            keyArgs: ['filters'],
            merge: (existing, incoming) => incoming || existing,
          },
          positions: {
            keyArgs: ['filters', 'organizationId'],
            merge: (existing, incoming) => incoming || existing,
          },
        },
      },
      Employee: {
        keyFields: ['id'],
        fields: {
          positionHistory: {
            keyArgs: false,
            merge(existing = { edges: [] }, incoming) {
              // Deduplicate position history entries
              const existingIds = new Set(existing.edges.map((edge: any) => edge.node?.id).filter(Boolean));
              const newEdges = incoming.edges.filter((edge: any) => 
                edge.node?.id && !existingIds.has(edge.node.id)
              );
              
              return {
                ...incoming,
                edges: [...existing.edges, ...newEdges],
                totalCount: incoming.totalCount,
              };
            },
          },
          // Cache computed fields to avoid recalculation
          fullName: {
            read(existing, { readField }) {
              if (existing) return existing;
              const firstName = readField('firstName');
              const lastName = readField('lastName');
              return firstName && lastName ? `${firstName} ${lastName}` : firstName || lastName || '';
            },
          },
        },
      },
      Organization: {
        keyFields: ['id'],
        fields: {
          children: {
            merge: (existing, incoming) => incoming || existing,
          },
          employeeCount: {
            // Cache employee count to reduce API calls
            merge: (existing, incoming) => incoming ?? existing,
          },
        },
      },
      Position: {
        keyFields: ['id'],
        fields: {
          occupancyRate: {
            // Cache computed occupancy rates
            merge: (existing, incoming) => incoming ?? existing,
          },
          employees: {
            keyArgs: ['status'],
            merge: (existing, incoming) => incoming || existing,
          },
        },
      },
      PositionHistoryConnection: {
        fields: {
          edges: {
            merge(existing = [], incoming) {
              // Smart deduplication for position history
              const existingIds = new Set(existing.map((edge: any) => edge.node?.id).filter(Boolean));
              const newEdges = incoming.filter((edge: any) => 
                edge.node?.id && !existingIds.has(edge.node.id)
              );
              return [...existing, ...newEdges];
            },
          },
        },
      },
    },
    
    // Garbage collection configuration
    possibleTypes: {
      Node: [
        'Employee',
        'Organization', 
        'Position',
        'PositionHistory',
        'User',
        'WorkflowExecution'
      ],
    },
  });
};

// Apollo Client configuration with enterprise-grade optimizations
export const apolloClient = new ApolloClient({
  link: splitLink,
  cache: createOptimizedCache(),
  // Enhanced default options for optimal performance
  defaultOptions: {
    watchQuery: {
      errorPolicy: 'all',
      notifyOnNetworkStatusChange: true,
      fetchPolicy: 'cache-first',
      // Reduce network requests with longer cache timeout
      nextFetchPolicy: 'cache-first',
      pollInterval: 0, // Disable automatic polling by default
    },
    query: {
      errorPolicy: 'all',
      fetchPolicy: 'cache-first',
      // Enable partial results for better UX
      partialRefetch: true,
    },
    mutate: {
      errorPolicy: 'all',
      // Optimistic updates configuration
      optimisticResponse: undefined, // Will be set per mutation
      // Refetch queries strategy
      refetchQueries: 'active',
      awaitRefetchQueries: false,
    },
  },
  connectToDevTools: process.env.NODE_ENV === 'development',
  
  // Performance optimizations
  queryDeduplication: true,
  assumeImmutableResults: true,
  
  // Cache management
  typeDefs: undefined, // Will be loaded dynamically if needed
  
  // Enhanced dev tools configuration
  devtools: {
    enabled: process.env.NODE_ENV === 'development',
    // Reduce dev tools overhead in production
  },
});

// Cache persistence utilities for offline support
export const persistCache = async () => {
  if (typeof window !== 'undefined' && 'localStorage' in window) {
    try {
      const cacheData = apolloClient.cache.extract();
      localStorage.setItem('apollo-cache', JSON.stringify(cacheData));
    } catch (error) {
      // Cache persistence failed - continue without persistence
    }
  }
};

export const restoreCache = async () => {
  if (typeof window !== 'undefined' && 'localStorage' in window) {
    try {
      const cacheData = localStorage.getItem('apollo-cache');
      if (cacheData) {
        apolloClient.cache.restore(JSON.parse(cacheData));
      }
    } catch (error) {
      // Cache restoration failed - start with empty cache
      localStorage.removeItem('apollo-cache');
    }
  }
};

// Helper function to set authentication token
export const setAuthToken = (token: string, tenantId: string) => {
  if (typeof window !== 'undefined') {
    localStorage.setItem('token', token);
    localStorage.setItem('tenantId', tenantId);
  }
};

// Helper function to clear authentication
export const clearAuth = () => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('token');
    localStorage.removeItem('tenantId');
  }
  apolloClient.clearStore();
};