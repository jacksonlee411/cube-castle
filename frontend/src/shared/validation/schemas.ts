import { z } from 'zod';

// 组织单元完整验证模式
export const OrganizationUnitSchema = z.object({
  code: z.string().regex(/^\d{7}$/, 'Organization code must be 7 digits'),
  name: z.string().min(1, 'Name is required').max(100, 'Name too long'),
  unitType: z.enum(['DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM']),
  status: z.enum(['ACTIVE', 'INACTIVE', 'PLANNED']),
  level: z.number().int().min(1).max(10),
  parentCode: z.string().regex(/^(0|\d{7})$/, 'Parent code must be "0" for root organizations or 7 digits for child organizations'),
  sortOrder: z.number().int().min(0).default(0),
  description: z.string().optional().or(z.literal('')),
  createdAt: z.string().datetime().optional().or(z.literal('')),
  updatedAt: z.string().datetime().optional().or(z.literal('')),
  path: z.string().optional().or(z.literal('')),
});

// 创建组织单元输入验证模式
export const CreateOrganizationInputSchema = z.object({
  code: z.string().regex(/^\d{7}$/).optional(), // 可选，由系统生成
  name: z.string().min(1, 'Name is required').max(100, 'Name too long'),
  unitType: z.enum(['DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM']),
  status: z.enum(['ACTIVE', 'INACTIVE', 'PLANNED']).default('ACTIVE'),
  level: z.number().int().min(1).max(10),
  parentCode: z.string().regex(/^(0|\d{7})$/, 'Parent code must be "0" for root organizations or 7 digits for child organizations'),
  sortOrder: z.number().int().min(0).default(0),
  description: z.string().optional().or(z.literal('')),
});

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

// 导出推导类型
export type ValidatedOrganizationUnit = z.infer<typeof OrganizationUnitSchema>;
export type ValidatedCreateOrganizationInput = z.infer<typeof CreateOrganizationInputSchema>;
export type ValidatedCreateOrganizationResponse = z.infer<typeof CreateOrganizationResponseSchema>;
export type ValidatedUpdateOrganizationInput = z.infer<typeof UpdateOrganizationInputSchema>;
export type ValidatedGraphQLVariables = z.infer<typeof GraphQLVariablesSchema>;
export type ValidatedGraphQLOrganizationResponse = z.infer<typeof GraphQLOrganizationResponseSchema>;