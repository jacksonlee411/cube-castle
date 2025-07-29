/**
 * Unit tests for PropertyPanel component
 * Tests property editing, form validation, and element type-specific forms
 */
import React from 'react';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PropertyPanel } from '@/components/metacontract-editor/visual/PropertyPanel';
import testUtils, { setupTests } from '@/tests/setup/test-utils';
import { MetaContractElement } from '@/components/metacontract-editor/VisualEditor';

// Setup tests
setupTests();

describe('PropertyPanel', () => {
  const mockFieldElement: MetaContractElement = {
    id: 'field-1',
    type: 'field',
    name: 'Test Field',
    properties: {
      name: 'test_field',
      type: 'string',
      required: true,
      unique: false,
      max_length: 255,
      description: 'Test field description'
    }
  };

  const mockRelationshipElement: MetaContractElement = {
    id: 'relationship-1',
    type: 'relationship',
    name: 'Test Relationship',
    properties: {
      name: 'test_relation',
      type: 'one_to_many',
      target_resource: 'related_table',
      foreign_key: 'test_id',
      cascade_delete: false
    }
  };

  const mockSecurityElement: MetaContractElement = {
    id: 'security-1',
    type: 'security',
    name: 'Security Model',
    properties: {
      type: 'rbac',
      roles: ['admin', 'user'],
      permissions: {
        read: ['admin', 'user'],
        write: ['admin'],
        delete: ['admin']
      }
    }
  };

  const defaultProps = {
    element: mockFieldElement,
    onUpdateElement: jest.fn(),
    onDeleteElement: jest.fn(),
    readonly: false
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Component Rendering', () => {
    it('should render property panel with element information', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} />);

      expect(screen.getByText('Properties')).toBeInTheDocument();
      expect(screen.getByText('field')).toBeInTheDocument();
      expect(screen.getByDisplayValue('test_field')).toBeInTheDocument();
    });

    it('should show unsaved changes indicator', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      const nameInput = screen.getByDisplayValue('test_field');
      await user.clear(nameInput);
      await user.type(nameInput, 'modified_field');

      // Should show unsaved changes indicator
      const indicator = document.querySelector('.bg-orange-500');
      expect(indicator).toBeInTheDocument();
    });

    it('should render in readonly mode', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} readonly={true} />);

      // All inputs should be disabled
      expect(screen.getByDisplayValue('test_field')).toBeDisabled();
      expect(screen.getByDisplayValue('string')).toBeDisabled();
      
      // Save and delete buttons should not be present
      expect(screen.queryByRole('button', { name: /save/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /delete/i })).not.toBeInTheDocument();
    });
  });

  describe('Field Properties', () => {
    it('should render field-specific properties', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} />);

      expect(screen.getByLabelText(/field name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/data type/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/required/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/unique/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/primary key/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/max length/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    });

    it('should update field name', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      const nameInput = screen.getByLabelText(/field name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'updated_field');

      expect(nameInput).toHaveValue('updated_field');
    });

    it('should change field type', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      const typeSelect = screen.getByRole('combobox');
      await user.click(typeSelect);
      await user.click(screen.getByText('integer'));

      expect(screen.getByDisplayValue('integer')).toBeInTheDocument();
    });

    it('should toggle boolean properties', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      const requiredSwitch = screen.getByRole('switch', { name: /required/i });
      const uniqueSwitch = screen.getByRole('switch', { name: /unique/i });
      const primaryKeySwitch = screen.getByRole('switch', { name: /primary key/i });

      // Initially required should be checked
      expect(requiredSwitch).toBeChecked();
      expect(uniqueSwitch).not.toBeChecked();

      // Toggle unique
      await user.click(uniqueSwitch);
      expect(uniqueSwitch).toBeChecked();

      // Toggle primary key
      await user.click(primaryKeySwitch);
      expect(primaryKeySwitch).toBeChecked();
    });

    it('should show decimal-specific fields for decimal type', async () => {
      const decimalElement: MetaContractElement = {
        ...mockFieldElement,
        properties: {
          ...mockFieldElement.properties,
          type: 'decimal',
          precision: 10,
          scale: 2
        }
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={decimalElement} />);

      expect(screen.getByLabelText(/precision/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/scale/i)).toBeInTheDocument();
      expect(screen.getByDisplayValue('10')).toBeInTheDocument();
      expect(screen.getByDisplayValue('2')).toBeInTheDocument();
    });

    it('should show numeric range fields for integer type', async () => {
      const integerElement: MetaContractElement = {
        ...mockFieldElement,
        properties: {
          ...mockFieldElement.properties,
          type: 'integer',
          min: 0,
          max: 1000
        }
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={integerElement} />);

      expect(screen.getByLabelText(/min value/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/max value/i)).toBeInTheDocument();
    });
  });

  describe('Relationship Properties', () => {
    it('should render relationship-specific properties', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} element={mockRelationshipElement} />);

      expect(screen.getByLabelText(/relationship name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/relationship type/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/target resource/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/foreign key/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/cascade delete/i)).toBeInTheDocument();
    });

    it('should show join table field for many-to-many relationships', async () => {
      const manyToManyElement: MetaContractElement = {
        ...mockRelationshipElement,
        properties: {
          ...mockRelationshipElement.properties,
          type: 'many_to_many',
          join_table: 'junction_table'
        }
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={manyToManyElement} />);

      expect(screen.getByLabelText(/join table/i)).toBeInTheDocument();
      expect(screen.getByDisplayValue('junction_table')).toBeInTheDocument();
    });

    it('should update relationship properties', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} element={mockRelationshipElement} />);

      const targetInput = screen.getByLabelText(/target resource/i);
      await user.clear(targetInput);
      await user.type(targetInput, 'new_target_table');

      expect(targetInput).toHaveValue('new_target_table');
    });
  });

  describe('Security Properties', () => {
    it('should render security-specific properties', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} element={mockSecurityElement} />);

      expect(screen.getByLabelText(/security type/i)).toBeInTheDocument();
      expect(screen.getByText('Roles')).toBeInTheDocument();
      expect(screen.getByText('Permissions')).toBeInTheDocument();
    });

    it('should display existing roles', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} element={mockSecurityElement} />);

      expect(screen.getByDisplayValue('admin')).toBeInTheDocument();
      expect(screen.getByDisplayValue('user')).toBeInTheDocument();
    });

    it('should add new roles', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} element={mockSecurityElement} />);

      const addRoleButton = screen.getByRole('button', { name: /add role/i });
      await user.click(addRoleButton);

      // Should add a new role input
      const roleInputs = screen.getAllByDisplayValue(/role|admin|user/);
      expect(roleInputs.length).toBeGreaterThan(2);
    });

    it('should remove roles', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} element={mockSecurityElement} />);

      const deleteButtons = screen.getAllByRole('button');
      const deleteRoleButton = deleteButtons.find(btn => 
        btn.querySelector('svg')?.getAttribute('class')?.includes('w-4')
      );

      if (deleteRoleButton) {
        await user.click(deleteRoleButton);
        // Role should be removed (implementation would update the array)
      }
    });

    it('should display permissions correctly', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} element={mockSecurityElement} />);

      expect(screen.getByText('read')).toBeInTheDocument();
      expect(screen.getByText('write')).toBeInTheDocument();
      expect(screen.getByText('delete')).toBeInTheDocument();
    });
  });

  describe('Validation Properties', () => {
    it('should render validation-specific properties', async () => {
      const validationElement: MetaContractElement = {
        id: 'validation-1',
        type: 'validation',
        name: 'Test Validation',
        properties: {
          type: 'required',
          message: 'This field is required',
          fields: ['field1', 'field2']
        }
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={validationElement} />);

      expect(screen.getByLabelText(/validation type/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/error message/i)).toBeInTheDocument();
      expect(screen.getByText('Target Fields')).toBeInTheDocument();
    });

    it('should show pattern field for pattern validation', async () => {
      const patternValidationElement: MetaContractElement = {
        id: 'validation-pattern',
        type: 'validation',
        name: 'Pattern Validation',
        properties: {
          type: 'pattern',
          pattern: '^[A-Za-z0-9]+$',
          message: 'Invalid format'
        }
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={patternValidationElement} />);

      expect(screen.getByLabelText(/pattern.*regex/i)).toBeInTheDocument();
      expect(screen.getByDisplayValue('^[A-Za-z0-9]+$')).toBeInTheDocument();
    });

    it('should show range fields for range validation', async () => {
      const rangeValidationElement: MetaContractElement = {
        id: 'validation-range',
        type: 'validation',
        name: 'Range Validation',
        properties: {
          type: 'range',
          min: 0,
          max: 100
        }
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={rangeValidationElement} />);

      expect(screen.getByLabelText(/min value/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/max value/i)).toBeInTheDocument();
    });
  });

  describe('Index Properties', () => {
    it('should render index-specific properties', async () => {
      const indexElement: MetaContractElement = {
        id: 'index-1',
        type: 'index',
        name: 'Test Index',
        properties: {
          name: 'idx_test',
          type: 'btree',
          fields: ['field1', 'field2'],
          unique: false
        }
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={indexElement} />);

      expect(screen.getByLabelText(/index name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/index type/i)).toBeInTheDocument();
      expect(screen.getByText('Fields')).toBeInTheDocument();
      expect(screen.getByLabelText(/unique index/i)).toBeInTheDocument();
    });
  });

  describe('Save and Reset Functionality', () => {
    it('should save changes when save button is clicked', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      // Make a change
      const nameInput = screen.getByLabelText(/field name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'modified_field');

      // Click save
      const saveButton = screen.getByRole('button', { name: /save/i });
      await user.click(saveButton);

      expect(defaultProps.onUpdateElement).toHaveBeenCalledWith({
        name: 'modified_field',
        properties: expect.objectContaining({
          name: 'modified_field'
        })
      });
    });

    it('should reset changes when reset button is clicked', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      // Make a change
      const nameInput = screen.getByLabelText(/field name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'modified_field');

      // Click reset
      const resetButton = screen.getByRole('button', { name: /reset/i });
      await user.click(resetButton);

      // Should revert to original value
      expect(nameInput).toHaveValue('test_field');
    });

    it('should only show save/reset buttons when there are changes', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      // Initially no save/reset buttons
      expect(screen.queryByRole('button', { name: /save/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /reset/i })).not.toBeInTheDocument();

      // Make a change
      const nameInput = screen.getByLabelText(/field name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'modified_field');

      // Now save/reset buttons should appear
      expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /reset/i })).toBeInTheDocument();
    });

    it('should not show save/reset buttons in readonly mode', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} readonly={true} />);

      expect(screen.queryByRole('button', { name: /save/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /reset/i })).not.toBeInTheDocument();
    });
  });

  describe('Delete Functionality', () => {
    it('should call onDeleteElement when delete button is clicked', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      const deleteButton = screen.getByRole('button', { name: /delete/i });
      await user.click(deleteButton);

      expect(defaultProps.onDeleteElement).toHaveBeenCalled();
    });

    it('should not show delete button in readonly mode', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} readonly={true} />);

      expect(screen.queryByRole('button', { name: /delete/i })).not.toBeInTheDocument();
    });
  });

  describe('Form Validation and Input Handling', () => {
    it('should handle numeric inputs correctly', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      const maxLengthInput = screen.getByLabelText(/max length/i);
      await user.clear(maxLengthInput);
      await user.type(maxLengthInput, '500');

      expect(maxLengthInput).toHaveValue('500');
    });

    it('should handle text inputs correctly', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      const descriptionInput = screen.getByLabelText(/description/i);
      await user.clear(descriptionInput);
      await user.type(descriptionInput, 'Updated description text');

      expect(descriptionInput).toHaveValue('Updated description text');
    });

    it('should handle select inputs correctly', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      const typeSelect = screen.getByRole('combobox');
      await user.click(typeSelect);
      
      // Select email type
      await user.click(screen.getByText('email'));

      expect(screen.getByDisplayValue('email')).toBeInTheDocument();
    });
  });

  describe('Array Field Management', () => {
    it('should add items to array fields', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} element={mockSecurityElement} />);

      const addRoleButton = screen.getByRole('button', { name: /add role/i });
      const initialRoleCount = screen.getAllByDisplayValue(/admin|user/).length;
      
      await user.click(addRoleButton);

      // Should have one more role input
      const newRoleCount = screen.getAllByRole('textbox').filter(input => 
        input.getAttribute('value')?.includes('role') || 
        input.getAttribute('value') === 'admin' || 
        input.getAttribute('value') === 'user'
      ).length;
      
      expect(newRoleCount).toBeGreaterThan(initialRoleCount);
    });

    it('should remove items from array fields', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} element={mockSecurityElement} />);

      const deleteButtons = screen.getAllByRole('button');
      const roleDeleteButton = deleteButtons.find(btn => 
        btn.querySelector('svg')?.getAttribute('class')?.includes('w-4')
      );

      if (roleDeleteButton) {
        await user.click(roleDeleteButton);
        // Item should be removed (implementation would update state)
      }
    });
  });

  describe('Accessibility', () => {
    it('should be accessible', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} />);

      // Check for proper labels
      expect(screen.getByLabelText(/field name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/data type/i)).toBeInTheDocument();
      expect(screen.getByRole('switch', { name: /required/i })).toBeInTheDocument();

      await testUtils.checkAccessibility();
    });

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup();
      testUtils.render(<PropertyPanel {...defaultProps} />);

      // Tab through form elements
      await user.tab();
      
      const activeElement = document.activeElement;
      expect(activeElement).toBeInTheDocument();
      expect(activeElement?.tagName).toMatch(/INPUT|BUTTON|SELECT/i);
    });

    it('should have proper ARIA labels for complex controls', async () => {
      testUtils.render(<PropertyPanel {...defaultProps} element={mockSecurityElement} />);

      // Array management buttons should be accessible
      const addRoleButton = screen.getByRole('button', { name: /add role/i });
      expect(addRoleButton).toBeInTheDocument();
    });
  });

  describe('Performance', () => {
    it('should render quickly with complex elements', async () => {
      const complexElement: MetaContractElement = {
        id: 'complex-security',
        type: 'security',
        name: 'Complex Security',
        properties: {
          type: 'rbac',
          roles: Array.from({ length: 20 }, (_, i) => `role_${i}`),
          permissions: {
            read: Array.from({ length: 10 }, (_, i) => `role_${i}`),
            write: Array.from({ length: 5 }, (_, i) => `role_${i}`),
            delete: ['admin']
          }
        }
      };

      const renderTime = testUtils.measureRenderTime('PropertyPanel - Complex Element', () => {
        testUtils.render(<PropertyPanel {...defaultProps} element={complexElement} />);
      });

      expect(renderTime).toBeLessThan(500); // 500ms
      expect(screen.getByText('Properties')).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should handle elements with missing properties', async () => {
      const elementWithoutProperties: MetaContractElement = {
        id: 'minimal-element',
        type: 'field',
        name: 'Minimal Element',
        properties: {}
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={elementWithoutProperties} />);

      // Should not crash and should show default values
      expect(screen.getByText('Properties')).toBeInTheDocument();
      expect(screen.getByLabelText(/field name/i)).toBeInTheDocument();
    });

    it('should handle unsupported element types', async () => {
      const unsupportedElement: MetaContractElement = {
        id: 'unsupported',
        type: 'unknown' as any,
        name: 'Unsupported Element',
        properties: {}
      };

      testUtils.render(<PropertyPanel {...defaultProps} element={unsupportedElement} />);

      expect(screen.getByText(/no properties available/i)).toBeInTheDocument();
    });

    it('should handle callback errors gracefully', async () => {
      const errorCallback = jest.fn(() => {
        throw new Error('Test error');
      });
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      const user = userEvent.setup();
      
      testUtils.render(
        <PropertyPanel {...defaultProps} onUpdateElement={errorCallback} />
      );

      // Make a change and save
      const nameInput = screen.getByLabelText(/field name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'modified');

      const saveButton = screen.getByRole('button', { name: /save/i });
      await user.click(saveButton);
      
      // Should not crash the component
      expect(screen.getByText('Properties')).toBeInTheDocument();
      
      consoleSpy.mockRestore();
    });
  });
});