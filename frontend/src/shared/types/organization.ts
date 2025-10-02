import type { JsonValue } from './json';

// 🎯 核心接口1: 组织单元主实体 (统一所有组织相关字段)
export interface OrganizationUnit {
  // 主要标识字段
  code: string;
  recordId?: string;  // UUID唯一标识符 (camelCase)
  parentCode: string;  // camelCase - 必填字段，根组织使用"0"
  
  // 基本属性
  name: string;
  unitType: 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM';  // camelCase
  status: 'ACTIVE' | 'INACTIVE' | 'PLANNED' | 'DELETED';
  level: number;
  path?: string | null;
  sortOrder: number;  // camelCase
  description?: string;
  childrenCount?: number;
  
  // 审计字段
  createdAt: string;  // camelCase
  updatedAt: string;  // camelCase
  tenantId?: string;  // 租户ID
  
  // 时态字段（支持时态和非时态场景）
  effectiveDate?: string;  // camelCase
  endDate?: string;  // camelCase
  isTemporal?: boolean;  // camelCase
  isCurrent?: boolean;  // camelCase
  version?: number;
  changeReason?: string;  // camelCase
  deletedAt?: string | null;  // 审计字段
  deletedBy?: string | null;
  deletionReason?: string | null;
  suspendedAt?: string | null;
  suspendedBy?: string | null;
  suspensionReason?: string | null;
  approvedBy?: string;  // camelCase
  approvedAt?: string;  // camelCase
}

// 🎯 核心接口2: 组织列表响应 (支持分页)
export interface OrganizationListResponse {
  organizations: OrganizationUnit[];
  totalCount: number;  // camelCase
  page?: number;
  pageSize?: number;  // camelCase
  totalPages?: number;  // camelCase
}


// 🎯 核心接口3: 组织查询参数 (统一查询场景)
export interface OrganizationQueryParams {
  // 搜索条件
  name?: string;
  searchText?: string;  // 通用搜索文本
  code?: string;
  parentCode?: string;  // camelCase - 查询参数保持可选
  
  // 过滤条件
  unitType?: string;  // camelCase
  status?: string;
  level?: number;
  
  // 时态查询
  effectiveDate?: string;  // 时态查询的基准日期
  asOfDate?: string;  // 别名，兼容不同命名习惯
  includeHistorical?: boolean;  // 是否包含历史记录
  
  // 分页排序
  page?: number;
  pageSize?: number;  // camelCase
  sortBy?: string;  // camelCase
  sortOrder?: 'ASC' | 'DESC';  // camelCase
}

// 🎯 核心接口4: 统一请求类型 (合并Create/Update/Operation)
export interface OrganizationRequest {
  // 基本字段
  code?: string;  // 创建时可选（支持自动生成）
  name?: string;
  unitType?: 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM';
  status?: 'ACTIVE' | 'INACTIVE' | 'PLANNED' | 'DELETED';
  parentCode?: string;
  description?: string;
  sortOrder?: number;
  level?: number;  // 添加缺少的level字段
  
  // 时态字段
  effectiveDate?: string;
  endDate?: string;  // 添加缺少的endDate字段
  changeReason?: string;
  
  // 操作相关
  operationType?: 'CREATE' | 'UPDATE' | 'SUSPEND' | 'REACTIVATE' | 'DELETE';
  operationReason?: string;  // 操作原因（兼容reason字段）
  reason?: string;  // 向后兼容
}

// 🎯 核心接口5: 统一响应类型 (替代所有响应接口)
export interface OrganizationResponse {
  // 必返字段
  code: string;
  name: string;
  status: 'ACTIVE' | 'INACTIVE' | 'PLANNED' | 'DELETED';
  
  // 操作相关响应
  operationType?: 'CREATE' | 'UPDATE' | 'SUSPEND' | 'REACTIVATE' | 'DELETE';
  createdAt?: string;
  updatedAt?: string;
  suspendedAt?: string;  // 操作时间戳
  reactivatedAt?: string;
  
  // 完整组织信息（可选，根据API返回）
  organization?: OrganizationUnit;
  
  // 变更信息
  changes?: Record<string, JsonValue>;
  reason?: string;  // 操作原因
  
  // 时态信息
  effectiveDate?: string;
  version?: number;
}

// 🎯 核心接口6: 组件Props统一接口 (替代所有组件Props)
export interface OrganizationComponentProps {
  // 组织数据
  organization?: OrganizationUnit;
  organizations?: OrganizationUnit[];
  
  // 表格/列表Props
  loading?: boolean;
  error?: Error | null;
  onSelect?: (organization: OrganizationUnit) => void;
  onEdit?: (organization: OrganizationUnit) => void;
  onDelete?: (code: string) => void;
  
  // 表单Props
  mode?: 'create' | 'edit' | 'view';
  onSubmit?: (data: OrganizationRequest) => void;
  onCancel?: () => void;
  initialData?: OrganizationUnit;
  
  // 过滤/搜索Props
  filters?: OrganizationQueryParams;
  onFiltersChange?: (filters: OrganizationQueryParams) => void;
  
  // 树形结构Props
  expandedNodes?: string[];
  onNodeToggle?: (code: string) => void;
  showRoot?: boolean;
  
  // 时态相关Props
  temporalMode?: 'current' | 'historical' | 'planning';
  asOfDate?: string;
  
  // 通用Props
  className?: string;
  title?: string;
  disabled?: boolean;
  readOnly?: boolean;
}

// 🎯 核心接口7: 操作上下文 (已存在，保持不变)
// 由于OrganizationOperationContext已在organizationPermissions.ts中定义，这里重新导出
// export { OrganizationOperationContext } from '../utils/organizationPermissions';

// 🎯 核心接口8: 验证错误接口 (统一验证场景)
export interface OrganizationValidationError {
  field: string;
  message: string;
  code: string;
  value?: JsonValue;
}
