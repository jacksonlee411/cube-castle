export interface TemporalStatusOption {
  value: 'ACTIVE' | 'PLANNED' | 'INACTIVE' | 'EXPIRED';
  label: string;
  description: string;
}

export const TEMPORAL_STATUS_OPTIONS: TemporalStatusOption[] = [
  {
    value: 'ACTIVE',
    label: '启用',
    description: '当前生效的组织'
  },
  {
    value: 'PLANNED', 
    label: '计划',
    description: '计划在未来生效的组织'
  },
  {
    value: 'INACTIVE',
    label: '停用',
    description: '已停用的组织'
  },
  {
    value: 'EXPIRED',
    label: '已过期',
    description: '生效期已结束的组织'
  }
];

export const TEMPORAL_STATUS_COLORS = {
  ACTIVE: 'positive' as const,
  PLANNED: 'caution' as const,
  INACTIVE: 'neutral' as const,
  EXPIRED: 'critical' as const
};

export type TemporalStatus = 'ACTIVE' | 'PLANNED' | 'INACTIVE' | 'EXPIRED';

// 时态状态工具函数
export const temporalStatusUtils = {
  // 获取状态标签
  getStatusLabel: (status: TemporalStatus): string => {
    const option = TEMPORAL_STATUS_OPTIONS.find(opt => opt.value === status);
    return option?.label || status;
  },

  // 获取状态描述
  getStatusDescription: (status: TemporalStatus): string => {
    const option = TEMPORAL_STATUS_OPTIONS.find(opt => opt.value === status);
    return option?.description || '';
  },

  // 根据日期计算状态
  calculateStatus: (effectiveDate?: string, endDate?: string): TemporalStatus => {
    const today = new Date().toISOString().split('T')[0];
    
    // 如果没有生效日期，默认为启用
    if (!effectiveDate) return 'ACTIVE';
    
    // 如果生效日期在未来，为计划状态
    if (effectiveDate > today) return 'PLANNED';
    
    // 如果有结束日期且已过期，为停用状态
    if (endDate && endDate < today) return 'INACTIVE';
    
    // 其他情况为启用状态
    return 'ACTIVE';
  },

  // 判断是否为时态组织
  isTemporal: (effectiveDate?: string, endDate?: string): boolean => {
    return !!(effectiveDate || endDate);
  },

  // 获取状态颜色
  getStatusColor: (status: TemporalStatus): string => {
    switch (status) {
      case 'ACTIVE': return '#00A844'; // 绿色
      case 'PLANNED': return '#0875E1'; // 蓝色  
      case 'INACTIVE': return '#999999'; // 灰色
      case 'EXPIRED': return '#D32F2F'; // 红色
      default: return '#333333';
    }
  },

  // 获取状态图标
  getStatusIcon: (status: TemporalStatus): string => {
    switch (status) {
      case 'ACTIVE': return '✓';
      case 'PLANNED': return '';
      case 'INACTIVE': return '';
      case 'EXPIRED': return '✕';
      default: return '•';
    }
  }
};