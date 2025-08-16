import { describe, it, expect } from 'vitest';
import {
  validateOrganizationUnit,
  validateCreateOrganizationInput,
  validateGraphQLOrganizationList,
  isGraphQLError,
  isGraphQLSuccessResponse,
  isAPIError,
  isValidationError,
  isNetworkError,
  safeTransformGraphQLToOrganizationUnit,
  safeTransformCreateInputToAPI,
  ValidationError
} from '../type-guards';

describe('Type Guards and Validators', () => {
  describe('validateOrganizationUnit', () => {
    it('should validate and return a valid organization unit', () => {
      const validData = {
        code: '1000001',
        name: 'Test Department',
        unit_type: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 2,
        parent_code: '1000000',
        sort_order: 1,
        description: 'Test description',
        created_at: '2025-08-08T12:00:00Z',
        updated_at: '2025-08-08T12:00:00Z',
        path: '/1000000/1000001'
      };

      const result = validateOrganizationUnit(validData);
      expect(result).toEqual(validData);
    });

    it('should throw ValidationError for invalid data', () => {
      const invalidData = {
        code: '123', // Invalid code
        name: 'Test Department',
        unit_type: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 2
      };

      expect(() => validateOrganizationUnit(invalidData)).toThrow(ValidationError);
    });
  });

  describe('validateCreateOrganizationInput', () => {
    it('should validate and return valid create input', () => {
      const validInput = {
        name: 'New Department',
        unit_type: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 3,
        parent_code: '1000001',
        sort_order: 5,
        description: 'New department description'
      };

      const result = validateCreateOrganizationInput(validInput);
      expect(result.name).toBe('New Department');
      expect(result.status).toBe('ACTIVE');
    });

    it('should apply default values', () => {
      const minimalInput = {
        name: 'New Department',
        unit_type: 'DEPARTMENT',
        level: 3
      };

      const result = validateCreateOrganizationInput(minimalInput);
      expect(result.status).toBe('ACTIVE');
      expect(result.sort_order).toBe(0);
    });
  });

  describe('validateGraphQLOrganizationList', () => {
    it('should validate an array of GraphQL organization responses', () => {
      const validList = [
        {
          code: '1000001',
          name: 'Department 1',
          unitType: 'DEPARTMENT',
          status: 'ACTIVE',
          level: 2
        },
        {
          code: '1000002',
          name: 'Department 2',
          unitType: 'DEPARTMENT',
          status: 'ACTIVE',
          level: 2
        }
      ];

      const result = validateGraphQLOrganizationList(validList);
      expect(result).toHaveLength(2);
      expect(result[0].code).toBe('1000001');
    });

    it('should throw ValidationError with index information for invalid items', () => {
      const invalidList = [
        {
          code: '1000001',
          name: 'Department 1',
          unitType: 'DEPARTMENT',
          status: 'ACTIVE',
          level: 2
        },
        {
          code: '1000002',
          // Missing required fields
        }
      ];

      expect(() => validateGraphQLOrganizationList(invalidList))
        .toThrow(/index 1/);
    });
  });

  describe('Type Guard Functions', () => {
    describe('isGraphQLError', () => {
      it('should return true for valid GraphQL error response', () => {
        const errorResponse = {
          errors: [
            {
              message: 'Field error',
              locations: [{ line: 1, column: 1 }],
              path: ['organizations']
            }
          ]
        };

        expect(isGraphQLError(errorResponse)).toBe(true);
      });

      it('should return false for successful GraphQL response', () => {
        const successResponse = {
          data: {
            organizations: []
          }
        };

        expect(isGraphQLError(successResponse)).toBe(false);
      });

      it('should return false for empty errors array', () => {
        const emptyErrorsResponse = {
          errors: []
        };

        expect(isGraphQLError(emptyErrorsResponse)).toBe(false);
      });
    });

    describe('isGraphQLSuccessResponse', () => {
      it('should return true for valid success response', () => {
        const successResponse = {
          data: {
            organizations: []
          }
        };

        expect(isGraphQLSuccessResponse(successResponse)).toBe(true);
      });

      it('should return false for response with null data', () => {
        const nullDataResponse = {
          data: null
        };

        expect(isGraphQLSuccessResponse(nullDataResponse)).toBe(false);
      });

      it('should return false for response without data field', () => {
        const noDataResponse = {
          result: 'success'
        };

        expect(isGraphQLSuccessResponse(noDataResponse)).toBe(false);
      });
    });

    describe('isAPIError', () => {
      it('should return true for API error with status and statusText', () => {
        const apiError = new Error('API failed') as Error & { status: number; statusText: string };
        apiError.status = 404;
        apiError.statusText = 'Not Found';

        expect(isAPIError(apiError)).toBe(true);
      });

      it('should return false for regular Error', () => {
        const regularError = new Error('Regular error');

        expect(isAPIError(regularError)).toBe(false);
      });
    });

    describe('isValidationError', () => {
      it('should return true for ValidationError instance', () => {
        const validationError = new ValidationError('Validation failed', []);

        expect(isValidationError(validationError)).toBe(true);
      });

      it('should return false for regular Error', () => {
        const regularError = new Error('Regular error');

        expect(isValidationError(regularError)).toBe(false);
      });
    });

    describe('isNetworkError', () => {
      it('should return true for TypeError with fetch message', () => {
        const networkError = new TypeError('fetch failed');

        expect(isNetworkError(networkError)).toBe(true);
      });

      it('should return false for TypeError without fetch message', () => {
        const typeError = new TypeError('type error');

        expect(isNetworkError(typeError)).toBe(false);
      });
    });
  });

  describe('Safe Transform Functions', () => {
    describe('safeTransformGraphQLToOrganizationUnit', () => {
      it('should transform GraphQL response to OrganizationUnit', () => {
        const graphqlResponse = {
          code: '1000001',
          name: 'Test Department',
          unitType: 'DEPARTMENT',
          status: 'ACTIVE',
          level: 2,
          parentCode: '1000000',
          path: '/1000000/1000001',
          sortOrder: 1,
          description: 'Test description',
          createdAt: '2025-08-08T12:00:00Z',
          updatedAt: '2025-08-08T12:00:00Z'
        };

        const result = safeTransformGraphQLToOrganizationUnit(graphqlResponse);

        expect(result).toEqual({
          code: '1000001',
          parent_code: '1000000',
          name: 'Test Department',
          unit_type: 'DEPARTMENT',
          status: 'ACTIVE',
          level: 2,
          path: '/1000000/1000001',
          sort_order: 1,
          description: 'Test description',
          created_at: '2025-08-08T12:00:00Z',
          updated_at: '2025-08-08T12:00:00Z'
        });
      });

      it('should handle null/undefined optional fields', () => {
        const graphqlResponse = {
          code: '1000001',
          name: 'Test Department',
          unitType: 'DEPARTMENT',
          status: 'ACTIVE',
          level: 2,
          parentCode: null,
          sortOrder: null,
          description: null,
          createdAt: null,
          updatedAt: null
        };

        const result = safeTransformGraphQLToOrganizationUnit(graphqlResponse);

        expect(result.parent_code).toBe('');
        expect(result.sort_order).toBe(0);
        expect(result.description).toBe('');
        expect(result.created_at).toBe('');
        expect(result.updated_at).toBe('');
      });
    });

    describe('safeTransformCreateInputToAPI', () => {
      it('should transform create input to API payload', () => {
        const createInput = {
          code: '1000001',
          name: 'New Department',
          unit_type: 'DEPARTMENT' as const,
          status: 'ACTIVE' as const,
          level: 3,
          parent_code: '1000000',
          sort_order: 5,
          description: 'New department description'
        };

        const result = safeTransformCreateInputToAPI(createInput);

        expect(result).toEqual({
          name: 'New Department',
          unit_type: 'DEPARTMENT',
          status: 'ACTIVE',
          level: 3,
          sort_order: 5,
          description: 'New department description',
          code: '1000001',
          parent_code: '1000000'
        });
      });

      it('should omit undefined optional fields', () => {
        const minimalInput = {
          name: 'New Department',
          unit_type: 'DEPARTMENT' as const,
          status: 'ACTIVE' as const,
          level: 3,
          sort_order: 0,
          description: ''
        };

        const result = safeTransformCreateInputToAPI(minimalInput);

        expect(result).toEqual({
          name: 'New Department',
          unit_type: 'DEPARTMENT',
          status: 'ACTIVE',
          level: 3,
          sort_order: 0,
          description: ''
        });
        expect(result).not.toHaveProperty('code');
        expect(result).not.toHaveProperty('parent_code');
      });

      it('should not include empty string parent_code', () => {
        const inputWithEmptyParent = {
          name: 'New Department',
          unit_type: 'DEPARTMENT' as const,
          status: 'ACTIVE' as const,
          level: 3,
          parent_code: '',
          sort_order: 0,
          description: ''
        };

        const result = safeTransformCreateInputToAPI(inputWithEmptyParent);

        expect(result).not.toHaveProperty('parent_code');
      });
    });
  });

  describe('ValidationError Class', () => {
    it('should create ValidationError with message and details', () => {
      const details = [
        { message: 'Invalid field', code: 'invalid', path: ['name'] }
      ];
      const error = new ValidationError('Validation failed', details);

      expect(error.message).toBe('Validation failed');
      expect(error.details).toEqual(details);
      expect(error.code).toBe('VALIDATION_ERROR');
      expect(error.name).toBe('ValidationError');
    });
  });
});