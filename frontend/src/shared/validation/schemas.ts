import { z } from 'zod';

// 统一验证系统 - 整合所有验证逻辑到Zod Schema
// 替代: simple-validation.ts, ValidationRules.ts, temporalValidation.ts

// 时态日期验证助手
const TemporalDateSchema = z.string().refine(
  (dateString) => {
    if (!dateString) return true; // 允许空值
    const date = new Date(dateString);
    return date instanceof Date && !isNaN(date.getTime()) && dateString === date.toISOString().split('T')[0];
  },
  { message: '无效的日期格式' }
);

// 未来日期验证
const FutureDateSchema = z.string().refine(
  (dateString) => {
    if (!dateString) return true;
    const date = new Date(dateString);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return date > today;
  },
  { message: '日期必须是未来日期' }
);

// 组织单元完整验证模式
export const OrganizationUnitSchema = z.object({
  code: z.string().regex(/^\d{7}$/, 'Organization code must be 7 digits'),
  name: z.string().min(1, '组织名称不能为空').max(100, '组织名称不能超过100个字符'),
  unitType: z.enum(['DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM'], { 
    message: '请选择有效的组织类型'
  }),
  status: z.enum(['ACTIVE', 'INACTIVE', 'PLANNED'], {
    message: '状态必须是 ACTIVE、INACTIVE 或 PLANNED'
  }),
  level: z.number().int().min(1, '组织层级必须大于0').max(10, '组织层级不能超过10'),
  parentCode: z.string().regex(/^(0|\d{7})$/, 'Parent code must be "0" for root organizations or 7 digits for child organizations'),
  sortOrder: z.number().int().min(0, '排序顺序必须为非负数').default(0),
  description: z.string().max(500, '描述不能超过500个字符').optional().or(z.literal('')),
  createdAt: z.string().datetime().optional().or(z.literal('')),
  updatedAt: z.string().datetime().optional().or(z.literal('')),
  path: z.string().optional().or(z.literal('')),
  // 时态字段支持
  effectiveDate: TemporalDateSchema.optional(),
  endDate: TemporalDateSchema.optional(),
  isTemporal: z.boolean().default(false),
});

// 时态表单验证Schema - 整合ValidationRules.ts的时态验证逻辑
export const TemporalFormSchema = z.object({
  effectiveFrom: z.string().optional(),
  effectiveTo: z.string().optional(), 
  changeReason: z.string().max(200, '变更原因不能超过200个字符').optional(),
  isTemporal: z.boolean().default(false)
}).refine(
  (data) => {
    if (!data.isTemporal) return true;
    
    // 时态模式下的验证
    if (!data.effectiveFrom) return false;
    if (!data.changeReason?.trim()) return false;
    
    // 验证日期范围
    if (data.effectiveFrom && data.effectiveTo) {
      const from = new Date(data.effectiveFrom);
      const to = new Date(data.effectiveTo);
      return to > from;
    }
    
    return true;
  },
  {
    message: '时态模式下必须提供生效时间和变更原因，且失效时间必须晚于生效时间',
    path: ['effectiveFrom']
  }
);

// 创建组织单元输入验证模式 - 整合所有验证逻辑
export const CreateOrganizationInputSchema = z.object({
  code: z.string().regex(/^\d{7}$/, '组织编码必须为7位数字').optional(), // 可选，由系统生成
  name: z.string().min(1, '组织名称不能为空').max(100, '组织名称不能超过100个字符'),
  unitType: z.enum(['DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM'], { 
    message: '请选择有效的组织类型'
  }),
  status: z.enum(['ACTIVE', 'INACTIVE', 'PLANNED'], {
    message: '状态必须是 ACTIVE、INACTIVE 或 PLANNED'
  }).default('ACTIVE'),
  level: z.number().int().min(1, '组织层级必须大于0').max(10, '组织层级不能超过10'),
  parentCode: z.string().regex(/^(0|\d{7})$/, 'Parent code must be "0" for root organizations or 7 digits for child organizations'),
  sortOrder: z.number().int().min(0, '排序顺序必须为非负数').default(0),
  description: z.string().max(500, '描述不能超过500个字符').optional().or(z.literal('')),
}).merge(TemporalFormSchema); // 合并时态验证

// 创建组织单元响应验证模式 (后端实际返回的字段)
export const CreateOrganizationResponseSchema = z.object({
  code: z.string().regex(/^\d{7}$/, 'Organization code must be 7 digits'),
  name: z.string(),
  unitType: z.enum(['DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM']),
  status: z.enum(['ACTIVE', 'INACTIVE', 'PLANNED']),
  createdAt: z.string().datetime(),
  // 注意：后端创建响应不包含这些字段：level, parentCode, sortOrder, description, updatedAt, path
});

