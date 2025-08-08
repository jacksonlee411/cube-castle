import { z } from 'zod';

// 组织单元完整验证模式
export const OrganizationUnitSchema = z.object({
  code: z.string().regex(/^\d{7}$/, 'Organization code must be 7 digits'),
  name: z.string().min(1, 'Name is required').max(100, 'Name too long'),
  unit_type: z.enum(['DEPARTMENT', 'COST_CENTER', 'COMPANY', 'PROJECT_TEAM']),
  status: z.enum(['ACTIVE', 'INACTIVE', 'PLANNED']),
  level: z.number().int().min(1).max(10),
  parent_code: z.string().regex(/^\d{7}$/).optional().or(z.literal('')),
  sort_order: z.number().int().min(0).default(0),
  description: z.string().optional().or(z.literal('')),
  created_at: z.string().datetime().optional().or(z.literal('')),
  updated_at: z.string().datetime().optional().or(z.literal('')),
  path: z.string().optional().or(z.literal('')),
});

// 创建组织单元输入验证模式
export const CreateOrganizationInputSchema = z.object({
  code: z.string().regex(/^\d{7}$/).optional(), // 可选，由系统生成
  name: z.string().min(1, 'Name is required').max(100, 'Name too long'),
  unit_type: z.enum(['DEPARTMENT', 'COST_CENTER', 'COMPANY', 'PROJECT_TEAM']),
  status: z.enum(['ACTIVE', 'INACTIVE', 'PLANNED']).default('ACTIVE'),
  level: z.number().int().min(1).max(10),
  parent_code: z.string().regex(/^\d{7}$/).optional().or(z.literal('')),
  sort_order: z.number().int().min(0).default(0),
  description: z.string().optional().or(z.literal('')),
});

// 更新组织单元输入验证模式
export const UpdateOrganizationInputSchema = CreateOrganizationInputSchema.partial().extend({
  code: z.string().regex(/^\d{7}$/, 'Organization code must be 7 digits'), // 更新时code必需
});

// GraphQL查询变量验证模式
export const GraphQLVariablesSchema = z.object({
  searchText: z.string().optional(),
  unitType: z.enum(['DEPARTMENT', 'COST_CENTER', 'COMPANY', 'PROJECT_TEAM']).optional(),
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
  parentCode: z.string().nullable().optional(),
  path: z.string().optional(),
  sortOrder: z.number().nullable().optional(),
  description: z.string().nullable().optional(),
  createdAt: z.string().nullable().optional(),
  updatedAt: z.string().nullable().optional(),
});

// 导出推导类型
export type ValidatedOrganizationUnit = z.infer<typeof OrganizationUnitSchema>;
export type ValidatedCreateOrganizationInput = z.infer<typeof CreateOrganizationInputSchema>;
export type ValidatedUpdateOrganizationInput = z.infer<typeof UpdateOrganizationInputSchema>;
export type ValidatedGraphQLVariables = z.infer<typeof GraphQLVariablesSchema>;
export type ValidatedGraphQLOrganizationResponse = z.infer<typeof GraphQLOrganizationResponseSchema>;