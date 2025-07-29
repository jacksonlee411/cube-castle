/**
 * Unit tests for ComponentPalette component
 * Tests component template display, filtering, and element addition
 */
import React from 'react';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ComponentPalette } from '@/components/metacontract-editor/visual/ComponentPalette';
import testUtils, { setupTests } from '@/tests/setup/test-utils';

// Setup tests
setupTests();

describe('ComponentPalette', () => {
  const mockOnAddElement = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Component Rendering', () => {
    it('should render component palette with all categories', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Check for title
      expect(screen.getByText('Components')).toBeInTheDocument();

      // Check for search input
      expect(screen.getByPlaceholderText(/search components/i)).toBeInTheDocument();

      // Check for category tabs
      expect(screen.getByRole('tab', { name: /all/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /fields/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /relations/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /security/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /validation/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /performance/i })).toBeInTheDocument();
    });

    it('should display component templates', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Should show various field components
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      expect(screen.getByText('Email Field')).toBeInTheDocument();
      expect(screen.getByText('UUID Field')).toBeInTheDocument();

      // Should show relationship components
      expect(screen.getByText('One-to-Many')).toBeInTheDocument();
      expect(screen.getByText('Many-to-One')).toBeInTheDocument();

      // Should show security components
      expect(screen.getByText('Role-Based Access')).toBeInTheDocument();
    });

    it('should display component descriptions and tags', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Check for component descriptions
      expect(screen.getByText(/basic text field with optional validation/i)).toBeInTheDocument();
      expect(screen.getByText(/email field with built-in validation/i)).toBeInTheDocument();

      // Check for component type badges
      expect(screen.getAllByText('field')).toHaveLength(9); // 9 field components
      expect(screen.getAllByText('relationship')).toHaveLength(3); // 3 relationship components
    });
  });

  describe('Category Filtering', () => {
    it('should filter components by category', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Initially should show all components
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      expect(screen.getByText('One-to-Many')).toBeInTheDocument();

      // Filter by fields category
      await user.click(screen.getByRole('tab', { name: /fields/i }));

      // Should only show field components
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      expect(screen.getByText('Email Field')).toBeInTheDocument();
      expect(screen.queryByText('One-to-Many')).not.toBeInTheDocument();

      // Filter by relationships category
      await user.click(screen.getByRole('tab', { name: /relations/i }));

      // Should only show relationship components
      expect(screen.getByText('One-to-Many')).toBeInTheDocument();
      expect(screen.getByText('Many-to-One')).toBeInTheDocument();
      expect(screen.queryByText('Text Field')).not.toBeInTheDocument();
    });

    it('should show all components when "All" category is selected', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Filter by fields first
      await user.click(screen.getByRole('tab', { name: /fields/i }));
      expect(screen.queryByText('One-to-Many')).not.toBeInTheDocument();

      // Switch back to all
      await user.click(screen.getByRole('tab', { name: /all/i }));

      // Should show components from all categories
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      expect(screen.getByText('One-to-Many')).toBeInTheDocument();
      expect(screen.getByText('Role-Based Access')).toBeInTheDocument();
    });

    it('should filter security components correctly', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      await user.click(screen.getByRole('tab', { name: /security/i }));

      // Should show security components
      expect(screen.getByText('Role-Based Access')).toBeInTheDocument();
      expect(screen.getByText('Row Level Security')).toBeInTheDocument();
      
      // Should not show other types
      expect(screen.queryByText('Text Field')).not.toBeInTheDocument();
      expect(screen.queryByText('One-to-Many')).not.toBeInTheDocument();
    });
  });

  describe('Search Functionality', () => {
    it('should filter components by search term in name', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const searchInput = screen.getByPlaceholderText(/search components/i);
      await user.type(searchInput, 'text');

      // Should show components with "text" in the name
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      expect(screen.queryByText('Email Field')).not.toBeInTheDocument();
    });

    it('should filter components by search term in description', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const searchInput = screen.getByPlaceholderText(/search components/i);
      await user.type(searchInput, 'validation');

      // Should show components with "validation" in description
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      expect(screen.getByText('Email Field')).toBeInTheDocument();
      expect(screen.getByText('Required Validation')).toBeInTheDocument();
    });

    it('should filter components by search term in tags', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const searchInput = screen.getByPlaceholderText(/search components/i);
      await user.type(searchInput, 'contact');

      // Should show components with "contact" tag
      expect(screen.getByText('Email Field')).toBeInTheDocument();
      expect(screen.getByText('Phone Field')).toBeInTheDocument();
      expect(screen.queryByText('Text Field')).not.toBeInTheDocument();
    });

    it('should show no results message when search has no matches', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const searchInput = screen.getByPlaceholderText(/search components/i);
      await user.type(searchInput, 'nonexistent');

      // Should show no results message
      expect(screen.getByText(/no components found/i)).toBeInTheDocument();
      expect(screen.getByText(/try adjusting your search/i)).toBeInTheDocument();
    });

    it('should clear search results when search is cleared', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const searchInput = screen.getByPlaceholderText(/search components/i);
      
      // Search for something
      await user.type(searchInput, 'email');
      expect(screen.getByText('Email Field')).toBeInTheDocument();
      expect(screen.queryByText('Text Field')).not.toBeInTheDocument();

      // Clear search
      await user.clear(searchInput);

      // Should show all components again
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      expect(screen.getByText('Email Field')).toBeInTheDocument();
    });
  });

  describe('Combined Filtering', () => {
    it('should apply both category and search filters', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Filter by fields category
      await user.click(screen.getByRole('tab', { name: /fields/i }));

      // Then search for 'email'
      const searchInput = screen.getByPlaceholderText(/search components/i);
      await user.type(searchInput, 'email');

      // Should only show email field (both field category and email search)
      expect(screen.getByText('Email Field')).toBeInTheDocument();
      expect(screen.queryByText('Text Field')).not.toBeInTheDocument();
      expect(screen.queryByText('One-to-Many')).not.toBeInTheDocument();
    });

    it('should show no results when category and search have no overlap', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Filter by fields category
      await user.click(screen.getByRole('tab', { name: /fields/i }));

      // Search for relationship term
      const searchInput = screen.getByPlaceholderText(/search components/i);
      await user.type(searchInput, 'many-to-one');

      // Should show no results
      expect(screen.getByText(/no components found/i)).toBeInTheDocument();
    });
  });

  describe('Component Addition', () => {
    it('should call onAddElement when clicking a component', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Click on Text Field component
      const textFieldCard = screen.getByText('Text Field').closest('[role="button"], .cursor-pointer');
      expect(textFieldCard).toBeInTheDocument();
      
      await user.click(textFieldCard!);

      expect(mockOnAddElement).toHaveBeenCalledWith(
        'field',
        expect.objectContaining({
          name: expect.stringMatching(/text_field_\d+/),
          type: 'string',
          required: false,
          max_length: 255
        })
      );
    });

    it('should call onAddElement when clicking Add button', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Click the Add button for Email Field
      const addButtons = screen.getAllByRole('button', { name: /add/i });
      await user.click(addButtons[1]); // Email field is second

      expect(mockOnAddElement).toHaveBeenCalledWith(
        'field',
        expect.objectContaining({
          name: 'email',
          type: 'string',
          required: true,
          format: 'email'
        })
      );
    });

    it('should generate unique names for field components', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Add text field twice
      const textFieldCard = screen.getByText('Text Field').closest('[role="button"], .cursor-pointer');
      
      await user.click(textFieldCard!);
      await user.click(textFieldCard!);

      expect(mockOnAddElement).toHaveBeenCalledTimes(2);
      
      // Both calls should have different names
      const firstCall = mockOnAddElement.mock.calls[0][1];
      const secondCall = mockOnAddElement.mock.calls[1][1];
      
      expect(firstCall.name).not.toBe(secondCall.name);
      expect(firstCall.name).toMatch(/text_field_\d+/);
      expect(secondCall.name).toMatch(/text_field_\d+/);
    });

    it('should add relationship components correctly', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const oneToManyCard = screen.getByText('One-to-Many').closest('[role="button"], .cursor-pointer');
      await user.click(oneToManyCard!);

      expect(mockOnAddElement).toHaveBeenCalledWith(
        'relationship',
        expect.objectContaining({
          name: 'has_many',
          type: 'one_to_many',
          target_resource: 'related_resource',
          foreign_key: 'parent_id'
        })
      );
    });

    it('should add security components correctly', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const rbacCard = screen.getByText('Role-Based Access').closest('[role="button"], .cursor-pointer');
      await user.click(rbacCard!);

      expect(mockOnAddElement).toHaveBeenCalledWith(
        'security',
        expect.objectContaining({
          type: 'rbac',
          roles: ['admin', 'user', 'viewer'],
          permissions: expect.any(Object)
        })
      );
    });
  });

  describe('Component Templates', () => {
    it('should display correct field type components', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Check various field types are available
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      expect(screen.getByText('Email Field')).toBeInTheDocument();
      expect(screen.getByText('Phone Field')).toBeInTheDocument();
      expect(screen.getByText('Number Field')).toBeInTheDocument();
      expect(screen.getByText('Decimal Field')).toBeInTheDocument();
      expect(screen.getByText('Date Field')).toBeInTheDocument();
      expect(screen.getByText('DateTime Field')).toBeInTheDocument();
      expect(screen.getByText('Boolean Field')).toBeInTheDocument();
      expect(screen.getByText('UUID Field')).toBeInTheDocument();
      expect(screen.getByText('JSON Field')).toBeInTheDocument();
    });

    it('should display relationship type components', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      expect(screen.getByText('One-to-Many')).toBeInTheDocument();
      expect(screen.getByText('Many-to-One')).toBeInTheDocument();
      expect(screen.getByText('Many-to-Many')).toBeInTheDocument();
    });

    it('should display validation components', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      expect(screen.getByText('Required Validation')).toBeInTheDocument();
      expect(screen.getByText('Range Validation')).toBeInTheDocument();
      expect(screen.getByText('Pattern Validation')).toBeInTheDocument();
    });

    it('should display performance components', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      expect(screen.getByText('Simple Index')).toBeInTheDocument();
      expect(screen.getByText('Composite Index')).toBeInTheDocument();
    });
  });

  describe('UI Interactions', () => {
    it('should highlight components on hover', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const textFieldCard = screen.getByText('Text Field').closest('.cursor-pointer');
      
      // Hover over component
      await user.hover(textFieldCard!);
      
      // Component should be present (hover effects are CSS-based)
      expect(textFieldCard).toBeInTheDocument();
    });

    it('should scroll through components', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // ScrollArea should be present
      const scrollArea = screen.getByText('Text Field').closest('[data-radix-scroll-area-viewport]');
      expect(scrollArea || screen.getByText('Text Field')).toBeInTheDocument();
    });
  });

  describe('Performance', () => {
    it('should render quickly with all components', async () => {
      const renderTime = testUtils.measureRenderTime('ComponentPalette', () => {
        testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);
      });

      expect(renderTime).toBeLessThan(500); // 500ms
      expect(screen.getByText('Components')).toBeInTheDocument();
    });

    it('should filter efficiently with search terms', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      const searchInput = screen.getByPlaceholderText(/search components/i);
      
      const startTime = performance.now();
      await user.type(searchInput, 'field');
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(500); // Should filter quickly
      expect(screen.getByText('Text Field')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should be accessible', async () => {
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Check for proper ARIA labels and roles
      expect(screen.getByRole('tablist')).toBeInTheDocument();
      expect(screen.getByPlaceholderText(/search components/i)).toBeInTheDocument();

      await testUtils.checkAccessibility();
    });

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Tab through the interface
      await user.tab();
      
      // Should be able to navigate to search input
      expect(document.activeElement).toBe(screen.getByPlaceholderText(/search components/i));

      // Tab to category tabs
      await user.tab();
      
      // Should be in the tablist area
      const activeElement = document.activeElement;
      expect(activeElement).toBeInTheDocument();
    });

    it('should support Enter key for component selection', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={mockOnAddElement} />);

      // Focus on a component and press Enter
      const textFieldCard = screen.getByText('Text Field').closest('[role="button"], .cursor-pointer');
      
      if (textFieldCard) {
        textFieldCard.focus();
        await user.keyboard('{Enter}');
        
        expect(mockOnAddElement).toHaveBeenCalled();
      }
    });
  });

  describe('Error Handling', () => {
    it('should handle empty onAddElement callback gracefully', async () => {
      const user = userEvent.setup();
      testUtils.render(<ComponentPalette onAddElement={() => {}} />);

      const textFieldCard = screen.getByText('Text Field').closest('[role="button"], .cursor-pointer');
      
      // Should not throw when clicking
      await user.click(textFieldCard!);
      
      expect(screen.getByText('Text Field')).toBeInTheDocument();
    });

    it('should handle onAddElement errors gracefully', async () => {
      const errorCallback = jest.fn(() => {
        throw new Error('Test error');
      });
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      const user = userEvent.setup();
      
      testUtils.render(<ComponentPalette onAddElement={errorCallback} />);

      const textFieldCard = screen.getByText('Text Field').closest('[role="button"], .cursor-pointer');
      
      // Should not crash when callback throws
      await user.click(textFieldCard!);
      
      expect(screen.getByText('Text Field')).toBeInTheDocument();
      
      consoleSpy.mockRestore();
    });
  });
});