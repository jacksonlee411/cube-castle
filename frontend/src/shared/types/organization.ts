export interface OrganizationUnit {
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
}

export interface OrganizationListResponse {
  organizations: OrganizationUnit[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface OrganizationStats {
  total_count: number;
  by_type: Record<string, number>;
  by_status: Record<string, number>;
}

// GraphQL API响应类型定义
export interface GraphQLOrganizationResponse {
  code: string;
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
  byType: GraphQLStatsTypeItem[];
  byStatus: GraphQLStatsStatusItem[];
}

export interface APIResponse<T> {
  data: T;
}

export interface OrganizationListAPIResponse {
  organizations: GraphQLOrganizationResponse[];
}

export interface OrganizationStatsAPIResponse {
  organizationStats: GraphQLStatsResponse;
}