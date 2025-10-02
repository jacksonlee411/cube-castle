import { describe, expect, it, vi, afterEach } from 'vitest';

import { createLogger } from '../logger';

const createSpies = () => ({
  debug: vi.spyOn(console, 'debug').mockImplementation(() => {}),
  info: vi.spyOn(console, 'info').mockImplementation(() => {}),
  log: vi.spyOn(console, 'log').mockImplementation(() => {}),
  warn: vi.spyOn(console, 'warn').mockImplementation(() => {}),
  error: vi.spyOn(console, 'error').mockImplementation(() => {}),
  group: vi.spyOn(console, 'group').mockImplementation(() => {}),
  groupEnd: vi.spyOn(console, 'groupEnd').mockImplementation(() => {})
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('logger', () => {
  it('emits debug logs only in development mode', () => {
    const spies = createSpies();
    const logger = createLogger({ DEV: true, MODE: 'development' });

    logger.debug('Debug message', { foo: 'bar' });
    expect(spies.debug).toHaveBeenCalled();
    const [firstArg, secondArg] = spies.debug.mock.calls[0] ?? [];
    expect(firstArg).toContain('[DEBUG]');
    expect(firstArg).toContain('Debug message');
    expect(secondArg).toEqual({ foo: 'bar' });

    spies.debug.mockClear();

    const testLogger = createLogger({ DEV: true, MODE: 'test' });
    testLogger.debug('Should be ignored');

    expect(spies.debug).not.toHaveBeenCalled();
  });

  it('always emits warnings and errors', () => {
    const spies = createSpies();
    const logger = createLogger({ DEV: false, MODE: 'production' });

    logger.warn('Warning message');
    logger.error('Error message');

    expect(spies.warn).toHaveBeenCalled();
    expect(spies.error).toHaveBeenCalled();

    const warnCall = spies.warn.mock.calls[0] ?? [];
    const errorCall = spies.error.mock.calls[0] ?? [];
    expect(warnCall[0]).toContain('[WARN]');
    expect(errorCall[0]).toContain('[ERROR]');
  });

  it('controls mutation logs through environment flag', () => {
    const spies = createSpies();
    const disabledLogger = createLogger({ DEV: false, MODE: 'production' });
    disabledLogger.mutation('Disabled');
    expect(spies.log).not.toHaveBeenCalled();

    spies.log.mockClear();

    const enabledLogger = createLogger({
      DEV: false,
      MODE: 'production',
      VITE_ENABLE_MUTATION_LOGS: 'true'
    });
    enabledLogger.mutation('Enabled mutation', { context: 'test' });

    expect(spies.log).toHaveBeenCalled();
    const mutationCall = spies.log.mock.calls[0] ?? [];
    expect(mutationCall[0]).toContain('[MUTATION]');
    expect(mutationCall[1]).toEqual({ context: 'test' });
  });

  it('groups logs only when verbose output is enabled', () => {
    const spies = createSpies();
    const verboseLogger = createLogger({ DEV: true, MODE: 'development' });

    verboseLogger.group('Group section', () => {
      verboseLogger.info('Grouped message');
    });

    expect(spies.group).toHaveBeenCalled();
    const groupCall = spies.group.mock.calls[0] ?? [];
    expect(groupCall[0]).toContain('[GROUP]');
    expect(spies.groupEnd).toHaveBeenCalled();

    spies.group.mockClear();
    spies.groupEnd.mockClear();

    const silentLogger = createLogger({ DEV: false, MODE: 'production' });
    silentLogger.group('Should not appear');
    silentLogger.groupEnd();

    expect(spies.group).not.toHaveBeenCalled();
    expect(spies.groupEnd).not.toHaveBeenCalled();
  });
});
