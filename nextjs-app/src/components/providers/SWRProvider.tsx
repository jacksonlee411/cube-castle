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
        // Global fetcher with monitoring
        fetcher: async (url: string) => {
          const startTime = Date.now();
          
          try {
            const response = await fetch(url);
            
            if (!response.ok) {
              const error = new Error(`HTTP ${response.status}: ${response.statusText}`);
              const duration = Date.now() - startTime;
              logger.trackSWRRequest(url, false, duration, error);
              throw error;
            }
            
            const data = await response.json();
            const duration = Date.now() - startTime;
            logger.trackSWRRequest(url, true, duration);
            
            return data;
          } catch (error) {
            const duration = Date.now() - startTime;
            logger.trackSWRRequest(url, false, duration, error as Error);
            throw error;
          }
        },
        
        // Global defaults optimized for performance
        dedupingInterval: 5000,        // 5 seconds default deduplication
        refreshInterval: 0,            // No automatic refresh by default
        revalidateOnFocus: false,      // Don't revalidate on focus by default
        revalidateOnReconnect: true,   // Revalidate on network reconnect
        revalidateIfStale: true,       // Revalidate stale data
        
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
        
        // Focus threshold
        focusThrottleInterval: 5000,   // Throttle focus revalidation to 5 seconds
      }}
    >
      {children}
    </SWRConfig>
  );
};