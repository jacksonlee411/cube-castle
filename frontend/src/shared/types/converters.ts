/**
 * TypeScript类型转换工具
 * 用于在前后端API之间进行一致的数据类型转换
 * 符合CLAUDE.md API一致性规范
 */

import type { OrganizationUnit } from './organization';
import type { TemporalOrganizationUnit } from './temporal';

// ============================================================================
// 前端 ↔ GraphQL 转换器 (camelCase ↔ camelCase)
// ============================================================================

/**
 * GraphQL响应数据接口 (实际返回字段)
 */
export interface GraphQLOrganizationData {
  // 基本字段
  code?: string;
  parentCode?: string;
  tenantId?: string;
  name?: string;
  unitType?: string;
  status?: string;
  level?: number;
  
  // 路径和排序
  codePath?: string;  // 新字段：层级路径
  namePath?: string;  // 新字段：名称路径
  path?: string;      // 兼容旧字段
  sortOrder?: number;
  
  // 描述和配置
  description?: string;
  profile?: string;
  
  // 时态字段
  effectiveDate?: string;
  endDate?: string;
  isCurrent?: boolean;
  isFuture?: boolean;
  
  // 审计字段
  createdAt?: string;
  updatedAt?: string;
  operationType?: string;
  operationReason?: string;
  recordId?: string;
  deletedAt?: string | null;
  deletedBy?: string | null;
  deletionReason?: string | null;
  suspendedAt?: string | null;
  suspendedBy?: string | null;
  suspensionReason?: string | null;
  
  // 兼容字段
  isTemporal?: boolean;
  version?: number;
  changeReason?: string;
  validFrom?: string;
  validTo?: string;
}

/**
 * 将GraphQL响应转换为前端OrganizationUnit类型
 * GraphQL已经使用camelCase，直接映射即可
 */
export function convertGraphQLToOrganizationUnit(
  data: GraphQLOrganizationData
): OrganizationUnit {
  return {
    code: data.code || '',
    recordId: data.tenantId,
    parentCode: data.parentCode || '0', // 根据API规范，parentCode必填，默认使用根组织"0"
    name: data.name || '',
    unitType: (data.unitType as OrganizationUnit['unitType']) || 'DEPARTMENT',
    status: (data.status as OrganizationUnit['status']) || 'ACTIVE',
    level: data.level || 1,
    path: data.codePath ?? data.path ?? undefined,  // 使用codePath字段
    sortOrder: data.sortOrder || 0,
    description: data.description || '',
    createdAt: data.createdAt || '',
    updatedAt: data.updatedAt || '',
    effectiveDate: data.effectiveDate,
    endDate: data.endDate,
    isTemporal: data.isTemporal,
    version: data.version,
    changeReason: data.changeReason || data.operationReason, // 支持operationReason字段
    isCurrent: data.isCurrent,
    deletedAt: data.deletedAt ?? null,
    deletedBy: data.deletedBy ?? null,
    deletionReason: data.deletionReason ?? null,
    suspendedAt: data.suspendedAt ?? null,
    suspendedBy: data.suspendedBy ?? null,
    suspensionReason: data.suspensionReason ?? null,
  };
}

/**
 * 将GraphQL响应转换为时态组织单元类型
 */
export function convertGraphQLToTemporalOrganizationUnit(
  data: GraphQLOrganizationData
): TemporalOrganizationUnit {
  return {
    code: data.code || '',
    parentCode: data.parentCode || '',
    name: data.name || '',
    unitType: (data.unitType as TemporalOrganizationUnit['unitType']) || 'DEPARTMENT',
    status: (data.status as TemporalOrganizationUnit['status']) || 'ACTIVE',
    level: data.level || 1,
    path: data.path ?? undefined,
    sortOrder: data.sortOrder || 0,
    description: data.description || '',
    createdAt: data.createdAt || '',
    updatedAt: data.updatedAt || '',
    effectiveDate: data.effectiveDate || '',
    endDate: data.endDate,
    isCurrent: data.isCurrent ?? true,
    changeReason: data.changeReason,
    approvedBy: undefined, // GraphQL中暂无此字段
    approvedAt: undefined  // GraphQL中暂无此字段
  };
}

// ============================================================================
// 前端 ↔ REST API 转换器 (camelCase ↔ camelCase)
// ============================================================================

// 使用统一的OrganizationRequest接口，无需重复定义
import type { OrganizationRequest } from './organization';

/**
 * REST API请求数据接口
 * 使用统一的OrganizationRequest类型
 */
export type RESTOrganizationRequest = OrganizationRequest;

/**
 * 将前端创建输入转换为REST API请求格式
 * 由于都使用camelCase，主要进行字段验证和清理
 */
export function convertCreateInputToREST(
  input: Partial<OrganizationUnit>
): OrganizationRequest {
  const request: OrganizationRequest = {
    name: input.name || '',
    unitType: input.unitType || 'DEPARTMENT',
    description: input.description || '',
    level: input.level || 1,
    sortOrder: input.sortOrder || 0,
    status: input.status || 'ACTIVE',
  };

  // 只添加有值的可选字段
  if (input.parentCode) {
    request.parentCode = input.parentCode;
  }
  if (input.effectiveDate) {
    request.effectiveDate = input.effectiveDate;
  }
  if (input.endDate) {
    request.endDate = input.endDate;
  }
  if (input.changeReason) {
    request.changeReason = input.changeReason;
  }

  return request;
}

/**
 * 将前端更新输入转换为REST API请求格式
 */
