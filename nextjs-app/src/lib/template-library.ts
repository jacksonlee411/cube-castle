/**
 * 企业级模板库实现
 * 包含预定义的行业模板、技术模式模板、安全模式模板和常用组件模板
 */

import { 
  IntelligentTemplate, 
  TemplateCategory, 
  TemplateComplexity,
  TemplateAuthor,
  TemplateCompatibility,
  TemplateQuality,
  MetaContractElement,
  MetaContractSchema 
} from '@/types/template';

// 系统默认作者
const SYSTEM_AUTHOR: TemplateAuthor = {
  id: 'system',
  name: 'System Templates',
  organization: 'Cube Castle Platform',
  verified: true
};

// 默认兼容性配置
const DEFAULT_COMPATIBILITY: TemplateCompatibility = {
  minSpecVersion: '1.0',
  supportedDatabases: ['postgresql', 'mysql', 'mongodb', 'sqlite'],
  supportedFrameworks: ['rest', 'graphql', 'grpc']
};

// 生成默认质量指标
const generateDefaultQuality = (performanceScore = 85, securityScore = 90): TemplateQuality => ({
  performanceScore,
  securityScore,
  maintainabilityScore: 88,
  bestPracticesScore: 92,
  communityRating: 4.5,
  usageCount: 0,
  lastValidated: new Date()
});

/**
 * 行业模板库
 */
export class IndustryTemplates {
  // HR管理模板
  static createEmployeeManagementTemplate(): IntelligentTemplate {
    const elements: MetaContractElement[] = [
      {
        id: 'field-id',
        type: 'field',
        name: 'id',
        properties: {
          name: 'id',
          type: 'uuid',
          required: true,
          description: 'Unique identifier for employee'
        }
      },
      {
        id: 'field-employee_id',
        type: 'field', 
        name: 'employee_id',
        properties: {
          name: 'employee_id',
          type: 'string',
          required: true,
          unique: true,
          validation: { pattern: '^EMP[0-9]{6}$' },
          description: 'Employee ID with format EMP123456'
        }
      },
      {
        id: 'field-first_name',
        type: 'field',
        name: 'first_name', 
        properties: {
          name: 'first_name',
          type: 'string',
          required: true,
          validation: { maxLength: 50 },
          description: 'Employee first name'
        }
      },
      {
        id: 'field-last_name',
        type: 'field',
        name: 'last_name',
        properties: {
          name: 'last_name', 
          type: 'string',
          required: true,
          validation: { maxLength: 50 },
          description: 'Employee last name'
        }
      },
      {
        id: 'field-email',
        type: 'field',
        name: 'email',
        properties: {
          name: 'email',
          type: 'email',
          required: true,
          unique: true,
          validation: { format: 'email' },
          description: 'Employee email address'
        }
      },
      {
        id: 'field-department_id',
        type: 'field',
        name: 'department_id',
        properties: {
          name: 'department_id',
          type: 'uuid',
          required: true,
          description: 'Department identifier'
        }
      },
      {
        id: 'field-position',
        type: 'field',
        name: 'position',
        properties: {
          name: 'position',
          type: 'string',
          required: true,
          validation: { maxLength: 100 },
          description: 'Job position title'
        }
      },
      {
        id: 'field-hire_date',
        type: 'field',
        name: 'hire_date',
        properties: {
          name: 'hire_date',
          type: 'date',
          required: true,
          description: 'Date when employee was hired'
        }
      },
      {
        id: 'field-salary',
        type: 'field',
        name: 'salary',
        properties: {
          name: 'salary',
          type: 'decimal',
          precision: 10,
          scale: 2,
          required: false,
          sensitive: true,
          description: 'Employee salary (sensitive data)'
        }
      },
      {
        id: 'field-status',
        type: 'field',
        name: 'status',
        properties: {
          name: 'status',
          type: 'enum',
          values: ['active', 'inactive', 'terminated', 'on_leave'],
          default: 'active',
          description: 'Employee status'
        }
      },
      {
        id: 'relationship-department',
        type: 'relationship',
        name: 'department',
        properties: {
          name: 'department',
          type: 'belongs_to',
          target: 'departments',
          foreign_key: 'department_id',
          description: 'Employee belongs to a department'
        }
      }
    ];

    const schema: MetaContractSchema = {
      specification_version: '1.0',
      api_id: 'employee-management-api',
      namespace: 'hr',
      resource_name: 'employees',
      data_structure: {
        primary_key: 'id',
        data_classification: 'confidential',
        fields: elements.filter(e => e.type === 'field').map(e => e.properties)
      },
      relationships: elements.filter(e => e.type === 'relationship').map(e => e.properties),
      security_model: {
        access_control: 'rbac',
        encryption: ['salary'],
        audit_trail: true,
        data_retention: '7_years'
      }
    };

    return {
      id: 'hr-employee-management',
      name: 'Employee Management System',
      description: 'Comprehensive employee management template with HR best practices',
      category: TemplateCategory.HR_MANAGEMENT,
      complexity: TemplateComplexity.INTERMEDIATE,
      version: '1.0.0',
      author: SYSTEM_AUTHOR,
      createdAt: new Date('2024-01-01'),
      updatedAt: new Date(),
      schema,
      elements,
      tags: ['hr', 'employee', 'management', 'personnel'],
      keywords: ['employee', 'hr', 'staff', 'personnel', 'management'],
      compatibility: {
        ...DEFAULT_COMPATIBILITY,
        dependencies: ['organization-structure']
      },
      quality: generateDefaultQuality(88, 95),
      configurable: true,
      configOptions: [
        {
          key: 'include_salary',
          name: 'Include Salary Field',
          description: 'Whether to include salary information',
          type: 'boolean',
          required: false,
          defaultValue: true
        },
        {
          key: 'department_structure',
          name: 'Department Structure',
          description: 'Type of department organization',
          type: 'select',
          required: true,
          defaultValue: 'hierarchical',
          options: [
            { label: 'Flat Structure', value: 'flat' },
            { label: 'Hierarchical', value: 'hierarchical' },
            { label: 'Matrix', value: 'matrix' }
          ]
        }
      ]
    };
  }

