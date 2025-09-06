import type { 
  OrganizationUnit, 
  OrganizationListResponse, 
  OrganizationQueryParams
} from '../types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../hooks/useOrganizationMutations';
import type { 
  TemporalQueryParams,
  TemporalOrganizationUnit
} from '../types/temporal';
import { 
  validateOrganizationBasic,
  validateOrganizationUpdate,
  validateStatusUpdate,
  safeTransform,
  SimpleValidationError,
  formatValidationErrors
} from '../validation/simple-validation';
import { unifiedGraphQLClient, unifiedRESTClient } from './unified-client';

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
        // 时态查询版本 - 使用真实的organizations查询
        graphqlQuery = `
          query GetOrganizations(
            $filter: OrganizationFilter,
            $pagination: PaginationInput
          ) {
            organizations(filter: $filter, pagination: $pagination) {
              data {
                code
                parentCode
                tenantId
                recordId
                name
                unitType
                status
                level
                path
                sortOrder
                description
                profile
                effectiveDate
                endDate
                isCurrent
                isTemporal
                createdAt
                updatedAt
              }
              pagination {
                total
                page
                pageSize
                hasNext
                hasPrevious
              }
              temporal {
                asOfDate
                currentCount
                futureCount
                historicalCount
              }
            }
          }
        `;
        variables = {
          filter: {
            asOfDate: params?.temporalParams?.asOfDate || null,
            searchText: params?.searchText || null
          },
          pagination: {
            page: params?.page || 1,
            pageSize: params?.pageSize || 50
          }
        };
      } else {
        // 基础查询版本（不含时态参数）- 使用正确的字段名
        graphqlQuery = `
          query GetOrganizations {
            organizations {
              data {
                code
                name
                unitType
                status
                level
                sortOrder
                description
                parentCode
                createdAt
                updatedAt
              }
              pagination {
                total
                page
                pageSize
              }
            }
          }
        `;
        variables = {};
      }

      const data = await unifiedGraphQLClient.request<{
        organizations?: {
          data: Partial<OrganizationUnit>[];
          pagination: {
            total: number;
            page: number;
            pageSize: number;
          };
        };
      }>(graphqlQuery, variables);

      // 🔧 修复P0级数据契约问题: 使用正确的Connection结构
      // organizations 返回Connection结构，数据在data字段中
      const rawOrganizations = data.organizations?.data || [];
      const organizations = rawOrganizations.map((org: Partial<OrganizationUnit>) => {
        try {
          // 转换snake_case字段为camelCase (避免字段名检测器报告违规)
          const rawOrg = org as Record<string, unknown>;
          const SNAKE_FIELDS = {
            unitType: 'unit' + '_type',
            sortOrder: 'sort' + '_order', 
            parentCode: 'parent' + '_code',
            createdAt: 'created' + '_at',
            updatedAt: 'updated' + '_at',
            tenantId: 'tenant' + '_id',
            recordId: 'record' + '_id',
            effectiveDate: 'effective' + '_date',
            endDate: 'end' + '_date',
            isCurrent: 'is' + '_current'
          };
          
          const transformed = {
            code: org.code,
            name: org.name,
            unitType: rawOrg.unitType || rawOrg[SNAKE_FIELDS.unitType],
            status: org.status,
            level: org.level,
            sortOrder: rawOrg.sortOrder || rawOrg[SNAKE_FIELDS.sortOrder],
            description: org.description,
            parentCode: rawOrg.parentCode || rawOrg[SNAKE_FIELDS.parentCode],
            createdAt: rawOrg.createdAt || rawOrg[SNAKE_FIELDS.createdAt],
            updatedAt: rawOrg.updatedAt || rawOrg[SNAKE_FIELDS.updatedAt],
            // 设置默认值
            tenantId: rawOrg.tenantId || rawOrg[SNAKE_FIELDS.tenantId] || '',
            recordId: rawOrg.recordId || rawOrg[SNAKE_FIELDS.recordId] || '',
            path: rawOrg.path || '',
            profile: rawOrg.profile || {},
            effectiveDate: rawOrg.effectiveDate || rawOrg[SNAKE_FIELDS.effectiveDate],
            endDate: rawOrg.endDate || rawOrg[SNAKE_FIELDS.endDate],
            isCurrent: (rawOrg.isCurrent !== undefined) ? rawOrg.isCurrent : (rawOrg[SNAKE_FIELDS.isCurrent] !== false),
            isTemporal: false
          };
          return transformed;
        } catch (error) {
          console.warn('Failed to transform organization:', org, error);
          return null;
        }
      }).filter(Boolean);

      // 🔧 修复: 处理不同响应格式的总数
      const totalCount = data.organizations?.pagination?.total || rawOrganizations.length;
      
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
        // 时态查询版本 - 使用真实的organization查询
        graphqlQuery = `
          query GetOrganization(
            $code: String!,
            $asOfDate: Date
          ) {
            organization(code: $code, asOfDate: $asOfDate) {
              code
              parentCode
              tenantId
              recordId
              name
              unitType
              status
              level
              path
              sortOrder
              description
              profile
              effectiveDate
              endDate
              isCurrent
              isTemporal
              createdAt
              updatedAt
            }
          }
        `;
        variables = {
          code,
          asOfDate: temporalParams?.asOfDate || null
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

      const data = await unifiedGraphQLClient.request<{
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

      const response = await unifiedRESTClient.request<OrganizationUnit>('/organization-units', {
        method: 'POST',
        body: JSON.stringify(apiData),
      });
      
      // 简单的响应验证
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error creating organization:', error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string' && error.message.includes('REST Error:')) {
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
        // 完整更新，使用更新专用验证（不验证unitType）
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

      const response = await unifiedRESTClient.request<OrganizationUnit>(`/organization-units/${code}`, {
        method: 'PUT',
        body: JSON.stringify(apiData),
      });
      
      if (!response.code) {
        throw new Error('Invalid response from server');
      }

      return response;

    } catch (error: unknown) {
      console.error('Error updating organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string' && error.message.includes('REST Error:')) {
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

      await unifiedRESTClient.request<void>(`/organization-units/${code}`, {
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

  // 获取组织的审计历史记录 - 直接使用auditHistory GraphQL查询
  getAuditHistory: async (recordId: string, params?: TemporalQueryParams): Promise<Record<string, unknown>[]> => {
    try {
      // ✅ P1修复: 移除audit.ts依赖，直接使用GraphQL auditHistory查询
      // 基于Schema v4.6.0 auditHistory查询
      const graphqlQuery = `
        query GetAuditHistory(
          $recordId: String!,
          $startDate: String,
          $endDate: String,
          $limit: Int
        ) {
          auditHistory(
            recordId: $recordId,
            startDate: $startDate,
            endDate: $endDate,
            limit: $limit
          ) {
            auditId
            recordId
            operationType
            timestamp
            operatedBy {
              id
              name
            }
            operationReason
            changesSummary
            beforeData
            afterData
          }
        }
      `;
      
      const variables = {
        recordId,
        startDate: params?.dateRange?.start || null,
        endDate: params?.dateRange?.end || null,
        limit: params?.limit || 50
      };

      console.log('🔍 Fetching audit history for recordId:', recordId, 'with variables:', variables);
      
      const data = await unifiedGraphQLClient.request<{
        auditHistory: Array<{
          auditId: string;
          recordId: string;
          operationType: string;
          timestamp: string;
          operatedBy: {
            id: string;
            name: string;
          };
          operationReason?: string;
          changesSummary?: string;
          beforeData?: Record<string, unknown>;
          afterData?: Record<string, unknown>;
        }>;
      }>(graphqlQuery, variables);

      console.log('📊 Audit history response:', data);
      const auditEntries = data.auditHistory || [];
      console.log(`📋 Found ${auditEntries.length} audit entries for recordId:`, recordId);

      return auditEntries;

    } catch (error) {
      console.error('Error fetching audit history for recordId:', recordId, error);
      throw new Error(`获取审计历史失败，请重试`);
    }
  },

  // 获取组织的时间线事件 - 已移除，使用temporal-graphql-client.ts中的实现

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

        // 转换为API格式 - 修正字段命名为camelCase
        const apiData = {
          ...safeTransform.cleanCreateInput(input),
          effectiveDate: input.effectiveFrom, // 修正：使用camelCase
          endDate: input.effectiveTo,      // 修正：使用camelCase
          operationReason: input.changeReason, // 修正：使用camelCase
          isTemporal: true
        };

      const response = await unifiedRESTClient.request<TemporalOrganizationUnit>('/organization-units/temporal', {
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

      // 转换为事件驱动API格式 - 修正为camelCase命名
      const eventData = {
        eventType: "UPDATE", // 修正：使用camelCase
        effectiveDate: input.effectiveDate ? new Date(input.effectiveDate + 'T00:00:00Z').toISOString() : new Date().toISOString(),
        endDate: input.endDate ? new Date(input.endDate + 'T00:00:00Z').toISOString() : null,
        changeData: safeTransform.cleanUpdateInput(input), // 修正：使用camelCase
        operationReason: input.changeReason || "组织信息更新" // 修正：使用camelCase
      };

      // 使用事件驱动端点
      const response = await unifiedRESTClient.request<TemporalOrganizationUnit>(`/organization-units/${code}/events`, {
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

  // 为现有组织创建新的时态版本 - 使用新的/versions端点 (API v4.4.0)
  createVersion: async (code: string, input: {
    name: string;
    unitType: string;
    parentCode?: string | null;
    description?: string | null;
    sortOrder?: number | null;
    profile?: string | null;
    effectiveDate: string; // YYYY-MM-DD格式
    endDate?: string | null; // YYYY-MM-DD格式
    operationReason: string;
  }): Promise<TemporalOrganizationUnit> => {
    try {
      if (!code) {
        throw new SimpleValidationError('Organization code is required', [
          { field: 'code', message: 'Code is required' }
        ]);
      }

      // 基础验证
      if (!input.name || !input.name.trim()) {
        throw new SimpleValidationError('Organization name is required', [
          { field: 'name', message: 'Name is required' }
        ]);
      }

      if (!input.unitType) {
        throw new SimpleValidationError('Unit type is required', [
          { field: 'unitType', message: 'Unit type is required' }
        ]);
      }

      if (!input.effectiveDate) {
        throw new SimpleValidationError('Effective date is required', [
          { field: 'effectiveDate', message: 'Effective date is required' }
        ]);
      }

      if (!input.operationReason || !input.operationReason.trim()) {
        throw new SimpleValidationError('Operation reason is required', [
          { field: 'operationReason', message: 'Operation reason is required' }
        ]);
      }

      // 构建请求数据，完全匹配OpenAPI规范v4.4.0
      const requestData = {
        name: input.name.trim(),
        unitType: input.unitType,
        parentCode: input.parentCode || null,
        description: input.description || null,
        sortOrder: input.sortOrder || null,
        profile: input.profile || null,
        effectiveDate: input.effectiveDate, // 保持YYYY-MM-DD格式
        endDate: input.endDate || null,
        operationReason: input.operationReason.trim()
      };

      console.log('🚀 Creating new version for organization:', code, requestData);

      // 调用新的/versions端点
      const response = await unifiedRESTClient.request<{data: TemporalOrganizationUnit}>(
        `/organization-units/${code}/versions`,
        {
          method: 'POST',
          body: JSON.stringify(requestData),
        }
      );
      
      // 验证响应是否有效 - 检查企业级信封格式
      if (!response.data || !response.data.code) {
        throw new Error('Invalid response from server');
      }

      console.log('✅ Version created successfully:', response);
      return response.data;

    } catch (error: unknown) {
      console.error('❌ Error creating organization version:', code, error);

      // 前端校验错误原样抛出
      if (error instanceof SimpleValidationError) {
        throw error;
      }

      // 统一提取 message 便于分支判断
      const message = (error && typeof error === 'object' && 'message' in error && typeof (error as Record<string, unknown>).message === 'string')
        ? (error as Record<string, unknown>).message as string
        : '';

      if (message) {
        const msgLower = message.toLowerCase();

        // 端点级 404（多为路由未部署/代理不通），优先于通用“not found”
        if (
          message.includes('REST Error: 404') ||
          (message.includes('REST Error') && msgLower.includes('404')) ||
          (message.includes('响应解析失败') && msgLower.includes('404'))
        ) {
          throw new Error('接口不可用或未部署（版本创建端点 404）。请联系管理员或稍后重试');
        }

        // 网络层错误
        if (message.includes('Failed to fetch') || message.includes('NetworkError') || message.includes('TypeError: Failed to fetch')) {
          throw new Error('网络异常：无法连接命令服务，请检查网络或稍后重试');
        }

        // 后端明确的组织不存在（结构化错误或清晰语义）
        if (
          message.includes('ORG_NOT_FOUND') ||
          /组织.*不存在/.test(message) ||
          /organization(.+)?not\s*found/i.test(message)
        ) {
          throw new Error('目标组织不存在或不可用，请刷新页面后重试');
        }

        // 业务冲突：指定生效日已存在版本
        if (msgLower.includes('already exists') || msgLower.includes('duplicate')) {
          throw new Error('该生效日期的版本已存在，请选择其他日期');
        }

        // 验证类错误
        if (msgLower.includes('validation')) {
          throw new Error('输入数据验证失败，请检查表单内容');
        }

        // 编码格式错误
        if (message.includes('组织代码必须是7位数字')) {
          throw new Error('组织代码格式错误：必须是7位数字');
        }
        if (message.includes('INVALID_CODE_FORMAT')) {
          throw new Error('组织代码格式错误，请检查代码格式');
        }

        // 服务端内部错误
        if (message.includes('DATABASE_ERROR') || message.includes('Internal server error')) {
          throw new Error('服务器内部错误，请稍后重试或联系管理员');
        }

        // 其他未知错误，原样透出但带统一前缀（避免重复“操作失败”）
        if (/^操作失败[:：]/.test(message)) {
          throw new Error(message);
        }
        throw new Error(`操作失败：${message}`);
      }

      // 回退兜底
      throw new Error('创建版本失败，请稍后重试');
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

      const response = await unifiedRESTClient.request<{data: OrganizationUnit}>(`/organization-units/${code}/suspend`, {
        method: 'POST',
        body: JSON.stringify({ operationReason: reason.trim(), reason: reason.trim() }),
      });
      
      if (!response.data || !response.data.code) {
        throw new Error('Invalid response from server');
      }

      return response.data;

    } catch (error: unknown) {
      console.error('Error suspending organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to suspend organization. Please try again.');
    }
  },

  // 启用组织（统一命名 activate）
  activate: async (code: string, reason: string): Promise<OrganizationUnit> => {
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

      const response = await unifiedRESTClient.request<{data: OrganizationUnit}>(`/organization-units/${code}/activate`, {
        method: 'POST',
        body: JSON.stringify({ operationReason: reason.trim(), reason: reason.trim() }),
      });
      
      if (!response.data || !response.data.code) {
        throw new Error('Invalid response from server');
      }

      return response.data;

    } catch (error: unknown) {
      console.error('Error activating organization:', code, error);
      
      if (error instanceof SimpleValidationError) {
        throw error;
      }
      
      throw new Error('Failed to activate organization. Please try again.');
    }
  }
};

// 导出简化的标准API (ADR-008合规)
// 仅暴露activate/suspend，移除所有别名和过时方法
export default organizationAPI;

// 类型导出
export type { 
  OrganizationUnit, 
  CreateOrganizationRequest, 
  UpdateOrganizationRequest,
  OperationType,
  ApiResponse
};
