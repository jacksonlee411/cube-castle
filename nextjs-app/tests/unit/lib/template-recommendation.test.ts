/**
 * Unit tests for template recommendation system
 * Tests intelligent recommendation engine, strategies, and search functionality
 */
import {
  IntelligentTemplateRecommendationEngine
} from '@/lib/template-recommendation';
import { EnterpriseTemplateLibrary } from '@/lib/template-library';
import {
  TemplateRecommendationContext,
  TemplateCategory,
  TemplateComplexity,
  MetaContractElement,
  TemplateSearchFilter
} from '@/types/template';

describe('Template Recommendation System', () => {
  let recommendationEngine: IntelligentTemplateRecommendationEngine;
  
  beforeEach(() => {
    // Initialize template library
    EnterpriseTemplateLibrary.initialize();
    recommendationEngine = new IntelligentTemplateRecommendationEngine();
  });

  describe('IntelligentTemplateRecommendationEngine', () => {
    describe('Recommendation Generation', () => {
      it('should generate recommendations based on context', async () => {
        const mockExistingElements: MetaContractElement[] = [
          {
            id: 'existing-field-1',
            type: 'field',
            name: 'employee_id',
            properties: {
              name: 'employee_id',
              type: 'string',
              required: true,
              description: 'Employee identifier'
            }
          },
          {
            id: 'existing-field-2',
            type: 'field',
            name: 'name',
            properties: {
              name: 'name',
              type: 'string',
              required: true,
              description: 'Employee name'
            }
          }
        ];

        const context: TemplateRecommendationContext = {
          existingElements: mockExistingElements,
          existingCategories: [TemplateCategory.HR_MANAGEMENT],
          projectType: 'hr',
          industry: 'business',
          teamSize: 5,
          userPreferences: {
            complexity: [TemplateComplexity.INTERMEDIATE],
            categories: [TemplateCategory.HR_MANAGEMENT]
          }
        };

        const recommendations = await recommendationEngine.getRecommendations(context, 5);
        
        expect(recommendations).toBeDefined();
        expect(recommendations.length).toBeGreaterThan(0);
        expect(recommendations.length).toBeLessThanOrEqual(5);
        
        // Check recommendation structure
        recommendations.forEach(rec => {
          expect(rec.template).toBeDefined();
          expect(rec.score).toBeGreaterThanOrEqual(0);
          expect(rec.score).toBeLessThanOrEqual(100);
          expect(rec.reasons).toBeInstanceOf(Array);
          expect(rec.compatibility).toMatch(/^(perfect|good|partial|poor)$/);
          expect(rec.conflictRisk).toMatch(/^(none|low|medium|high)$/);
        });
        
        // Recommendations should be sorted by score (highest first)
        for (let i = 1; i < recommendations.length; i++) {
          expect(recommendations[i - 1].score).toBeGreaterThanOrEqual(recommendations[i].score);
        }
      });

      it('should prioritize HR templates for HR context', async () => {
        const context: TemplateRecommendationContext = {
          existingElements: [
            {
              id: 'hr-field',
              type: 'field',
              name: 'employee_name',
              properties: { name: 'employee_name', type: 'string' }
            }
          ],
          existingCategories: [TemplateCategory.HR_MANAGEMENT],
          projectType: 'hr',
          industry: 'business'
        };

        const recommendations = await recommendationEngine.getRecommendations(context, 10);
        
        expect(recommendations.length).toBeGreaterThan(0);
        
        // Should have HR-related templates in top recommendations
        const topRecommendations = recommendations.slice(0, 3);
        const hasHRTemplate = topRecommendations.some(rec => 
          rec.template.category === TemplateCategory.HR_MANAGEMENT ||
          rec.template.tags.includes('hr') ||
          rec.template.keywords.includes('employee')
        );
        
        expect(hasHRTemplate).toBe(true);
      });

      it('should handle empty existing elements', async () => {
        const context: TemplateRecommendationContext = {
          existingElements: [],
          existingCategories: [],
          projectType: 'general'
        };

        const recommendations = await recommendationEngine.getRecommendations(context, 5);
        
        expect(recommendations).toBeDefined();
        expect(recommendations.length).toBeGreaterThan(0);
        
        // Should still provide recommendations based on other factors
        recommendations.forEach(rec => {
          expect(rec.score).toBeGreaterThan(0);
        });
      });

      it('should limit recommendations to specified number', async () => {
        const context: TemplateRecommendationContext = {
          existingElements: [],
          existingCategories: []
        };

        const recommendations3 = await recommendationEngine.getRecommendations(context, 3);
        const recommendations10 = await recommendationEngine.getRecommendations(context, 10);
        
        expect(recommendations3.length).toBeLessThanOrEqual(3);
        expect(recommendations10.length).toBeLessThanOrEqual(10);
        
        if (recommendations3.length === 3 && recommendations10.length >= 3) {
          // Top 3 should be the same in both results
          for (let i = 0; i < 3; i++) {
            expect(recommendations3[i].template.id).toBe(recommendations10[i].template.id);
          }
        }
      });

      it('should generate meaningful recommendation reasons', async () => {
        const context: TemplateRecommendationContext = {
          existingElements: [
            {
              id: 'user-field',
              type: 'field',
              name: 'user_id',
              properties: { name: 'user_id', type: 'uuid' }
            }
          ],
          existingCategories: [TemplateCategory.USER_MANAGEMENT],
          userPreferences: {
            complexity: [TemplateComplexity.BASIC, TemplateComplexity.INTERMEDIATE]
          }
        };

        const recommendations = await recommendationEngine.getRecommendations(context, 5);
        
        expect(recommendations.length).toBeGreaterThan(0);
        
        recommendations.forEach(rec => {
          expect(rec.reasons).toBeDefined();
          expect(rec.reasons.length).toBeGreaterThan(0);
          expect(rec.reasons.length).toBeLessThanOrEqual(3);
          
          // Reasons should be meaningful strings
          rec.reasons.forEach(reason => {
            expect(typeof reason).toBe('string');
            expect(reason.length).toBeGreaterThan(0);
          });
        });
      });

      it('should assess conflict risk correctly', async () => {
        const conflictingElements: MetaContractElement[] = [
          {
            id: 'conflict-field',
            type: 'field',
            name: 'id', // This will conflict with template ID fields
            properties: { name: 'id', type: 'integer' }
          },
          {
            id: 'conflict-field-2',
            type: 'field',
            name: 'name',
            properties: { name: 'name', type: 'text' }
          }
        ];

        const context: TemplateRecommendationContext = {
          existingElements: conflictingElements,
          existingCategories: [TemplateCategory.HR_MANAGEMENT]
        };

        const recommendations = await recommendationEngine.getRecommendations(context, 5);
        
        expect(recommendations.length).toBeGreaterThan(0);
        
        // Some recommendations should have conflict risk detected
        const hasConflictRisk = recommendations.some(rec => rec.conflictRisk !== 'none');
        expect(hasConflictRisk).toBe(true);
      });
    });

    describe('Template Search', () => {
      it('should search templates with basic filter', async () => {
        const filter: TemplateSearchFilter = {
          query: 'employee',
          limit: 5,
          offset: 0
        };

        const result = await recommendationEngine.searchTemplates(filter);
        
        expect(result).toBeDefined();
        expect(result.templates).toBeInstanceOf(Array);
        expect(result.total).toBeGreaterThanOrEqual(result.templates.length);
        expect(result.facets).toBeDefined();
        
        // Should find employee-related templates
        const hasEmployeeTemplate = result.templates.some(t => 
          t.name.toLowerCase().includes('employee') ||
          t.description.toLowerCase().includes('employee') ||
          t.tags.includes('employee') ||
          t.keywords.includes('employee')
        );
        expect(hasEmployeeTemplate).toBe(true);
      });

      it('should filter by categories', async () => {
        const filter: TemplateSearchFilter = {
          categories: [TemplateCategory.HR_MANAGEMENT, TemplateCategory.FINANCIAL_SERVICES],
          limit: 10
        };

        const result = await recommendationEngine.searchTemplates(filter);
        
        expect(result.templates.length).toBeGreaterThan(0);
        
        // All templates should match specified categories
        result.templates.forEach(template => {
          expect(filter.categories).toContain(template.category);
        });
      });

      it('should filter by complexity', async () => {
        const filter: TemplateSearchFilter = {
          complexity: [TemplateComplexity.BASIC, TemplateComplexity.INTERMEDIATE],
          limit: 10
        };

        const result = await recommendationEngine.searchTemplates(filter);
        
        expect(result.templates.length).toBeGreaterThan(0);
        
        // All templates should match specified complexity levels
        result.templates.forEach(template => {
          expect(filter.complexity).toContain(template.complexity);
        });
      });

      it('should filter by tags', async () => {
        const filter: TemplateSearchFilter = {
          tags: ['hr', 'security'],
          limit: 10
        };

        const result = await recommendationEngine.searchTemplates(filter);
        
        expect(result.templates.length).toBeGreaterThan(0);
        
        // All templates should have at least one of the specified tags
        result.templates.forEach(template => {
          const hasMatchingTag = filter.tags!.some(tag => template.tags.includes(tag));
          expect(hasMatchingTag).toBe(true);
        });
      });

      it('should filter by rating range', async () => {
        const filter: TemplateSearchFilter = {
          minRating: 4.0,
          maxRating: 5.0,
          limit: 10
        };

        const result = await recommendationEngine.searchTemplates(filter);
        
        // All templates should be within rating range
        result.templates.forEach(template => {
          expect(template.quality.communityRating).toBeGreaterThanOrEqual(4.0);
          expect(template.quality.communityRating).toBeLessThanOrEqual(5.0);
        });
      });

      it('should sort templates correctly', async () => {
        const ratingFilter: TemplateSearchFilter = {
          sortBy: 'rating',
          sortOrder: 'desc',
          limit: 5
        };

        const ratingResult = await recommendationEngine.searchTemplates(ratingFilter);
        
        // Should be sorted by rating descending
        for (let i = 1; i < ratingResult.templates.length; i++) {
          expect(ratingResult.templates[i - 1].quality.communityRating)
            .toBeGreaterThanOrEqual(ratingResult.templates[i].quality.communityRating);
        }

        const nameFilter: TemplateSearchFilter = {
          sortBy: 'name',
          sortOrder: 'asc',
          limit: 5
        };

        const nameResult = await recommendationEngine.searchTemplates(nameFilter);
        
        // Should be sorted by name ascending
        for (let i = 1; i < nameResult.templates.length; i++) {
          expect(nameResult.templates[i - 1].name.localeCompare(nameResult.templates[i].name))
            .toBeLessThanOrEqual(0);
        }
      });

      it('should handle pagination correctly', async () => {
        const page1Filter: TemplateSearchFilter = {
          limit: 3,
          offset: 0
        };

        const page2Filter: TemplateSearchFilter = {
          limit: 3,
          offset: 3
        };

        const page1 = await recommendationEngine.searchTemplates(page1Filter);
        const page2 = await recommendationEngine.searchTemplates(page2Filter);
        
        expect(page1.templates.length).toBeLessThanOrEqual(3);
        expect(page2.templates.length).toBeLessThanOrEqual(3);
        
        // Pages should have different templates (assuming more than 3 total)
        if (page1.total > 3) {
          const page1Ids = new Set(page1.templates.map(t => t.id));
          const page2Ids = new Set(page2.templates.map(t => t.id));
          
          // Should have no overlap
          const intersection = [...page1Ids].filter(id => page2Ids.has(id));
          expect(intersection.length).toBe(0);
        }
      });

      it('should generate facets correctly', async () => {
        const filter: TemplateSearchFilter = {
          limit: 100 // Get many templates to ensure good facet data
        };

        const result = await recommendationEngine.searchTemplates(filter);
        
        expect(result.facets).toBeDefined();
        expect(result.facets.categories).toBeInstanceOf(Array);
        expect(result.facets.complexity).toBeInstanceOf(Array);
        expect(result.facets.tags).toBeInstanceOf(Array);
        expect(result.facets.authors).toBeInstanceOf(Array);
        
        // Check facet structure
        if (result.facets.categories.length > 0) {
          result.facets.categories.forEach(facet => {
            expect(facet.category).toBeDefined();
            expect(facet.count).toBeGreaterThan(0);
          });
        }
        
        if (result.facets.tags.length > 0) {
          result.facets.tags.forEach(facet => {
            expect(facet.tag).toBeDefined();
            expect(facet.count).toBeGreaterThan(0);
          });
          
          // Tags should be sorted by count (descending)
          for (let i = 1; i < result.facets.tags.length; i++) {
            expect(result.facets.tags[i - 1].count)
              .toBeGreaterThanOrEqual(result.facets.tags[i].count);
          }
        }
      });

      it('should combine multiple filters correctly', async () => {
        const combinedFilter: TemplateSearchFilter = {
          query: 'management',
          categories: [TemplateCategory.HR_MANAGEMENT],
          complexity: [TemplateComplexity.INTERMEDIATE],
          minRating: 4.0,
          limit: 5
        };

        const result = await recommendationEngine.searchTemplates(combinedFilter);
        
        // All returned templates should match all criteria
        result.templates.forEach(template => {
          // Should match query
          const matchesQuery = 
            template.name.toLowerCase().includes('management') ||
            template.description.toLowerCase().includes('management') ||
            template.tags.some(tag => tag.toLowerCase().includes('management')) ||
            template.keywords.some(keyword => keyword.toLowerCase().includes('management'));
          expect(matchesQuery).toBe(true);
          
          // Should match category
          expect(template.category).toBe(TemplateCategory.HR_MANAGEMENT);
          
          // Should match complexity
          expect(template.complexity).toBe(TemplateComplexity.INTERMEDIATE);
          
          // Should match rating
          expect(template.quality.communityRating).toBeGreaterThanOrEqual(4.0);
        });
      });
    });

    describe('Edge Cases and Error Handling', () => {
      it('should handle empty search results gracefully', async () => {
        const filter: TemplateSearchFilter = {
          query: 'nonexistenttemplate12345',
          limit: 10
        };

        const result = await recommendationEngine.searchTemplates(filter);
        
        expect(result.templates).toEqual([]);
        expect(result.total).toBe(0);
        expect(result.facets).toBeDefined();
      });

      it('should handle invalid filter values gracefully', async () => {
        const filter: TemplateSearchFilter = {
          categories: ['invalid_category' as any],
          complexity: ['invalid_complexity' as any],
          limit: 10
        };

        const result = await recommendationEngine.searchTemplates(filter);
        
        // Should return empty results for invalid filters
        expect(result.templates).toEqual([]);
        expect(result.total).toBe(0);
      });

      it('should handle context with technical constraints', async () => {
        const context: TemplateRecommendationContext = {
          existingElements: [],
          existingCategories: [],
          technicalConstraints: {
            database: 'postgresql',
            framework: 'rest',
            specVersion: '1.0'
          }
        };

        const recommendations = await recommendationEngine.getRecommendations(context, 5);
        
        expect(recommendations.length).toBeGreaterThan(0);
        
        // Should consider technical constraints in compatibility scoring
        recommendations.forEach(rec => {
          expect(rec.compatibility).toBeDefined();
        });
      });

      it('should handle large context with many existing elements', async () => {
        const manyElements: MetaContractElement[] = Array.from({ length: 50 }, (_, i) => ({
          id: `element-${i}`,
          type: 'field',
          name: `field_${i}`,
          properties: {
            name: `field_${i}`,
            type: 'string',
            description: `Field ${i}`
          }
        }));

        const context: TemplateRecommendationContext = {
          existingElements: manyElements,
          existingCategories: [TemplateCategory.CUSTOM]
        };

        const startTime = performance.now();
        const recommendations = await recommendationEngine.getRecommendations(context, 5);
        const endTime = performance.now();
        
        expect(recommendations.length).toBeGreaterThan(0);
        expect(endTime - startTime).toBeLessThan(1000); // Should complete within 1 second
      });
    });

    describe('Performance Tests', () => {
      it('should generate recommendations quickly', async () => {
        const context: TemplateRecommendationContext = {
          existingElements: [
            {
              id: 'test-field',
              type: 'field',
              name: 'test_field',
              properties: { name: 'test_field', type: 'string' }
            }
          ],
          existingCategories: [TemplateCategory.CUSTOM]
        };

        const startTime = performance.now();
        const recommendations = await recommendationEngine.getRecommendations(context, 10);
        const endTime = performance.now();
        
        expect(recommendations.length).toBeGreaterThan(0);
        expect(endTime - startTime).toBeLessThan(500); // Should complete within 500ms
      });

      it('should search templates efficiently', async () => {
        const filter: TemplateSearchFilter = {
          query: 'template',
          limit: 20
        };

        const startTime = performance.now();
        const result = await recommendationEngine.searchTemplates(filter);
        const endTime = performance.now();
        
        expect(result.templates).toBeDefined();
        expect(endTime - startTime).toBeLessThan(100); // Should complete within 100ms
      });
    });

    describe('Recommendation Quality', () => {
      it('should provide diverse recommendations', async () => {
        const context: TemplateRecommendationContext = {
          existingElements: [],
          existingCategories: []
        };

        const recommendations = await recommendationEngine.getRecommendations(context, 10);
        
        if (recommendations.length >= 5) {
          const categories = new Set(recommendations.map(r => r.template.category));
          const complexities = new Set(recommendations.map(r => r.template.complexity));
          
          // Should have some diversity in categories and complexities
          expect(categories.size).toBeGreaterThan(1);
          expect(complexities.size).toBeGreaterThan(1);
        }
      });

      it('should score high-quality templates higher', async () => {
        const context: TemplateRecommendationContext = {
          existingElements: [],
          existingCategories: []
        };

        const recommendations = await recommendationEngine.getRecommendations(context, 10);
        
        expect(recommendations.length).toBeGreaterThan(0);
        
        // Top recommendations should generally have high quality scores
        const topRecommendations = recommendations.slice(0, Math.min(3, recommendations.length));
        const avgQualityScore = topRecommendations.reduce((sum, rec) => {
          const templateQuality = (
            rec.template.quality.performanceScore +
            rec.template.quality.securityScore +
            rec.template.quality.maintainabilityScore +
            rec.template.quality.bestPracticesScore
          ) / 4;
          return sum + templateQuality;
        }, 0) / topRecommendations.length;
        
        expect(avgQualityScore).toBeGreaterThan(75); // Top recommendations should have good quality
      });
    });
  });
});