  // 财务服务模板
  static createFinancialAccountTemplate(): IntelligentTemplate {
    const elements: MetaContractElement[] = [
      {
        id: 'field-account_id',
        type: 'field',
        name: 'account_id',
        properties: {
          name: 'account_id',
          type: 'string',
          required: true,
          unique: true,
          validation: { pattern: '^ACC[0-9]{10}$' },
          description: 'Account number with format ACC1234567890'
        }
      },
      {
        id: 'field-account_type',
        type: 'field',
        name: 'account_type',
        properties: {
          name: 'account_type',
          type: 'enum',
          values: ['checking', 'savings', 'credit', 'loan', 'investment'],
          required: true,
          description: 'Type of financial account'
        }
      },
      {
        id: 'field-balance',
        type: 'field',
        name: 'balance',
        properties: {
          name: 'balance',
          type: 'decimal',
          precision: 15,
          scale: 2,
          required: true,
          default: 0.00,
          description: 'Current account balance'
        }
      },
      {
        id: 'field-currency',
        type: 'field',
        name: 'currency',
        properties: {
          name: 'currency',
          type: 'string',
          required: true,
          validation: { pattern: '^[A-Z]{3}$' },
          default: 'USD',
          description: 'Currency code (ISO 4217)'
        }
      },
      {
        id: 'field-customer_id',
        type: 'field',
        name: 'customer_id',
        properties: {
          name: 'customer_id',
          type: 'uuid',
          required: true,
          description: 'Customer identifier'
        }
      }
    ];

    const schema: MetaContractSchema = {
      specification_version: '1.0',
      api_id: 'financial-account-api',
      namespace: 'finance',
      resource_name: 'accounts',
      data_structure: {
        primary_key: 'id',
        data_classification: 'restricted',
        fields: elements.map(e => e.properties)
      },
      security_model: {
        access_control: 'attribute_based',
        encryption: ['balance', 'account_id'],
        audit_trail: true,
        compliance: ['PCI_DSS', 'SOX'],
        data_retention: '10_years'
      }
    };

    return {
      id: 'finance-account-management',
      name: 'Financial Account Management',
      description: 'Secure financial account management with compliance features',
      category: TemplateCategory.FINANCIAL_SERVICES,
      complexity: TemplateComplexity.ADVANCED,
      version: '1.0.0',
      author: SYSTEM_AUTHOR,
      createdAt: new Date('2024-01-01'),
      updatedAt: new Date(),
      schema,
      elements,
      tags: ['finance', 'account', 'banking', 'money'],
      keywords: ['account', 'balance', 'finance', 'banking', 'money'],
      compatibility: {
        ...DEFAULT_COMPATIBILITY,
        dependencies: ['customer-management', 'audit-trail']
      },
      quality: generateDefaultQuality(85, 98),
      configurable: true
    };
  }

