import { z } from 'zod';
import { 
  OrganizationUnitSchema, 
  CreateOrganizationInputSchema, 
  CreateOrganizationResponseSchema,
  UpdateOrganizationInputSchema,
  GraphQLVariablesSchema,
  GraphQLOrganizationResponseSchema
} from '../validation/schemas';
import type { CreateOrganizationInput } from '../hooks/useOrganizationMutations';
import type { 
  ValidatedOrganizationUnit,
  ValidatedCreateOrganizationInput,
  ValidatedCreateOrganizationResponse,
  ValidatedUpdateOrganizationInput,
  ValidatedGraphQLVariables,
  ValidatedGraphQLOrganizationResponse 
} from '../validation/schemas';
import type { OrganizationUnit } from '../types/organization';
import type { GraphQLResponse, GraphQLError } from '../types/api';

// 验证错误类
export class ValidationError extends Error {
  public readonly code: string = 'VALIDATION_ERROR';
  public readonly details: z.ZodIssue[];
  
  constructor(message: string, details: z.ZodIssue[]) {
    super(message);
    this.name = 'ValidationError';
    this.details = details;
  }
}

// 组织单元验证函数
export const validateOrganizationUnit = (data: unknown): ValidatedOrganizationUnit => {
  const result = OrganizationUnitSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid organization unit data', result.error.issues);
  }
  return result.data;
};

// 创建组织输入验证函数
export const validateCreateOrganizationInput = (data: unknown): ValidatedCreateOrganizationInput => {
  const result = CreateOrganizationInputSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid create organization input', result.error.issues);
  }
  return result.data;
};

// 更新组织输入验证函数
export const validateUpdateOrganizationInput = (data: unknown): ValidatedUpdateOrganizationInput => {
  const result = UpdateOrganizationInputSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid update organization input', result.error.issues);
  }
  return result.data;
};

// 创建组织响应验证函数
export const validateCreateOrganizationResponse = (data: unknown): ValidatedCreateOrganizationResponse => {
  const result = CreateOrganizationResponseSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid create organization response', result.error.issues);
  }
  return result.data;
};

// GraphQL变量验证函数
export const validateGraphQLVariables = (data: unknown): ValidatedGraphQLVariables => {
  const result = GraphQLVariablesSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid GraphQL variables', result.error.issues);
  }
  return result.data;
};

// GraphQL组织响应验证函数
export const validateGraphQLOrganizationResponse = (data: unknown): ValidatedGraphQLOrganizationResponse => {
  const result = GraphQLOrganizationResponseSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid GraphQL organization response', result.error.issues);
  }
  return result.data;
};

// 批量验证GraphQL组织响应
export const validateGraphQLOrganizationList = (data: unknown[]): ValidatedGraphQLOrganizationResponse[] => {
  return data.map((item, index) => {
    try {
      return validateGraphQLOrganizationResponse(item);
    } catch (error) {
      throw new ValidationError(`Invalid organization data at index ${index}`, 
        error instanceof ValidationError ? error.details : [{ message: 'Unknown validation error', code: 'custom', path: [] }]);
    }
  });
};

// GraphQL响应类型守卫
export const isGraphQLError = (response: unknown): response is { errors: GraphQLError[] } => {
  return (
    typeof response === 'object' && 
    response !== null && 
    'errors' in response &&
    Array.isArray((response as Record<string, unknown>).errors) &&
    (response as Record<string, unknown>).errors.length > 0
  );
};

// GraphQL成功响应类型守卫
export const isGraphQLSuccessResponse = <T>(
  response: unknown
): response is GraphQLResponse<T> => {
  return (
    typeof response === 'object' && 
    response !== null && 
    'data' in response &&
    (response as Record<string, unknown>).data !== null &&
    (response as Record<string, unknown>).data !== undefined
  );
};

// API错误类型守卫
export const isAPIError = (error: unknown): error is Error & { status: number; statusText: string } => {
  return (
    error instanceof Error && 
    'status' in error && 
    'statusText' in error &&
    typeof (error as Record<string, unknown>).status === 'number' &&
    typeof (error as Record<string, unknown>).statusText === 'string'
  );
};

// 验证错误类型守卫
export const isValidationError = (error: unknown): error is ValidationError => {
  return error instanceof ValidationError;
};

// 网络错误类型守卫
export const isNetworkError = (error: unknown): error is TypeError => {
  return error instanceof TypeError && error.message.includes('fetch');
};

// 安全的类型转换函数 - 将GraphQL响应转换为前端期望的格式
export const safeTransformGraphQLToOrganizationUnit = (
  graphqlOrg: ValidatedGraphQLOrganizationResponse
): OrganizationUnit => {
  return {
    code: graphqlOrg.code,
    parent_code: graphqlOrg.parentCode || '',
    name: graphqlOrg.name,
    unit_type: graphqlOrg.unitType as OrganizationUnit['unit_type'],
    status: graphqlOrg.status as OrganizationUnit['status'],
    level: graphqlOrg.level,
    path: graphqlOrg.path || '',
    sort_order: graphqlOrg.sortOrder || 0,
    description: graphqlOrg.description || '',
    created_at: graphqlOrg.createdAt || '',
    updated_at: graphqlOrg.updatedAt || '',
  };
};

// 安全的类型转换函数 - 将前端输入转换为API格式
export const safeTransformCreateInputToAPI = (
  input: CreateOrganizationInput
): Record<string, unknown> => {
  const validated = validateCreateOrganizationInput(input);
  
  const apiPayload: Record<string, unknown> = {
    name: validated.name,
    unit_type: validated.unit_type,
    status: validated.status,
    level: validated.level,
    sort_order: validated.sort_order,
    description: validated.description,
  };

  // 只添加有值的可选字段
  if (validated['code'] !== undefined) {
    apiPayload.code = validated['code'];
  }
  if (validated['parent_code'] !== undefined && validated['parent_code'] !== '') {
    apiPayload.parent_code = validated['parent_code'];
  }

  return apiPayload;
};