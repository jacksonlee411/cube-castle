// @vitest-environment jsdom
import { describe, it, expect } from 'vitest';
import { coerceOrganizationLevel, getDisplayLevel } from '../organization-helpers';

describe('organization-helpers', () => {
  describe('coerceOrganizationLevel', () => {
    it('returns numeric level when valid number provided', () => {
      expect(coerceOrganizationLevel(3)).toBe(3);
    });

    it('parses numeric string values correctly', () => {
      expect(coerceOrganizationLevel('4')).toBe(4);
    });

    it('prefers primary value even when fallback exists', () => {
      expect(coerceOrganizationLevel(0, 2)).toBe(0);
    });

    it('uses fallback when primary value is null or undefined', () => {
      expect(coerceOrganizationLevel(null, 5)).toBe(5);
      expect(coerceOrganizationLevel(undefined, '6')).toBe(6);
    });

    it('falls back to 0 when both primary and fallback are invalid', () => {
      expect(coerceOrganizationLevel('invalid', 'NaN')).toBe(0);
    });
  });

  describe('getDisplayLevel', () => {
    it('offsets level by default value', () => {
      expect(getDisplayLevel(0, 1)).toBe(1);
      expect(getDisplayLevel(2, 1)).toBe(3);
    });

    it('handles non-finite values by treating them as 0', () => {
      expect(getDisplayLevel(Number.NaN, 1)).toBe(1);
    });
  });
});

