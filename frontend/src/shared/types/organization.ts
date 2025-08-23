export interface OrganizationUnit {
  code: string;
  recordId?: string;  // UUID唯一标识符 (camelCase)
  parentCode?: string;  // camelCase
  name: string;
  unitType: 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM';  // camelCase
  status: 'ACTIVE' | 'SUSPENDED' | 'PLANNED' | 'DELETED';
  level: number;
  path: string;
  sortOrder: number;  // camelCase
  description?: string;
  createdAt: string;  // camelCase
  updatedAt: string;  // camelCase
  // 组织详情字段 (camelCase 统一命名)
  effectiveDate?: string;  // camelCase
  endDate?: string;  // camelCase
  isTemporal?: boolean;  // camelCase
  version?: number;
  changeReason?: string;  // camelCase
  isCurrent?: boolean;  // camelCase
}

export interface OrganizationListResponse {
  organizations: OrganizationUnit[];
  totalCount: number;  // camelCase
  page: number;
  pageSize: number;  // camelCase
  totalPages: number;  // camelCase
}

export interface OrganizationStats {
  totalCount: number;  // camelCase
  byType: Record<string, number>;  // camelCase
  byStatus: Record<string, number>;  // camelCase
}

// 组织查询参数 (camelCase)
export interface OrganizationQueryParams {
  name?: string;
  unitType?: string;  // camelCase
  status?: string;
  parentCode?: string;  // camelCase
  level?: number;
  page?: number;
  pageSize?: number;  // camelCase
  sortBy?: string;  // camelCase
  sortOrder?: 'ASC' | 'DESC';  // camelCase
}

// GraphQL API响应类型定义
export interface GraphQLOrganizationResponse {
  code: string;
  record_id?: string;  // UUID唯一标识符
  parentCode?: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  path: string;
  sortOrder: number;
  description?: string;
  createdAt: string;
  updatedAt: string;
  // 组织详情字段 (统一命名)
  effectiveDate?: string;
  endDate?: string;
  isTemporal?: boolean;
  version?: number;
  changeReason?: string;
  isCurrent?: boolean;
}

export interface GraphQLStatsTypeItem {
  type: string;
  count: number;
}

export interface GraphQLStatsStatusItem {
  status: string;
  count: number;
}

export interface GraphQLStatsResponse {
  totalCount: number;
  byType?: GraphQLStatsTypeItem[];
  byStatus?: GraphQLStatsStatusItem[];
}

export interface OrganizationListAPIResponse {
  organizations: GraphQLOrganizationResponse[];
}

export interface OrganizationStatsAPIResponse {
  organizationStats: GraphQLStatsResponse;  // camelCase
}

// 命令API响应类型 - 用于创建和更新操作 (camelCase)
export interface CreateOrganizationResponse {
  code: string;
  name: string;
  unitType: string;  // camelCase
  status: string;
  createdAt: string;  // camelCase
  level?: number;
  parentCode?: string;  // camelCase
  sortOrder?: number;  // camelCase
  description?: string;
  path?: string;
  updatedAt?: string;  // camelCase
  // 组织详情字段 (camelCase 统一命名)
  effectiveDate?: string;  // camelCase
  endDate?: string;  // camelCase
  isTemporal?: boolean;  // camelCase
  version?: number;
  changeReason?: string;  // camelCase
  isCurrent?: boolean;  // camelCase
}

export interface UpdateOrganizationResponse {
  code: string;
  updatedAt: string;  // camelCase
  changes: Record<string, unknown>;
  // 组织详情字段 (camelCase 统一命名)
  effectiveDate?: string;  // camelCase
  endDate?: string;  // camelCase
  version?: number;
  changeReason?: string;  // camelCase
}

// 组织操作请求和响应类型 (操作驱动状态管理)
export interface SuspendOrganizationRequest {
  reason: string;
}

export interface ReactivateOrganizationRequest {
  reason: string;
}

export interface SuspendOrganizationResponse {
  code: string;
  name: string;
  status: 'SUSPENDED';
  suspendedAt: string;  // camelCase
  reason: string;
}

export interface ReactivateOrganizationResponse {
  code: string;
  name: string;
  status: 'ACTIVE';
  reactivatedAt: string;  // camelCase
  reason: string;
}