  // 电商产品模板
  static createEcommerceProductTemplate(): IntelligentTemplate {
    const elements: MetaContractElement[] = [
      {
        id: 'field-sku',
        type: 'field',
        name: 'sku',
        properties: {
          name: 'sku',
          type: 'string',
          required: true,
          unique: true,
          validation: { maxLength: 50 },
          description: 'Stock Keeping Unit identifier'
        }
      },
      {
        id: 'field-name',
        type: 'field',
        name: 'name',
        properties: {
          name: 'name',
          type: 'string',
          required: true,
          validation: { maxLength: 200 },
          description: 'Product name'
        }
      },
      {
        id: 'field-description',
        type: 'field',
        name: 'description',
        properties: {
          name: 'description',
          type: 'text',
          required: false,
          description: 'Product description'
        }
      },
      {
        id: 'field-price',
        type: 'field',
        name: 'price',
        properties: {
          name: 'price',
          type: 'decimal',
          precision: 10,
          scale: 2,
          required: true,
          validation: { min: 0 },
          description: 'Product price'
        }
      },
      {
        id: 'field-category_id',
        type: 'field',
        name: 'category_id',
        properties: {
          name: 'category_id',
          type: 'uuid',
          required: true,
          description: 'Product category identifier'
        }
      },
      {
        id: 'field-inventory_count',
        type: 'field',
        name: 'inventory_count',
        properties: {
          name: 'inventory_count',
          type: 'integer',
          required: true,
          default: 0,
          validation: { min: 0 },
          description: 'Available inventory count'
        }
      }
    ];

    const schema: MetaContractSchema = {
      specification_version: '1.0',
      api_id: 'ecommerce-product-api',
      namespace: 'ecommerce',
      resource_name: 'products',
      data_structure: {
        primary_key: 'id',
        data_classification: 'public',
        fields: elements.map(e => e.properties)
      }
    };

    return {
      id: 'ecommerce-product-catalog',
      name: 'E-commerce Product Catalog',
      description: 'Product catalog template for e-commerce applications',
      category: TemplateCategory.ECOMMERCE,
      complexity: TemplateComplexity.BASIC,
      version: '1.0.0',
      author: SYSTEM_AUTHOR,
      createdAt: new Date('2024-01-01'),
      updatedAt: new Date(),
      schema,
      elements,
      tags: ['ecommerce', 'product', 'catalog', 'inventory'],
      keywords: ['product', 'sku', 'price', 'inventory', 'catalog'],
      compatibility: DEFAULT_COMPATIBILITY,
      quality: generateDefaultQuality(90, 85),
      configurable: true
    };
  }
}

/**
 * 技术模式模板库
 */
