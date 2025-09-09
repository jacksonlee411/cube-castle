/**
 * DEPRECATED: 该文件功能已迁移到 shared/utils/temporal-converter.ts
 * 
 * 请使用：
 * import { TemporalConverter, TemporalUtils } from '@/shared/utils/temporal-converter';
 * 
 * 新的统一API映射：
 * - isValidDate -> TemporalConverter.validateTemporalRecord
 * - isFutureDate -> 自定义逻辑使用 TemporalUtils
 * - isDateInRange -> TemporalUtils.isInRange 
 * - isEndDateAfterStartDate -> TemporalConverter.validateTemporalRecord
 * - getTodayString -> TemporalUtils.today
 * - formatDateDisplay -> TemporalUtils.formatDate
 */

// 临时兼容封装，在所有引用替换完成后删除
import { TemporalConverter, TemporalUtils } from '../../../shared/utils/temporal-converter';

export const validateTemporalDate = {
  // DEPRECATED: 使用 TemporalConverter.validateTemporalRecord
  isValidDate: (dateString: string): boolean => {
    try {
      return TemporalConverter.validateTemporalRecord({ effectiveDate: dateString });
    } catch {
      return false;
    }
  },

  // DEPRECATED: 自定义逻辑，使用 TemporalUtils
  isFutureDate: (dateString: string): boolean => {
    const date = new Date(dateString);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return date > today;
  },

  // DEPRECATED: 使用 TemporalUtils.isInRange
  isDateInRange: (dateString: string, startDate?: string, endDate?: string): boolean => {
    if (!startDate && !endDate) return true;
    if (!startDate || !endDate) return true;
    return TemporalUtils.isInRange(dateString, startDate, endDate);
  },

  // DEPRECATED: 使用 TemporalConverter.validateTemporalRecord
  isEndDateAfterStartDate: (startDate: string, endDate: string): boolean => {
    return TemporalConverter.validateTemporalRecord({ effectiveDate: startDate, endDate });
  },

  // DEPRECATED: 使用 TemporalUtils.today
  getTodayString: (): string => {
    return TemporalUtils.today();
  },

  // DEPRECATED: 使用 TemporalUtils.formatDate
  formatDateDisplay: (dateString: string): string => {
    if (!dateString) return '';
    return TemporalUtils.formatDate(dateString);
  }
};

// TODO-TEMPORARY: 该文件将在 2025-09-16 后完全删除
// 所有引用都应替换为 shared/utils/temporal-converter.ts