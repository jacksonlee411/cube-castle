import { beforeAll, afterAll, describe, expect, it, vi } from 'vitest';
import { TemporalConverter } from '../temporal-converter';
import { validateTemporalDate } from '../temporal-validation-adapter';

describe('validateTemporalDate adapter', () => {
  const fixedToday = new Date('2025-03-18T10:20:30.000Z');

  beforeAll(() => {
    vi.useFakeTimers();
    vi.setSystemTime(fixedToday);
  });

  afterAll(() => {
    vi.useRealTimers();
  });

  it('validates ISO date strings via TemporalConverter', () => {
    expect(validateTemporalDate.isValidDate('2025-03-18')).toBe(true);
    expect(validateTemporalDate.isValidDate('invalid-date')).toBe(false);
  });

  it('checks future dates relative to today', () => {
    const tomorrow = new Date(fixedToday);
    tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);
    const yesterday = new Date(fixedToday);
    yesterday.setUTCDate(yesterday.getUTCDate() - 1);

    expect(validateTemporalDate.isFutureDate(tomorrow.toISOString())).toBe(true);
    expect(validateTemporalDate.isFutureDate(yesterday.toISOString())).toBe(false);
    expect(validateTemporalDate.isFutureDate('invalid-date')).toBe(false);
  });

  it('checks date range with optional boundaries', () => {
    expect(validateTemporalDate.isDateInRange('2025-03-18', undefined, undefined)).toBe(true);
    expect(validateTemporalDate.isDateInRange('2025-03-18', '2025-03-10', '2025-03-20')).toBe(true);
    expect(validateTemporalDate.isDateInRange('2025-03-18', '2025-03-19', '2025-03-20')).toBe(false);
    expect(validateTemporalDate.isDateInRange('2025-03-18', '2025-03-10')).toBe(true);
    expect(validateTemporalDate.isDateInRange('2025-03-08', '2025-03-10')).toBe(false);
    expect(validateTemporalDate.isDateInRange('2025-03-18', undefined, '2025-03-19')).toBe(true);
    expect(validateTemporalDate.isDateInRange('2025-03-21', undefined, '2025-03-19')).toBe(false);
  });

  it('validates end date ordering via TemporalConverter', () => {
    expect(
      validateTemporalDate.isEndDateAfterStartDate('2025-03-18', '2025-03-19')
    ).toBe(true);
    expect(
      validateTemporalDate.isEndDateAfterStartDate('2025-03-18', '2025-03-17')
    ).toBe(false);
  });

  it('returns today string consistent with TemporalConverter', () => {
    expect(validateTemporalDate.getTodayString()).toBe(
      TemporalConverter.getCurrentDateString()
    );
  });

  it('formats date display via TemporalUtils', () => {
    const expected = TemporalConverter.formatForDisplay('2025-03-18');
    expect(validateTemporalDate.formatDateDisplay('2025-03-18')).toBe(expected);
    expect(validateTemporalDate.formatDateDisplay('')).toBe('');
  });
});
