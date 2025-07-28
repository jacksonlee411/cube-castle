/**
 * Unit tests for template library system
 * Tests enterprise template library, industry templates, and pattern templates
 */
import {
  EnterpriseTemplateLibrary,
  IndustryTemplates,
  TechnicalPatternTemplates,
  SecurityPatternTemplates
} from '@/lib/template-library';
import { TemplateCategory, TemplateComplexity } from '@/types/template';

describe('Template Library System', () => {
  beforeEach(() => {
    // Reset template library before each test
    EnterpriseTemplateLibrary.initialize();
  });

  describe('EnterpriseTemplateLibrary', () => {
    describe('Library Management', () => {
      it('should initialize with all predefined templates', async () => {
        const allTemplates = EnterpriseTemplateLibrary.getAllTemplates();
        
        expect(allTemplates.length).toBeGreaterThan(0);
        
        // Should contain industry templates
        expect(allTemplates.some(t => t.id === 'hr-employee-management')).toBe(true);
        expect(allTemplates.some(t => t.id === 'finance-account-management')).toBe(true);
        expect(allTemplates.some(t => t.id === 'ecommerce-product-catalog')).toBe(true);
        
        // Should contain technical pattern templates
        expect(allTemplates.some(t => t.id === 'technical-audit-trail')).toBe(true);
        expect(allTemplates.some(t => t.id === 'technical-soft-delete')).toBe(true);
        
        // Should contain security pattern templates
        expect(allTemplates.some(t => t.id === 'security-rbac-roles')).toBe(true);
      });

      it('should register new templates correctly', async () => {
        const initialCount = EnterpriseTemplateLibrary.getAllTemplates().length;
        
        const customTemplate = {
          id: 'custom-test-template',
          name: 'Test Template',
          description: 'Test template for unit testing',
          category: TemplateCategory.CUSTOM,
          complexity: TemplateComplexity.BASIC,
          version: '1.0.0',
          author: {
            id: 'test-author',
            name: 'Test Author',
            organization: 'Test Org',
            verified: false
          },
          createdAt: new Date(),
          updatedAt: new Date(),
          schema: {
            specification_version: '1.0',
            api_id: 'test-api',
            namespace: 'test',
            resource_name: 'test_resource',
            data_structure: {
              primary_key: 'id',
              data_classification: 'public',
              fields: []
            }
          },
          elements: [],
          tags: ['test'],
          keywords: ['test'],
          compatibility: {
            minSpecVersion: '1.0',
            supportedDatabases: ['postgresql'],
            supportedFrameworks: ['rest']
          },
          quality: {
            performanceScore: 80,
            securityScore: 80,
            maintainabilityScore: 80,
            bestPracticesScore: 80,
            communityRating: 4.0,
            usageCount: 0,
            lastValidated: new Date()
          },
          configurable: false
        };

        EnterpriseTemplateLibrary.registerTemplate(customTemplate);
        
        const afterCount = EnterpriseTemplateLibrary.getAllTemplates().length;
        expect(afterCount).toBe(initialCount + 1);
        
        const retrievedTemplate = EnterpriseTemplateLibrary.getTemplateById('custom-test-template');
        expect(retrievedTemplate).toBeDefined();
        expect(retrievedTemplate?.name).toBe('Test Template');
      });

      it('should retrieve template by ID correctly', async () => {
        const template = EnterpriseTemplateLibrary.getTemplateById('hr-employee-management');
        
        expect(template).toBeDefined();
        expect(template?.id).toBe('hr-employee-management');
        expect(template?.name).toBe('Employee Management System');
        expect(template?.category).toBe(TemplateCategory.HR_MANAGEMENT);
      });

      it('should return undefined for non-existent template ID', async () => {
        const template = EnterpriseTemplateLibrary.getTemplateById('non-existent-template');
        expect(template).toBeUndefined();
      });
    });

    describe('Template Filtering', () => {
      it('should filter templates by category', async () => {
        const hrTemplates = EnterpriseTemplateLibrary.getTemplatesByCategory(TemplateCategory.HR_MANAGEMENT);
        
        expect(hrTemplates.length).toBeGreaterThan(0);
        hrTemplates.forEach(template => {
          expect(template.category).toBe(TemplateCategory.HR_MANAGEMENT);
        });
      });

      it('should filter templates by complexity', async () => {
        const basicTemplates = EnterpriseTemplateLibrary.getTemplatesByComplexity(TemplateComplexity.BASIC);
        
        expect(basicTemplates.length).toBeGreaterThan(0);
        basicTemplates.forEach(template => {
          expect(template.complexity).toBe(TemplateComplexity.BASIC);
        });
      });

      it('should return empty array for non-existent category', async () => {
        const templates = EnterpriseTemplateLibrary.getTemplatesByCategory('non_existent_category' as any);
        expect(templates).toEqual([]);
      });
    });

    describe('Template Search', () => {
      it('should search templates by name', async () => {
        const results = EnterpriseTemplateLibrary.searchTemplates('Employee');
        
        expect(results.length).toBeGreaterThan(0);
        expect(results.some(t => t.name.toLowerCase().includes('employee'))).toBe(true);
      });

      it('should search templates by description', async () => {
        const results = EnterpriseTemplateLibrary.searchTemplates('management');
        
        expect(results.length).toBeGreaterThan(0);
        expect(results.some(t => t.description.toLowerCase().includes('management'))).toBe(true);
      });

      it('should search templates by tags', async () => {
        const results = EnterpriseTemplateLibrary.searchTemplates('hr');
        
        expect(results.length).toBeGreaterThan(0);
        expect(results.some(t => t.tags.includes('hr'))).toBe(true);
      });

      it('should search templates by keywords', async () => {
        const results = EnterpriseTemplateLibrary.searchTemplates('employee');
        
        expect(results.length).toBeGreaterThan(0);
        expect(results.some(t => t.keywords.includes('employee'))).toBe(true);
      });

      it('should return empty array for non-matching search', async () => {
        const results = EnterpriseTemplateLibrary.searchTemplates('nonexistentterm12345');
        expect(results).toEqual([]);
      });

      it('should be case insensitive', async () => {
        const lowerResults = EnterpriseTemplateLibrary.searchTemplates('employee');
        const upperResults = EnterpriseTemplateLibrary.searchTemplates('EMPLOYEE');
        const mixedResults = EnterpriseTemplateLibrary.searchTemplates('Employee');
        
        expect(lowerResults.length).toBe(upperResults.length);
        expect(upperResults.length).toBe(mixedResults.length);
      });
    });
  });

  describe('IndustryTemplates', () => {
    describe('Employee Management Template', () => {
      it('should create valid employee management template', async () => {
        const template = IndustryTemplates.createEmployeeManagementTemplate();
        
        expect(template.id).toBe('hr-employee-management');
        expect(template.name).toBe('Employee Management System');
        expect(template.category).toBe(TemplateCategory.HR_MANAGEMENT);
        expect(template.complexity).toBe(TemplateComplexity.INTERMEDIATE);
        
        // Check elements
        expect(template.elements.length).toBeGreaterThan(0);
        
        // Should have required fields
        const idField = template.elements.find(e => e.properties.name === 'id');
        const employeeIdField = template.elements.find(e => e.properties.name === 'employee_id');
        const emailField = template.elements.find(e => e.properties.name === 'email');
        
        expect(idField).toBeDefined();
        expect(idField?.properties.type).toBe('uuid');
        expect(idField?.properties.required).toBe(true);
        
        expect(employeeIdField).toBeDefined();
        expect(employeeIdField?.properties.unique).toBe(true);
        expect(employeeIdField?.properties.validation?.pattern).toBe('^EMP[0-9]{6}$');
        
        expect(emailField).toBeDefined();
        expect(emailField?.properties.type).toBe('email');
        expect(emailField?.properties.unique).toBe(true);
      });

      it('should have proper schema structure', async () => {
        const template = IndustryTemplates.createEmployeeManagementTemplate();
        
        expect(template.schema.specification_version).toBe('1.0');
        expect(template.schema.api_id).toBe('employee-management-api');
        expect(template.schema.namespace).toBe('hr');
        expect(template.schema.resource_name).toBe('employees');
        expect(template.schema.data_structure.primary_key).toBe('id');
        expect(template.schema.data_structure.data_classification).toBe('confidential');
      });

      it('should have security model', async () => {
        const template = IndustryTemplates.createEmployeeManagementTemplate();
        
        expect(template.schema.security_model).toBeDefined();
        expect(template.schema.security_model?.access_control).toBe('rbac');
        expect(template.schema.security_model?.audit_trail).toBe(true);
        expect(template.schema.security_model?.encryption).toContain('salary');
        expect(template.schema.security_model?.data_retention).toBe('7_years');
      });

      it('should be configurable with options', async () => {
        const template = IndustryTemplates.createEmployeeManagementTemplate();
        
        expect(template.configurable).toBe(true);
        expect(template.configOptions).toBeDefined();
        expect(template.configOptions?.length).toBeGreaterThan(0);
        
        const salaryOption = template.configOptions?.find(opt => opt.key === 'include_salary');
        expect(salaryOption).toBeDefined();
        expect(salaryOption?.type).toBe('boolean');
        expect(salaryOption?.defaultValue).toBe(true);
      });
    });

    describe('Financial Account Template', () => {
      it('should create valid financial account template', async () => {
        const template = IndustryTemplates.createFinancialAccountTemplate();
        
        expect(template.id).toBe('finance-account-management');
        expect(template.name).toBe('Financial Account Management');
        expect(template.category).toBe(TemplateCategory.FINANCIAL_SERVICES);
        expect(template.complexity).toBe(TemplateComplexity.ADVANCED);
        
        // Check critical financial fields
        const accountIdField = template.elements.find(e => e.properties.name === 'account_id');
        const balanceField = template.elements.find(e => e.properties.name === 'balance');
        const currencyField = template.elements.find(e => e.properties.name === 'currency');
        
        expect(accountIdField?.properties.validation?.pattern).toBe('^ACC[0-9]{10}$');
        expect(balanceField?.properties.type).toBe('decimal');
        expect(balanceField?.properties.precision).toBe(15);
        expect(balanceField?.properties.scale).toBe(2);
        expect(currencyField?.properties.validation?.pattern).toBe('^[A-Z]{3}$');
      });

      it('should have high security requirements', async () => {
        const template = IndustryTemplates.createFinancialAccountTemplate();
        
        expect(template.schema.data_structure.data_classification).toBe('restricted');
        expect(template.schema.security_model?.access_control).toBe('attribute_based');
        expect(template.schema.security_model?.encryption).toContain('balance');
        expect(template.schema.security_model?.encryption).toContain('account_id');
        expect(template.schema.security_model?.compliance).toContain('PCI_DSS');
        expect(template.schema.security_model?.compliance).toContain('SOX');
        expect(template.quality.securityScore).toBe(98);
      });
    });

    describe('E-commerce Product Template', () => {
      it('should create valid e-commerce product template', async () => {
        const template = IndustryTemplates.createEcommerceProductTemplate();
        
        expect(template.id).toBe('ecommerce-product-catalog');
        expect(template.name).toBe('E-commerce Product Catalog');
        expect(template.category).toBe(TemplateCategory.ECOMMERCE);
        expect(template.complexity).toBe(TemplateComplexity.BASIC);
        
        // Check product-specific fields
        const skuField = template.elements.find(e => e.properties.name === 'sku');
        const priceField = template.elements.find(e => e.properties.name === 'price');
        const inventoryField = template.elements.find(e => e.properties.name === 'inventory_count');
        
        expect(skuField?.properties.unique).toBe(true);
        expect(priceField?.properties.type).toBe('decimal');
        expect(priceField?.properties.validation?.min).toBe(0);
        expect(inventoryField?.properties.type).toBe('integer');
        expect(inventoryField?.properties.validation?.min).toBe(0);
      });

      it('should have public data classification', async () => {
        const template = IndustryTemplates.createEcommerceProductTemplate();
        
        expect(template.schema.data_structure.data_classification).toBe('public');
      });
    });
  });

  describe('TechnicalPatternTemplates', () => {
    describe('Audit Trail Template', () => {
      it('should create valid audit trail template', async () => {
        const template = TechnicalPatternTemplates.createAuditTrailTemplate();
        
        expect(template.id).toBe('technical-audit-trail');
        expect(template.name).toBe('Audit Trail System');
        expect(template.category).toBe(TemplateCategory.AUDIT_TRAIL);
        expect(template.complexity).toBe(TemplateComplexity.INTERMEDIATE);
        
        // Check audit-specific fields
        const auditIdField = template.elements.find(e => e.properties.name === 'audit_id');
        const actionField = template.elements.find(e => e.properties.name === 'action');
        const timestampField = template.elements.find(e => e.properties.name === 'timestamp');
        
        expect(auditIdField?.properties.type).toBe('uuid');
        expect(actionField?.properties.type).toBe('enum');
        expect(actionField?.properties.values).toContain('create');
        expect(actionField?.properties.values).toContain('read');
        expect(actionField?.properties.values).toContain('update');
        expect(actionField?.properties.values).toContain('delete');
        expect(timestampField?.properties.default).toBe('now()');
      });

      it('should have immutable security model', async () => {
        const template = TechnicalPatternTemplates.createAuditTrailTemplate();
        
        expect(template.schema.security_model?.immutable).toBe(true);
        expect(template.schema.security_model?.retention_policy).toBe('indefinite');
        expect(template.quality.securityScore).toBe(98);
      });
    });

    describe('Soft Delete Template', () => {
      it('should create valid soft delete template', async () => {
        const template = TechnicalPatternTemplates.createSoftDeleteTemplate();
        
        expect(template.id).toBe('technical-soft-delete');
        expect(template.name).toBe('Soft Delete Pattern');
        expect(template.category).toBe(TemplateCategory.SOFT_DELETE);
        expect(template.complexity).toBe(TemplateComplexity.BASIC);
        
        // Check soft delete fields
        const deletedAtField = template.elements.find(e => e.properties.name === 'deleted_at');
        const deletedByField = template.elements.find(e => e.properties.name === 'deleted_by');
        const deleteReasonField = template.elements.find(e => e.properties.name === 'delete_reason');
        
        expect(deletedAtField?.properties.type).toBe('datetime');
        expect(deletedAtField?.properties.nullable).toBe(true);
        expect(deletedByField?.properties.type).toBe('uuid');
        expect(deletedByField?.properties.nullable).toBe(true);
        expect(deleteReasonField?.properties.validation?.maxLength).toBe(500);
      });

      it('should have recovery-enabled business logic', async () => {
        const template = TechnicalPatternTemplates.createSoftDeleteTemplate();
        
        expect(template.schema.business_logic?.default_scope).toBe('WHERE deleted_at IS NULL');
        expect(template.schema.business_logic?.recovery_enabled).toBe(true);
      });
    });
  });

  describe('SecurityPatternTemplates', () => {
    describe('RBAC Template', () => {
      it('should create valid RBAC template', async () => {
        const template = SecurityPatternTemplates.createRBACTemplate();
        
        expect(template.id).toBe('security-rbac-roles');
        expect(template.name).toBe('RBAC Role Management');
        expect(template.category).toBe(TemplateCategory.RBAC);
        expect(template.complexity).toBe(TemplateComplexity.INTERMEDIATE);
        
        // Check RBAC-specific fields
        const roleIdField = template.elements.find(e => e.properties.name === 'role_id');
        const roleNameField = template.elements.find(e => e.properties.name === 'role_name');
        const permissionsField = template.elements.find(e => e.properties.name === 'permissions');
        const isSystemRoleField = template.elements.find(e => e.properties.name === 'is_system_role');
        
        expect(roleIdField?.properties.type).toBe('uuid');
        expect(roleNameField?.properties.unique).toBe(true);
        expect(permissionsField?.properties.type).toBe('json');
        expect(isSystemRoleField?.properties.type).toBe('boolean');
        expect(isSystemRoleField?.properties.default).toBe(false);
      });

      it('should have admin-only security model', async () => {
        const template = SecurityPatternTemplates.createRBACTemplate();
        
        expect(template.schema.security_model?.access_control).toBe('admin_only');
        expect(template.schema.security_model?.audit_trail).toBe(true);
        expect(template.schema.security_model?.immutable_system_roles).toBe(true);
        expect(template.compatibility.dependencies).toContain('user-management');
      });
    });
  });

  describe('Template Quality and Metadata', () => {
    it('should have consistent quality metrics across templates', async () => {
      const allTemplates = EnterpriseTemplateLibrary.getAllTemplates();
      
      allTemplates.forEach(template => {
        expect(template.quality.performanceScore).toBeGreaterThanOrEqual(0);
        expect(template.quality.performanceScore).toBeLessThanOrEqual(100);
        expect(template.quality.securityScore).toBeGreaterThanOrEqual(0);
        expect(template.quality.securityScore).toBeLessThanOrEqual(100);
        expect(template.quality.maintainabilityScore).toBeGreaterThanOrEqual(0);
        expect(template.quality.maintainabilityScore).toBeLessThanOrEqual(100);
        expect(template.quality.bestPracticesScore).toBeGreaterThanOrEqual(0);
        expect(template.quality.bestPracticesScore).toBeLessThanOrEqual(100);
        expect(template.quality.communityRating).toBeGreaterThanOrEqual(0);
        expect(template.quality.communityRating).toBeLessThanOrEqual(5);
      });
    });

    it('should have valid versioning', async () => {
      const allTemplates = EnterpriseTemplateLibrary.getAllTemplates();
      
      allTemplates.forEach(template => {
        expect(template.version).toMatch(/^\d+\.\d+\.\d+$/);
        expect(template.createdAt).toBeInstanceOf(Date);
        expect(template.updatedAt).toBeInstanceOf(Date);
        expect(template.author.id).toBeTruthy();
        expect(template.author.name).toBeTruthy();
      });
    });

    it('should have consistent compatibility information', async () => {
      const allTemplates = EnterpriseTemplateLibrary.getAllTemplates();
      
      allTemplates.forEach(template => {
        expect(template.compatibility.minSpecVersion).toBeTruthy();
        expect(template.compatibility.supportedDatabases).toBeInstanceOf(Array);
        expect(template.compatibility.supportedFrameworks).toBeInstanceOf(Array);
        expect(template.compatibility.supportedDatabases.length).toBeGreaterThan(0);
        expect(template.compatibility.supportedFrameworks.length).toBeGreaterThan(0);
      });
    });

    it('should have meaningful tags and keywords', async () => {
      const allTemplates = EnterpriseTemplateLibrary.getAllTemplates();
      
      allTemplates.forEach(template => {
        expect(template.tags).toBeInstanceOf(Array);
        expect(template.keywords).toBeInstanceOf(Array);
        expect(template.tags.length).toBeGreaterThan(0);
        expect(template.keywords.length).toBeGreaterThan(0);
        
        // Tags and keywords should be non-empty strings
        template.tags.forEach(tag => {
          expect(typeof tag).toBe('string');
          expect(tag.length).toBeGreaterThan(0);
        });
        
        template.keywords.forEach(keyword => {
          expect(typeof keyword).toBe('string');
          expect(keyword.length).toBeGreaterThan(0);
        });
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle empty search queries gracefully', async () => {
      const results = EnterpriseTemplateLibrary.searchTemplates('');
      expect(results).toEqual([]);
    });

    it('should handle whitespace-only search queries', async () => {
      const results = EnterpriseTemplateLibrary.searchTemplates('   ');
      expect(results).toEqual([]);
    });

    it('should handle special characters in search', async () => {
      const results = EnterpriseTemplateLibrary.searchTemplates('!@#$%^&*()');
      expect(results).toEqual([]);
    });
  });

  describe('Performance', () => {
    it('should initialize library quickly', async () => {
      const startTime = performance.now();
      EnterpriseTemplateLibrary.initialize();
      const endTime = performance.now();
      
      expect(endTime - startTime).toBeLessThan(100); // Should initialize in under 100ms
    });

    it('should search templates efficiently', async () => {
      const startTime = performance.now();
      EnterpriseTemplateLibrary.searchTemplates('employee');
      const endTime = performance.now();
      
      expect(endTime - startTime).toBeLessThan(50); // Should search in under 50ms
    });

    it('should handle large result sets efficiently', async () => {
      const startTime = performance.now();
      const allTemplates = EnterpriseTemplateLibrary.getAllTemplates();
      const endTime = performance.now();
      
      expect(endTime - startTime).toBeLessThan(10); // Should retrieve all in under 10ms
      expect(allTemplates.length).toBeGreaterThan(0);
    });
  });
});