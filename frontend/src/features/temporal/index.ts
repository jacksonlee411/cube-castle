/**
 * 时态功能模块入口文件
 */

// 主要组件
export { default as TemporalDashboard } from './TemporalDashboard';
export type { TemporalDashboardProps } from './TemporalDashboard';

// 组织详情组件
export { TemporalDatePicker, validateTemporalDate } from './components/TemporalDatePicker';
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

  isTemporalOrganization: (org: { effective_from?: string; effective_to?: string; is_temporal?: boolean }): boolean => {
    return !!(org.effective_from || org.effective_to || org.is_temporal);
  },

  getTemporalStatus: (org: { effective_from?: string; effective_to?: string }, asOfDate?: string) => {
    const now = asOfDate ? new Date(asOfDate) : new Date();
    const effectiveFrom = org.effective_from ? new Date(org.effective_from) : null;
    const effectiveTo = org.effective_to ? new Date(org.effective_to) : null;

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