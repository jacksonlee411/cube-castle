/**
 * 时态类型转换工具类
 * 统一处理Date和string之间的转换，解决时态系统中的类型不一致问题
 */

// 时态转换器工具类
export class TemporalConverter {
  /**
   * 将Date对象转换为ISO字符串格式
   */
  static dateToIso(date: Date | string): string {
    if (typeof date === 'string') {
      // 验证字符串格式是否为有效的ISO日期
      const parsed = new Date(date);
      if (isNaN(parsed.getTime())) {
        throw new Error(`Invalid date string: ${date}`);
      }
      return date;
    }
    // 如果是Date对象
    if (date && typeof date === 'object' && typeof date.toISOString === 'function') {
      return date.toISOString();
    }
    throw new Error(`Invalid date input: ${date}`);
  }

  /**
   * 将ISO字符串转换为Date对象
   */
  static isoToDate(iso: string): Date {
    const date = new Date(iso);
    if (isNaN(date.getTime())) {
      throw new Error(`Invalid ISO date string: ${iso}`);
    }
    return date;
  }

  /**
   * 将Date对象转换为日期字符串 (YYYY-MM-DD)
   */
  static dateToDateString(date: Date | string): string {
    if (typeof date === 'string') {
      // 如果已经是字符串，尝试提取日期部分
      if (date.includes('T')) {
        return date.split('T')[0];
      }
      // 验证是否为有效的日期字符串
      const parsed = new Date(date);
      if (isNaN(parsed.getTime())) {
        throw new Error(`Invalid date string: ${date}`);
      }
      return date.slice(0, 10);
    }
    // 如果是Date对象
    if (date && typeof date === 'object' && typeof date.toISOString === 'function') {
      return date.toISOString().slice(0, 10);
    }
    throw new Error(`Invalid date input: ${date}`);
  }

  /**
   * 将日期字符串转换为Date对象
   */
  static dateStringToDate(dateString: string): Date {
    // 如果是完整的ISO字符串，直接转换
    if (dateString.includes('T')) {
      return new Date(dateString);
    }
    // 如果是日期字符串 (YYYY-MM-DD)，添加时间部分
    const date = new Date(`${dateString}T00:00:00.000Z`);
    if (isNaN(date.getTime())) {
      throw new Error(`Invalid date string: ${dateString}`);
    }
    return date;
  }

  /**
   * 标准化时态字段 - 确保所有时态字段使用统一的字符串格式
   */
  static normalizeTemporalFields<T extends Record<string, any>>(
    obj: T,
    temporalFields: (keyof T)[]
  ): T {
    const normalized = { ...obj };
    
    temporalFields.forEach(field => {
      const value = normalized[field];
      if (value instanceof Date) {
        normalized[field] = value.toISOString() as T[keyof T];
      } else if (typeof value === 'string' && value) {
        // 验证字符串格式
        const date = new Date(value);
        if (isNaN(date.getTime())) {
          throw new Error(`Invalid date string in field ${String(field)}: ${value}`);
        }
        normalized[field] = date.toISOString() as T[keyof T];
      }
    });
    
    return normalized;
  }

  /**
   * 转换时态查询参数为API期望的格式
   */
  static normalizeTemporalQueryParams(params: {
    asOfDate?: Date | string;
    dateRange?: {
      start: Date | string;
      end: Date | string;
    };
    effectiveFrom?: Date | string;
    effectiveTo?: Date | string;
  }): {
    asOfDate?: string;
    dateRange?: {
      start: string;
      end: string;
    };
    effectiveFrom?: string;
    effectiveTo?: string;
  } {
    const normalized: any = {};

    if (params.asOfDate) {
      normalized.asOfDate = this.dateToIso(params.asOfDate);
    }

    if (params.dateRange) {
      normalized.dateRange = {
        start: this.dateToIso(params.dateRange.start),
        end: this.dateToIso(params.dateRange.end)
      };
    }

    if (params.effectiveFrom) {
      normalized.effectiveFrom = this.dateToIso(params.effectiveFrom);
    }

    if (params.effectiveTo) {
      normalized.effectiveTo = this.dateToIso(params.effectiveTo);
    }

    return normalized;
  }

  /**
   * 格式化时态日期为用户显示格式
   */
  static formatForDisplay(date: Date | string, format: 'date' | 'datetime' | 'short' = 'date'): string {
    let dateObj: Date;
    
    if (typeof date === 'string') {
      dateObj = new Date(date);
    } else if (date && typeof date === 'object' && typeof date.getTime === 'function') {
      dateObj = date;
    } else {
      return '无效日期';
    }
    
    if (isNaN(dateObj.getTime())) {
      return '无效日期';
    }

    const options: Intl.DateTimeFormatOptions = {
      timeZone: 'Asia/Shanghai'
    };

    switch (format) {
      case 'date':
        options.year = 'numeric';
        options.month = '2-digit';
        options.day = '2-digit';
        break;
      case 'datetime':
        options.year = 'numeric';
        options.month = '2-digit';
        options.day = '2-digit';
        options.hour = '2-digit';
        options.minute = '2-digit';
        options.second = '2-digit';
        break;
      case 'short':
        options.month = 'short';
        options.day = 'numeric';
        break;
    }

    return dateObj.toLocaleDateString('zh-CN', options);
  }

  /**
   * 检查两个日期是否相等（忽略时间部分）
   */
  static isSameDate(date1: Date | string, date2: Date | string): boolean {
    const d1 = this.dateToDateString(date1);
    const d2 = this.dateToDateString(date2);
    return d1 === d2;
  }

  /**
   * 检查日期是否在指定范围内
   */
  static isDateInRange(
    date: Date | string,
    start: Date | string,
    end: Date | string
  ): boolean {
    const dateStr = this.dateToDateString(date);
    const startStr = this.dateToDateString(start);
    const endStr = this.dateToDateString(end);
    
    return dateStr >= startStr && dateStr <= endStr;
  }

  /**
   * 获取当前日期字符串 (YYYY-MM-DD)
   */
  static getCurrentDateString(): string {
    return new Date().toISOString().slice(0, 10);
  }

  /**
   * 获取当前ISO字符串
   */
  static getCurrentISOString(): string {
    return new Date().toISOString();
  }

  /**
   * 验证时态记录的有效性
   */
  static validateTemporalRecord(record: {
    effective_date: string;
    end_date?: string;
  }): boolean {
    try {
      const effectiveDate = new Date(record.effective_date);
      if (isNaN(effectiveDate.getTime())) {
        return false;
      }

      if (record.end_date) {
        const endDate = new Date(record.end_date);
        if (isNaN(endDate.getTime())) {
          return false;
        }
        // 结束日期必须晚于生效日期
        return endDate > effectiveDate;
      }

      return true;
    } catch {
      return false;
    }
  }
}

// 导出常用的时态工具函数
export const TemporalUtils = {
  today: () => TemporalConverter.getCurrentDateString(),
  now: () => TemporalConverter.getCurrentISOString(),
  formatDate: (date: Date | string) => TemporalConverter.formatForDisplay(date, 'date'),
  formatDateTime: (date: Date | string) => TemporalConverter.formatForDisplay(date, 'datetime'),
  isSameDate: TemporalConverter.isSameDate,
  isInRange: TemporalConverter.isDateInRange,
  normalize: TemporalConverter.normalizeTemporalFields,
  toISO: TemporalConverter.dateToIso,
  toDate: TemporalConverter.isoToDate,
  toDateString: TemporalConverter.dateToDateString,
  validate: TemporalConverter.validateTemporalRecord
};

export default TemporalConverter;