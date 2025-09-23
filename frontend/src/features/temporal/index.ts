/**
 * 时态功能模块入口文件
 */

// 主要组件 - 移除已删除的组件
// export { default as TemporalDashboard } from './TemporalDashboard'; // 已删除
// export type { TemporalDashboardProps } from './TemporalDashboard'; // 已删除

// 组织详情组件
export { TemporalDatePicker } from './components/TemporalDatePicker';
export { validateTemporalDate } from '@/shared/utils/temporal-validation-adapter';
export { 
  TemporalStatusSelector, 
  temporalStatusUtils,
  TEMPORAL_STATUS_OPTIONS 
} from './components/TemporalStatusSelector';
export type { TemporalStatus } from './components/TemporalStatusSelector';
export { PlannedOrganizationForm } from './components/PlannedOrganizationForm';
export { 
  TemporalInfoDisplay, 
  TemporalStatusBadge,
  TemporalDateRange 
} from './components/TemporalInfoDisplay';

// 组织详情相关类型
export type { PlannedOrganizationData } from './components/PlannedOrganizationForm';
export type { TemporalInfo } from './components/TemporalInfoDisplay';

// 子组件（原有）
export * from './components';

// 工具函数和常量
export const TEMPORAL_CONSTANTS = {
  CACHE_DURATION: 5 * 60 * 1000, // 5分钟
  MAX_HISTORY_VERSIONS: 50,
  MAX_TIMELINE_EVENTS: 100,
  DEFAULT_PAGE_SIZE: 20
} as const;

// 工具函数
export const temporalUtils = {
  formatTemporalDate: (dateStr: string, options?: Intl.DateTimeFormatOptions) => {
    try {
      return new Date(dateStr).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        ...options
      });
    } catch {
      return dateStr;
    }
  },

  isTemporalOrganization: (org: { effectiveFrom?: string; effectiveTo?: string; isTemporal?: boolean }): boolean => {
    return !!(org.effectiveFrom || org.effectiveTo || org.isTemporal);
  },

  getTemporalStatus: (org: { effectiveFrom?: string; effectiveTo?: string }, asOfDate?: string) => {
    const now = asOfDate ? new Date(asOfDate) : new Date();
    const effectiveFrom = org.effectiveFrom ? new Date(org.effectiveFrom) : null;
    const effectiveTo = org.effectiveTo ? new Date(org.effectiveTo) : null;

    if (effectiveFrom && now < effectiveFrom) {
      return 'future'; // 未来生效
    }
    if (effectiveTo && now > effectiveTo) {
      return 'expired'; // 已失效
    }
    return 'active'; // 当前有效
  },

  buildTemporalQueryKey: (organizationCode: string, params: Record<string, unknown>) => {
    return `temporal:${organizationCode}:${JSON.stringify(params)}`;
  }
};
