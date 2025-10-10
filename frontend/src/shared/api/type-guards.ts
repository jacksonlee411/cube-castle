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

// GraphQL响应类型守卫
export const isGraphQLError = (
  response: JsonValue,
): response is { errors: GraphQLError[] } => {
  return (
    typeof response === 'object' && 
    response !== null && 
    'errors' in response &&
    Array.isArray((response as JsonObject).errors) &&
    ((response as JsonObject).errors as JsonValue[]).length > 0
  );
};

// GraphQL成功响应类型守卫
export const isGraphQLSuccessResponse = <T>(
  response: JsonValue
): response is GraphQLResponse<T> => {
  if (typeof response !== 'object' || response === null) {
    return false;
  }
  
  const obj = response as JsonObject;
  return (
    'data' in obj &&
    obj.data !== null &&
    obj.data !== undefined
  );
};

type PossibleError = APIError | ValidationError | Error | null | undefined;

// API错误类型守卫
export const isAPIError = (error: PossibleError): error is APIError => {
  return (
    error instanceof Error &&
    'status' in error &&
    'statusText' in error &&
    typeof (error as APIError).status === 'number' &&
    typeof (error as APIError).statusText === 'string'
  );
};

// 验证错误类型守卫
export const isValidationError = (error: PossibleError): error is ValidationError => {
  return error instanceof ValidationError;
};

// 网络错误类型守卫
export const isNetworkError = (error: Error | null | undefined): error is TypeError => {
  return error instanceof TypeError && error.message.includes('fetch');
};

// 安全的类型转换函数 - 将GraphQL响应转换为前端期望的格式
export const safeTransformGraphQLToOrganizationUnit = (
  graphqlOrg: ValidatedGraphQLOrganizationResponse
): OrganizationUnit => {
  return {
    code: graphqlOrg.code,
    parentCode: graphqlOrg.parentCode || '',
    name: graphqlOrg.name,
    unitType: graphqlOrg.unitType as OrganizationUnit['unitType'],
    status: graphqlOrg.status as OrganizationUnit['status'],
    level: graphqlOrg.level,
    codePath: graphqlOrg.codePath ?? undefined,
    namePath: graphqlOrg.namePath ?? undefined,
    path: graphqlOrg.codePath ?? graphqlOrg.path ?? undefined,
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
  const validated = validateCreateOrganizationInput(sanitized);
  
  const apiPayload: JsonObject = {
    name: validated.name,
    unitType: validated.unitType,
    status: validated.status,
    level: validated.level,
    sortOrder: validated.sortOrder,
    description: validated.description,
  };

  // 只添加有值的可选字段
  if (validated['code'] !== undefined) {
    apiPayload.code = validated['code'];
  }
  if (validated['parentCode'] !== undefined && validated['parentCode'] !== '') {
    apiPayload.parentCode = validated['parentCode'];
  }

  return apiPayload;
};
