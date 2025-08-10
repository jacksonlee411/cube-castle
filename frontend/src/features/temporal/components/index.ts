/**
 * 时态组件库入口文件
 * 导出所有时态相关组件和钩子
 */

// 核心组件
export { default as TemporalNavbar } from './TemporalNavbar';
export { default as DateTimePicker } from './DateTimePicker';
export { default as Timeline } from './Timeline';
export { default as VersionComparison } from './VersionComparison';
export { default as TemporalTable } from './TemporalTable';
export { default as TemporalSettings } from './TemporalSettings';

// 添加时态状态选择器的导出
export { TemporalStatusSelector, temporalStatusUtils, TEMPORAL_STATUS_OPTIONS } from './TemporalStatusSelector';
export type { TemporalStatus, TemporalStatusSelectorProps } from './TemporalStatusSelector';

// 组件Props类型
export type { TemporalNavbarProps } from './TemporalNavbar';
export type { DateTimePickerProps } from './DateTimePicker';
export type { TimelineProps } from './Timeline';
export type { VersionComparisonProps } from './VersionComparison';
export type { TemporalTableProps } from './TemporalTable';
export type { TemporalSettingsProps } from './TemporalSettings';

// 重新导出核心钩子和存储
export {
  useTemporalOrganizations,
  useTemporalOrganization,
  useOrganizationHistory,
  useOrganizationTimeline,
  useTemporalMode,
  useTemporalQueryState,
  useTemporalPreloader,
  useTemporalUtils
} from '../../shared/hooks/useTemporalQuery';

export {
  useTemporalStore,
  useTemporalActions,
  temporalSelectors
} from '../../shared/stores/temporalStore';

// 重新导出类型定义
export type {
  TemporalMode,
  TemporalQueryParams,
  DateRange,
  TemporalOrganizationUnit,
  TimelineEvent,
  EventType,
  EventStatus,
  TemporalContext,
  TimelineViewConfig
} from '../../shared/types/temporal';