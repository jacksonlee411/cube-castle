import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationStats,
  GraphQLResponse,
  OrganizationQueryParams
} from '../types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../hooks/useOrganizationMutations';
import type { 
  TemporalQueryParams,
  TemporalOrganizationUnit,
  TimelineEvent
} from '../types/temporal';
import { 
  validateOrganizationBasic,
  validateOrganizationUpdate,
  validateStatusUpdate,
  safeTransform,
  SimpleValidationError,
  formatValidationErrors
} from '../validation/simple-validation';
import { authManager } from './auth';

// GraphQL统计响应接口
interface GraphQLStatsResponse {
  totalCount: number;
  byType?: Array<{ unitType: string; count: number }>;
  byStatus?: Array<{ status: string; count: number }>;
  byLevel?: Array<{ level: number; count: number }>;
}

// GraphQL客户端 - 使用正确的端口8090
const GRAPHQL_ENDPOINT = 'http://localhost:8090/graphql';

const graphqlClient = {
  async request<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
    // 获取OAuth访问令牌
    const accessToken = await authManager.getAccessToken();
    
    const response = await fetch(GRAPHQL_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`,
      },
      body: JSON.stringify({
        query,
        variables
      }),
    });

    if (!response.ok) {
      throw new Error(`GraphQL Error: ${response.status} ${response.statusText}`);
    }

    const result = await response.json() as GraphQLResponse<T>;
    
    if (result.errors && result.errors.length > 0) {
      throw new Error(`GraphQL Error: ${result.errors[0].message}`);
    }

    if (!result.data) {
      throw new Error('No data returned from GraphQL');
    }

    return result.data;
  }
};

// REST API客户端 - 使用命令服务端口9090
const REST_ENDPOINT = 'http://localhost:9090/api/v1';

const restClient = {
  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    // 获取OAuth访问令牌
    const accessToken = await authManager.getAccessToken();
    
    const url = `${REST_ENDPOINT}${endpoint}`;
    
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`,
        'X-Tenant-ID': '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        ...options.headers,
      },
      ...options,
    });

    if (!response.ok) {
      throw new Error(`REST Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }
};

// 扩展查询参数以支持时态查询
interface ExtendedOrganizationQueryParams extends OrganizationQueryParams {
  searchText?: string;
  pageSize?: number;
  temporalParams?: TemporalQueryParams;
}

export const organizationAPI = {
  // 获取组织单元列表 - 使用GraphQL (修复getByCode问题)
  getAll: async (params?: ExtendedOrganizationQueryParams): Promise<OrganizationListResponse> => {
    try {
      // 轻量级参数验证
      if (params) {
        // 简化的参数验证，依赖后端详细验证
        if (params.page && params.page < 1) {
          throw new SimpleValidationError('页码必须大于0', [
            { field: 'page', message: '页码必须大于0' }
          ]);
        }
        if (params.pageSize && (params.pageSize < 1 || params.pageSize > 100)) {
          throw new SimpleValidationError('页面大小必须在1-100之间', [
            { field: 'pageSize', message: '页面大小必须在1-100之间' }
          ]);
        }
      }

      // 构建GraphQL查询和变量 (基础版本，不含时态参数)
      const useTemporalQuery = params?.temporalParams && Object.keys(params.temporalParams).length > 0;
      
      let graphqlQuery, variables;
      
      if (useTemporalQuery) {
        // 时态查询版本
        graphqlQuery = `
          query GetOrganizations(
            $first: Int, 
            $offset: Int, 
            $searchText: String,
            $asOfDate: String,
            $effectiveFrom: String,
            $effectiveTo: String,
            $temporalMode: String
          ) {
            organizations(
              first: $first, 
              offset: $offset, 
              searchText: $searchText
            ) {
              code
              name
              unitType
              status
              level
              path
              sortOrder
              description
              parentCode
              createdAt
              updatedAt
              effectiveDate
              endDate
              isTemporal
            }
            organizationStats {
              totalCount
            }
          }
        `;
        variables = {
          first: params?.pageSize || 50,
          offset: ((params?.page || 1) - 1) * (params?.pageSize || 50),
          searchText: params?.searchText || null,
          asOfDate: params?.temporalParams?.asOfDate || null,
          effectiveFrom: params?.temporalParams?.dateRange?.start || null,
          effectiveTo: params?.temporalParams?.dateRange?.end || null,
          temporalMode: params?.temporalParams?.mode || 'current'
        };
      } else {
        // 基础查询版本（不含时态参数）
        graphqlQuery = `
          query GetOrganizations(
            $first: Int, 
            $offset: Int, 
            $searchText: String
          ) {
            organizations(
              first: $first, 
              offset: $offset, 
              searchText: $searchText
            ) {
              code
              name
              unitType
              status
              level
              path
              sortOrder
              description
              parentCode
              createdAt
              updatedAt
            }
            organizationStats {
              totalCount
            }
          }
        `;
        variables = {
          first: params?.pageSize || 50,
          offset: ((params?.page || 1) - 1) * (params?.pageSize || 50),
          searchText: params?.searchText || null
        };
      }

      const data = await graphqlClient.request<{
        organizations: {
          data: Partial<OrganizationUnit>[];
          totalCount: number;
          hasMore: boolean;
        };
        organizationStats?: {
          totalCount: number;
        };
      }>(graphqlQuery, variables);

      // 简化的数据转换 - 使用正确的响应结构
      const organizations = data.organizations.data.map((org: Partial<OrganizationUnit>) => {
        try {
          return safeTransform.graphqlToOrganization ? 
            safeTransform.graphqlToOrganization(org) : 
            org; // 直接返回原始数据，依赖后端格式
        } catch (error) {
          console.warn('Failed to transform organization:', org, error);
          return null;
        }
      }).filter(Boolean);

      // 🔧 修复: 使用正确的总数来源
      const totalCount = data.organizations.totalCount;
      
      return {
        organizations: organizations.filter((org): org is OrganizationUnit => org !== null),
        totalCount: totalCount,
        page: params?.page || 1,
        pageSize: organizations.length,
        totalPages: Math.ceil(totalCount / (params?.pageSize || 50))
      };

    } catch (error) {
      console.error('Error fetching organizations:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to fetch organizations. Please try again.');
    }
  },

  // 根据代码获取单个组织 - ✅ 修复协议违反，统一使用GraphQL (支持时态查询)
  getByCode: async (code: string, temporalParams?: TemporalQueryParams): Promise<OrganizationUnit> => {
    try {
      if (!code || typeof code !== 'string') {
        throw new SimpleValidationError('Invalid organization code', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // ✅ 使用GraphQL查询，遵循"查询统一用GraphQL"原则 (基础版本)
      const useTemporalQuery = temporalParams && Object.keys(temporalParams).length > 0;
      
      let graphqlQuery, variables;
      
      if (useTemporalQuery) {
        // 时态查询版本
        graphqlQuery = `
          query GetOrganization(
            $code: String!, 
            $asOfDate: String,
            $temporalMode: String
          ) {
            organization(
              code: $code, 
              asOfDate: $asOfDate,
              temporalMode: $temporalMode
            ) {
              code
              recordId
              name
              unitType
              status
              level
              path
              sortOrder
              description
              parentCode
              createdAt
              updatedAt
              effectiveDate
              endDate
              isTemporal
            }
          }
        `;
        variables = {
          code,
          asOfDate: temporalParams?.asOfDate || null,
          temporalMode: temporalParams?.mode || 'current'
        };
      } else {
        // 基础查询版本（不含时态参数）
        graphqlQuery = `
          query GetOrganization($code: String!) {
            organization(code: $code) {
              code
              recordId
              name
              unitType
              status
              level
              path
              sortOrder
              description
              parentCode
              createdAt
              updatedAt
            }
          }
        `;
        variables = { code };
      }

      const data = await graphqlClient.request<{
        organization: Partial<OrganizationUnit>;
      }>(graphqlQuery, variables);

      const organization = data.organization;
      if (!organization) {
        throw new Error(`组织 ${code} 不存在`);
      }

      // 简单数据转换，依赖后端格式
      if (safeTransform.graphqlToOrganization) {
        const transformed = safeTransform.graphqlToOrganization(organization) as unknown as OrganizationUnit;
        // 确保转换后的对象包含所有必需字段
        if (transformed && typeof transformed === 'object' && 'code' in transformed && 'name' in transformed) {
          return transformed as OrganizationUnit;
        }
      }
      
      return organization as OrganizationUnit;

    } catch (error: unknown) {
      console.error('Error fetching organization by code:', code, error);
      
      if (error instanceof Error && 'response' in error && 
          (error as Error & { response?: { status?: number } }).response?.status === 404) {
        throw new Error(`组织 ${code} 不存在`);
      }
      
      throw new Error(`获取组织 ${code} 失败，请重试`);
    }
  },

  // 获取组织统计信息 - 使用GraphQL
  getStats: async (): Promise<OrganizationStats> => {
    try {
      const graphqlQuery = `
        query GetOrganizationStats {
          organizations(first: 1000) {
            totalCount
            data {
              unitType
              status
            }
          }
        }
      `;

      const data = await graphqlClient.request<{
        organizations: {
          totalCount: number;
          data: Array<{ unitType: string; status: string }>;
        };
      }>(graphqlQuery);

      const organizations = data.organizations;
      if (!organizations) {
        throw new Error('No statistics data returned');
      }

      // 计算统计信息
      const byType: Record<string, number> = {};
      const byStatus: Record<string, number> = {};
      
      organizations.data.forEach(org => {
        byType[org.unitType] = (byType[org.unitType] || 0) + 1;
        byStatus[org.status] = (byStatus[org.status] || 0) + 1;
      });

      return {
        totalCount: organizations.totalCount,
        byType,
        byStatus
      };

    } catch (error) {
      console.error('Error fetching organization stats:', error);
      throw new Error('Failed to fetch organization statistics. Please try again.');
    }
  },

  // 创建组织 - 依赖后端统一验证
  create: async (input: CreateOrganizationInput): Promise<OrganizationUnit> => {
    try {
      // 基础前端验证 (用户体验)
      const validationResult = validateOrganizationBasic(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors), 
          validationResult.errors
        );
      }

      // 转换为API格式
      const apiData = safeTransform.cleanCreateInput(input);

      const response = await restClient.request<OrganizationUnit>('/organization-units', {
        method: 'POST',
        body: JSON.stringify(apiData),
      });
      
      // 简单的响应验证
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: any) {
      console.error('Error creating organization:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      if (error?.message?.includes('REST Error:')) {
        // 服务器端验证错误
        const serverMessage = error.message;
        throw new Error(serverMessage || 'Failed to create organization');
      }
      
      throw new Error('Failed to create organization. Please try again.');
    }
  },

  // 更新组织 - 智能验证，根据更新内容选择合适的验证策略
  update: async (code: string, input: UpdateOrganizationInput): Promise<OrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // 智能验证策略：根据更新的字段选择验证方法
      let validationResult;
      
      const inputKeys = Object.keys(input);
      const isStatusOnlyUpdate = inputKeys.length === 1 && inputKeys[0] === 'status';
      
      if (isStatusOnlyUpdate) {
        // 仅状态更新，使用状态专用验证
        console.log('[API] Status-only update detected, using validateStatusUpdate');
        validationResult = validateStatusUpdate(input);
      } else {
        // 完整更新，使用更新专用验证（不验证unit_type）
        console.log('[API] Full update detected, using validateOrganizationUpdate');
        validationResult = validateOrganizationUpdate(input);
      }
      
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors),
          validationResult.errors
        );
      }

      // 转换为API格式
      const apiData = safeTransform.cleanUpdateInput(input);

      const response = await restClient.request<OrganizationUnit>(`/organization-units/${code}`, {
        method: 'PUT',
        body: JSON.stringify(apiData),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: any) {
      console.error('Error updating organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      if (error?.message?.includes('REST Error:')) {
        const serverMessage = error.message;
        throw new Error(serverMessage || 'Failed to update organization');
      }
      
      throw new Error('Failed to update organization. Please try again.');
    }
  },

  // 删除组织
  delete: async (code: string): Promise<void> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      await restClient.request<void>(`/organization-units/${code}`, {
        method: 'DELETE'
      });

    } catch (error) {
      console.error('Error deleting organization:', code, error);
      
      if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string' && error.message.includes('REST Error:')) {
        const serverMessage = error.message;
        throw new Error(serverMessage || 'Failed to delete organization');
      }
      
      throw new Error('Failed to delete organization. Please try again.');
    }
  },

  // ====== 组织详情API方法 ======

  // 获取组织的历史版本
  getHistory: async (code: string, params?: TemporalQueryParams): Promise<TemporalOrganizationUnit[]> => {
    try {
      if (!code || typeof code !== 'string') {
        throw new SimpleValidationError('Invalid organization code', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      const graphqlQuery = `
        query GetOrganizationHistory(
          $code: String!,
          $dateFrom: String,
          $dateTo: String,
          $limit: Int
        ) {
          organizationHistory(
            code: $code,
            dateFrom: $dateFrom,
            dateTo: $dateTo,
            limit: $limit
          ) {
            code
            name
            unitType
            status
            level
            path
            sortOrder
            description
            parentCode
            effectiveFrom
            effectiveTo
            isTemporal
            changeReason
            changedBy
            createdAt
            updatedAt
          }
        }
      `;

      const variables = {
        code,
        dateFrom: params?.dateRange?.start || null,
        dateTo: params?.dateRange?.end || null,
        limit: params?.limit || 50
      };

      const data = await graphqlClient.request<{
        organizationHistory: TemporalOrganizationUnit[];
      }>(graphqlQuery, variables);

      return data.organizationHistory || [];

    } catch (error) {
      console.error('Error fetching organization history:', code, error);
      throw new Error(`获取组织 ${code} 历史记录失败，请重试`);
    }
  },

  // 获取组织的时间线事件
  getTimeline: async (code: string, params?: TemporalQueryParams): Promise<TimelineEvent[]> => {
    try {
      if (!code || typeof code !== 'string') {
        throw new SimpleValidationError('Invalid organization code', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      const graphqlQuery = `
        query GetOrganizationTimeline(
          $code: String!,
          $dateFrom: String,
          $dateTo: String,
          $eventTypes: [String],
          $limit: Int
        ) {
          organizationTimeline(
            code: $code,
            dateFrom: $dateFrom,
            dateTo: $dateTo,
            eventTypes: $eventTypes,
            limit: $limit
          ) {
            id
            organizationCode
            eventType
            eventDate
            effectiveDate
            status
            title
            description
            metadata
            previousValue
            newValue
            triggeredBy
            approvedBy
            createdAt
          }
        }
      `;

      const variables = {
        code,
        dateFrom: params?.dateRange?.start || null,
        dateTo: params?.dateRange?.end || null,
        eventTypes: params?.eventTypes || null,
        limit: params?.limit || 100
      };

      const data = await graphqlClient.request<{
        organizationTimeline: TimelineEvent[];
      }>(graphqlQuery, variables);

      return data.organizationTimeline || [];

    } catch (error) {
      console.error('Error fetching organization timeline:', code, error);
      throw new Error(`获取组织 ${code} 时间线失败，请重试`);
    }
  },

  // 创建时态组织记录
  createTemporal: async (input: CreateOrganizationInput & { 
    effectiveFrom: string; 
    effectiveTo?: string; 
    changeReason?: string;
  }): Promise<TemporalOrganizationUnit> => {
    try {
      // 基础前端验证
      const validationResult = validateOrganizationBasic(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors), 
          validationResult.errors
        );
      }

      // 转换为API格式
      const apiData = {
        ...safeTransform.cleanCreateInput(input),
        effective_date: input.effectiveFrom, // 修正：字段名映射
        end_date: input.effectiveTo,      // 修正：字段名映射
        change_reason: input.changeReason,
        is_temporal: true
      };

      const response = await restClient.request<TemporalOrganizationUnit>('/organization-units/temporal', {
        method: 'POST',
        body: JSON.stringify(apiData),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error creating temporal organization:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to create temporal organization. Please try again.');
    }
  },

  // 更新时态组织记录 - 使用事件驱动API
  updateTemporal: async (code: string, input: UpdateOrganizationInput & {
    effectiveDate?: string;
    endDate?: string;
    changeReason?: string;
  }): Promise<TemporalOrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // 智能验证策略
      const validationResult = validateOrganizationUpdate(input);
      if (!validationResult.isValid) {
        throw new SimpleValidationError(
          '输入验证失败：' + formatValidationErrors(validationResult.errors),
          validationResult.errors
        );
      }

      // 转换为事件驱动API格式 - 修复日期格式
      const eventData = {
        event_type: "UPDATE",
        effective_date: input.effectiveDate ? new Date(input.effectiveDate + 'T00:00:00Z').toISOString() : new Date().toISOString(),
        end_date: input.endDate ? new Date(input.endDate + 'T00:00:00Z').toISOString() : null,
        change_data: safeTransform.cleanUpdateInput(input),
        change_reason: input.changeReason || "组织信息更新"
      };

      // 使用事件驱动端点
      const response = await restClient.request<TemporalOrganizationUnit>(`/organization-units/${code}/events`, {
        method: 'POST',
        body: JSON.stringify(eventData),
      });
      
      // 验证响应是否有效 - 修正：检查核心字段而非event_id
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error updating temporal organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to update temporal organization. Please try again.');
    }
  },

  // === 新增：操作驱动状态管理API ===

  // 停用组织
  suspend: async (code: string, reason: string): Promise<OrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      if (!reason || !reason.trim()) {
        throw new SimpleValidationError('Suspend reason is required', [
          { field: 'reason', message: 'Reason is required' }
        ]);
      }

      const response = await restClient.request<OrganizationUnit>(`/organization-units/${code}/suspend`, {
        method: 'POST',
        body: JSON.stringify({ reason: reason.trim() }),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: any) {
      console.error('Error suspending organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to suspend organization. Please try again.');
    }
  },

  // 重新启用组织
  reactivate: async (code: string, reason: string): Promise<OrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      if (!reason || !reason.trim()) {
        throw new SimpleValidationError('Reactivate reason is required', [
          { field: 'reason', message: 'Reason is required' }
        ]);
      }

      const response = await restClient.request<OrganizationUnit>(`/organization-units/${code}/reactivate`, {
        method: 'POST',
        body: JSON.stringify({ reason: reason.trim() }),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: any) {
      console.error('Error reactivating organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to reactivate organization. Please try again.');
    }
  }
};

// 导出简化的API
export default organizationAPI;