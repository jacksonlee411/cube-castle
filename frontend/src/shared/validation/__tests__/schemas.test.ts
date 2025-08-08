import { describe, it, expect, beforeEach } from 'vitest';
import {
  OrganizationUnitSchema,
  CreateOrganizationInputSchema,
  UpdateOrganizationInputSchema,
  GraphQLVariablesSchema,
  GraphQLOrganizationResponseSchema
} from '../schemas';

describe('Validation Schemas', () => {
  describe('OrganizationUnitSchema', () => {
    it('should validate a complete valid organization unit', () => {
      const validOrg = {
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

      expect(() => OrganizationUnitSchema.parse(validOrg)).not.toThrow();
      const result = OrganizationUnitSchema.parse(validOrg);
      expect(result.code).toBe('1000001');
      expect(result.name).toBe('Test Department');
    });

    it('should reject organization with invalid code format', () => {
      const invalidOrg = {
        code: '123', // Too short
        name: 'Test Department',
        unit_type: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 2,
        created_at: '2025-08-08T12:00:00Z',
        updated_at: '2025-08-08T12:00:00Z',
      };

      expect(() => OrganizationUnitSchema.parse(invalidOrg)).toThrow();
    });

    it('should reject organization with invalid unit_type', () => {
      const invalidOrg = {
        code: '1000001',
        name: 'Test Department',
        unit_type: 'INVALID_TYPE',
        status: 'ACTIVE',
        level: 2,
        created_at: '2025-08-08T12:00:00Z',
        updated_at: '2025-08-08T12:00:00Z',
      };

      expect(() => OrganizationUnitSchema.parse(invalidOrg)).toThrow();
    });

    it('should reject organization with level out of range', () => {
      const invalidOrg = {
        code: '1000001',
        name: 'Test Department',
        unit_type: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 15, // Too high
        created_at: '2025-08-08T12:00:00Z',
        updated_at: '2025-08-08T12:00:00Z',
      };

      expect(() => OrganizationUnitSchema.parse(invalidOrg)).toThrow();
    });

    it('should handle optional fields correctly', () => {
      const minimalOrg = {
        code: '1000001',
        name: 'Test Department',
        unit_type: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 2,
        parent_code: '',
        sort_order: 0,
        description: '',
        created_at: '',
        updated_at: '',
        path: ''
      };

      expect(() => OrganizationUnitSchema.parse(minimalOrg)).not.toThrow();
    });
  });

  describe('CreateOrganizationInputSchema', () => {
    it('should validate a complete create input', () => {
      const validInput = {
        name: 'New Department',
        unit_type: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 3,
        parent_code: '1000001',
        sort_order: 5,
        description: 'New department description'
      };

      expect(() => CreateOrganizationInputSchema.parse(validInput)).not.toThrow();
    });

    it('should use default values for optional fields', () => {
      const minimalInput = {
        name: 'New Department',
        unit_type: 'DEPARTMENT',
        level: 3
      };

      const result = CreateOrganizationInputSchema.parse(minimalInput);
      expect(result.status).toBe('ACTIVE');
      expect(result.sort_order).toBe(0);
    });

    it('should reject input with empty name', () => {
      const invalidInput = {
        name: '',
        unit_type: 'DEPARTMENT',
        level: 3
      };

      expect(() => CreateOrganizationInputSchema.parse(invalidInput)).toThrow();
    });

    it('should reject input with name too long', () => {
      const invalidInput = {
        name: 'a'.repeat(101), // Too long
        unit_type: 'DEPARTMENT',
        level: 3
      };

      expect(() => CreateOrganizationInputSchema.parse(invalidInput)).toThrow();
    });
  });

  describe('UpdateOrganizationInputSchema', () => {
    it('should validate partial update input', () => {
      const validUpdate = {
        code: '1000001',
        name: 'Updated Department',
        status: 'INACTIVE'
      };

      expect(() => UpdateOrganizationInputSchema.parse(validUpdate)).not.toThrow();
    });

    it('should require code field for updates', () => {
      const invalidUpdate = {
        name: 'Updated Department',
        status: 'INACTIVE'
        // Missing code
      };

      expect(() => UpdateOrganizationInputSchema.parse(invalidUpdate)).toThrow();
    });

    it('should allow empty update (all fields optional except code)', () => {
      const emptyUpdate = {
        code: '1000001'
      };

      expect(() => UpdateOrganizationInputSchema.parse(emptyUpdate)).not.toThrow();
    });
  });

  describe('GraphQLVariablesSchema', () => {
    it('should validate complete query variables', () => {
      const validVars = {
        searchText: 'test',
        unitType: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 2,
        page: 1,
        pageSize: 20
      };

      expect(() => GraphQLVariablesSchema.parse(validVars)).not.toThrow();
    });

    it('should validate empty variables object', () => {
      const emptyVars = {};

      expect(() => GraphQLVariablesSchema.parse(emptyVars)).not.toThrow();
    });

    it('should reject invalid pageSize', () => {
      const invalidVars = {
        pageSize: 150 // Too large
      };

      expect(() => GraphQLVariablesSchema.parse(invalidVars)).toThrow();
    });

    it('should reject negative page number', () => {
      const invalidVars = {
        page: 0 // Should be at least 1
      };

      expect(() => GraphQLVariablesSchema.parse(invalidVars)).toThrow();
    });
  });

  describe('GraphQLOrganizationResponseSchema', () => {
    it('should validate typical GraphQL response', () => {
      const validResponse = {
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

      expect(() => GraphQLOrganizationResponseSchema.parse(validResponse)).not.toThrow();
    });

    it('should handle null optional fields', () => {
      const responseWithNulls = {
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

      expect(() => GraphQLOrganizationResponseSchema.parse(responseWithNulls)).not.toThrow();
    });

    it('should require core fields', () => {
      const incompleteResponse = {
        code: '1000001',
        name: 'Test Department'
        // Missing unitType, status, level
      };

      expect(() => GraphQLOrganizationResponseSchema.parse(incompleteResponse)).toThrow();
    });
  });
});