export function convertUpdateInputToREST(
  input: Partial<OrganizationUnit>
): Partial<OrganizationRequest> {
  const request: Partial<OrganizationRequest> = {};

  // 只包含需要更新的字段
  if (input.name !== undefined) request.name = input.name;
  if (input.unitType !== undefined) request.unitType = input.unitType;
  if (input.parentCode !== undefined) request.parentCode = input.parentCode;
  if (input.description !== undefined) request.description = input.description;
  if (input.level !== undefined) request.level = input.level;
  if (input.sortOrder !== undefined) request.sortOrder = input.sortOrder;
  if (input.status !== undefined) request.status = input.status;
  if (input.effectiveDate !== undefined) request.effectiveDate = input.effectiveDate;
  if (input.endDate !== undefined) request.endDate = input.endDate;
  if (input.changeReason !== undefined) request.changeReason = input.changeReason;

  return request;
}

// ============================================================================
// 数据验证和清理工具
// ============================================================================

/**
 * DEPRECATED: 使用 shared/api/type-guards.ts 的 validateOrganizationUnit
 * 该函数已被 Zod 验证系统替代，提供更强的类型安全和验证
 */
// import { validateOrganizationUnit } from '../api/type-guards';
// 用于验证和转换未知数据为组织单元类型

/**
 * DEPRECATED: 使用 shared/api/type-guards.ts 的批量验证函数
 * 该函数已被 Zod 验证系统替代
 */
// import { validateGraphQLOrganizationList } from '../api/type-guards';
// 用于批量验证组织单元数据

// ============================================================================
// 类型同步检查工具
// ============================================================================

/**
 * 检查前端类型与API响应的一致性
 * 用于开发期间的类型同步验证
 */
export function checkTypeConsistency(
  apiResponse: unknown,
  expectedFields: string[]
): {
  isConsistent: boolean;
  missingFields: string[];
  extraFields: string[];
  report: string;
} {
  if (!apiResponse || typeof apiResponse !== 'object') {
    return {
      isConsistent: false,
      missingFields: expectedFields,
      extraFields: [],
      report: 'API响应不是有效对象'
    };
  }

  const obj = apiResponse as Record<string, unknown>;
  const actualFields = Object.keys(obj);
  
  const missingFields = expectedFields.filter(field => !(field in obj));
  const extraFields = actualFields.filter(field => !expectedFields.includes(field));
  
  const isConsistent = missingFields.length === 0 && extraFields.length === 0;
  
  const report = [
    `类型一致性检查报告:`,
    `- 预期字段: ${expectedFields.length}个`,
    `- 实际字段: ${actualFields.length}个`,
    `- 缺失字段: ${missingFields.length}个 ${missingFields.length > 0 ? `[${missingFields.join(', ')}]` : ''}`,
    `- 额外字段: ${extraFields.length}个 ${extraFields.length > 0 ? `[${extraFields.join(', ')}]` : ''}`,
    `- 一致性状态: ${isConsistent ? '✅ 通过' : '❌ 不一致'}`
  ].join('\n');

  return {
    isConsistent,
    missingFields,
    extraFields,
    report
  };
}

/**
 * 组织单元字段列表 - 用于类型同步检查
 */
export const ORGANIZATION_UNIT_FIELDS: string[] = [
  'code', 'recordId', 'parentCode', 'name', 'unitType', 'status',
  'level', 'path', 'sortOrder', 'description', 'createdAt', 'updatedAt',
  'effectiveDate', 'endDate', 'isTemporal', 'version', 'changeReason', 'isCurrent'
];

/**
 * 时态组织单元字段列表 - 用于类型同步检查
 */
export const TEMPORAL_ORGANIZATION_UNIT_FIELDS: string[] = [
  'code', 'parentCode', 'name', 'unitType', 'status',
  'level', 'path', 'sortOrder', 'description', 'createdAt', 'updatedAt',
  'effectiveDate', 'endDate', 'isCurrent', 'changeReason', 'approvedBy', 'approvedAt'
];

// ============================================================================
// 开发工具函数
// ============================================================================

/**
 * 生成类型定义代码 - 基于API响应自动生成TypeScript接口
 * 用于开发期间快速同步类型定义
 */
export function generateTypeDefinition(
  apiResponse: unknown,
  interfaceName: string
): string {
  if (!apiResponse || typeof apiResponse !== 'object') {
    return `// 无法生成类型定义: API响应无效`;
  }

  const obj = apiResponse as Record<string, unknown>;
  const fields: string[] = [];

  Object.entries(obj).forEach(([key, value]) => {
    const type = Array.isArray(value) 
      ? 'unknown[]'
      : value === null || value === undefined
      ? 'unknown'
      : typeof value;
    
    const optional = value === null || value === undefined ? '?' : '';
    fields.push(`  ${key}${optional}: ${type};`);
  });

  return [
    `export interface ${interfaceName} {`,
    ...fields,
    `}`
  ].join('\n');
}

/**
 * 输出类型同步报告到控制台
 * 用于开发期间监控类型一致性
 */
export function logTypeSyncReport(
  context: string,
  apiResponse: unknown,
  expectedFields: string[]
): void {
  const result = checkTypeConsistency(apiResponse, expectedFields);
  
  console.group(`[TypeSync] ${context}`);
  console.log(result.report);
  
  if (!result.isConsistent) {
    console.warn('类型不一致可能导致运行时错误，建议更新类型定义');
    
    if (result.missingFields.length > 0) {
      console.error('缺失字段可能导致undefined错误:', result.missingFields);
    }
    
    if (result.extraFields.length > 0) {
      console.info('额外字段暂未使用:', result.extraFields);
      console.log('建议的类型定义更新:');
      console.log(generateTypeDefinition(apiResponse, 'UpdatedInterface'));
    }
  }
  
  console.groupEnd();
}
