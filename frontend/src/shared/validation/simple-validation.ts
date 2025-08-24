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

// 组织单元基础验证 - 依赖后端统一验证
export function validateOrganizationBasic(data: Record<string, unknown>): ValidationResult {
  const errors: ValidationError[] = [];

  // 仅保留关键的用户体验验证
  if (!basicValidation.required(data['name'])) {
    errors.push({ field: 'name', message: '组织名称不能为空' });
  }

  if (data['name'] && typeof data['name'] === 'string' && !basicValidation.maxLength(data['name'], 100)) {
    errors.push({ field: 'name', message: '组织名称不能超过100个字符' });
  }

  if (!basicValidation.required(data['unitType'])) {
    errors.push({ field: 'unitType', message: '请选择组织类型' });
  }

  if (data['sortOrder'] !== undefined && typeof data['sortOrder'] === 'number' && !basicValidation.positiveNumber(data['sortOrder'])) {
    errors.push({ field: 'sortOrder', message: '排序顺序必须为非负数' });
  }

  return {
    isValid: errors.length === 0,
    errors
  };
}

// 组织单元更新验证 - 用于编辑模式（支持所有字段编辑，除了组织编码）
export function validateOrganizationUpdate(data: Record<string, unknown>): ValidationResult {
  const errors: ValidationError[] = [];

  // 仅保留关键的用户体验验证
  if (data['name'] && !basicValidation.required(data['name'])) {
    errors.push({ field: 'name', message: '组织名称不能为空' });
  }

  if (data['name'] && typeof data['name'] === 'string' && !basicValidation.maxLength(data['name'], 100)) {
    errors.push({ field: 'name', message: '组织名称不能超过100个字符' });
  }

  // 编辑模式下也需要验证unitType
  if (data['unitType'] && !basicValidation.required(data['unitType'])) {
    errors.push({ field: 'unitType', message: '请选择组织类型' });
  }

  // 验证level字段
  if (data['level'] !== undefined && typeof data['level'] === 'number' && !basicValidation.positiveNumber(data['level'])) {
    errors.push({ field: 'level', message: '组织层级必须为正数' });
  }

  if (data['level'] && typeof data['level'] === 'number' && (data['level'] < 1 || data['level'] > 10)) {
    errors.push({ field: 'level', message: '组织层级必须在1-10之间' });
  }

  if (data['sortOrder'] !== undefined && typeof data['sortOrder'] === 'number' && !basicValidation.positiveNumber(data['sortOrder'])) {
    errors.push({ field: 'sortOrder', message: '排序顺序必须为非负数' });
  }

  return {
    isValid: errors.length === 0,
    errors
  };
}

// 验证API响应格式 - 确保是完整的组织单元对象
export function validateOrganizationResponse(data: Record<string, unknown>): ValidationResult {
  const errors: ValidationError[] = [];

  // 验证必需字段
  const requiredFields = ['code', 'name', 'unitType', 'status', 'level'];
  for (const field of requiredFields) {
    if (!basicValidation.required(data[field])) {
      errors.push({ field, message: `${field} 字段不能为空` });
    }
  }

  // 验证状态枚举
  if (data['status'] && typeof data['status'] === 'string' && !['ACTIVE', 'SUSPENDED', 'PLANNED'].includes(data['status'])) {
    errors.push({ field: 'status', message: '状态值无效' });
  }

  // 验证类型枚举  
  if (data['unitType'] && typeof data['unitType'] === 'string' && !['DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM'].includes(data['unitType'])) {
    errors.push({ field: 'unitType', message: '组织类型无效' });
  }

  return {
    isValid: errors.length === 0,
    errors
  };
}

// 企业级错误处理 - 前后端协同验证
export class SimpleValidationError extends Error {
  public readonly fieldErrors: ValidationError[];
  
  constructor(message: string, errors: ValidationError[] = []) {
    super(message);
    this.name = 'SimpleValidationError';
    this.fieldErrors = errors;
  }
}

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
      // 这是REST API响应格式，直接验证并返回
      const basicValidation = validateOrganizationResponse(orgData);
      if (basicValidation.isValid) {
        return orgData;
      }
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

// 状态更新验证 - 仅验证状态相关字段
export function validateStatusUpdate(data: Record<string, unknown>): ValidationResult {
  const errors: ValidationError[] = [];

  // 仅验证状态字段
  if (!basicValidation.required(data['status'])) {
    errors.push({ field: 'status', message: '状态不能为空' });
  }

  if (data['status'] && typeof data['status'] === 'string' && !['ACTIVE', 'SUSPENDED', 'PLANNED'].includes(data['status'])) {
    errors.push({ field: 'status', message: '状态值无效，必须是 ACTIVE、SUSPENDED 或 PLANNED' });
  }

  return {
    isValid: errors.length === 0,
    errors
  };
}

// 向后兼容的导出 (用于逐步迁移)
export const validateCreateOrganizationInput = validateOrganizationBasic;
export const validateUpdateOrganizationInput = validateOrganizationBasic;