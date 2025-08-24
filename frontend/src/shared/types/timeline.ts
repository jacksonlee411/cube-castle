/**
 * 时间线操作和管理类型定义
 * 支持时间线的可视化、导航和操作
 */

import type { TimelineEvent, EventType, EventStatus } from './temporal';

// 时间线视图配置
export interface TimelineViewConfig {
  mode: TimelineDisplayMode;
  zoomLevel: TimelineZoomLevel;
  showEvents: EventType[];
  dateRange: {
    start: Date;
    end: Date;
  };
  groupBy?: TimelineGrouping;
  filterBy?: TimelineFilter;
}

// 时间线显示模式
export type TimelineDisplayMode = 'compact' | 'detailed' | 'gantt' | 'calendar';

// 时间线缩放级别
export type TimelineZoomLevel = 'year' | 'quarter' | 'month' | 'week' | 'day';

// 时间线分组方式
export type TimelineGrouping = 'none' | 'eventType' | 'organization' | 'author' | 'status';

// 时间线筛选器
export interface TimelineFilter {
  eventTypes?: EventType[];
  statuses?: EventStatus[];
  authors?: string[];
  organizationCodes?: string[];
  dateRange?: {
    start: Date;
    end: Date;
  };
}

// 时间线数据点
export interface TimelineDataPoint {
  id: string;
  timestamp: Date;
  organizationCode: string;
  organizationName: string;
  event: TimelineEvent;
  position: TimelinePosition;
  visual: TimelineVisualConfig;
  interactions: TimelineInteraction[];
}

// 时间线位置信息
export interface TimelinePosition {
  x: number;          // 时间轴位置
  y: number;          // 垂直位置 (多轨道显示)
  width?: number;     // 事件持续时间宽度
  height?: number;    // 显示高度
  track: number;      // 轨道编号
}

// 时间线视觉配置
export interface TimelineVisualConfig {
  color: string;
  icon?: string;
  size: 'small' | 'medium' | 'large';
  shape: 'circle' | 'square' | 'diamond' | 'triangle';
  opacity?: number;
  highlighted?: boolean;
  selected?: boolean;
}

// 时间线交互配置
export interface TimelineInteraction {
  type: InteractionType;
  enabled: boolean;
  handler?: (event: TimelineEvent) => void;
}

export type InteractionType = 
  | 'click'
  | 'hover'
  | 'double_click'
  | 'context_menu'
  | 'drag'
  | 'select';

// 时间线导航器
export interface TimelineNavigator {
  currentDate: Date;
  viewportStart: Date;
  viewportEnd: Date;
  totalRange: {
    start: Date;
    end: Date;
  };
  bookmarks: TimelineBookmark[];
  quickJumps: QuickJumpOption[];
}

// 时间线书签
export interface TimelineBookmark {
  id: string;
  name: string;
  date: Date;
  description?: string;
  color?: string;
  icon?: string;
  organizationCode?: string;
}

// 快速跳转选项
export interface QuickJumpOption {
  id: string;
  label: string;
  date: Date;
  type: 'preset' | 'bookmark' | 'recent';
}

// 时间线操作接口
export interface TimelineOperations {
  // 时间线导航
  jumpToDate: (date: Date) => void;
  jumpToEvent: (eventId: string) => void;
  zoomIn: () => void;
  zoomOut: () => void;
  resetZoom: () => void;
  
  // 时间线编辑
  createEvent: (event: Partial<TimelineEvent>) => Promise<TimelineEvent>;
  updateEvent: (eventId: string, updates: Partial<TimelineEvent>) => Promise<TimelineEvent>;
  deleteEvent: (eventId: string) => Promise<void>;
  
  // 批量操作
  moveEvents: (eventIds: string[], newDate: Date) => Promise<void>;
  duplicateEvents: (eventIds: string[]) => Promise<TimelineEvent[]>;
  
  // 时间线校正
  correctTimeline: (correction: TimelineCorrection) => Promise<void>;
  undoCorrection: (correctionId: string) => Promise<void>;
}

// 时间线校正
export interface TimelineCorrection {
  id: string;
  type: CorrectionType;
  targetEventId: string;
  newTimestamp: Date;
  reason: string;
  approvedBy?: string;
  appliedAt?: Date;
  previousState: TimelineEvent;
}

export type CorrectionType = 'timestamp' | 'eventType' | 'rollback' | 'merge' | 'split';

// 时间线状态管理
export interface TimelineState {
  // 视图状态
  config: TimelineViewConfig;
  navigator: TimelineNavigator;
  selectedEvents: string[];
  highlightedEvents: string[];
  
  // 数据状态
  events: TimelineDataPoint[];
  loading: boolean;
  error?: string;
  
  // 交互状态
  dragging?: {
    eventId: string;
    startPosition: { x: number; y: number };
  };
  contextMenu?: {
    eventId: string;
    position: { x: number; y: number };
    visible: boolean;
  };
}

// 时间线分析指标
export interface TimelineAnalytics {
  totalEvents: number;
  eventsByType: Record<EventType, number>;
  eventsByStatus: Record<EventStatus, number>;
  averageEventDuration: number;  // 分钟
  peakActivityPeriods: {
    period: string;
    eventCount: number;
  }[];
  changePatterns: {
    pattern: string;
    frequency: number;
    description: string;
  }[];
}

// 时间线导出配置
export interface TimelineExportConfig {
  format: 'json' | 'csv' | 'pdf' | 'png' | 'svg';
  dateRange?: {
    start: Date;
    end: Date;
  };
  includeMetadata: boolean;
  includeVisuals: boolean;
  resolution?: 'low' | 'medium' | 'high';  // 图像导出分辨率
  pageSize?: 'A4' | 'A3' | 'letter';       // PDF页面大小
}

// 时间线性能配置
export interface TimelinePerformanceConfig {
  virtualScrolling: boolean;      // 虚拟滚动
  lazyLoading: boolean;          // 懒加载事件
  maxVisibleEvents: number;       // 最大可见事件数
  renderThreshold: number;        // 渲染阈值
  cacheStrategy: 'memory' | 'disk' | 'hybrid';
  prefetchMargin: number;         // 预取边距 (毫秒)
}