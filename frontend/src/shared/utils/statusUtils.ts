import { colors } from '@workday/canvas-kit-react/tokens';
import type { OrganizationStatus } from '../types/contract_gen';
import { TEMPORAL_ENTITY_STATUS_META } from '@/features/temporal/entity/statusMeta';

// 状态配置 - 复用 Temporal Entity 元数据，保持单一事实来源
export const STATUS_CONFIG = TEMPORAL_ENTITY_STATUS_META.organization;

/**
 * 状态工具函数
 */
export const statusUtils = {
  // 获取状态标签
  getStatusLabel: (status: OrganizationStatus): string => {
    return STATUS_CONFIG[status]?.label || status;
  },

  // 获取状态颜色
  getStatusColor: (status: OrganizationStatus): string => {
    return STATUS_CONFIG[status]?.color || colors.licorice400;
  },

  // 验证状态有效性
  isValidStatus: (status: string): status is OrganizationStatus => {
    return status in STATUS_CONFIG;
  },

  // 获取所有状态选项
  getAllStatusOptions: () => {
    return Object.entries(STATUS_CONFIG).map(([value, config]) => ({
      value: value as OrganizationStatus,
      label: config.label,
      description: config.description
    }));
  },

  // 判断是否为激活状态
  isActive: (status: OrganizationStatus): boolean => {
    return status === 'ACTIVE';
  },

  // 判断是否可以操作
  canOperate: (status: OrganizationStatus): boolean => {
    return status === 'ACTIVE' || status === 'INACTIVE';
  },

  // 获取可用操作
  getAvailableActions: (status: OrganizationStatus): string[] => {
    switch (status) {
      case 'ACTIVE':
        return ['UPDATE', 'SUSPEND'];
      case 'INACTIVE':
        return ['REACTIVATE'];
      case 'PLANNED':
        return ['UPDATE'];
      case 'DELETED':
        return [];
      default:
        return [];
    }
  }
};
