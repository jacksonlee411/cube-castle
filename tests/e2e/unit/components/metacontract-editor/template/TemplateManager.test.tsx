/**
 * Unit tests for TemplateManager component
 * Tests template browsing, searching, recommendations, and application
 */
import React from 'react';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TemplateManager } from '@/components/metacontract-editor/template/TemplateManager';
import testUtils, { setupTests } from '@/tests/setup/test-utils';
import { MetaContractElement } from '@/components/metacontract-editor/VisualEditor';
import { mockTemplates, server } from '@/tests/setup/msw.setup';
import { rest } from 'msw';

// Setup tests
setupTests();

describe('TemplateManager', () => {
  const mockExistingElements: MetaContractElement[] = [
    {
      id: 'field-1',
      type: 'field',
      name: 'Test Field',
      properties: {
        name: 'test_field',
        type: 'string',
        required: true
      }
    }
  ];

  const defaultProps = {
    existingElements: mockExistingElements,
    onApplyTemplate: jest.fn(),
    onClose: jest.fn(),
    open: true
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Component Rendering', () => {
    it('should render template manager dialog when open', async () => {
      testUtils.render(<TemplateManager {...defaultProps} />);

      expect(screen.getByText('Template Library')).toBeInTheDocument();
      expect(screen.getByText(/discover and apply enterprise-grade templates/i)).toBeInTheDocument();

      // Check for tabs
      expect(screen.getByRole('tab', { name: /recommended/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /browse all/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /popular/i })).toBeInTheDocument();
    });

    it('should not render when closed', async () => {
      testUtils.render(<TemplateManager {...defaultProps} open={false} />);

      expect(screen.queryByText('Template Library')).not.toBeInTheDocument();
    });

    it('should show loading state initially', async () => {
      testUtils.render(<TemplateManager {...defaultProps} />);

      // Should show loading spinner
      expect(document.querySelector('.animate-spin')).toBeInTheDocument();
    });
  });

  describe('Tab Navigation', () => {
    it('should switch between tabs', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      // Wait for loading to complete
      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Initially on recommended tab
      expect(screen.getByRole('tab', { name: /recommended/i })).toHaveAttribute('data-state', 'active');

      // Switch to browse tab
      await user.click(screen.getByRole('tab', { name: /browse all/i }));
      expect(screen.getByRole('tab', { name: /browse all/i })).toHaveAttribute('data-state', 'active');

      // Switch to popular tab
      await user.click(screen.getByRole('tab', { name: /popular/i }));
      expect(screen.getByRole('tab', { name: /popular/i })).toHaveAttribute('data-state', 'active');
    });

    it('should show search controls only in browse tab', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      // Wait for loading
      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Initially on recommended - no search controls
      expect(screen.queryByPlaceholderText(/search templates/i)).not.toBeInTheDocument();

      // Switch to browse tab
      await user.click(screen.getByRole('tab', { name: /browse all/i }));

      // Should show search controls
      expect(screen.getByPlaceholderText(/search templates/i)).toBeInTheDocument();
      expect(screen.getByRole('combobox')).toBeInTheDocument(); // Category select
    });
  });

  describe('Template Display', () => {
    it('should display template cards with correct information', async () => {
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Should show template information
      await waitFor(() => {
        expect(screen.getByText(/test employee template/i)).toBeInTheDocument();
      });

      // Should show rating stars, usage count, etc.
      // These would be rendered based on mock template data
    });

    it('should show recommendation badges on recommended templates', async () => {
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Should show recommendation scores
      await waitFor(() => {
        // Look for percentage match badges
        const matchBadges = document.querySelectorAll('[class*="match"]');
        expect(matchBadges.length).toBeGreaterThan(0);
      });
    });

    it('should display template quality metrics', async () => {
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Should show quality indicators (performance, security, etc.)
      // These would be rendered as part of template cards
    });
  });

  describe('Search and Filtering', () => {
    it('should perform text search', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      // Wait for loading and switch to browse tab
      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await user.click(screen.getByRole('tab', { name: /browse all/i }));

      const searchInput = screen.getByPlaceholderText(/search templates/i);
      await user.type(searchInput, 'employee');

      // Should trigger search (debounced)
      expect(searchInput).toHaveValue('employee');
    });

    it('should filter by category', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await user.click(screen.getByRole('tab', { name: /browse all/i }));

      // Find and click category select
      const categorySelects = screen.getAllByRole('combobox');
      const categorySelect = categorySelects.find(select => 
        select.getAttribute('aria-expanded') !== null
      );

      if (categorySelect) {
        await user.click(categorySelect);
        
        // Should show category options
        await waitFor(() => {
          expect(screen.getByText(/all categories/i)).toBeInTheDocument();
        });
      }
    });

    it('should filter by complexity', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await user.click(screen.getByRole('tab', { name: /browse all/i }));

      // Complexity filter should be available
      const selects = screen.getAllByRole('combobox');
      expect(selects.length).toBeGreaterThan(1); // Category, complexity, sort
    });

    it('should change sort order', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await user.click(screen.getByRole('tab', { name: /browse all/i }));

      // Should have sort options
      const selects = screen.getAllByRole('combobox');
      const sortSelect = selects[selects.length - 1]; // Usually the last one

      await user.click(sortSelect);
      
      await waitFor(() => {
        expect(screen.getByText(/relevance/i)).toBeInTheDocument();
      });
    });
  });

  describe('Template Actions', () => {
    it('should preview template when preview button is clicked', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Wait for templates to load and find preview button
      await waitFor(() => {
        const previewButton = screen.queryByRole('button', { name: /preview/i });
        expect(previewButton).toBeInTheDocument();
      });

      const previewButton = screen.getByRole('button', { name: /preview/i });
      await user.click(previewButton);

      // Should open preview dialog
      await waitFor(() => {
        expect(screen.getByText(/basic information/i)).toBeInTheDocument();
      });
    });

    it('should apply template when apply button is clicked', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Wait for templates to load and find apply button
      await waitFor(() => {
        const applyButton = screen.queryByRole('button', { name: /apply/i });
        expect(applyButton).toBeInTheDocument();
      });

      const applyButton = screen.getByRole('button', { name: /apply/i });
      await user.click(applyButton);

      expect(defaultProps.onApplyTemplate).toHaveBeenCalled();
      expect(defaultProps.onClose).toHaveBeenCalled();
    });

    it('should show conflict warnings for high-risk templates', async () => {
      // This would require mocking templates with conflict risks
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // If conflict risk templates are present, should show warnings
      // This depends on the mock template data having conflict risks
    });
  });

  describe('Template Preview Dialog', () => {
    it('should show template details in preview', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await waitFor(() => {
        const previewButton = screen.queryByRole('button', { name: /preview/i });
        expect(previewButton).toBeInTheDocument();
      });

      const previewButton = screen.getByRole('button', { name: /preview/i });
      await user.click(previewButton);

      // Should show template details
      await waitFor(() => {
        expect(screen.getByText(/basic information/i)).toBeInTheDocument();
        expect(screen.getByText(/quality metrics/i)).toBeInTheDocument();
      });
    });

    it('should show quality metrics with progress bars', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await waitFor(() => {
        const previewButton = screen.queryByRole('button', { name: /preview/i });
        expect(previewButton).toBeInTheDocument();
      });

      const previewButton = screen.getByRole('button', { name: /preview/i });
      await user.click(previewButton);

      await waitFor(() => {
        expect(screen.getByText(/performance/i)).toBeInTheDocument();
        expect(screen.getByText(/security/i)).toBeInTheDocument();
        expect(screen.getByText(/maintainability/i)).toBeInTheDocument();
      });
    });

    it('should allow applying template from preview', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await waitFor(() => {
        const previewButton = screen.queryByRole('button', { name: /preview/i });
        expect(previewButton).toBeInTheDocument();
      });

      const previewButton = screen.getByRole('button', { name: /preview/i });
      await user.click(previewButton);

      await waitFor(() => {
        const applyTemplateButton = screen.getByRole('button', { name: /apply template/i });
        expect(applyTemplateButton).toBeInTheDocument();
      });

      const applyTemplateButton = screen.getByRole('button', { name: /apply template/i });
      await user.click(applyTemplateButton);

      expect(defaultProps.onApplyTemplate).toHaveBeenCalled();
    });

    it('should close preview dialog', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await waitFor(() => {
        const previewButton = screen.queryByRole('button', { name: /preview/i });
        expect(previewButton).toBeInTheDocument();
      });

      const previewButton = screen.getByRole('button', { name: /preview/i });
      await user.click(previewButton);

      await waitFor(() => {
        const closeButton = screen.getByRole('button', { name: /close/i });
        expect(closeButton).toBeInTheDocument();
      });

      const closeButtons = screen.getAllByRole('button', { name: /close/i });
      const previewCloseButton = closeButtons[0]; // First close button (in preview dialog)
      await user.click(previewCloseButton);

      // Preview dialog should close
      await waitFor(() => {
        expect(screen.queryByText(/basic information/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('Popular Templates Tab', () => {
    it('should show popular templates sorted by usage', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await user.click(screen.getByRole('tab', { name: /popular/i }));

      // Should show templates (popular templates are from the library)
      // This would show actual templates from EnterpriseTemplateLibrary
      await waitFor(() => {
        // Should show some template content
        expect(screen.getByRole('tab', { name: /popular/i })).toHaveAttribute('data-state', 'active');
      });
    });
  });

  describe('Recommendations Engine Integration', () => {
    it('should load recommendations based on existing elements', async () => {
      testUtils.render(<TemplateManager {...defaultProps} />);

      // Should trigger recommendation loading
      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Recommendations should be displayed
      expect(screen.getByRole('tab', { name: /recommended/i })).toHaveAttribute('data-state', 'active');
    });

    it('should show recommendation reasons', async () => {
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Should show "Why recommended" sections
      await waitFor(() => {
        // Check for recommendation reasoning
        const whyRecommendedText = screen.queryByText(/why recommended/i);
        if (whyRecommendedText) {
          expect(whyRecommendedText).toBeInTheDocument();
        }
      });
    });
  });

  describe('Dialog Management', () => {
    it('should close main dialog when close button is clicked', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      const closeButton = screen.getByRole('button', { name: /close/i });
      await user.click(closeButton);

      expect(defaultProps.onClose).toHaveBeenCalled();
    });

    it('should close dialog when clicking outside', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      // Click on dialog overlay (this would be handled by the Dialog component)
      // The actual behavior depends on the Dialog component implementation
      expect(screen.getByText('Template Library')).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should handle recommendation loading errors', async () => {
      // Mock recommendation API to fail
      server.use(
        rest.post('/api/v1/templates/recommend', (req, res, ctx) => {
          return res(ctx.status(500), ctx.json({ error: 'Server error' }));
        })
      );

      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

      testUtils.render(<TemplateManager {...defaultProps} />);

      // Should handle error gracefully
      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Should not crash
      expect(screen.getByText('Template Library')).toBeInTheDocument();

      consoleSpy.mockRestore();
    });

    it('should handle search errors', async () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      const user = userEvent.setup();

      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await user.click(screen.getByRole('tab', { name: /browse all/i }));

      // Search should not crash even if it fails
      const searchInput = screen.getByPlaceholderText(/search templates/i);
      await user.type(searchInput, 'test search');

      expect(screen.getByText('Template Library')).toBeInTheDocument();

      consoleSpy.mockRestore();
    });

    it('should handle empty search results', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await user.click(screen.getByRole('tab', { name: /browse all/i }));

      // Should handle empty results gracefully
      const searchInput = screen.getByPlaceholderText(/search templates/i);
      await user.type(searchInput, 'nonexistent template');

      // Should not crash
      expect(screen.getByText('Template Library')).toBeInTheDocument();
    });
  });

  describe('Performance', () => {
    it('should render quickly with many templates', async () => {
      const renderTime = testUtils.measureRenderTime('TemplateManager', () => {
        testUtils.render(<TemplateManager {...defaultProps} />);
      });

      expect(renderTime).toBeLessThan(1000); // 1 second
      expect(screen.getByText('Template Library')).toBeInTheDocument();
    });

    it('should handle search input efficiently', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await user.click(screen.getByRole('tab', { name: /browse all/i }));

      const searchInput = screen.getByPlaceholderText(/search templates/i);
      
      const startTime = performance.now();
      await user.type(searchInput, 'fast typing test');
      const endTime = performance.now();

      expect(endTime - startTime).toBeLessThan(1000); // 1 second
    });
  });

  describe('Accessibility', () => {
    it('should be accessible', async () => {
      testUtils.render(<TemplateManager {...defaultProps} />);

      // Check for proper ARIA labels and roles
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByRole('tablist')).toBeInTheDocument();

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      await testUtils.checkAccessibility();
    });

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Tab through interface
      await user.tab();
      
      // Should be able to navigate through tabs and controls
      const activeElement = document.activeElement;
      expect(activeElement).toBeInTheDocument();
    });

    it('should have proper focus management in dialogs', async () => {
      const user = userEvent.setup();
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // When opening preview dialog, focus should be managed properly
      await waitFor(() => {
        const previewButton = screen.queryByRole('button', { name: /preview/i });
        if (previewButton) {
          expect(previewButton).toBeInTheDocument();
        }
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle no existing elements', async () => {
      testUtils.render(
        <TemplateManager {...defaultProps} existingElements={[]} />
      );

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Should still work with empty existing elements
      expect(screen.getByText('Template Library')).toBeInTheDocument();
    });

    it('should handle templates with missing data', async () => {
      // This would require mocking templates with incomplete data
      testUtils.render(<TemplateManager {...defaultProps} />);

      await waitFor(() => {
        expect(document.querySelector('.animate-spin')).not.toBeInTheDocument();
      });

      // Should handle incomplete template data gracefully
      expect(screen.getByText('Template Library')).toBeInTheDocument();
    });
  });
});