export class TechnicalPatternTemplates {
  // 审计追踪模板
  static createAuditTrailTemplate(): IntelligentTemplate {
    const elements: MetaContractElement[] = [
      {
        id: 'field-audit_id',
        type: 'field',
        name: 'audit_id',
        properties: {
          name: 'audit_id',
          type: 'uuid',
          required: true,
          description: 'Unique audit log identifier'
        }
      },
      {
        id: 'field-entity_type',
        type: 'field',
        name: 'entity_type',
        properties: {
          name: 'entity_type',
          type: 'string',
          required: true,
          validation: { maxLength: 100 },
          description: 'Type of entity being audited'
        }
      },
      {
        id: 'field-entity_id',
        type: 'field',
        name: 'entity_id',
        properties: {
          name: 'entity_id',
          type: 'string',
          required: true,
          description: 'Identifier of the audited entity'
        }
      },
      {
        id: 'field-action',
        type: 'field',
        name: 'action',
        properties: {
          name: 'action',
          type: 'enum',
          values: ['create', 'read', 'update', 'delete', 'login', 'logout'],
          required: true,
          description: 'Action performed on the entity'
        }
      },
      {
        id: 'field-user_id',
        type: 'field',
        name: 'user_id',
        properties: {
          name: 'user_id',
          type: 'uuid',
          required: true,
          description: 'User who performed the action'
        }
      },
      {
        id: 'field-timestamp',
        type: 'field',
        name: 'timestamp',
        properties: {
          name: 'timestamp',
          type: 'datetime',
          required: true,
          default: 'now()',
          description: 'When the action occurred'
        }
      },
      {
        id: 'field-old_values',
        type: 'field',
        name: 'old_values',
        properties: {
          name: 'old_values',
          type: 'json',
          required: false,
          description: 'Previous values before change'
        }
      },
      {
        id: 'field-new_values',
        type: 'field',
        name: 'new_values',
        properties: {
          name: 'new_values',
          type: 'json',
          required: false,
          description: 'New values after change'
        }
      }
    ];

    const schema: MetaContractSchema = {
      specification_version: '1.0',
      api_id: 'audit-trail-api',
      namespace: 'system',
      resource_name: 'audit_logs',
      data_structure: {
        primary_key: 'audit_id',
        data_classification: 'internal',
        fields: elements.map(e => e.properties)
      },
      security_model: {
        access_control: 'strict',
        immutable: true,
        retention_policy: 'indefinite'
      }
    };

    return {
      id: 'technical-audit-trail',
      name: 'Audit Trail System',
      description: 'Comprehensive audit trail template for compliance and security',
      category: TemplateCategory.AUDIT_TRAIL,
      complexity: TemplateComplexity.INTERMEDIATE,
      version: '1.0.0',
      author: SYSTEM_AUTHOR,
      createdAt: new Date('2024-01-01'),
      updatedAt: new Date(),
      schema,
      elements,
      tags: ['audit', 'trail', 'compliance', 'security', 'logging'],
      keywords: ['audit', 'log', 'trail', 'compliance', 'tracking'],
      compatibility: DEFAULT_COMPATIBILITY,
      quality: generateDefaultQuality(92, 98),
      configurable: true
    };
  }

  // 软删除模板
  static createSoftDeleteTemplate(): IntelligentTemplate {
    const elements: MetaContractElement[] = [
      {
        id: 'field-deleted_at',
        type: 'field',
        name: 'deleted_at',
        properties: {
          name: 'deleted_at',
          type: 'datetime',
          required: false,
          nullable: true,
          description: 'Timestamp when record was soft deleted'
        }
      },
      {
        id: 'field-deleted_by',
        type: 'field',
        name: 'deleted_by',
        properties: {
          name: 'deleted_by',
          type: 'uuid',
          required: false,
          nullable: true,
          description: 'User who performed the soft delete'
        }
      },
      {
        id: 'field-delete_reason',
        type: 'field',
        name: 'delete_reason',
        properties: {
          name: 'delete_reason',
          type: 'string',
          required: false,
          nullable: true,
          validation: { maxLength: 500 },
          description: 'Reason for deletion'
        }
      }
    ];

    const schema: MetaContractSchema = {
      specification_version: '1.0',
      api_id: 'soft-delete-api',
      namespace: 'system',
      resource_name: 'soft_deletable',
      data_structure: {
        primary_key: 'id',
        data_classification: 'internal',
        fields: elements.map(e => e.properties)
      },
      business_logic: {
        default_scope: 'WHERE deleted_at IS NULL',
        recovery_enabled: true
      }
    };

    return {
      id: 'technical-soft-delete',
      name: 'Soft Delete Pattern',
      description: 'Soft delete implementation with recovery capability',
      category: TemplateCategory.SOFT_DELETE,
      complexity: TemplateComplexity.BASIC,
      version: '1.0.0',
      author: SYSTEM_AUTHOR,
      createdAt: new Date('2024-01-01'),
      updatedAt: new Date(),
      schema,
      elements,
      tags: ['soft-delete', 'recovery', 'data-protection'],
      keywords: ['delete', 'soft', 'recovery', 'restore'],
      compatibility: DEFAULT_COMPATIBILITY,
      quality: generateDefaultQuality(95, 90),
      configurable: true
    };
  }
}

/**
 * 安全模式模板库
 */
