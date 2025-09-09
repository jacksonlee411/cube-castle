/**
 * 统一类型导出体系 - 消除类型重复定义
 * 🎯 单一真源：所有TypeScript类型的权威来源
 * 🔒 避免重复：消除组件中的重复类型定义
 */

// 🎯 核心业务类型
export * from './organization';
export * from './temporal';
export * from './api';

// 🎯 类型转换工具
export * from './converters';

// 🎯 验证系统类型 - 从统一验证系统导入
export type {
  ValidationError as ValidatorError,
  ValidationResult,
  ValidatedOrganizationUnit,
  ValidatedCreateOrganizationInput,
  ValidatedUpdateOrganizationInput,
  ValidatedGraphQLVariables,
  ValidatedGraphQLOrganizationResponse
} from '../validation/schemas';

// 🎯 错误处理类型 - 从统一错误处理系统导入  
export type {
  ApiErrorCode,
  SuccessMessageKey,
  FormValidationErrors
} from '../api/error-handling';

// 🎯 配置系统类型
export type {
  ServicePortKey,
  CQRSEndpointKey
} from '../config/ports';

// 🎯 组件Props类型 - 统一组件接口定义
export interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
}

export interface LoadingProps extends BaseComponentProps {
  isLoading?: boolean;
  loadingText?: string;
}

export interface ErrorProps extends BaseComponentProps {
  error?: string | null;
  onRetry?: () => void;
}

// 🎯 Hook返回类型
export interface UseOrganizationsResult {
  organizations: OrganizationUnit[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
  fetchMore?: (page: number) => void;
}

// 🎯 通用工具类型
export type Nullable<T> = T | null;
export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;

// 🎯 异步操作类型
export type AsyncState<T> = {
  data: T | null;
  loading: boolean;
  error: string | null;
};

// 📋 类型系统使用指南
export const TYPE_USAGE_GUIDE = {
  '🎯 业务实体': 'import type { OrganizationUnit } from "@/shared/types"',
  '🎯 API操作': 'import type { APIResponse, OrganizationQueryParams } from "@/shared/types"',
  '🎯 验证相关': 'import type { ValidationResult } from "@/shared/types"',
  '🎯 错误处理': 'import type { ApiErrorCode } from "@/shared/types"',
  '🎯 配置相关': 'import type { ServicePortKey } from "@/shared/types"'
} as const;