import { describe, expect, it, vi, beforeEach } from 'vitest';
import type { QueryCacheNotifyEvent } from '@tanstack/query-core';
import {
  __internal,
  createQueryError,
  queryClient,
  queryMetrics,
  type QueryMetricsSnapshot,
} from '../queryClient';

const buildEvent = (options: { queryHash: string; stale?: boolean }): QueryCacheNotifyEvent => {
  const { queryHash, stale = false } = options;

  const query = {
    queryHash,
    getObserversCount: () => 0,
    getObservers: () => [],
    isStale: () => stale,
    promise: Promise.resolve(),
    options: {},
    meta: {},
    setData: vi.fn(),
    invalidate: vi.fn(),
    fetch: vi.fn(),
    reset: vi.fn(),
    cancel: vi.fn(),
    destroy: vi.fn(),
    state: {
      data: stale ? undefined : { ok: true },
      dataUpdatedAt: Date.now(),
      dataUpdateCount: 0,
      error: null,
      errorUpdatedAt: 0,
      errorUpdateCount: 0,
      fetchFailureCount: 0,
      fetchFailureReason: null,
      fetchMeta: undefined,
      fetchStatus: 'idle' as const,
      isInvalidated: false,
      status: 'success' as const,
    },
  };

  return {
    type: 'observerAdded',
    query: query as unknown as QueryCacheNotifyEvent['query'],
  } as QueryCacheNotifyEvent;
};

describe('queryClient shared configuration', () => {
  it('applies unified default options for queries and mutations', () => {
    const defaults = queryClient.getDefaultOptions();
    expect(defaults.queries?.staleTime).toBe(5 * 60 * 1000);
    expect(defaults.queries?.gcTime).toBe(30 * 60 * 1000);
    expect(defaults.queries?.refetchOnWindowFocus).toBe(false);
    expect(defaults.mutations?.retry).toBe(1);
  });
});

describe('createQueryError', () => {
  it('attaches requestId and code metadata', () => {
    const error = createQueryError('测试错误', {
      code: 'GRAPHQL_ERROR',
      requestId: 'req-123',
      details: { field: 'name' },
    });

    expect(error.message).toBe('测试错误');
    expect(error.code).toBe('GRAPHQL_ERROR');
    expect(error.requestId).toBe('req-123');
    expect(error.details).toEqual({ field: 'name' });
  });
});

describe('queryMetrics tracker', () => {
  beforeEach(() => {
    queryMetrics.reset();
  });

  it('records cache hits and misses via tracker events', () => {
    const tracker = new __internal.QueryMetricsTracker({
      enableLogging: false,
      logIntervalMs: 0,
    });

    tracker.handleEvent(buildEvent({ queryHash: 'cache-hit' }));
    tracker.handleEvent(buildEvent({ queryHash: 'cache-miss', stale: true }));

    const snapshot = tracker.snapshot();
    expect(snapshot.totalRequests).toBe(2);
    expect(snapshot.cacheHits).toBe(1);
    expect(snapshot.cacheMisses).toBe(1);
    expect(snapshot.hitRate).toBe(0.5);
  });

  it('logs metrics when forced', () => {
    const tracker = new __internal.QueryMetricsTracker({
      enableLogging: true,
      logIntervalMs: 0,
    });
    const spy = vi.spyOn(__internal, 'safeQueryHash').mockReturnValue('forced-log');

    tracker.handleEvent(buildEvent({ queryHash: 'forced-log' }));

    tracker.logNow('test');
    const snapshot: QueryMetricsSnapshot = tracker.snapshot();

    expect(snapshot.totalRequests).toBe(1);
    expect(snapshot.errorCount).toBe(0);

    spy.mockRestore();
  });
});
