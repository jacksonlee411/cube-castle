import { z } from 'zod';
import {
  OrganizationUnitSchema,
  CreateOrganizationInputSchema,
  CreateOrganizationResponseSchema,
  UpdateOrganizationInputSchema,
  GraphQLVariablesSchema,
  GraphQLOrganizationResponseSchema,
} from '../validation/schemas';
import type { OrganizationRequest as CreateOrganizationInput } from '../types/organization';
import type {
  ValidatedOrganizationUnit,
  ValidatedCreateOrganizationInput,
  ValidatedCreateOrganizationResponse,
  ValidatedUpdateOrganizationInput,
  ValidatedGraphQLVariables,
  ValidatedGraphQLOrganizationResponse,
} from '../validation/schemas';
import type { OrganizationUnit } from '../types/organization';
import type { GraphQLResponse, GraphQLError } from '../types/api';
import type { JsonObject, JsonValue } from '../types/json';
import type { APIError } from './error-handling';

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
export const validateOrganizationUnit = (data: JsonValue): ValidatedOrganizationUnit => {
  const result = OrganizationUnitSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid organization unit data', result.error.issues);
  }
  return result.data;
};

// 创建组织输入验证函数
export const validateCreateOrganizationInput = (
  data: JsonValue,
): ValidatedCreateOrganizationInput => {
  const result = CreateOrganizationInputSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid create organization input', result.error.issues);
  }
  return result.data;
};

// 更新组织输入验证函数
export const validateUpdateOrganizationInput = (
  data: JsonValue,
): ValidatedUpdateOrganizationInput => {
  const result = UpdateOrganizationInputSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid update organization input', result.error.issues);
  }
  return result.data;
};

// 创建组织响应验证函数
export const validateCreateOrganizationResponse = (
  data: JsonValue,
): ValidatedCreateOrganizationResponse => {
  const result = CreateOrganizationResponseSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid create organization response', result.error.issues);
  }
  return result.data;
};

// GraphQL变量验证函数
export const validateGraphQLVariables = (data: JsonValue): ValidatedGraphQLVariables => {
  const result = GraphQLVariablesSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid GraphQL variables', result.error.issues);
  }
  return result.data;
};

// GraphQL组织响应验证函数
export const validateGraphQLOrganizationResponse = (
  data: JsonValue,
): ValidatedGraphQLOrganizationResponse => {
  const result = GraphQLOrganizationResponseSchema.safeParse(data);
  if (!result.success) {
    throw new ValidationError('Invalid GraphQL organization response', result.error.issues);
  }
  return result.data;
};

// 批量验证GraphQL组织响应
export const validateGraphQLOrganizationList = (
  data: JsonValue[],
): ValidatedGraphQLOrganizationResponse[] => {
  return data.map((item, index) => {
    try {
      return validateGraphQLOrganizationResponse(item);
    } catch (error) {
      throw new ValidationError(`Invalid organization data at index ${index}`, 
        error instanceof ValidationError ? error.details : [{ message: 'Unknown validation error', code: 'custom', path: [] }]);
    }
  });
};

const isGraphQLErrorEntry = (value: unknown): value is GraphQLError => {
  return (
    typeof value === 'object' &&
    value !== null &&
    'message' in value &&
    typeof (value as { message?: unknown }).message === 'string'
  );
};

// GraphQL响应类型守卫
export const isGraphQLError = (
  response: unknown,
): response is { errors: GraphQLError[] } => {
  if (typeof response !== 'object' || response === null) {
    return false;
  }

  const errors = (response as { errors?: unknown }).errors;
  return Array.isArray(errors) && errors.length > 0 && errors.every(isGraphQLErrorEntry);
};

// GraphQL成功响应类型守卫
export const isGraphQLSuccessResponse = <T>(
  response: unknown
): response is GraphQLResponse<T> => {
  if (typeof response !== 'object' || response === null) {
    return false;
  }
  
  const obj = response as Record<string, unknown>;
  return 'data' in obj && obj.data !== null && obj.data !== undefined;
};

type PossibleError = unknown;

// API错误类型守卫
export const isAPIError = (error: PossibleError): error is APIError => {
  if (typeof error !== 'object' || error === null) {
    return false;
  }

  return (
    error instanceof Error &&
    'status' in error &&
    'statusText' in error &&
    typeof (error as { status?: unknown }).status === 'number' &&
    typeof (error as { statusText?: unknown }).statusText === 'string'
  );
};

// 验证错误类型守卫
export const isValidationError = (error: PossibleError): error is ValidationError => {
  return error instanceof ValidationError;
};

// 网络错误类型守卫
export const isNetworkError = (error: PossibleError): error is TypeError => {
  return error instanceof TypeError && typeof error.message === 'string' && error.message.includes('fetch');
};

// 安全的类型转换函数 - 将GraphQL响应转换为前端期望的格式
export const safeTransformGraphQLToOrganizationUnit = (
  graphqlOrg: ValidatedGraphQLOrganizationResponse
): OrganizationUnit => {
  const codePath =
    'codePath' in graphqlOrg && typeof (graphqlOrg as { codePath?: string | null }).codePath === 'string'
      ? (graphqlOrg as { codePath?: string | null }).codePath ?? undefined
      : undefined;
  const namePath =
    'namePath' in graphqlOrg && typeof (graphqlOrg as { namePath?: string | null }).namePath === 'string'
      ? (graphqlOrg as { namePath?: string | null }).namePath ?? undefined
      : undefined;
  const pathValue =
    'path' in graphqlOrg && typeof (graphqlOrg as { path?: string | null }).path === 'string'
      ? (graphqlOrg as { path?: string | null }).path ?? undefined
      : undefined;

  return {
    code: graphqlOrg.code,
    parentCode: graphqlOrg.parentCode || '',
    name: graphqlOrg.name,
    unitType: graphqlOrg.unitType as OrganizationUnit['unitType'],
    status: graphqlOrg.status as OrganizationUnit['status'],
    level: graphqlOrg.level,
    codePath,
    namePath,
    path: pathValue ?? codePath,
    sortOrder: graphqlOrg.sortOrder || 0,
    description: graphqlOrg.description || '',
    createdAt: graphqlOrg.createdAt || '',
    updatedAt: graphqlOrg.updatedAt || '',
  };
};

// 安全的类型转换函数 - 将前端输入转换为API格式
export const safeTransformCreateInputToAPI = (
  input: CreateOrganizationInput,
): JsonObject => {
  // 预清洗：空字符串的 parentCode 视为未提供
  const sanitized: CreateOrganizationInput = {
    ...input,
    parentCode: input.parentCode === '' ? undefined : input.parentCode,
  };
  const validated = validateCreateOrganizationInput(sanitized as JsonValue);

  const apiPayload: JsonObject = {
    name: validated.name,
    unitType: validated.unitType,
    status: validated.status,
    level: validated.level,
    sortOrder: validated.sortOrder,
  };

  if (typeof validated.description === 'string') {
    apiPayload.description = validated.description;
  }

  // 只添加有值的可选字段
  if (validated['code'] !== undefined) {
    apiPayload.code = validated['code'];
  }
  if (validated['parentCode'] !== undefined && validated['parentCode'] !== '') {
    apiPayload.parentCode = validated['parentCode'];
  }

  return apiPayload;
};
