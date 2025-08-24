/**
 * 时态组件库入口文件
 * 导出所有时态相关组件和钩子
 */

// 核心组件 - 移除已删除的组件
// export { default as TemporalNavbar } from './TemporalNavbar'; // 已删除
export { default as DateTimePicker } from './DateTimePicker';
// export { default as Timeline } from './Timeline'; // 已删除
// export { default as VersionComparison } from './VersionComparison'; // 已删除 - Canvas Kit v13兼容性问题
// export { default as TemporalTable } from './TemporalTable'; // 已删除
export { default as TemporalSettings } from './TemporalSettings';

// 添加时态状态选择器的导出
export { TemporalStatusSelector, temporalStatusUtils, TEMPORAL_STATUS_OPTIONS } from './TemporalStatusSelector';
export type { TemporalStatus, TemporalStatusSelectorProps } from './TemporalStatusSelector';

// 组件Props类型 - 移除已删除组件的类型
// export type { TemporalNavbarProps } from './TemporalNavbar'; // 已删除
export type { DateTimePickerProps } from './DateTimePicker';
// export type { TimelineProps } from './Timeline'; // 已删除
// export type { VersionComparisonProps } from './VersionComparison'; // 已删除
// export type { TemporalTableProps } from './TemporalTable'; // 已删除
export type { TemporalSettingsProps } from './TemporalSettings';

// 重新导出核心钩子和存储 - 已移除不存在的钩子
// export {
//   useTemporalOrganizations,
//   useTemporalOrganization,
//   useOrganizationHistory,
//   useOrganizationTimeline,
//   useTemporalMode,
//   useTemporalQueryState,
//   useTemporalPreloader,
//   useTemporalUtils
// } from '../../../shared/hooks/useTemporalQuery';

export {
  useTemporalStore,
  useTemporalActions,
  temporalSelectors
} from '../../../shared/stores/temporalStore';

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
  TemporalTimelineViewConfig
} from '../../../shared/types/temporal';