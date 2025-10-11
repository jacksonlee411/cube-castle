import { QueryCache, QueryClient, MutationCache } from '@tanstack/react-query';
import type { QueryCacheNotifyEvent } from '@tanstack/query-core';
import { logger } from '@/shared/utils/logger';
import { env } from '../config/environment';
import { UnifiedErrorHandler } from './error-handling';

const LOG_INTERVAL_MS = 10_000;
const DEFAULT_STALE_TIME = 5 * 60 * 1000;

export interface QueryMetricsSnapshot {
  totalRequests: number;
  cacheHits: number;
  cacheMisses: number;
  errorCount: number;
  hitRate: number;
}

interface QueryMetricsTrackerOptions {
  enableLogging: boolean;
  logIntervalMs: number;
}

type QueryLike = {
  queryHash?: string;
  isStale?: () => boolean;
  state?: {
    data?: unknown;
    dataUpdatedAt?: number;
  };
};

const toPercentage = (value: number) => Math.round(value * 1000) / 10;

const safeIsQueryFresh = (query: QueryLike): boolean => {
  if (!query || typeof query !== 'object') {
    return false;
  }

  const hasData = Boolean(query.state && typeof query.state === 'object' && 'data' in query.state);
  if (!hasData) {
    return false;
  }

  if (typeof query.isStale === 'function') {
    try {
      return !query.isStale();
    } catch {
      return false;
    }
  }

  const dataUpdatedAt = query.state?.dataUpdatedAt;
  if (typeof dataUpdatedAt === 'number' && Number.isFinite(dataUpdatedAt)) {
    return Date.now() - dataUpdatedAt < DEFAULT_STALE_TIME;
  }

  return false;
};

const safeQueryHash = (query: QueryLike): string => {
  if (query && typeof query === 'object' && typeof query.queryHash === 'string') {
    return query.queryHash;
  }
  return 'unknown';
};

class QueryMetricsTracker {
  #stats: QueryMetricsSnapshot = {
    totalRequests: 0,
    cacheHits: 0,
    cacheMisses: 0,
    errorCount: 0,
    hitRate: 0,
  };

  #lastLoggedAt = 0;
  #options: QueryMetricsTrackerOptions;

  constructor(options: QueryMetricsTrackerOptions) {
    this.#options = options;
  }

  handleEvent(event: QueryCacheNotifyEvent): void {
    if (event.type === 'observerAdded') {
      this.#recordObservation(safeIsQueryFresh(event.query as QueryLike), safeQueryHash(event.query as QueryLike));
      return;
    }

    if (event.type === 'updated' && event.action.type === 'error') {
      this.recordError();
    }
  }

  recordError(): void {
    this.#stats.errorCount += 1;
    this.#logNow('error');
  }

  reset(): void {
    this.#stats = {
      totalRequests: 0,
      cacheHits: 0,
      cacheMisses: 0,
      errorCount: 0,
      hitRate: 0,
    };
    this.#lastLoggedAt = 0;
  }

  snapshot(): QueryMetricsSnapshot {
    return { ...this.#stats };
  }

  logNow(reason: string = 'manual'): void {
    this.#logNow(reason, { force: true });
  }

  #recordObservation(hit: boolean, queryHash: string): void {
    this.#stats.totalRequests += 1;
    if (hit) {
      this.#stats.cacheHits += 1;
    } else {
      this.#stats.cacheMisses += 1;
    }

    this.#stats.hitRate =
      this.#stats.totalRequests === 0
        ? 0
        : this.#stats.cacheHits / this.#stats.totalRequests;

    this.#logNow('observation', { queryHash });
  }

  #logNow(reason: string, { force = false, queryHash }: { force?: boolean; queryHash?: string } = {}): void {
    if (!this.#options.enableLogging) {
      return;
    }

    const now = Date.now();
    if (!force && now - this.#lastLoggedAt < this.#options.logIntervalMs) {
      return;
    }

    this.#lastLoggedAt = now;
    const snapshot = this.snapshot();
    const percentage = toPercentage(snapshot.hitRate);

    logger.info(
      `[ReactQuery] 查询缓存命中率 ${percentage}% (${snapshot.cacheHits}/${snapshot.totalRequests})`,
      {
        misses: snapshot.cacheMisses,
        errors: snapshot.errorCount,
        reason,
        queryHash,
      },
    );
  }
}

const metricsTracker = new QueryMetricsTracker({
  enableLogging: env.isDevelopment,
  logIntervalMs: LOG_INTERVAL_MS,
});

const queryCache = new QueryCache({
  onError: (error, query) => {
    metricsTracker.recordError();
    UnifiedErrorHandler.logError('React Query查询失败', error, {
      queryHash: safeQueryHash(query as QueryLike),
    });
  },
  onSuccess: (_data, query) => {
    logger.debug(
      '[ReactQuery] 查询成功',
      { queryHash: safeQueryHash(query as QueryLike) },
    );
  },
});

const mutationCache = new MutationCache({
  onError: (error, _variables, _context, mutation) => {
    UnifiedErrorHandler.logError('React Query变更失败', error, {
      mutationKey: mutation?.options?.mutationKey,
    });
  },
});

const shouldRetry = (failureCount: number, error: unknown): boolean => {
  if (failureCount >= 2) {
    return false;
  }

  if (error && typeof error === 'object' && 'code' in error) {
    const code = (error as { code?: string | undefined }).code;
    if (code === 'VALIDATION_ERROR' || code === 'NOT_FOUND' || code === 'FORBIDDEN') {
      return false;
    }
  }

  return true;
};

export const queryClient = new QueryClient({
  queryCache,
  mutationCache,
  defaultOptions: {
    queries: {
      gcTime: 30 * 60 * 1000,
      staleTime: 5 * 60 * 1000,
      retry: shouldRetry,
      refetchOnReconnect: true,
      refetchOnWindowFocus: false,
      refetchOnMount: false,
    },
    mutations: {
      retry: 1,
    },
  },
});

queryCache.subscribe((event) => metricsTracker.handleEvent(event));

if (env.isDevelopment) {
  (globalThis as Record<string, unknown>).__cubeCastleQueryMetrics__ = {
    getSnapshot: () => metricsTracker.snapshot(),
  };
}

export const queryMetrics = {
  getSnapshot: (): QueryMetricsSnapshot => metricsTracker.snapshot(),
  reset: () => metricsTracker.reset(),
  logNow: () => metricsTracker.logNow('manual'),
};

export interface QueryErrorDetail {
  code?: string;
  requestId?: string;
  details?: unknown;
}

export const createQueryError = (
  message: string,
  detail: QueryErrorDetail = {},
): Error & QueryErrorDetail => {
  const error = new Error(message) as Error & QueryErrorDetail;
  if (detail.code) {
    error.code = detail.code;
  }
  if (detail.requestId) {
    error.requestId = detail.requestId;
  }
  if (detail.details !== undefined) {
    error.details = detail.details;
  }
  return error;
};

export const __internal = {
  QueryMetricsTracker,
  safeIsQueryFresh,
  safeQueryHash,
};
