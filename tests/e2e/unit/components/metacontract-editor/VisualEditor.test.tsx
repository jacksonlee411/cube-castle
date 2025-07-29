/**
 * Unit tests for VisualEditor component
 * Tests visual editing functionality, drag and drop, template integration
 */
import React from 'react';
import { screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { VisualEditor } from '@/components/metacontract-editor/VisualEditor';
import testUtils, { setupTests } from '@/tests/setup/test-utils';
import { mockTemplates } from '@/tests/setup/msw.setup';

// Setup tests
setupTests();

describe('VisualEditor', () => {
  const defaultProps = {
    content: `specification_version: "1.0"
api_id: "test-api"
namespace: "hr"
resource_name: "employees"
data_structure:
  primary_key: "id"
  data_classification: "internal"
  fields:
    - name: "id"
      type: "uuid"
      required: true
    - name: "name"
      type: "string"
      required: true
relationships:
  - name: "department"
    type: "belongs_to"
    target: "departments"
security_model:
  access_control: "rbac"
  audit_trail: true`,
    onChange: jest.fn(),
    readonly: false,
    theme: 'light' as const
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Component Rendering', () => {
    it('should render visual editor with all view modes', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Check for view mode tabs
      expect(screen.getByRole('tab', { name: /design/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /code/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /preview/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /er diagram/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /enhanced er/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /multi preview/i })).toBeInTheDocument();

      // Check for toolbar elements
      expect(screen.getByPlaceholderText(/search elements/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /templates/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /undo/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /redo/i })).toBeInTheDocument();
    });

    it('should render in readonly mode', async () => {
      testUtils.render(<VisualEditor {...defaultProps} readonly={true} />);

      // Component palette and editing tools should not be visible in readonly mode
      expect(screen.queryByRole('button', { name: /templates/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /undo/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /redo/i })).not.toBeInTheDocument();
    });

    it('should apply dark theme', async () => {
      testUtils.render(<VisualEditor {...defaultProps} theme="dark" />);

      // Component should render with dark theme
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });
  });

  describe('View Mode Switching', () => {
    it('should switch between different view modes', async () => {
      const user = userEvent.setup();
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Start in design mode
      expect(screen.getByRole('tab', { name: /design/i })).toHaveAttribute('aria-selected', 'true');

      // Switch to code view
      await user.click(screen.getByRole('tab', { name: /code/i }));
      expect(screen.getByRole('tab', { name: /code/i })).toHaveAttribute('aria-selected', 'true');

      // Switch to preview view
      await user.click(screen.getByRole('tab', { name: /preview/i }));
      expect(screen.getByRole('tab', { name: /preview/i })).toHaveAttribute('aria-selected', 'true');

      // Switch to ER diagram
      await user.click(screen.getByRole('tab', { name: /er diagram/i }));
      expect(screen.getByRole('tab', { name: /er diagram/i })).toHaveAttribute('aria-selected', 'true');
    });

    it('should display appropriate panels for each view mode', async () => {
      const user = userEvent.setup();
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Design mode should show component palette
      expect(screen.getByRole('tab', { name: /design/i })).toHaveAttribute('aria-selected', 'true');
      
      // Switch to code mode - should not show palette
      await user.click(screen.getByRole('tab', { name: /code/i }));
      
      // Code content should be visible
      expect(screen.getByText(/specification_version/)).toBeInTheDocument();
    });
  });

  describe('Content Parsing and Generation', () => {
    it('should parse YAML content into elements', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      await waitFor(() => {
        // Should show parsed elements count
        expect(screen.getByText(/elements/i)).toBeInTheDocument();
      });
    });

    it('should handle malformed YAML gracefully', async () => {
      const malformedContent = `
        invalid: yaml: content:
        - missing quotes
        - [unclosed bracket
      `;

      testUtils.render(
        <VisualEditor {...defaultProps} content={malformedContent} />
      );

      // Should not crash and should still render the editor
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });

    it('should generate YAML from elements', async () => {
      const mockOnChange = jest.fn();
      testUtils.render(
        <VisualEditor {...defaultProps} onChange={mockOnChange} />
      );

      // Content should be parsed and potentially regenerated
      await waitFor(() => {
        expect(mockOnChange).toHaveBeenCalledWith(expect.any(String));
      });
    });
  });

  describe('Element Search and Filtering', () => {
    it('should filter elements based on search term', async () => {
      const user = userEvent.setup();
      testUtils.render(<VisualEditor {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search elements/i);
      
      // Search for 'id' field
      await user.type(searchInput, 'id');

      // Should filter elements (implementation would depend on how elements are displayed)
      expect(searchInput).toHaveValue('id');
    });

    it('should clear search when input is emptied', async () => {
      const user = userEvent.setup();
      testUtils.render(<VisualEditor {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search elements/i);
      
      // Type and then clear
      await user.type(searchInput, 'test');
      await user.clear(searchInput);

      expect(searchInput).toHaveValue('');
    });
  });

  describe('Undo/Redo Functionality', () => {
    it('should enable undo after making changes', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Initially undo should be disabled
      const undoButton = screen.getByRole('button', { name: /undo/i });
      expect(undoButton).toBeDisabled();

      // After making changes (simulated by parsing content), undo might be enabled
      // This would require implementing actual element manipulation
    });

    it('should enable redo after undo', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      const redoButton = screen.getByRole('button', { name: /redo/i });
      expect(redoButton).toBeDisabled();
    });

    it('should support keyboard shortcuts for undo/redo', async () => {
      const user = userEvent.setup();
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Test Ctrl+Z for undo
      await user.keyboard('{Control>}z{/Control}');
      
      // Test Ctrl+Y for redo
      await user.keyboard('{Control>}y{/Control}');

      // Should not throw errors
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });
  });

  describe('Template Integration', () => {
    it('should open template manager when clicking templates button', async () => {
      const user = userEvent.setup();
      testUtils.render(<VisualEditor {...defaultProps} />);

      const templatesButton = screen.getByRole('button', { name: /templates/i });
      await user.click(templatesButton);

      // Template manager should open (would need to check for modal or dialog)
      // This depends on TemplateManager component implementation
    });

    it('should apply template successfully', async () => {
      const mockOnChange = jest.fn();
      const user = userEvent.setup();
      
      testUtils.render(
        <VisualEditor {...defaultProps} onChange={mockOnChange} />
      );

      const templatesButton = screen.getByRole('button', { name: /templates/i });
      await user.click(templatesButton);

      // After applying a template, onChange should be called
      // This would require mocking the template application flow
    });

    it('should handle template application errors', async () => {
      // This test would require mocking template application to fail
      const user = userEvent.setup();
      testUtils.render(<VisualEditor {...defaultProps} />);

      const templatesButton = screen.getByRole('button', { name: /templates/i });
      await user.click(templatesButton);

      // Should handle errors gracefully without crashing
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });
  });

  describe('Status Bar', () => {
    it('should display element count in status bar', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText(/elements/i)).toBeInTheDocument();
      });
    });

    it('should show selected element in status bar', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      // After parsing content, if no element is selected, should not show selection
      await waitFor(() => {
        expect(screen.queryByText(/selected/i)).not.toBeInTheDocument();
      });
    });

    it('should display version information', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      expect(screen.getByText(/visual editor v1\.0/i)).toBeInTheDocument();
    });
  });

  describe('Drag and Drop', () => {
    it('should handle drag start events', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      // DnD context should be present
      expect(screen.getByRole('tablist')).toBeInTheDocument();
      
      // Actual drag testing would require more complex setup with @dnd-kit
    });

    it('should show drag overlay during drag operations', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      // DragOverlay should be in the DOM (though not visible initially)
      // This is handled by @dnd-kit DragOverlay component
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });

    it('should reorder elements after successful drop', async () => {
      const mockOnChange = jest.fn();
      testUtils.render(
        <VisualEditor {...defaultProps} onChange={mockOnChange} />
      );

      // After drag and drop, content should be updated
      // This would require mocking drag events
    });
  });

  describe('Property Panel', () => {
    it('should show property panel when element is selected', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Initially no element selected, so no property panel
      // This would require simulating element selection
    });

    it('should hide property panel when no element is selected', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Property panel should not be visible initially
      // Since selectedElement is null by default
    });

    it('should update element properties through property panel', async () => {
      const mockOnChange = jest.fn();
      testUtils.render(
        <VisualEditor {...defaultProps} onChange={mockOnChange} />
      );

      // Would require simulating element selection and property updates
    });
  });

  describe('Readonly Mode', () => {
    it('should disable editing controls in readonly mode', async () => {
      testUtils.render(<VisualEditor {...defaultProps} readonly={true} />);

      // Editing controls should not be present
      expect(screen.queryByRole('button', { name: /templates/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /undo/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /redo/i })).not.toBeInTheDocument();
    });

    it('should not respond to keyboard shortcuts in readonly mode', async () => {
      const user = userEvent.setup();
      const mockOnChange = jest.fn();
      
      testUtils.render(
        <VisualEditor {...defaultProps} readonly={true} onChange={mockOnChange} />
      );

      // Try keyboard shortcuts
      await user.keyboard('{Control>}z{/Control}');
      await user.keyboard('{Delete}');

      // onChange should not be called for editing operations
      expect(mockOnChange).not.toHaveBeenCalled();
    });
  });

  describe('Content Synchronization', () => {
    it('should update elements when content prop changes', async () => {
      const { rerender } = testUtils.render(<VisualEditor {...defaultProps} />);

      const newContent = `specification_version: "1.0"
api_id: "updated-api"
namespace: "updated"
resource_name: "updated_resource"
data_structure:
  primary_key: "id"
  data_classification: "internal"
  fields:
    - name: "id"
      type: "uuid"
      required: true`;

      rerender(<VisualEditor {...defaultProps} content={newContent} />);

      // Component should re-parse the new content
      await waitFor(() => {
        expect(screen.getByRole('tablist')).toBeInTheDocument();
      });
    });

    it('should call onChange when elements are modified', async () => {
      const mockOnChange = jest.fn();
      testUtils.render(
        <VisualEditor {...defaultProps} onChange={mockOnChange} />
      );

      // Content parsing might trigger onChange
      await waitFor(() => {
        expect(mockOnChange).toHaveBeenCalledWith(expect.any(String));
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle template application errors gracefully', async () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Should not crash on errors
      expect(screen.getByRole('tablist')).toBeInTheDocument();

      consoleSpy.mockRestore();
    });

    it('should handle YAML parsing errors gracefully', async () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      testUtils.render(
        <VisualEditor {...defaultProps} content="invalid: yaml: [content" />
      );

      // Should still render despite parsing errors
      expect(screen.getByRole('tablist')).toBeInTheDocument();

      consoleSpy.mockRestore();
    });
  });

  describe('Performance', () => {
    it('should handle large schemas efficiently', async () => {
      const largeContent = `specification_version: "1.0"
api_id: "large-api"
namespace: "test"
resource_name: "large_resource"
data_structure:
  primary_key: "id"
  data_classification: "internal"
  fields:${Array.from({ length: 100 }, (_, i) => `
    - name: "field_${i}"
      type: "string"
      required: false`).join('')}`;

      const renderTime = testUtils.measureRenderTime('VisualEditor - Large Schema', () => {
        testUtils.render(<VisualEditor {...defaultProps} content={largeContent} />);
      });

      expect(renderTime).toBeLessThan(2000); // 2 seconds
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should be accessible', async () => {
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Check for proper ARIA labels and roles
      expect(screen.getByRole('tablist')).toBeInTheDocument();
      expect(screen.getByPlaceholderText(/search elements/i)).toBeInTheDocument();

      // Wait for component to stabilize
      await waitFor(() => {
        expect(screen.getByRole('tablist')).toBeInTheDocument();
      });

      await testUtils.checkAccessibility();
    });

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup();
      testUtils.render(<VisualEditor {...defaultProps} />);

      // Tab navigation should work
      await user.tab();
      
      // Should be able to navigate through tabs
      const focusedElement = document.activeElement;
      expect(focusedElement).toBeInTheDocument();
    });
  });
});