export class SecurityPatternTemplates {
  // RBAC权限控制模板
  static createRBACTemplate(): IntelligentTemplate {
    const elements: MetaContractElement[] = [
      {
        id: 'field-role_id',
        type: 'field',
        name: 'role_id',
        properties: {
          name: 'role_id',
          type: 'uuid',
          required: true,
          description: 'Role identifier'
        }
      },
      {
        id: 'field-role_name',
        type: 'field',
        name: 'role_name',
        properties: {
          name: 'role_name',
          type: 'string',
          required: true,
          unique: true,
          validation: { maxLength: 100 },
          description: 'Role name'
        }
      },
      {
        id: 'field-permissions',
        type: 'field',
        name: 'permissions',
        properties: {
          name: 'permissions',
          type: 'json',
          required: true,
          description: 'Array of permissions assigned to role'
        }
      },
      {
        id: 'field-description',
        type: 'field',
        name: 'description',
        properties: {
          name: 'description',
          type: 'text',
          required: false,
          description: 'Role description'
        }
      },
      {
        id: 'field-is_system_role',
        type: 'field',
        name: 'is_system_role',
        properties: {
          name: 'is_system_role',
          type: 'boolean',
          required: true,
          default: false,
          description: 'Whether this is a system-defined role'
        }
      }
    ];

    const schema: MetaContractSchema = {
      specification_version: '1.0',
      api_id: 'rbac-role-api',
      namespace: 'security',
      resource_name: 'roles',
      data_structure: {
        primary_key: 'role_id',
        data_classification: 'confidential',
        fields: elements.map(e => e.properties)
      },
      security_model: {
        access_control: 'admin_only',
        audit_trail: true,
        immutable_system_roles: true
      }
    };

    return {
      id: 'security-rbac-roles',
      name: 'RBAC Role Management',
      description: 'Role-based access control template with permission management',
      category: TemplateCategory.RBAC,
      complexity: TemplateComplexity.INTERMEDIATE,
      version: '1.0.0',
      author: SYSTEM_AUTHOR,
      createdAt: new Date('2024-01-01'),
      updatedAt: new Date(),
      schema,
      elements,
      tags: ['rbac', 'roles', 'permissions', 'security', 'access-control'],
      keywords: ['role', 'permission', 'access', 'control', 'security'],
      compatibility: {
        ...DEFAULT_COMPATIBILITY,
        dependencies: ['user-management']
      },
      quality: generateDefaultQuality(88, 98),
      configurable: true
    };
  }
}

/**
 * 企业级模板库管理器
 */
export class EnterpriseTemplateLibrary {
  private static templates: Map<string, IntelligentTemplate> = new Map();

  // 初始化所有预定义模板
  static initialize(): void {
    // 行业模板
    this.registerTemplate(IndustryTemplates.createEmployeeManagementTemplate());
    this.registerTemplate(IndustryTemplates.createFinancialAccountTemplate());
    this.registerTemplate(IndustryTemplates.createEcommerceProductTemplate());
    
    // 技术模式模板
    this.registerTemplate(TechnicalPatternTemplates.createAuditTrailTemplate());
    this.registerTemplate(TechnicalPatternTemplates.createSoftDeleteTemplate());
    
    // 安全模式模板
    this.registerTemplate(SecurityPatternTemplates.createRBACTemplate());
  }

  // 注册模板
  static registerTemplate(template: IntelligentTemplate): void {
    this.templates.set(template.id, template);
  }

  // 获取所有模板
  static getAllTemplates(): IntelligentTemplate[] {
    return Array.from(this.templates.values());
  }

  // 根据分类获取模板
  static getTemplatesByCategory(category: TemplateCategory): IntelligentTemplate[] {
    return this.getAllTemplates().filter(t => t.category === category);
  }

  // 根据复杂度获取模板
  static getTemplatesByComplexity(complexity: TemplateComplexity): IntelligentTemplate[] {
    return this.getAllTemplates().filter(t => t.complexity === complexity);
  }

  // 根据ID获取模板
  static getTemplateById(id: string): IntelligentTemplate | undefined {
    return this.templates.get(id);
  }

  // 搜索模板
  static searchTemplates(query: string): IntelligentTemplate[] {
    const lowercaseQuery = query.toLowerCase();
    return this.getAllTemplates().filter(template => 
      template.name.toLowerCase().includes(lowercaseQuery) ||
      template.description.toLowerCase().includes(lowercaseQuery) ||
      template.tags.some(tag => tag.toLowerCase().includes(lowercaseQuery)) ||
      template.keywords.some(keyword => keyword.toLowerCase().includes(lowercaseQuery))
    );
  }
}

// 初始化模板库
EnterpriseTemplateLibrary.initialize();