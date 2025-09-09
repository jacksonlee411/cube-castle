/**
 * 统一验证系统导出文件
 * 基于Zod的企业级验证架构，消除重复代码
 * 
 * 这个文件是验证系统的统一入口，所有验证功能都应该从这里导入
 */

// ============================================================================
// 核心验证功能 - 来自 type-guards.ts (Zod-based)
// ============================================================================

export {
  // 验证函数
  validateOrganizationUnit,
  validateCreateOrganizationInput,
  validateUpdateOrganizationInput,
  validateCreateOrganizationResponse,
  validateGraphQLVariables,
  validateGraphQLOrganizationResponse,
  validateGraphQLOrganizationList,
  
  // 类型守卫
  isGraphQLError,
  isGraphQLSuccessResponse,
  isAPIError,
  isValidationError,
  isNetworkError,
  
  // 安全转换函数
  safeTransformGraphQLToOrganizationUnit,
  safeTransformCreateInputToAPI,
  
  // 验证错误类
  ValidationError
} from '../api/type-guards';

// ============================================================================
// Zod Schema定义 - 来自 schemas.ts
// ============================================================================

export {
  // 核心Schema
  OrganizationUnitSchema,
  CreateOrganizationInputSchema,
  CreateOrganizationResponseSchema,
  UpdateOrganizationInputSchema,
  GraphQLVariablesSchema,
  GraphQLOrganizationResponseSchema,
  
  // 验证类型
  type ValidatedOrganizationUnit,
  type ValidatedCreateOrganizationInput,
  type ValidatedCreateOrganizationResponse,
  type ValidatedUpdateOrganizationInput,
  type ValidatedGraphQLVariables,
  type ValidatedGraphQLOrganizationResponse,
  
  // 工具函数
  ValidationUtils
} from './schemas';

// ============================================================================
// 废弃警告 - 引导开发者使用正确的验证系统
// ============================================================================

/**
 * 废弃文件警告：
 * 
 * ❌ 不要使用 simple-validation.ts - 已废弃
 * ❌ 不要使用 converters.ts 中的验证函数 - 已废弃
 * 
 * ✅ 使用本文件导出的统一验证系统：
 * 
 * import { validateOrganizationUnit, ValidationError } from '@/shared/validation';
 * 
 * 优势：
 * - Zod 提供强类型安全
 * - 统一错误处理
 * - 企业级验证规则
 * - 消除代码重复
 */

// ============================================================================
// 开发工具 - 用于验证系统迁移
// ============================================================================

export const VALIDATION_SYSTEM_INFO = {
  version: '2.0.0',
  architecture: 'Zod-based unified validation',
  coreFiles: [
    'shared/api/type-guards.ts',
    'shared/validation/schemas.ts',
    'shared/validation/index.ts'
  ],
  deprecatedFiles: [
    'shared/validation/simple-validation.ts',
    'shared/types/converters.ts (validation functions only)'
  ],
  migrationStatus: 'COMPLETED',
  duplicateCodeEliminated: true
} as const;