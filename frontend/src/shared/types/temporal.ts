/**
 * 时态管理核心类型定义
 * 支持组织架构的时间维度查询和操作
 */

// 时态查询模式
export type TemporalMode = 'current' | 'historical' | 'planning';

// 时间范围定义
export interface DateRange {
  start: Date;
  end: Date;
}

// 时态查询参数
export interface TemporalQueryParams {
  asOfDate?: Date;           // 查询特定时间点的数据
  dateRange?: DateRange;     // 查询时间范围
  includeHistory?: boolean;  // 是否包含历史版本
  includeFuture?: boolean;   // 是否包含未来规划
}

// 版本信息
export interface VersionInfo {
  id: string;
  timestamp: Date;
  type: 'creation' | 'modification' | 'deletion' | 'status_change';
  description: string;
  author?: string;
  changes?: Record<string, { old: unknown; new: unknown }>;
}

// 时态组织单元 - 扩展原有OrganizationUnit
export interface TemporalOrganizationUnit {
  // 基础字段 (继承自OrganizationUnit)
  code: string;
  parent_code?: string;
  name: string;
  unit_type: 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM';
  status: 'ACTIVE' | 'INACTIVE' | 'PLANNED';
  level: number;
  path: string;
  sort_order: number;
  description?: string;
  created_at: string;
  updated_at: string;

  // 时态扩展字段
  effective_date: string;     // 生效日期
  end_date?: string;          // 结束日期 (null表示当前有效)
  is_current: boolean;        // 是否为当前版本
  version_number: number;     // 版本号
  predecessor_id?: string;    // 前一版本ID
  successor_id?: string;      // 后一版本ID
  change_reason?: string;     // 变更原因
  approved_by?: string;       // 批准人
  approved_at?: string;       // 批准时间
}

// 组织历史版本列表
export interface OrganizationHistory {
  organizationCode: string;
  versions: TemporalOrganizationUnit[];
  totalVersions: number;
  timelineEvents: TimelineEvent[];
}

// 时间线事件
export interface TimelineEvent {
  id: string;
  organizationCode: string;
  timestamp: Date;
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
  compareVersions?: string[];  // 需要对比的版本ID列表
  includeMetadata?: boolean;
  maxResults?: number;
}

// 时态统计信息
export interface TemporalStats {
  totalVersions: number;
  activeVersions: number;
  plannedChanges: number;
  lastModified: Date;
  averageLifespanDays: number;
  changeFrequency: number;  // 每月变更次数
}

// 批量时态操作
export interface BatchTemporalOperation {
  operationId: string;
  type: 'bulk_update' | 'bulk_delete' | 'bulk_plan';
  organizationCodes: string[];
  effectiveDate: Date;
  endDate?: Date;
  changes: Record<string, unknown>;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress?: number;  // 0-100
}

// 时态缓存配置
export interface TemporalCacheConfig {
  currentDataTTL: number;     // 当前数据缓存时长 (秒)
  historicalDataTTL: number;  // 历史数据缓存时长 (秒) 
  maxVersionsCache: number;   // 最大缓存版本数
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