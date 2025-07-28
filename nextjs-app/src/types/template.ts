/**
 * 智能模板系统的核心类型定义
 * 支持企业级模板库、智能推荐、自定义管理等功能
 */

import { MetaContractElement, MetaContractSchema } from '@/components/metacontract-editor/VisualEditor';

// 模板分类枚举
export enum TemplateCategory {
  // 行业模板
  HR_MANAGEMENT = 'hr_management',
  FINANCIAL_SERVICES = 'financial_services', 
  HEALTHCARE = 'healthcare',
  ECOMMERCE = 'ecommerce',
  EDUCATION = 'education',
  MANUFACTURING = 'manufacturing',
  
  // 技术模式模板
  AUDIT_TRAIL = 'audit_trail',
  SOFT_DELETE = 'soft_delete',
  MULTI_TENANT = 'multi_tenant',
  VERSION_CONTROL = 'version_control',
  CACHING_STRATEGY = 'caching_strategy',
  EVENT_SOURCING = 'event_sourcing',
  
  // 安全模式模板  
  RBAC = 'rbac',
  DATA_ENCRYPTION = 'data_encryption',
  PII_PROTECTION = 'pii_protection',
  GDPR_COMPLIANCE = 'gdpr_compliance',
  OAUTH_INTEGRATION = 'oauth_integration',
  API_SECURITY = 'api_security',
  
  // 常用组件模板
  EMPLOYEE_PROFILE = 'employee_profile',
  ORGANIZATION = 'organization',
  PERMISSION_MANAGEMENT = 'permission_management',
  USER_MANAGEMENT = 'user_management',
  PRODUCT_CATALOG = 'product_catalog',
  ORDER_MANAGEMENT = 'order_management'
}

// 模板复杂度等级
export enum TemplateComplexity {
  BASIC = 'basic',       // 基础模板，5个字段以内
  INTERMEDIATE = 'intermediate', // 中等模板，5-15个字段  
  ADVANCED = 'advanced', // 高级模板，15-30个字段
  ENTERPRISE = 'enterprise' // 企业级模板，30+字段
}

// 模板兼容性标记
export interface TemplateCompatibility {
  // 最小/最大规范版本
  minSpecVersion: string;
  maxSpecVersion?: string;
  
  // 依赖的其他模板
  dependencies?: string[];
  
  // 冲突的模板
  conflicts?: string[];
  
  // 支持的数据库类型
  supportedDatabases?: string[];
  
  // 支持的框架
  supportedFrameworks?: string[];
}

// 模板质量指标
export interface TemplateQuality {
  // 性能评分 (0-100)
  performanceScore: number;
  
  // 安全评分 (0-100) 
  securityScore: number;
  
  // 可维护性评分 (0-100)
  maintainabilityScore: number;
  
  // 最佳实践符合度 (0-100)
  bestPracticesScore: number;
  
  // 社区评分 (0-5)
  communityRating: number;
  
  // 使用统计
  usageCount: number;
  
  // 最后验证时间
  lastValidated: Date;
}

// 模板作者信息
export interface TemplateAuthor {
  id: string;
  name: string;
  email?: string;
  organization?: string;
  verified: boolean; // 是否经过验证的作者
}

// 智能模板核心定义
export interface IntelligentTemplate {
  // 基本信息
  id: string;
  name: string;
  description: string;
  category: TemplateCategory;
  complexity: TemplateComplexity;
  version: string;
  
  // 作者和创建信息
  author: TemplateAuthor;
  createdAt: Date;
  updatedAt: Date;
  
  // 模板内容
  schema: MetaContractSchema;
  elements: MetaContractElement[];
  
  // 预览和文档
  previewImage?: string;
  documentation?: string;
  exampleUsage?: string;
  
  // 标签和搜索
  tags: string[];
  keywords: string[];
  
  // 兼容性和质量
  compatibility: TemplateCompatibility;
  quality: TemplateQuality;
  
  // 自定义配置
  configurable: boolean;
  configOptions?: TemplateConfigOption[];
  
  // 本地化支持
  localization?: {
    [locale: string]: {
      name: string;
      description: string;
      documentation?: string;
    };
  };
}