// 更新组织单元输入验证模式
export const UpdateOrganizationInputSchema = CreateOrganizationInputSchema.partial().extend({
  code: z.string().regex(/^\d{7}$/, 'Organization code must be 7 digits'), // 更新时code必需
});

// GraphQL查询变量验证模式
export const GraphQLVariablesSchema = z.object({
  searchText: z.string().optional(),
  unitType: z.enum(['DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM']).optional(),
  status: z.enum(['ACTIVE', 'INACTIVE', 'PLANNED']).optional(),
  level: z.number().int().min(1).max(10).optional(),
  page: z.number().int().min(1).optional(),
  pageSize: z.number().int().min(1).max(100).optional(),
});

// GraphQL组织响应验证模式
export const GraphQLOrganizationResponseSchema = z.object({
  code: z.string(),
  name: z.string(),
  unitType: z.string(),
  status: z.string(),
  level: z.number(),
  parentCode: z.string(), // 必填字段，根组织使用"0"
  path: z.string().optional(),
  sortOrder: z.number().nullable().optional(),
  description: z.string().nullable().optional(),
  createdAt: z.string().nullable().optional(),
  updatedAt: z.string().nullable().optional(),
});

// 统一验证工具函数 - 替代simple-validation.ts
export const ValidationUtils = {
  // 替代validateOrganizationBasic
  validateCreateInput: (data: unknown) => {
    const result = CreateOrganizationInputSchema.safeParse(data);
    return {
      isValid: result.success,
      errors: result.success ? [] : result.error.issues.map(e => ({
        field: e.path.join('.'),
        message: e.message
      }))
    };
  },

  // 替代validateOrganizationUpdate  
  validateUpdateInput: (data: unknown) => {
    const result = UpdateOrganizationInputSchema.safeParse(data);
    return {
      isValid: result.success,
      errors: result.success ? [] : result.error.issues.map(e => ({
        field: e.path.join('.'),
        message: e.message
      }))
    };
  },

  // 替代validateOrganizationResponse
  validateAPIResponse: (data: unknown) => {
    const result = GraphQLOrganizationResponseSchema.safeParse(data);
    return {
      isValid: result.success,
      errors: result.success ? [] : result.error.issues.map(e => ({
        field: e.path.join('.'),
        message: e.message
      }))
    };
  },

  // 替代validateForm (来自ValidationRules.ts)
  validateForm: (formData: unknown, isEditing = false) => {
    const schema = isEditing ? UpdateOrganizationInputSchema : CreateOrganizationInputSchema;
    const result = schema.safeParse(formData);
    
    if (result.success) return {};
    
    const errors: Record<string, string> = {};
    result.error.issues.forEach(e => {
      const field = e.path.join('.');
      errors[field] = e.message;
    });
    return errors;
  },

  // 时态验证工具 - 替代temporalValidation.ts
  temporal: {
    isValidDate: (dateString: string): boolean => {
      return TemporalDateSchema.safeParse(dateString).success;
    },
    
    isFutureDate: (dateString: string): boolean => {
      return FutureDateSchema.safeParse(dateString).success;
    },
    
    isEndDateAfterStartDate: (startDate: string, endDate: string): boolean => {
      return new Date(endDate) > new Date(startDate);
    },
    
    // DEPRECATED: 使用 TemporalConverter.getCurrentDateString()
    getTodayString: (): string => {
      return new Date().toISOString().split('T')[0];
    },
    
    // DEPRECATED: 使用 TemporalConverter.formatForDisplay()
    formatDateDisplay: (dateString: string): string => {
      if (!dateString) return '';
      const date = new Date(dateString);
      return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: 'long', 
        day: 'numeric'
      });
    }
  }
};

// 向后兼容导出 - 用于逐步迁移现有代码
export const validateCreateOrganizationInput = ValidationUtils.validateCreateInput;
export const validateUpdateOrganizationInput = ValidationUtils.validateUpdateInput;
export const validateOrganizationResponse = ValidationUtils.validateAPIResponse;
export const validateForm = ValidationUtils.validateForm;

// 导出推导类型
export type ValidatedOrganizationUnit = z.infer<typeof OrganizationUnitSchema>;
export type ValidatedCreateOrganizationInput = z.infer<typeof CreateOrganizationInputSchema>;
export type ValidatedCreateOrganizationResponse = z.infer<typeof CreateOrganizationResponseSchema>;
export type ValidatedUpdateOrganizationInput = z.infer<typeof UpdateOrganizationInputSchema>;
export type ValidatedGraphQLVariables = z.infer<typeof GraphQLVariablesSchema>;
export type ValidatedGraphQLOrganizationResponse = z.infer<typeof GraphQLOrganizationResponseSchema>;

// 验证错误类型
export interface ValidationError {
  field: string;
  message: string;
}

export interface ValidationResult {
  isValid: boolean;
  errors: ValidationError[];
}