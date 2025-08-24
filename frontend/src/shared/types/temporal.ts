/**
 * 组织详情核心类型定义
 * 支持组织架构的时间维度查询和操作
 */

// 时态查询模式
export type TemporalMode = 'current' | 'historical' | 'planning';

// 时间范围定义
export interface DateRange {
  start: string;
  end: string;
}

// 时态查询参数
export interface TemporalQueryParams {
  asOfDate?: string;         // 查询特定时间点的数据
  dateRange?: DateRange;     // 查询时间范围
  includeHistory?: boolean;  // 是否包含历史版本
  includeFuture?: boolean;   // 是否包含未来规划
  includeInactive?: boolean; // 是否包含停用数据
  mode?: TemporalMode;       // 查询模式
  limit?: number;           // 查询数量限制
  eventTypes?: EventType[]; // 事件类型过滤
}

// 变更信息 (纯日期生效模型 - 统一字符串类型)
export interface ChangeInfo {
  id: string;
  timestamp: string;  // 统一为字符串类型
  type: 'creation' | 'modification' | 'deletion' | 'status_change';
  description: string;
  author?: string;
  changes?: Record<string, { old: unknown; new: unknown }>;
}

// 时态组织单元 - 扩展原有OrganizationUnit (符合camelCase规范)
export interface TemporalOrganizationUnit {
  // 基础字段 (继承自OrganizationUnit)
  code: string;
  parentCode?: string;        // camelCase
  name: string;
  unitType: 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM';  // camelCase
  status: 'ACTIVE' | 'SUSPENDED' | 'PLANNED' | 'DELETED';
  level: number;
  path: string;
  sortOrder: number;          // camelCase
  description?: string;
  createdAt: string;          // camelCase
  updatedAt: string;          // camelCase
  tenantId?: string;          // 租户ID

  // 时态扩展字段 (纯日期生效模型)
  effectiveDate: string;      // 生效日期 camelCase
  endDate?: string;           // 结束日期 (可选，undefined表示当前有效) camelCase
  isCurrent: boolean;         // 是否为当前有效记录 camelCase
  changeReason?: string;      // 变更原因 camelCase
  approvedBy?: string;        // 批准人 camelCase
  approvedAt?: string;        // 批准时间 camelCase
}

// 组织历史记录列表 (纯日期生效模型)
export interface OrganizationHistory {
  organizationCode: string;
  records: TemporalOrganizationUnit[];  // 改名为records，去掉版本概念
  totalRecords: number;                 // 改名为totalRecords
  timelineEvents: TimelineEvent[];
}

// 时间线事件
export interface TimelineEvent {
  id: string;
  organizationCode: string;
  timestamp: string;
  type: EventType;
  title: string;
  description?: string;
  changes?: FieldChange[];
  metadata?: Record<string, unknown>;
  author?: string;
  status: EventStatus;
}

// 事件类型
export type EventType = 
  | 'organization_created'
  | 'organization_updated' 
  | 'organization_deleted'
  | 'status_changed'
  | 'hierarchy_changed'
  | 'metadata_updated'
  | 'planned_change'
  | 'change_cancelled';

// 事件状态
export type EventStatus = 'planned' | 'active' | 'completed' | 'cancelled';

// 字段变更详情
export interface FieldChange {
  field: string;
  fieldLabel: string;
  oldValue: unknown;
  newValue: unknown;
  changeType: 'added' | 'modified' | 'removed';
}

// 时态查询选项
export interface TemporalQueryOptions {
  mode: TemporalMode;
  selectedDate?: Date;
  dateRange?: DateRange;
  compareRecords?: string[];  // 需要对比的记录ID列表 (纯日期模型)
  includeMetadata?: boolean;
  maxResults?: number;
}

// 时态统计信息 (纯日期生效模型)
export interface TemporalStats {
  totalRecords: number;          // 总记录数
  activeRecords: number;         // 当前有效记录数
  plannedChanges: number;        // 计划中的变更数
  lastModified: string;            // 最后修改时间 (统一为字符串)
  averageLifespanDays: number;   // 平均生命周期(天)
  changeFrequency: number;       // 每月变更次数
}

// 批量时态操作 (统一字符串类型)
export interface BatchTemporalOperation {
  operationId: string;
  type: 'bulk_update' | 'bulk_delete' | 'bulk_plan';
  organizationCodes: string[];
  effectiveDate: string;  // 统一为字符串
  endDate?: string;       // 统一为字符串
  changes: Record<string, unknown>;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress?: number;  // 0-100
}

// 时态缓存配置 (纯日期生效模型)
export interface TemporalCacheConfig {
  currentDataTTL: number;     // 当前数据缓存时长 (秒)
  historicalDataTTL: number;  // 历史数据缓存时长 (秒) 
  maxRecordsCache: number;    // 最大缓存记录数 (去掉版本概念)
  enablePrefetch: boolean;    // 是否启用预取
}

// 时态权限配置
export interface TemporalPermissions {
  canViewHistory: boolean;
  canViewFuture: boolean;
  canCreatePlannedChanges: boolean;
  canModifyHistory: boolean;  // 时间线校正权限
  canCancelPlannedChanges: boolean;
  maxHistoryViewDays?: number;
}

// 时间线视图配置 (纯日期生效模型) - 重命名为TemporalTimelineViewConfig避免与timeline.ts冲突
export interface TemporalTimelineViewConfig {
  showEvents: boolean;          // 显示事件
  showRecords: boolean;         // 显示历史记录 (替换showVersions)
  dateFormat: string;           // 日期格式
  timeRange: DateRange;         // 时间范围
  eventTypes: EventType[];      // 事件类型
}

// 时态上下文 (统一字符串类型)
export interface TemporalContext {
  mode: TemporalMode;
  currentDate: string;  // 统一为字符串
  viewConfig: TemporalTimelineViewConfig;
  permissions: TemporalPermissions;
  cacheConfig: TemporalCacheConfig;
}