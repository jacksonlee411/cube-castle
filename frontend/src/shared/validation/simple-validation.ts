// 企业级前端验证系统 - Canvas Kit v13兼容
// 采用健壮方案，完整的前端验证体系
// 符合API契约v4.2.1标准，使用camelCase命名

export interface ValidationError {
  field: string;
  message: string;
}

export interface ValidationResult {
  isValid: boolean;
  errors: ValidationError[];
}

// 基础验证函数 - 仅用于即时用户体验反馈
export const basicValidation = {
  required: (value: unknown): boolean => {
    if (typeof value === 'string') {
      return value.trim() !== '';
    }
    return value != null && value !== undefined;
  },

  maxLength: (value: string, max: number): boolean => {
    return !value || value.length <= max;
  },

  minLength: (value: string, min: number): boolean => {
    return !value || value.length >= min;
  },

  pattern: (value: string, regex: RegExp): boolean => {
    return !value || regex.test(value);
  },

  positiveNumber: (value: number): boolean => {
    return typeof value === 'number' && value >= 0;
  }
};

// DEPRECATED: 使用 shared/api/type-guards.ts 的 validateCreateOrganizationInput
// 该函数已被 Zod 验证系统替代

// DEPRECATED: 使用 shared/api/type-guards.ts 的 validateUpdateOrganizationInput
// 该函数已被 Zod 验证系统替代

// DEPRECATED: 使用 shared/api/type-guards.ts 的 validateOrganizationUnit
// 该函数已被 Zod 验证系统替代

// DEPRECATED: 使用 shared/api/type-guards.ts 的 ValidationError
// 该类已被 Zod 验证系统替代

// 格式化错误消息
export function formatValidationErrors(errors: ValidationError[]): string {
  return errors.map(error => error.message).join('; ');
}

// 获取字段错误
export function getFieldError(errors: ValidationError[], fieldName: string): string | undefined {
  const error = errors.find(e => e.field === fieldName);
  return error?.message;
}

// 健壮的数据转换 - 完整类型安全保证
export const safeTransform = {
  // GraphQL到前端格式转换 (兼容REST API响应格式)
  graphqlToOrganization: (orgData: Record<string, unknown>) => {
    // 兼容处理: REST API响应直接返回OrganizationUnit格式
    if (orgData.unitType && orgData.createdAt) {
      // 注意：原验证逻辑已迁移到 type-guards.ts
      return orgData;
    }
    
    // GraphQL格式转换 (支持下划线命名约定)
    return {
      code: orgData.code || '',
      recordId: orgData.recordId || '',  // UUID唯一标识符
      name: orgData.name || '',
      unitType: orgData.unitType || orgData.unitType || '',  // 支持两种命名方式
      status: orgData.status || 'ACTIVE',
      level: orgData.level || 1,
      parentCode: orgData.parentCode || orgData.parentCode || null, // 修复：使用null而不是空字符串
      path: orgData.path || '',
      sortOrder: orgData.sortOrder || orgData.sortOrder || 0,
      description: orgData.description || '',
      createdAt: orgData.createdAt || orgData.createdAt || '',
      updatedAt: orgData.updatedAt || orgData.updatedAt || '',
      // 时态字段（如果存在）
      effectiveDate: orgData.effectiveDate || orgData.effectiveDate || null,
      endDate: orgData.endDate || orgData.endDate || null,
      is_temporal: orgData.is_temporal || orgData.isTemporal || false
    };
  },

  // 简单的数据清理，依赖后端验证
  cleanCreateInput: (input: Record<string, unknown>) => ({
    name: input['name'] && typeof input['name'] === 'string' ? input['name'].trim() : '',
    unitType: input['unitType'],
    parentCode: input['parentCode'] || null,
    sortOrder: input['sortOrder'] || 0,
    description: input['description'] && typeof input['description'] === 'string' ? input['description'].trim() : '',
  }),

  cleanUpdateInput: (input: Record<string, unknown>) => {
    const result: Record<string, unknown> = {};
    if (input['name'] !== undefined && typeof input['name'] === 'string') result['name'] = input['name'].trim();
    if (input['unitType'] !== undefined) result['unitType'] = input['unitType'];
    if (input['status'] !== undefined) result['status'] = input['status'];
    if (input['sortOrder'] !== undefined) result['sortOrder'] = input['sortOrder'];
    if (input['description'] !== undefined && typeof input['description'] === 'string') result['description'] = input['description'].trim();
    return result;
  }
};

// DEPRECATED: 使用 shared/api/type-guards.ts 的统一验证
// 状态验证已整合到 Zod Schema 中

// DEPRECATED: 使用 shared/api/type-guards.ts 的统一验证函数
// 所有验证功能已迁移到 Zod 验证系统