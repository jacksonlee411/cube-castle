// 简化的前端验证系统
// 移除Zod依赖，使用基础JavaScript验证

// ===== 基础验证函数 =====

export const basicValidation = {
  // 必填验证
  required: (value: any, fieldName: string = 'Field'): string | null => {
    if (value === null || value === undefined || value === '') {
      return `${fieldName} is required`;
    }
    if (typeof value === 'string' && value.trim() === '') {
      return `${fieldName} cannot be empty`;
    }
    return null;
  },

  // 最大长度验证
  maxLength: (value: string, max: number, fieldName: string = 'Field'): string | null => {
    if (!value) return null;
    if (value.length > max) {
      return `${fieldName} cannot exceed ${max} characters`;
    }
    return null;
  },

  // 最小长度验证
  minLength: (value: string, min: number, fieldName: string = 'Field'): string | null => {
    if (!value) return null;
    if (value.length < min) {
      return `${fieldName} must be at least ${min} characters`;
    }
    return null;
  },

  // 正则表达式验证
  pattern: (value: string, regex: RegExp, fieldName: string = 'Field', errorMsg?: string): string | null => {
    if (!value) return null;
    if (!regex.test(value)) {
      return errorMsg || `${fieldName} format is invalid`;
    }
    return null;
  },

  // 数字范围验证
  numberRange: (value: number, min: number, max: number, fieldName: string = 'Field'): string | null => {
    if (value < min || value > max) {
      return `${fieldName} must be between ${min} and ${max}`;
    }
    return null;
  },

  // 枚举值验证
  enum: (value: string, validValues: string[], fieldName: string = 'Field'): string | null => {
    if (!validValues.includes(value)) {
      return `${fieldName} must be one of: ${validValues.join(', ')}`;
    }
    return null;
  }
};

// ===== 组织单元特定验证 =====

export interface ValidationResult {
  isValid: boolean;
  errors: { [key: string]: string };
}

export const organizationValidation = {
  // 验证创建组织输入
  validateCreateInput: (input: any): ValidationResult => {
    const errors: { [key: string]: string } = {};

    // 名称验证
    const nameError = basicValidation.required(input.name, 'Organization name') ||
                     basicValidation.maxLength(input.name, 100, 'Organization name');
    if (nameError) errors.name = nameError;

    // 组织类型验证
    const unitTypeError = basicValidation.required(input.unit_type, 'Unit type') ||
                         basicValidation.enum(input.unit_type, ['COMPANY', 'DEPARTMENT', 'TEAM'], 'Unit type');
    if (unitTypeError) errors.unit_type = unitTypeError;

    // 排序顺序验证
    if (input.sort_order !== undefined && input.sort_order < 0) {
      errors.sort_order = 'Sort order cannot be negative';
    }

    // 描述长度验证
    if (input.description && input.description.length > 500) {
      errors.description = 'Description cannot exceed 500 characters';
    }

    return {
      isValid: Object.keys(errors).length === 0,
      errors
    };
  },

  // 验证更新组织输入
  validateUpdateInput: (input: any): ValidationResult => {
    const errors: { [key: string]: string } = {};

    // 只验证提供的字段
    if (input.name !== undefined) {
      const nameError = basicValidation.required(input.name, 'Organization name') ||
                       basicValidation.maxLength(input.name, 100, 'Organization name');
      if (nameError) errors.name = nameError;
    }

    if (input.unit_type !== undefined) {
      const unitTypeError = basicValidation.enum(input.unit_type, ['COMPANY', 'DEPARTMENT', 'TEAM'], 'Unit type');
      if (unitTypeError) errors.unit_type = unitTypeError;
    }

    if (input.status !== undefined) {
      const statusError = basicValidation.enum(input.status, ['ACTIVE', 'INACTIVE', 'PLANNED'], 'Status');
      if (statusError) errors.status = statusError;
    }

    if (input.sort_order !== undefined && input.sort_order < 0) {
      errors.sort_order = 'Sort order cannot be negative';
    }

    if (input.description && input.description.length > 500) {
      errors.description = 'Description cannot exceed 500 characters';
    }

    return {
      isValid: Object.keys(errors).length === 0,
      errors
    };
  },

  // 验证查询参数
  validateQueryParams: (params: any): ValidationResult => {
    const errors: { [key: string]: string } = {};

    if (params.unit_type && !['COMPANY', 'DEPARTMENT', 'TEAM'].includes(params.unit_type)) {
      errors.unit_type = 'Invalid unit type';
    }

    if (params.status && !['ACTIVE', 'INACTIVE', 'PLANNED'].includes(params.status)) {
      errors.status = 'Invalid status';
    }

    if (params.level !== undefined) {
      const levelError = basicValidation.numberRange(params.level, 1, 10, 'Level');
      if (levelError) errors.level = levelError;
    }

    if (params.page !== undefined && params.page < 1) {
      errors.page = 'Page must be positive';
    }

    if (params.pageSize !== undefined) {
      const pageSizeError = basicValidation.numberRange(params.pageSize, 1, 100, 'Page size');
      if (pageSizeError) errors.pageSize = pageSizeError;
    }

    return {
      isValid: Object.keys(errors).length === 0,
      errors
    };
  }
};

