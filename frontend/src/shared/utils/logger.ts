/* eslint-disable no-console -- Logger bridges console under controlled policy */

import type { JsonValue } from '@/shared/types/json';

type LoggerEnv = {
  DEV?: boolean;
  MODE?: string;
  VITE_ENABLE_MUTATION_LOGS?: string;
};

type LogValue = JsonValue | Error | Date | RegExp | bigint;
type LogArguments = [message: string, ...optional: LogValue[]];

const buildTimestamp = () => new Date().toISOString();

const formatMessage = (level: string, message: string) =>
  `[${level}] ${buildTimestamp()} - ${message}`;

const shouldEmitVerbose = (env: LoggerEnv) => Boolean(env.DEV) && env.MODE !== 'test';

const shouldEmitMutation = (env: LoggerEnv) =>
  Boolean(env.DEV) || env.VITE_ENABLE_MUTATION_LOGS === 'true';

const emit = (
  method: (...args: LogValue[]) => void,
  level: string,
  env: LoggerEnv,
  { force }: { force?: boolean } = {}
) =>
  (...args: LogArguments) => {
    if (!force && !shouldEmitVerbose(env)) {
      return;
    }

    const [message, ...rest] = args;
    method(formatMessage(level, message), ...rest);
  };

export const createLogger = (env: LoggerEnv) => {
  const verbose = shouldEmitVerbose(env);

  return {
    debug: emit(console.debug, 'DEBUG', env),
    info: emit(console.info, 'INFO', env),
    log: emit(console.log, 'LOG', env),
    warn: emit(console.warn, 'WARN', env, { force: true }),
    error: emit(console.error, 'ERROR', env, { force: true }),
    group: (label: string, callback?: () => void) => {
      if (!verbose) {
        return;
      }

      console.group(formatMessage('GROUP', label));
      if (callback) {
        try {
          callback();
        } finally {
          console.groupEnd();
        }
      }
    },
    groupEnd: () => {
      if (!verbose) {
        return;
      }

      console.groupEnd();
    },
    mutation: (...args: LogArguments) => {
      if (!shouldEmitMutation(env)) {
        return;
      }

      const [message, ...rest] = args;
      console.log(formatMessage('MUTATION', message), ...rest);
    }
  } as const;
};

export const logger = createLogger(import.meta.env);

export type Logger = typeof logger;
