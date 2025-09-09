// 企业级信封响应结构 - 符合API一致性规范11.1
export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
  message?: string;
  timestamp: string;
  requestId?: string;
}

// 分页响应结构 - 使用camelCase命名
export interface PaginatedResponse<T> {
  items: T[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

// GraphQL specific types with strict typing
export interface GraphQLResponse<T> {
  data?: T;
  errors?: GraphQLError[];
}

export interface GraphQLError {
  message: string;
  locations?: Array<{
    line: number;
    column: number;
  }>;
  path?: Array<string | number>;
  extensions?: Record<string, unknown>;
}

// Strict GraphQL variables interface
export interface GraphQLVariables {
  searchText?: string;
  unitType?: OrganizationUnitType;
  status?: OrganizationStatus;
  level?: number;
  page?: number;
  pageSize?: number;
}

// API Error types
/**
 * DEPRECATED: 使用 shared/api/error-handling.ts 的统一错误类
 * 
 * 新的导入方式：
 * import { APIError, ValidationError } from '../api/error-handling';
 * import type { APIError, FormValidationErrors } from '../api/error-handling';
 */

// 临时兼容导出，避免破坏现有引用
import { APIError as _APIError } from '../api/error-handling';
import { ValidationError as _ValidationError } from '../api/type-guards';

// TODO-TEMPORARY: 该导出将在 2025-09-16 后删除
export const APIError = _APIError;
export const ValidationError = _ValidationError;

// 新的统一类型定义在 error-handling.ts 中
export interface ValidationIssue {
  field: string;
  message: string;
  code: string;
}

// Type guards for API responses
export const isGraphQLResponse = <T>(response: unknown): response is GraphQLResponse<T> => {
  return (
    typeof response === 'object' &&
    response !== null &&
    ('data' in response || 'errors' in response)
  );
};

export const hasGraphQLErrors = <T>(response: GraphQLResponse<T>): response is GraphQLResponse<T> & { errors: GraphQLError[] } => {
  return Array.isArray(response.errors) && response.errors.length > 0;
};

// DEPRECATED: 使用 shared/api/type-guards.ts 的统一类型守卫
import { isAPIError as _isAPIError, isValidationError as _isValidationError } from '../api/type-guards';

// TODO-TEMPORARY: 该导出将在 2025-09-16 后删除
export const isAPIError = _isAPIError;
export const isValidationError = _isValidationError;

// Utility types for API operations
export type APIResult<T> = Promise<T>;
export type APIOperation<TInput, TOutput> = (input: TInput) => APIResult<TOutput>;

// HTTP method types
export type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

// Request configuration
export interface RequestConfig {
  method: HTTPMethod;
  headers?: Record<string, string>;
  params?: Record<string, string | number | boolean>;
  timeout?: number;
}

// Organization-specific API types
export type OrganizationUnitType = 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM';
export type OrganizationStatus = 'ACTIVE' | 'INACTIVE' | 'PLANNED';