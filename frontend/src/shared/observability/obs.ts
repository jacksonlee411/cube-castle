/**
 * Minimal Observability Emitter for UI events
 * - Single source for `[OBS] <event>` emission
 * - Honors env gates: VITE_OBS_ENABLED, DEV, VITE_ENABLE_MUTATION_LOGS
 * - Provides small helpers for performance marks and simple de-duplication
 *
 * Note: Keep this file tiny and dependency-free to avoid coupling.
 */
import { logger } from '@/shared/utils/logger';

type Json = Record<string, unknown>;

const env =
  typeof import.meta !== 'undefined' && (import.meta as { env?: Record<string, string> }).env
    ? ((import.meta as { env: Record<string, string | boolean | undefined> }).env)
    : { DEV: false, MODE: 'test' };

const isDev = Boolean(env.DEV);
const obsEnabled = (env.VITE_OBS_ENABLED as unknown as string) === 'true' || isDev === true;
const mutationEnabled = (env.VITE_ENABLE_MUTATION_LOGS as unknown as string) === 'true';

const onceKeys = new Set<string>();
let lastTabFromTo: { from?: string; to?: string } | null = null;

export const obs = {
  enabled(): boolean {
    return obsEnabled;
  },
  emit(event: string, payload: Json = {}): void {
    if (!obsEnabled) return;
    const ts = new Date().toISOString();
    const body = { ...payload, ts, source: 'ui' };
    const msg = `[OBS] ${event}`;
    if (mutationEnabled && !isDev) {
      // CI channel: ensure visibility even if verbose is disabled
      logger.mutation(msg, body);
    } else {
      logger.info(msg, body);
    }
  },
  emitOnce(event: string, key: string, payload: Json = {}): void {
    if (!obsEnabled) return;
    const k = `${event}::${key}`;
    if (onceKeys.has(k)) return;
    onceKeys.add(k);
    obs.emit(event, payload);
  },
  emitTabChangeOnce(from: string | undefined, to: string | undefined, payload: Json = {}): void {
    if (!obsEnabled) return;
    const f = from ?? '';
    const t = to ?? '';
    if (f === t) return;
    // Avoid immediate duplicate of same from/to
    if (lastTabFromTo && lastTabFromTo.from === f && lastTabFromTo.to === t) return;
    lastTabFromTo = { from: f, to: t };
    obs.emit('position.tab.change', { tabFrom: f, tabTo: t, ...payload });
  },
  // Performance helpers
  markStart(name: string): void {
    if (typeof performance?.mark !== 'function') return;
    try {
      performance.mark(`${name}:start`);
    } catch {
      // ignore
    }
  },
  markEndAndMeasure(name: string): number | null {
    if (typeof performance?.mark !== 'function' || typeof performance?.measure !== 'function') {
      return null;
    }
    const start = `${name}:start`;
    const end = `${name}:end`;
    const duration = `${name}:duration`;
    try {
      performance.mark(end);
      const measure = performance.measure(duration, start, end);
      // Clean up to avoid leaking entries in long-lived pages
      performance.clearMarks(start);
      performance.clearMarks(end);
      performance.clearMeasures(duration);
      return typeof measure.duration === 'number' ? Math.round(measure.duration) : null;
    } catch {
      return null;
    }
  },
};

export type Obs = typeof obs;

