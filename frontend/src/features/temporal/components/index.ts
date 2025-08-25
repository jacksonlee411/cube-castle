/**
 * 时态组件库入口文件
 * 导出所有时态相关组件和钩子
 */

// 核心组件 - 添加新的健壮版组件
export { default as DateTimePicker } from './DateTimePicker';
export { TimelineComponent, type TimelineComponentProps, type TimelineVersion } from './TimelineComponent'; // 新增健壮版时间轴组件
export { default as TemporalSettings } from './TemporalSettings';

// 添加时态状态选择器的导出
export { TemporalStatusSelector, temporalStatusUtils, TEMPORAL_STATUS_OPTIONS } from './TemporalStatusSelector';
export type { TemporalStatus, TemporalStatusSelectorProps } from './TemporalStatusSelector';

// 组件Props类型
export type { DateTimePickerProps } from './DateTimePicker';
export type { TemporalSettingsProps } from './TemporalSettings';
// TimelineComponentProps 和 TimelineVersion 已在上面导出

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