// ===== 简化的错误处理 =====

export class SimpleValidationError extends Error {
  public readonly code: string = 'VALIDATION_ERROR';
  public readonly fieldErrors: { [key: string]: string };
  
  constructor(message: string, fieldErrors: { [key: string]: string }) {
    super(message);
    this.name = 'SimpleValidationError';
    this.fieldErrors = fieldErrors;
  }
}

export const createValidationError = (result: ValidationResult, message: string = 'Validation failed') => {
  if (!result.isValid) {
    return new SimpleValidationError(message, result.errors);
  }
  return null;
};

// ===== 类型守卫（简化版） =====

export const isSimpleValidationError = (error: any): error is SimpleValidationError => {
  return error instanceof SimpleValidationError;
};

export const isNetworkError = (error: any): boolean => {
  return error?.name === 'NetworkError' || error?.code === 'NETWORK_ERROR';
};

export const isAPIError = (error: any): boolean => {
  return error?.response && error?.response?.status >= 400;
};

// ===== 数据转换辅助函数 =====

export const safeTransform = {
  // 安全转换GraphQL响应到前端格式
  graphqlToOrganization: (data: any) => {
    if (!data || typeof data !== 'object') {
      throw new SimpleValidationError('Invalid organization data', { data: 'Data must be an object' });
    }

    return {
      code: data.code || '',
      name: data.name || '',
      unit_type: data.unitType || data.unit_type || 'DEPARTMENT',
      status: data.status || 'ACTIVE',
      level: data.level || 1,
      path: data.path || '',
      sort_order: data.sortOrder || data.sort_order || 0,
      description: data.description || '',
      parent_code: data.parentCode || data.parent_code || null,
      created_at: data.createdAt || data.created_at || new Date().toISOString(),
      updated_at: data.updatedAt || data.updated_at || new Date().toISOString(),
    };
  },

  // 转换创建输入为API格式
  createInputToAPI: (input: any) => {
    return {
      name: input.name?.trim(),
      unit_type: input.unit_type,
      parent_code: input.parent_code || null,
      sort_order: input.sort_order || 0,
      description: input.description?.trim() || '',
    };
  },

  // 转换更新输入为API格式
  updateInputToAPI: (input: any) => {
    const result: any = {};
    
    if (input.name !== undefined) result.name = input.name?.trim();
    if (input.unit_type !== undefined) result.unit_type = input.unit_type;
    if (input.status !== undefined) result.status = input.status;
    if (input.sort_order !== undefined) result.sort_order = input.sort_order;
    if (input.description !== undefined) result.description = input.description?.trim() || '';

    return result;
  }
};

// ===== 导出简化的验证接口 =====

export {
  basicValidation as validation,
  organizationValidation as orgValidation,
  SimpleValidationError as ValidationError,
  createValidationError,
  safeTransform
};

// 向后兼容的函数名（逐步迁移）
export const validateCreateOrganizationInput = (input: any) => {
  const result = organizationValidation.validateCreateInput(input);
  const error = createValidationError(result);
  if (error) throw error;
  return input; // 返回原始输入，不进行类型转换
};

export const validateUpdateOrganizationInput = (input: any) => {
  const result = organizationValidation.validateUpdateInput(input);
  const error = createValidationError(result);
  if (error) throw error;
  return input;
};

export const validateOrganizationUnit = (data: any) => {
  // 简化版本 - 只做基础检查
  if (!data || typeof data !== 'object') {
    throw new SimpleValidationError('Invalid organization data', { data: 'Must be an object' });
  }
  return safeTransform.graphqlToOrganization(data);
};