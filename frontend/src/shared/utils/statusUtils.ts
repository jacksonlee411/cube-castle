import { colors } from '@workday/canvas-kit-react/tokens';
import { 
  checkCircleIcon, 
  clockPauseIcon, 
  clockIcon
} from '@workday/canvas-system-icons-web';

import type { OrganizationStatus } from '../types/contract_gen';

// 状态配置 - 符合API契约规范
export const STATUS_CONFIG = {
  ACTIVE: {
    label: '启用',
    color: colors.greenApple600,
    icon: checkCircleIcon,
    backgroundColor: colors.greenApple100,
    borderColor: colors.greenApple300,
    description: '正常运行状态'
  },
  INACTIVE: {
    label: '停用',
    color: colors.cantaloupe600,
    icon: clockPauseIcon,
    backgroundColor: colors.cantaloupe100,
    borderColor: colors.cantaloupe300,
    description: '非活跃状态（等价于停用/暂停）'
  },
  PLANNED: {
    label: '计划中',
    color: colors.blueberry600,
    icon: clockIcon,
    backgroundColor: colors.blueberry100,
    borderColor: colors.blueberry300,
    description: '计划启用状态'
  },
  DELETED: {
    label: '已删除',
    color: colors.soap600,
    icon: clockPauseIcon,
    backgroundColor: colors.soap100,
    borderColor: colors.soap300,
    description: '软删除记录，仅用于审计展示'
  }
} as const;

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
