import { SWRConfig } from 'swr';
import { ReactNode } from 'react';
import { logger } from '@/lib/logger';

interface SWRProviderProps {
  children: ReactNode;
}

// Global SWR configuration with intelligent defaults
export const SWRProvider: React.FC<SWRProviderProps> = ({ children }) => {
  return (
    <SWRConfig
      value={{
        // Global defaults optimized for data fetching
        dedupingInterval: 0,           // DISABLE deduplication to ensure fetches occur
        refreshInterval: 0,            // No automatic refresh by default
        revalidateOnFocus: true,       // ENABLE revalidate on focus to trigger fetches
        revalidateOnReconnect: true,   // Revalidate on network reconnect
        revalidateIfStale: true,       // Revalidate stale data
        revalidateOnMount: true,       // ENABLE revalidate on mount to trigger initial fetch
        
        // Error handling
        errorRetryCount: 2,            // Conservative retry count
        errorRetryInterval: 1500,      // 1.5s between retries
        shouldRetryOnError: true,      // Retry on network errors
        
        // Performance optimizations
        refreshWhenHidden: false,      // Don't refresh when page is hidden
        refreshWhenOffline: false,     // Don't refresh when offline
        
        // Global event handlers
        onSuccess: (data, key) => {
          // Track successful requests globally
          if (process.env.NODE_ENV === 'development') {
            console.log(`🎯 SWR Global Success: ${key}`);
          }
        },
        
        onError: (error, key) => {
          // Track errors globally
          if (process.env.NODE_ENV === 'development') {
            console.error(`🚨 SWR Global Error: ${key}`, error.message);
          }
          
          // In production, you might want to send to error tracking service
          // Example: Sentry.captureException(error, { tags: { swrKey: key } });
        },
        
        onLoadingSlow: (key) => {
          // Track slow requests globally
          logger.warn('SWR', key, 'Global slow request detected');
          
          if (process.env.NODE_ENV === 'development') {
            console.warn(`⏳ SWR Global Slow: ${key}`);
          }
        },
        
        // Cache provider configuration
        provider: () => new Map(),     // Use Map for better performance than default
        
        // Loading timeout
        loadingTimeout: 3000,          // 3 seconds loading timeout
        
        // Compare function for smart updates (only for critical data)
        compare: (a, b) => {
          // Basic comparison - can be overridden by individual hooks
          return JSON.stringify(a) === JSON.stringify(b);
        },
        
        // Fallback data
        fallback: {},
        
        // Focus threshold - disable throttling for immediate response
        focusThrottleInterval: 0,      // DISABLE focus throttling to ensure immediate fetches
      }}
    >
      {children}
    </SWRConfig>
  );
};