// 时态日期验证工具函数
export const validateTemporalDate = {
  // 验证日期格式
  isValidDate: (dateString: string): boolean => {
    const date = new Date(dateString);
    return date instanceof Date && !isNaN(date.getTime()) && dateString === date.toISOString().split('T')[0];
  },

  // 验证未来日期 (用于计划组织)
  isFutureDate: (dateString: string): boolean => {
    const date = new Date(dateString);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return date > today;
  },

  // 验证日期范围
  isDateInRange: (dateString: string, startDate?: string, endDate?: string): boolean => {
    if (!startDate && !endDate) return true;
    
    const date = new Date(dateString);
    
    if (startDate && date < new Date(startDate)) return false;
    if (endDate && date > new Date(endDate)) return false;
    
    return true;
  },

  // 验证结束日期在开始日期之后
  isEndDateAfterStartDate: (startDate: string, endDate: string): boolean => {
    return new Date(endDate) > new Date(startDate);
  },

  // 获取今天的日期字符串
  getTodayString: (): string => {
    return new Date().toISOString().split('T')[0];
  },

  // 格式化日期显示
  formatDateDisplay: (dateString: string): string => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long', 
      day: 'numeric'
    });
  }
};