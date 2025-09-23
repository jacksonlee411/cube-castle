/**
 * temporal-validation-adapter.ts
 *
 * 为遗留调用提供与 `validateTemporalDate` 等价的包装，便于逐步迁移至统一工具。
 */

import { TemporalConverter, TemporalUtils } from './temporal-converter';

export const validateTemporalDate = {
  isValidDate: (dateString: string): boolean => {
    try {
      return TemporalConverter.validateTemporalRecord({ effectiveDate: dateString });
    } catch {
      return false;
    }
  },

  isFutureDate: (dateString: string): boolean => {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      return false;
    }

    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return date > today;
  },

  isDateInRange: (dateString: string, startDate?: string, endDate?: string): boolean => {
    if (!startDate && !endDate) {
      return true;
    }
    return TemporalUtils.isInRange(dateString, startDate, endDate);
  },

  isEndDateAfterStartDate: (startDate: string, endDate: string): boolean => {
    return TemporalConverter.validateTemporalRecord({
      effectiveDate: startDate,
      endDate
    });
  },

  getTodayString: (): string => {
    return TemporalUtils.today();
  },

  formatDateDisplay: (dateString: string): string => {
    if (!dateString) {
      return '';
    }
    return TemporalUtils.formatDate(dateString);
  }
};

export default validateTemporalDate;
