/**
 * Unit tests for template application system
 * Tests template application engine, conflict detection, and resolution strategies
 */
import {
  IntelligentTemplateApplicationEngine,
  ConflictResolutionStrategy,
  FieldConflictType
} from '@/lib/template-application';
import { IndustryTemplates } from '@/lib/template-library';
import {
  IntelligentTemplate,
  MetaContractElement,
  TemplateApplicationConfig
} from '@/types/template';

describe('Template Application System', () => {
  let applicationEngine: IntelligentTemplateApplicationEngine;
  let mockTemplate: IntelligentTemplate;
  let mockExistingElements: MetaContractElement[];
  
  beforeEach(() => {
    applicationEngine = new IntelligentTemplateApplicationEngine();
    mockTemplate = IndustryTemplates.createEmployeeManagementTemplate();
    
    mockExistingElements = [
      {
        id: 'existing-field-1',
        type: 'field',
        name: 'id',
        properties: {
          name: 'id',
          type: 'uuid',
          required: true,
          description: 'Existing ID field'
        }
      },
      {
        id: 'existing-field-2',
        type: 'field',
        name: 'status',
        properties: {
          name: 'status',
          type: 'string',
          required: false,
          description: 'Existing status field'
        }
      }
    ];
  });

  describe('IntelligentTemplateApplicationEngine', () => {
    describe('Basic Template Application', () => {
      it('should apply template successfully with no conflicts', async () => {
        const nonConflictingElements: MetaContractElement[] = [
          {
            id: 'unique-field-1',
            type: 'field',
            name: 'unique_field',
            properties: {
              name: 'unique_field',
              type: 'string',
              description: 'A unique field'
            }
          }
        ];

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          nonConflictingElements
        );
        
        expect(result.success).toBe(true);
        expect(result.template).toBe(mockTemplate);
        expect(result.appliedElements.length).toBe(mockTemplate.elements.length);
        expect(result.error).toBeUndefined();
        expect(result.backup).toBeDefined();
        expect(result.backup).toEqual(nonConflictingElements);
      });

      it('should generate backup by default', async () => {
        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements
        );
        
        expect(result.backup).toBeDefined();
        expect(result.backup).toEqual(mockExistingElements);
        expect(result.backup).not.toBe(mockExistingElements); // Should be a copy
      });

      it('should skip backup when configured', async () => {
        const config: TemplateApplicationConfig = {
          generateBackup: false
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements,
          config
        );
        
        expect(result.backup).toBeUndefined();
      });

      it('should include performance impact assessment', async () => {
        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements
        );
        
        expect(result.performance).toBeDefined();
        expect(result.performance?.estimatedImpact).toMatch(/^(low|medium|high)$/);
        expect(result.performance?.metrics).toBeDefined();
        expect(result.performance?.metrics.elementCountChange).toBeDefined();
        expect(result.performance?.metrics.fieldCountChange).toBeDefined();
        expect(result.performance?.metrics.relationshipCountChange).toBeDefined();
        expect(result.performance?.metrics.complexityIncrease).toBeDefined();
      });

      it('should validate applied template by default', async () => {
        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements
        );
        
        expect(result.validationResults).toBeDefined();
        expect(result.validationResults.valid).toBe(true);
      });

      it('should skip validation when configured', async () => {
        const config: TemplateApplicationConfig = {
          validateAfterApply: false
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements,
          config
        );
        
        expect(result.validationResults).toEqual({});
      });
    });

    describe('Element Filtering', () => {
      it('should include only specified elements', async () => {
        const config: TemplateApplicationConfig = {
          includeElements: ['field-id', 'field-first_name']
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements,
          config
        );
        
        expect(result.success).toBe(true);
        expect(result.appliedElements.length).toBe(2);
        
        const appliedIds = result.appliedElements.map(e => e.id);
        expect(appliedIds).toContain('field-id');
        expect(appliedIds).toContain('field-first_name');
      });

      it('should exclude specified elements', async () => {
        const templateElementCount = mockTemplate.elements.length;
        const config: TemplateApplicationConfig = {
          excludeElements: ['field-salary', 'field-status']
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements,
          config
        );
        
        expect(result.success).toBe(true);
        expect(result.appliedElements.length).toBe(templateElementCount - 2);
        
        const appliedIds = result.appliedElements.map(e => e.id);
        expect(appliedIds).not.toContain('field-salary');
        expect(appliedIds).not.toContain('field-status');
      });

      it('should handle both include and exclude filters', async () => {
        const config: TemplateApplicationConfig = {
          includeElements: ['field-id', 'field-first_name', 'field-last_name', 'field-email'],
          excludeElements: ['field-email'] // Exclude from included set
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements,
          config
        );
        
        expect(result.success).toBe(true);
        expect(result.appliedElements.length).toBe(3);
        
        const appliedIds = result.appliedElements.map(e => e.id);
        expect(appliedIds).toContain('field-id');
        expect(appliedIds).toContain('field-first_name');
        expect(appliedIds).toContain('field-last_name');
        expect(appliedIds).not.toContain('field-email');
      });
    });

    describe('Conflict Detection', () => {
      it('should detect name conflicts', async () => {
        const conflictingElements: MetaContractElement[] = [
          {
            id: 'different-id',
            type: 'field',
            name: 'first_name', // Same name as template element
            properties: {
              name: 'first_name',
              type: 'text', // Different type
              required: false,
              description: 'Existing first name field'
            }
          }
        ];

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          conflictingElements
        );
        
        expect(result.success).toBe(true);
        expect(result.conflicts).toBeDefined();
        expect(result.conflicts!.length).toBeGreaterThan(0);
        
        // Should detect the name conflict
        const nameConflict = result.conflicts!.find(c => 
          c.existing.properties.name === 'first_name'
        );
        expect(nameConflict).toBeDefined();
      });

      it('should detect type mismatches', async () => {
        const typeMismatchElements: MetaContractElement[] = [
          {
            id: 'field-id', // Same ID as template
            type: 'field',
            name: 'id',
            properties: {
              name: 'id',
              type: 'integer', // Different type than template (uuid)
              required: true,
              description: 'Different type ID field'
            }
          }
        ];

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          typeMismatchElements
        );
        
        expect(result.success).toBe(true);
        expect(result.conflicts).toBeDefined();
        expect(result.conflicts!.length).toBeGreaterThan(0);
      });

      it('should detect constraint conflicts', async () => {
        const constraintConflictElements: MetaContractElement[] = [
          {
            id: 'different-id',
            type: 'field',
            name: 'email',
            properties: {
              name: 'email',
              type: 'email',
              required: false, // Template requires true
              unique: false,   // Template requires true
              description: 'Non-unique optional email'
            }
          }
        ];

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          constraintConflictElements
        );
        
        expect(result.success).toBe(true);
        expect(result.conflicts).toBeDefined();
        expect(result.conflicts!.length).toBeGreaterThan(0);
      });
    });

    describe('Conflict Resolution', () => {
      it('should resolve conflicts using merge strategy', async () => {
        const conflictingElements: MetaContractElement[] = [
          {
            id: 'existing-email',
            type: 'field',
            name: 'email',
            properties: {
              name: 'email',
              type: 'email',
              required: false,
              description: 'Existing email field'
            }
          }
        ];

        const config: TemplateApplicationConfig = {
          conflictResolution: {
            'name_conflict_field-email': ConflictResolutionStrategy.MERGE
          }
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          conflictingElements,
          config
        );
        
        expect(result.success).toBe(true);
        expect(result.warnings).toBeDefined();
      });

      it('should resolve conflicts using replace strategy', async () => {
        const conflictingElements: MetaContractElement[] = [
          {
            id: 'existing-field',
            type: 'field',
            name: 'status',
            properties: {
              name: 'status',
              type: 'boolean', // Different from template
              description: 'Existing status field'
            }
          }
        ];

        const config: TemplateApplicationConfig = {
          conflictResolution: {
            'name_conflict_field-status': ConflictResolutionStrategy.REPLACE
          }
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          conflictingElements,
          config
        );
        
        expect(result.success).toBe(true);
        expect(result.appliedElements.length).toBeGreaterThan(0);
      });

      it('should resolve conflicts using rename strategy', async () => {
        const conflictingElements: MetaContractElement[] = [
          {
            id: 'existing-name',
            type: 'field',
            name: 'first_name',
            properties: {
              name: 'first_name',
              type: 'text',
              description: 'Existing first name'
            }
          }
        ];

        const config: TemplateApplicationConfig = {
          conflictResolution: {
            'name_conflict_field-first_name': ConflictResolutionStrategy.RENAME
          }
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          conflictingElements,
          config
        );
        
        expect(result.success).toBe(true);
        expect(result.warnings).toBeDefined();
        expect(result.warnings!.some(w => w.includes('Renamed'))).toBe(true);
      });

      it('should resolve conflicts using skip strategy', async () => {
        const conflictingElements: MetaContractElement[] = [
          {
            id: 'existing-field',
            type: 'field',
            name: 'email',
            properties: {
              name: 'email',
              type: 'string',
              description: 'Existing email field'
            }
          }
        ];

        const config: TemplateApplicationConfig = {
          conflictResolution: {
            'name_conflict_field-email': ConflictResolutionStrategy.SKIP
          }
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          conflictingElements,
          config
        );
        
        expect(result.success).toBe(true);
        expect(result.warnings).toBeDefined();
        expect(result.warnings!.some(w => w.includes('Skipped'))).toBe(true);
        
        // Skipped element should not be in applied elements
        const emailElement = result.appliedElements.find(e => e.properties.name === 'email');
        expect(emailElement).toBeUndefined();
      });
    });

    describe('Merge Strategies', () => {
      it('should use additive merge strategy', async () => {
        const config: TemplateApplicationConfig = {
          mergeStrategy: 'additive',
          preserveExisting: true
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements,
          config
        );
        
        expect(result.success).toBe(true);
        
        // Should preserve existing elements and add template elements
        const resultIds = result.appliedElements.map(e => e.id);
        mockExistingElements.forEach(existing => {
          expect(resultIds).toContain(existing.id);
        });
      });

      it('should handle replace merge strategy', async () => {
        const config: TemplateApplicationConfig = {
          mergeStrategy: 'replace',
          preserveExisting: false
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements,
          config
        );
        
        expect(result.success).toBe(true);
        expect(result.appliedElements.length).toBeGreaterThan(0);
      });
    });

    describe('Template Preview', () => {
      it('should preview template application effects', async () => {
        const preview = await applicationEngine.previewTemplateApplication(
          mockTemplate,
          mockExistingElements
        );
        
        expect(preview).toBeDefined();
        expect(preview.conflicts).toBeInstanceOf(Array);
        expect(preview.previewElements).toBeInstanceOf(Array);
        expect(preview.impact).toBeDefined();
        
        expect(preview.impact.added).toBeGreaterThanOrEqual(0);
        expect(preview.impact.modified).toBeGreaterThanOrEqual(0);
        expect(preview.impact.conflicts).toBeGreaterThanOrEqual(0);
        expect(preview.impact.estimatedPerformanceImpact).toMatch(/^(low|medium|high)$/);
      });

      it('should detect conflicts in preview', async () => {
        const conflictingElements: MetaContractElement[] = [
          {
            id: 'conflict-field',
            type: 'field',
            name: 'first_name',
            properties: {
              name: 'first_name',
              type: 'text',
              description: 'Conflicting first name'
            }
          }
        ];

        const preview = await applicationEngine.previewTemplateApplication(
          mockTemplate,
          conflictingElements
        );
        
        expect(preview.conflicts.length).toBeGreaterThan(0);
        expect(preview.impact.conflicts).toBeGreaterThan(0);
        
        // Check conflict details
        preview.conflicts.forEach(conflict => {
          expect(conflict.id).toBeDefined();
          expect(conflict.type).toBeDefined();
          expect(conflict.existing).toBeDefined();
          expect(conflict.template).toBeDefined();
          expect(conflict.severity).toMatch(/^(low|medium|high|critical)$/);
          expect(conflict.description).toBeDefined();
          expect(conflict.suggestedResolution).toBeDefined();
          expect(conflict.resolutionOptions).toBeInstanceOf(Array);
        });
      });

      it('should calculate impact metrics correctly', async () => {
        const preview = await applicationEngine.previewTemplateApplication(
          mockTemplate,
          mockExistingElements
        );
        
        const totalAdded = preview.impact.added;
        const totalModified = preview.impact.modified;
        const totalConflicts = preview.impact.conflicts;
        
        expect(totalAdded).toBeGreaterThanOrEqual(0);
        expect(totalModified).toBeGreaterThanOrEqual(0);
        expect(totalConflicts).toBeGreaterThanOrEqual(0);
        
        // Impact should be reasonable
        expect(totalAdded + totalModified).toBeGreaterThan(0);
      });
    });

    describe('Rollback Functionality', () => {
      it('should rollback template application', async () => {
        const originalElements = JSON.parse(JSON.stringify(mockExistingElements));
        
        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements
        );
        
        expect(result.success).toBe(true);
        expect(result.backup).toBeDefined();
        
        const rolledBackElements = await applicationEngine.rollbackTemplateApplication(
          result.backup!
        );
        
        expect(rolledBackElements).toEqual(originalElements);
        expect(rolledBackElements).not.toBe(result.backup); // Should be a new copy
      });

      it('should throw error when no backup available', async () => {
        await expect(
          applicationEngine.rollbackTemplateApplication(undefined as any)
        ).rejects.toThrow('No backup available for rollback');
      });
    });

    describe('Application Reports', () => {
      it('should generate application report', async () => {
        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements
        );
        
        const report = applicationEngine.generateApplicationReport(result);
        
        expect(typeof report).toBe('string');
        expect(report.length).toBeGreaterThan(0);
        
        // Report should contain key information
        expect(report).toContain(mockTemplate.name);
        expect(report).toContain(mockTemplate.version);
        expect(report).toContain(result.success ? 'SUCCESS' : 'FAILED');
        expect(report).toContain(`Applied Elements: ${result.appliedElements.length}`);
        
        if (result.conflicts && result.conflicts.length > 0) {
          expect(report).toContain('Conflicts Resolved');
        }
        
        if (result.warnings && result.warnings.length > 0) {
          expect(report).toContain('Warnings');
        }
        
        if (result.performance) {
          expect(report).toContain('Performance Impact');
        }
      });

      it('should generate report for failed application', async () => {
        // Create a mock that will cause an error
        const invalidTemplate = {
          ...mockTemplate,
          elements: null as any // This should cause an error
        };

        const result = await applicationEngine.applyTemplate(
          invalidTemplate,
          mockExistingElements
        );
        
        expect(result.success).toBe(false);
        expect(result.error).toBeDefined();
        
        const report = applicationEngine.generateApplicationReport(result);
        
        expect(report).toContain('FAILED');
        expect(report).toContain(invalidTemplate.name);
      });
    });

    describe('Error Handling', () => {
      it('should handle template application errors gracefully', async () => {
        const invalidTemplate = {
          ...mockTemplate,
          elements: null as any
        };

        const result = await applicationEngine.applyTemplate(
          invalidTemplate,
          mockExistingElements
        );
        
        expect(result.success).toBe(false);
        expect(result.error).toBeDefined();
        expect(result.appliedElements).toEqual([]);
        expect(result.template).toBe(invalidTemplate);
      });

      it('should handle empty template elements', async () => {
        const emptyTemplate = {
          ...mockTemplate,
          elements: []
        };

        const result = await applicationEngine.applyTemplate(
          emptyTemplate,
          mockExistingElements
        );
        
        expect(result.success).toBe(true);
        expect(result.appliedElements).toEqual([]);
      });

      it('should handle empty existing elements', async () => {
        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          []
        );
        
        expect(result.success).toBe(true);
        expect(result.appliedElements.length).toBe(mockTemplate.elements.length);
      });

      it('should handle invalid conflict resolution config', async () => {
        const config: TemplateApplicationConfig = {
          conflictResolution: {
            'non-existent-conflict': ConflictResolutionStrategy.MERGE
          }
        };

        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements,
          config
        );
        
        expect(result.success).toBe(true);
        // Should not cause errors even with invalid conflict resolution keys
      });
    });

    describe('Performance Tests', () => {
      it('should apply template quickly', async () => {
        const startTime = performance.now();
        
        const result = await applicationEngine.applyTemplate(
          mockTemplate,
          mockExistingElements
        );
        
        const endTime = performance.now();
        
        expect(result.success).toBe(true);
        expect(endTime - startTime).toBeLessThan(1000); // Should complete within 1 second
      });

      it('should handle large templates efficiently', async () => {
        const largeTemplate = {
          ...mockTemplate,
          elements: Array.from({ length: 100 }, (_, i) => ({
            id: `large-field-${i}`,
            type: 'field' as const,
            name: `large_field_${i}`,
            properties: {
              name: `large_field_${i}`,
              type: 'string',
              description: `Large field ${i}`
            }
          }))
        };

        const startTime = performance.now();
        
        const result = await applicationEngine.applyTemplate(
          largeTemplate,
          mockExistingElements
        );
        
        const endTime = performance.now();
        
        expect(result.success).toBe(true);
        expect(result.appliedElements.length).toBe(100);
        expect(endTime - startTime).toBeLessThan(2000); // Should complete within 2 seconds
      });

      it('should preview template application quickly', async () => {
        const startTime = performance.now();
        
        const preview = await applicationEngine.previewTemplateApplication(
          mockTemplate,
          mockExistingElements
        );
        
        const endTime = performance.now();
        
        expect(preview).toBeDefined();
        expect(endTime - startTime).toBeLessThan(500); // Should complete within 500ms
      });
    });
  });
});