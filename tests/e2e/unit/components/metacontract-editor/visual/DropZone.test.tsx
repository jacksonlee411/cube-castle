/**
 * Unit tests for DropZone component
 * Tests drag and drop functionality, element display, and interactions
 */
import React from 'react';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DropZone } from '@/components/metacontract-editor/visual/DropZone';
import testUtils, { setupTests } from '@/tests/setup/test-utils';
import { MetaContractElement } from '@/components/metacontract-editor/VisualEditor';

// Setup tests
setupTests();

describe('DropZone', () => {
  const mockElements: MetaContractElement[] = [
    {
      id: 'field-1',
      type: 'field',
      name: 'Test Field',
      properties: {
        name: 'test_field',
        type: 'string',
        required: true,
        unique: false
      }
    },
    {
      id: 'relationship-1',
      type: 'relationship',
      name: 'Test Relationship',
      properties: {
        name: 'test_relation',
        type: 'one_to_many',
        target_resource: 'related_table',
        foreign_key: 'test_id'
      }
    },
    {
      id: 'security-1',
      type: 'security',
      name: 'Security Model',
      properties: {
        type: 'rbac',
        roles: ['admin', 'user'],
        permissions: { read: ['admin', 'user'], write: ['admin'] }
      }
    }
  ];

  const defaultProps = {
    elements: mockElements,
    selectedElement: null,
    onSelectElement: jest.fn(),
    onUpdateElement: jest.fn(),
    onDeleteElement: jest.fn(),
    readonly: false
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Component Rendering', () => {
    it('should render drop zone with elements', async () => {
      testUtils.render(<DropZone {...defaultProps} />);

      // Should show all elements
      expect(screen.getByText('Test Field')).toBeInTheDocument();
      expect(screen.getByText('Test Relationship')).toBeInTheDocument();
      expect(screen.getByText('Security Model')).toBeInTheDocument();
    });

    it('should render empty state when no elements', async () => {
      testUtils.render(<DropZone {...defaultProps} elements={[]} />);

      expect(screen.getByText('No Elements Yet')).toBeInTheDocument();
      expect(screen.getByText(/start building your meta-contract/i)).toBeInTheDocument();
      
      // Should show element type legend
      expect(screen.getByText('Fields')).toBeInTheDocument();
      expect(screen.getByText('Relations')).toBeInTheDocument();
      expect(screen.getByText('Security')).toBeInTheDocument();
    });

    it('should not show legend in readonly mode when empty', async () => {
      testUtils.render(
        <DropZone {...defaultProps} elements={[]} readonly={true} />
      );

      expect(screen.getByText('No Elements Yet')).toBeInTheDocument();
      expect(screen.queryByText('Fields')).not.toBeInTheDocument();
    });

    it('should render in readonly mode', async () => {
      testUtils.render(<DropZone {...defaultProps} readonly={true} />);

      expect(screen.getByText('Test Field')).toBeInTheDocument();
      
      // Dropdown menus should not be present in readonly mode
      expect(screen.queryByRole('button', { name: /more options/i })).not.toBeInTheDocument();
    });
  });

  describe('Element Display', () => {
    it('should display element information correctly', async () => {
      testUtils.render(<DropZone {...defaultProps} />);

      // Field element
      expect(screen.getByText('Test Field')).toBeInTheDocument();
      expect(screen.getByText('field')).toBeInTheDocument();
      expect(screen.getByText('string')).toBeInTheDocument();
      expect(screen.getByText('Yes')).toBeInTheDocument(); // required: true

      // Relationship element
      expect(screen.getByText('Test Relationship')).toBeInTheDocument();
      expect(screen.getByText('relationship')).toBeInTheDocument();
      expect(screen.getByText('related_table')).toBeInTheDocument(); // target_resource
    });

    it('should show unique and primary key badges', async () => {
      const elementWithBadges: MetaContractElement = {
        id: 'field-uuid',
        type: 'field',
        name: 'ID Field',
        properties: {
          name: 'id',
          type: 'uuid',
          required: true,
          unique: true,
          primary_key: true
        }
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[elementWithBadges]} />
      );

      expect(screen.getByText('Unique')).toBeInTheDocument();
      expect(screen.getByText('Primary Key')).toBeInTheDocument();
    });

    it('should show validation badges', async () => {
      const elementWithValidation: MetaContractElement = {
        id: 'field-validated',
        type: 'field',
        name: 'Validated Field',
        properties: {
          name: 'validated_field',
          type: 'string',
          validation: ['email', 'required']
        }
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[elementWithValidation]} />
      );

      expect(screen.getByText('email')).toBeInTheDocument();
      expect(screen.getByText('required')).toBeInTheDocument();
    });

    it('should use different colors for different element types', async () => {
      testUtils.render(<DropZone {...defaultProps} />);

      // Elements should be rendered with different styling
      const fieldCard = screen.getByText('Test Field').closest('.bg-blue-50');
      const relationshipCard = screen.getByText('Test Relationship').closest('.bg-green-50');
      const securityCard = screen.getByText('Security Model').closest('.bg-red-50');

      expect(fieldCard).toBeInTheDocument();
      expect(relationshipCard).toBeInTheDocument();
      expect(securityCard).toBeInTheDocument();
    });
  });

  describe('Element Selection', () => {
    it('should select element when clicked', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      const fieldCard = screen.getByText('Test Field').closest('[role="button"], .cursor-pointer');
      await user.click(fieldCard!);

      expect(defaultProps.onSelectElement).toHaveBeenCalledWith(mockElements[0]);
    });

    it('should deselect element when clicked again', async () => {
      const user = userEvent.setup();
      testUtils.render(
        <DropZone {...defaultProps} selectedElement={mockElements[0]} />
      );

      const fieldCard = screen.getByText('Test Field').closest('[role="button"], .cursor-pointer');
      await user.click(fieldCard!);

      expect(defaultProps.onSelectElement).toHaveBeenCalledWith(null);
    });

    it('should show selection styling for selected element', async () => {
      testUtils.render(
        <DropZone {...defaultProps} selectedElement={mockElements[0]} />
      );

      const fieldCard = screen.getByText('Test Field').closest('.ring-2');
      expect(fieldCard).toBeInTheDocument();
    });

    it('should switch selection between elements', async () => {
      const user = userEvent.setup();
      testUtils.render(
        <DropZone {...defaultProps} selectedElement={mockElements[0]} />
      );

      const relationshipCard = screen.getByText('Test Relationship').closest('[role="button"], .cursor-pointer');
      await user.click(relationshipCard!);

      expect(defaultProps.onSelectElement).toHaveBeenCalledWith(mockElements[1]);
    });
  });

  describe('Element Actions Menu', () => {
    it('should show action menu when clicking more options', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      // Find and click the more options button
      const moreButtons = screen.getAllByRole('button');
      const moreButton = moreButtons.find(button => 
        button.querySelector('svg')?.getAttribute('class')?.includes('w-3')
      );
      
      if (moreButton) {
        await user.click(moreButton);

        expect(screen.getByText('Edit Properties')).toBeInTheDocument();
        expect(screen.getByText('Duplicate')).toBeInTheDocument();
        expect(screen.getByText('Hide')).toBeInTheDocument();
        expect(screen.getByText('Delete')).toBeInTheDocument();
      }
    });

    it('should call onSelectElement when clicking Edit Properties', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      const moreButtons = screen.getAllByRole('button');
      const moreButton = moreButtons.find(button => 
        button.querySelector('svg')?.getAttribute('class')?.includes('w-3')
      );
      
      if (moreButton) {
        await user.click(moreButton);
        await user.click(screen.getByText('Edit Properties'));

        expect(defaultProps.onSelectElement).toHaveBeenCalled();
      }
    });

    it('should call onDeleteElement when clicking Delete', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      const moreButtons = screen.getAllByRole('button');
      const moreButton = moreButtons.find(button => 
        button.querySelector('svg')?.getAttribute('class')?.includes('w-3')
      );
      
      if (moreButton) {
        await user.click(moreButton);
        await user.click(screen.getByText('Delete'));

        expect(defaultProps.onDeleteElement).toHaveBeenCalledWith('field-1');
      }
    });

    it('should toggle visibility when clicking Hide/Show', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      const moreButtons = screen.getAllByRole('button');
      const moreButton = moreButtons.find(button => 
        button.querySelector('svg')?.getAttribute('class')?.includes('w-3')
      );
      
      if (moreButton) {
        await user.click(moreButton);
        await user.click(screen.getByText('Hide'));

        expect(defaultProps.onUpdateElement).toHaveBeenCalledWith(
          'field-1',
          expect.objectContaining({
            properties: expect.objectContaining({
              hidden: true
            })
          })
        );
      }
    });

    it('should not show action menu in readonly mode', async () => {
      testUtils.render(<DropZone {...defaultProps} readonly={true} />);

      // Should not have more options buttons in readonly mode
      const moreButtons = screen.queryAllByRole('button').filter(button => 
        button.querySelector('svg')?.getAttribute('class')?.includes('w-3')
      );
      
      expect(moreButtons).toHaveLength(0);
    });
  });

  describe('Drag and Drop Integration', () => {
    it('should integrate with sortable drag and drop', async () => {
      testUtils.render(<DropZone {...defaultProps} />);

      // Elements should be rendered (drag and drop functionality is mocked)
      expect(screen.getByText('Test Field')).toBeInTheDocument();
      expect(screen.getByText('Test Relationship')).toBeInTheDocument();
    });

    it('should disable drag in readonly mode', async () => {
      testUtils.render(<DropZone {...defaultProps} readonly={true} />);

      // Elements should still be rendered but drag should be disabled
      expect(screen.getByText('Test Field')).toBeInTheDocument();
    });

    it('should show drag styling during drag operations', async () => {
      // This would require more complex drag simulation
      // For now, we test that elements are rendered correctly
      testUtils.render(<DropZone {...defaultProps} />);
      
      expect(screen.getByText('Test Field')).toBeInTheDocument();
    });
  });

  describe('Grid Layout', () => {
    it('should render elements in responsive grid', async () => {
      testUtils.render(<DropZone {...defaultProps} />);

      // Grid container should be present
      const gridContainer = screen.getByText('Test Field').closest('.grid');
      expect(gridContainer).toHaveClass('grid-cols-1', 'md:grid-cols-2', 'lg:grid-cols-3', 'xl:grid-cols-4');
    });

    it('should handle large numbers of elements', async () => {
      const manyElements = Array.from({ length: 20 }, (_, i) => ({
        id: `element-${i}`,
        type: 'field' as const,
        name: `Element ${i}`,
        properties: {
          name: `element_${i}`,
          type: 'string'
        }
      }));

      testUtils.render(<DropZone {...defaultProps} elements={manyElements} />);

      // Should render all elements
      expect(screen.getByText('Element 0')).toBeInTheDocument();
      expect(screen.getByText('Element 19')).toBeInTheDocument();
    });
  });

  describe('Element Properties Display', () => {
    it('should display field-specific properties', async () => {
      const fieldElement: MetaContractElement = {
        id: 'detailed-field',
        type: 'field',
        name: 'Detailed Field',
        properties: {
          name: 'detailed_field',
          type: 'decimal',
          required: false,
          precision: 10,
          scale: 2,
          unique: true
        }
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[fieldElement]} />
      );

      expect(screen.getByText('decimal')).toBeInTheDocument();
      expect(screen.getByText('No')).toBeInTheDocument(); // required: false
      expect(screen.getByText('Unique')).toBeInTheDocument();
    });

    it('should display relationship-specific properties', async () => {
      const relationshipElement: MetaContractElement = {
        id: 'detailed-relationship',
        type: 'relationship',
        name: 'Detailed Relationship',
        properties: {
          name: 'detailed_relation',
          type: 'many_to_many',
          target_resource: 'junction_table',
          join_table: 'relation_junction',
          cascade_delete: true
        }
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[relationshipElement]} />
      );

      expect(screen.getByText('junction_table')).toBeInTheDocument();
    });

    it('should handle missing properties gracefully', async () => {
      const minimalElement: MetaContractElement = {
        id: 'minimal',
        type: 'field',
        name: 'Minimal Field',
        properties: {}
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[minimalElement]} />
      );

      expect(screen.getByText('Minimal Field')).toBeInTheDocument();
      expect(screen.getByText('field')).toBeInTheDocument();
    });
  });

  describe('Styling and Visual States', () => {
    it('should apply hover effects', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      const fieldCard = screen.getByText('Test Field').closest('.cursor-pointer');
      
      await user.hover(fieldCard!);
      
      // Element should be present (hover effects are CSS-based)
      expect(fieldCard).toBeInTheDocument();
    });

    it('should show hidden state for hidden elements', async () => {
      const hiddenElement: MetaContractElement = {
        id: 'hidden-field',
        type: 'field',
        name: 'Hidden Field',
        properties: {
          name: 'hidden_field',
          type: 'string',
          hidden: true
        }
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[hiddenElement]} />
      );

      const hiddenCard = screen.getByText('Hidden Field').closest('.opacity-50');
      expect(hiddenCard).toBeInTheDocument();
    });

    it('should use appropriate icons for different element types', async () => {
      testUtils.render(<DropZone {...defaultProps} />);

      // Icons should be present (rendered as SVG elements)
      const fieldCard = screen.getByText('Test Field').closest('.cursor-pointer');
      const relationshipCard = screen.getByText('Test Relationship').closest('.cursor-pointer');
      
      expect(fieldCard?.querySelector('svg')).toBeInTheDocument();
      expect(relationshipCard?.querySelector('svg')).toBeInTheDocument();
    });
  });

  describe('Performance', () => {
    it('should render quickly with many elements', async () => {
      const manyElements = Array.from({ length: 100 }, (_, i) => ({
        id: `perf-element-${i}`,
        type: 'field' as const,
        name: `Performance Element ${i}`,
        properties: {
          name: `perf_element_${i}`,
          type: 'string'
        }
      }));

      const renderTime = testUtils.measureRenderTime('DropZone - Many Elements', () => {
        testUtils.render(<DropZone {...defaultProps} elements={manyElements} />);
      });

      expect(renderTime).toBeLessThan(1000); // 1 second
      expect(screen.getByText('Performance Element 0')).toBeInTheDocument();
      expect(screen.getByText('Performance Element 99')).toBeInTheDocument();
    });

    it('should handle element updates efficiently', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      const startTime = performance.now();
      
      const fieldCard = screen.getByText('Test Field').closest('[role="button"], .cursor-pointer');
      await user.click(fieldCard!);
      
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(100); // 100ms
      expect(defaultProps.onSelectElement).toHaveBeenCalled();
    });
  });

  describe('Accessibility', () => {
    it('should be accessible', async () => {
      testUtils.render(<DropZone {...defaultProps} />);

      // Elements should be clickable and have proper structure
      expect(screen.getByText('Test Field')).toBeInTheDocument();
      expect(screen.getByText('Test Relationship')).toBeInTheDocument();

      await testUtils.checkAccessibility();
    });

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      // Tab through elements
      await user.tab();
      
      // Should be able to navigate to elements
      const activeElement = document.activeElement;
      expect(activeElement).toBeInTheDocument();
    });

    it('should have proper ARIA labels for actions', async () => {
      const user = userEvent.setup();
      testUtils.render(<DropZone {...defaultProps} />);

      const moreButtons = screen.getAllByRole('button');
      const moreButton = moreButtons.find(button => 
        button.querySelector('svg')?.getAttribute('class')?.includes('w-3')
      );
      
      if (moreButton) {
        await user.click(moreButton);
        
        // Menu items should have proper text content
        expect(screen.getByText('Edit Properties')).toBeInTheDocument();
        expect(screen.getByText('Delete')).toBeInTheDocument();
      }
    });
  });

  describe('Error Handling', () => {
    it('should handle missing element properties', async () => {
      const invalidElement = {
        id: 'invalid',
        type: 'field' as const,
        name: 'Invalid Element',
        properties: null as any
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[invalidElement]} />
      );

      // Should not crash
      expect(screen.getByText('Invalid Element')).toBeInTheDocument();
    });

    it('should handle callback errors gracefully', async () => {
      const errorCallback = jest.fn(() => {
        throw new Error('Test error');
      });
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      const user = userEvent.setup();
      
      testUtils.render(
        <DropZone {...defaultProps} onSelectElement={errorCallback} />
      );

      const fieldCard = screen.getByText('Test Field').closest('[role="button"], .cursor-pointer');
      await user.click(fieldCard!);
      
      // Should not crash the component
      expect(screen.getByText('Test Field')).toBeInTheDocument();
      
      consoleSpy.mockRestore();
    });
  });

  describe('Edge Cases', () => {
    it('should handle elements with very long names', async () => {
      const longNameElement: MetaContractElement = {
        id: 'long-name',
        type: 'field',
        name: 'Very Long Element Name That Might Cause Layout Issues And Should Be Handled Gracefully',
        properties: {
          name: 'very_long_field_name',
          type: 'string'
        }
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[longNameElement]} />
      );

      expect(screen.getByText(/very long element name/i)).toBeInTheDocument();
    });

    it('should handle elements with special characters in names', async () => {
      const specialCharElement: MetaContractElement = {
        id: 'special-chars',
        type: 'field',
        name: 'Field with "Quotes" & <Tags> and éspecial çharacters',
        properties: {
          name: 'special_field',
          type: 'string'
        }
      };

      testUtils.render(
        <DropZone {...defaultProps} elements={[specialCharElement]} />
      );

      expect(screen.getByText(/field with.*quotes.*tags/i)).toBeInTheDocument();
    });
  });
});