// 模板配置选项
export interface TemplateConfigOption {
  key: string;
  name: string;
  description: string;
  type: 'string' | 'number' | 'boolean' | 'select' | 'multiselect';
  required: boolean;
  defaultValue?: any;
  options?: Array<{ label: string; value: any }>;
  validation?: {
    min?: number;
    max?: number;
    pattern?: string;
    custom?: string; // JavaScript expression
  };
}

// 模板推荐上下文
export interface TemplateRecommendationContext {
  // 项目信息
  projectType?: string;
  industry?: string;
  teamSize?: number;
  
  // 现有元素分析
  existingElements: MetaContractElement[];
  existingCategories: TemplateCategory[];
  
  // 用户偏好
  userPreferences?: {
    complexity: TemplateComplexity[];
    categories: TemplateCategory[];
    author?: string;
    minRating?: number;
  };
  
  // 技术约束
  technicalConstraints?: {
    database?: string;
    framework?: string;
    specVersion?: string;
  };
}

// 模板推荐结果
export interface TemplateRecommendation {
  template: IntelligentTemplate;
  score: number; // 推荐得分 (0-100)
  reasons: string[]; // 推荐理由
  compatibility: 'perfect' | 'good' | 'partial' | 'poor';
  conflictRisk: 'none' | 'low' | 'medium' | 'high';
}

// 模板应用结果
export interface TemplateApplicationResult {
  success: boolean;
  template: IntelligentTemplate;
  appliedElements: MetaContractElement[];
  conflicts?: Array<{
    type: 'field' | 'relationship' | 'security' | 'validation';
    existing: MetaContractElement;
    template: MetaContractElement;
    resolution: 'merge' | 'replace' | 'rename' | 'skip';
  }>;
  warnings?: string[];
  performance?: {
    estimatedImpact: 'low' | 'medium' | 'high';
    metrics: Record<string, number>;
  };
}

// 模板集合（模板包）
export interface TemplateCollection {
  id: string;
  name: string;
  description: string;
  author: TemplateAuthor;
  templates: string[]; // template IDs
  category: TemplateCategory;
  version: string;
  createdAt: Date;
  updatedAt: Date;
  
  // 集合特有属性
  isBundle: boolean; // 是否为模板包
  installOrder?: string[]; // 安装顺序
  sharedConfiguration?: Record<string, any>;
}

// 模板搜索过滤器
export interface TemplateSearchFilter {
  query?: string;
  categories?: TemplateCategory[];
  complexity?: TemplateComplexity[];
  tags?: string[];
  author?: string;
  minRating?: number;
  maxRating?: number;
  compatibility?: {
    specVersion?: string;
    database?: string;
    framework?: string;
  };
  sortBy?: 'relevance' | 'rating' | 'usage' | 'recent' | 'name';
  sortOrder?: 'asc' | 'desc';
  limit?: number;
  offset?: number;
}

// 模板搜索结果
export interface TemplateSearchResult {
  templates: IntelligentTemplate[];
  total: number;
  facets: {
    categories: Array<{ category: TemplateCategory; count: number }>;
    complexity: Array<{ complexity: TemplateComplexity; count: number }>;
    tags: Array<{ tag: string; count: number }>;
    authors: Array<{ author: string; count: number }>;
  };
}

// 模板验证结果
export interface TemplateValidationResult {
  valid: boolean;
  errors: Array<{
    code: string;
    message: string;
    severity: 'error' | 'warning' | 'info';
    location?: {
      element?: string;
      field?: string;
      line?: number;
    };
  }>;
  performance?: {
    score: number;
    recommendations: string[];
  };
  security?: {
    score: number;
    vulnerabilities: Array<{
      type: string;
      severity: 'critical' | 'high' | 'medium' | 'low';
      description: string;
      mitigation?: string;
    }>;
  };
}

// 模板使用统计
export interface TemplateUsageStats {
  templateId: string;
  usageCount: number;
  lastUsed: Date;
  userCount: number;
  avgRating: number;
  ratingCount: number;
  
  // 使用趋势
  weeklyUsage: number[];
  monthlyUsage: number[];
  
  // 用户反馈统计
  feedback: {
    positive: number;
    negative: number;
    suggestions: string[];
  };
}

// 导出所有类型
export type {
  MetaContractElement,
  MetaContractSchema
};