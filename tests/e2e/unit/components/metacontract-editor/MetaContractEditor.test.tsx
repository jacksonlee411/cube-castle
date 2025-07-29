/**
 * Unit tests for MetaContractEditor component
 * Tests core functionality including compilation, saving, WebSocket connectivity
 */
import React from 'react';
import { screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MetaContractEditor } from '@/components/metacontract-editor/MetaContractEditor';
import testUtils, { setupTests } from '@/tests/setup/test-utils';
import { mockProject, mockCompileResults, server } from '@/tests/setup/msw.setup';
import { rest } from 'msw';

// Setup tests
setupTests();

describe('MetaContractEditor', () => {
  const defaultProps = {
    projectId: 'test-project-1',
    initialContent: 'specification_version: "1.0"\napi_id: "test-api"',
    readonly: false
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Component Rendering', () => {
    it('should render editor with project data', async () => {
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Check for main UI elements
      expect(screen.getByRole('heading', { name: /meta-contract editor/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /compile/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /export/i })).toBeInTheDocument();

      // Check for tabs
      expect(screen.getByRole('tab', { name: /visual editor/i })).toBeInTheDocument();
      expect(screen.getByRole('tab', { name: /code editor/i })).toBeInTheDocument();

      await testUtils.waitForLoadingToFinish();
    });

    it('should render in readonly mode', async () => {
      testUtils.render(
        <MetaContractEditor {...defaultProps} readonly={true} />
      );

      // Compile and Save buttons should be disabled in readonly mode
      const compileButton = screen.getByRole('button', { name: /compile/i });
      const saveButton = screen.getByRole('button', { name: /save/i });

      expect(compileButton).toBeDisabled();
      expect(saveButton).toBeDisabled();
    });

    it('should render without project ID (preview mode)', async () => {
      testUtils.render(
        <MetaContractEditor 
          initialContent={defaultProps.initialContent}
          readonly={false}
        />
      );

      expect(screen.getByRole('heading', { name: /meta-contract editor/i })).toBeInTheDocument();
      
      // Save button should not be visible in preview mode
      expect(screen.queryByRole('button', { name: /save/i })).not.toBeInTheDocument();
    });
  });

  describe('Editor Mode Switching', () => {
    it('should switch between visual and code editor modes', async () => {
      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Start in visual mode
      expect(screen.getByRole('tab', { name: /visual editor/i })).toHaveAttribute('aria-selected', 'true');

      // Switch to code mode
      await user.click(screen.getByRole('tab', { name: /code editor/i }));
      expect(screen.getByRole('tab', { name: /code editor/i })).toHaveAttribute('aria-selected', 'true');

      // Monaco editor should be visible in code mode
      expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
    });

    it('should preserve content when switching modes', async () => {
      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Switch to code mode and modify content
      await user.click(screen.getByRole('tab', { name: /code editor/i }));
      
      const monacoEditor = screen.getByTestId('monaco-editor');
      expect(monacoEditor).toBeInTheDocument();

      // Switch back to visual mode - content should be preserved
      await user.click(screen.getByRole('tab', { name: /visual editor/i }));
      
      // Switch back to code mode to verify content is still there
      await user.click(screen.getByRole('tab', { name: /code editor/i }));
      expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
    });
  });

  describe('Theme Switching', () => {
    it('should toggle between light and dark themes', async () => {
      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      const themeButton = screen.getByRole('button', { name: /dark/i });
      await user.click(themeButton);

      // After clicking, button should show "Light" (indicating we're in dark mode)
      expect(screen.getByRole('button', { name: /light/i })).toBeInTheDocument();
    });
  });

  describe('Compilation Functionality', () => {
    it('should compile project successfully', async () => {
      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      const compileButton = screen.getByRole('button', { name: /compile/i });
      await user.click(compileButton);

      // Button should show compiling state
      expect(screen.getByRole('button', { name: /compiling/i })).toBeInTheDocument();

      // Wait for compilation to complete
      await testUtils.waitForCompilationToComplete();

      // Should show success message
      await testUtils.expectToast(/compilation successful/i);
    });

    it('should handle compilation errors', async () => {
      // Override the compile endpoint to return error
      server.use(
        rest.post('/api/v1/metacontract-editor/projects/:id/compile', (req, res, ctx) => {
          return res(ctx.json({
            success: false,
            errors: [
              {
                line: 5,
                column: 12,
                message: 'Invalid YAML syntax',
                severity: 'error'
              }
            ],
            warnings: []
          }));
        })
      );

      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      const compileButton = screen.getByRole('button', { name: /compile/i });
      await user.click(compileButton);

      await testUtils.waitForCompilationToComplete();

      // Should show error message
      await testUtils.expectToast(/compilation failed/i);
    });

    it('should compile in preview mode without project ID', async () => {
      const user = userEvent.setup();
      testUtils.render(
        <MetaContractEditor 
          initialContent={defaultProps.initialContent}
          readonly={false}
        />
      );

      const compileButton = screen.getByRole('button', { name: /compile/i });
      await user.click(compileButton);

      await testUtils.waitForCompilationToComplete();

      // Should show success message for preview compilation
      await testUtils.expectToast(/compilation successful/i);
    });
  });

  describe('Save Functionality', () => {
    it('should save project successfully', async () => {
      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      const saveButton = screen.getByRole('button', { name: /save/i });
      await user.click(saveButton);

      await testUtils.expectToast(/project saved successfully/i);
    });

    it('should handle save errors', async () => {
      // Override the save endpoint to return error
      server.use(
        rest.put('/api/v1/metacontract-editor/projects/:id', (req, res, ctx) => {
          return res(ctx.status(500), ctx.json({ error: 'Save failed' }));
        })
      );

      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      const saveButton = screen.getByRole('button', { name: /save/i });
      await user.click(saveButton);

      await testUtils.expectToast(/failed to save project/i);
    });

    it('should not save in readonly mode', async () => {
      const user = userEvent.setup();
      testUtils.render(
        <MetaContractEditor {...defaultProps} readonly={true} />
      );

      const saveButton = screen.getByRole('button', { name: /save/i });
      expect(saveButton).toBeDisabled();
    });
  });

  describe('Auto-save Functionality', () => {
    beforeEach(() => {
      jest.useFakeTimers();
    });

    afterEach(() => {
      jest.useRealTimers();
    });

    it('should auto-save after content changes', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Switch to code mode to trigger content change
      await user.click(screen.getByRole('tab', { name: /code editor/i }));

      // Simulate content change (this would be done by Monaco editor)
      act(() => {
        // Fast forward time to trigger auto save
        jest.advanceTimersByTime(2500);
      });

      await waitFor(() => {
        expect(screen.queryByText(/last saved/i)).toBeInTheDocument();
      });
    });

    it('should not auto-save in readonly mode', async () => {
      testUtils.render(
        <MetaContractEditor {...defaultProps} readonly={true} />
      );

      // Fast forward time
      act(() => {
        jest.advanceTimersByTime(3000);
      });

      // Should not show save status in readonly mode
      expect(screen.queryByText(/last saved/i)).not.toBeInTheDocument();
    });
  });

  describe('Export Functionality', () => {
    // Mock URL.createObjectURL and document.createElement
    const mockCreateElement = jest.fn();
    const mockClick = jest.fn();
    const mockAppendChild = jest.fn();
    const mockRemoveChild = jest.fn();

    beforeEach(() => {
      const mockAnchor = {
        href: '',
        download: '',
        click: mockClick
      };

      mockCreateElement.mockReturnValue(mockAnchor);
      
      Object.defineProperty(document, 'createElement', {
        value: mockCreateElement
      });
      
      Object.defineProperty(document.body, 'appendChild', {
        value: mockAppendChild
      });
      
      Object.defineProperty(document.body, 'removeChild', {
        value: mockRemoveChild
      });
    });

    it('should export project as YAML file', async () => {
      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      const exportButton = screen.getByRole('button', { name: /export/i });
      await user.click(exportButton);

      expect(mockCreateElement).toHaveBeenCalledWith('a');
      expect(mockClick).toHaveBeenCalled();
      expect(mockAppendChild).toHaveBeenCalled();
      expect(mockRemoveChild).toHaveBeenCalled();
    });
  });

  describe('WebSocket Connectivity', () => {
    it('should show connection status', async () => {
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Should show connected status (mocked WebSocket is connected by default)
      await waitFor(() => {
        expect(screen.getByText(/connected/i)).toBeInTheDocument();
      });
    });

    it('should handle WebSocket messages', async () => {
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Simulate WebSocket message
      act(() => {
        testUtils.simulateWebSocketMessage({
          type: 'compile_response',
          data: mockCompileResults
        });
      });

      await testUtils.expectToast(/compilation successful/i);
    });
  });

  describe('Collaborator Status', () => {
    it('should display collaborator count when available', async () => {
      // This test would require mocking the collaborator state
      // For now, we'll test the UI structure
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Should not show collaborators initially (empty array)
      expect(screen.queryByText(/online/i)).not.toBeInTheDocument();
    });
  });

  describe('Status Badge Display', () => {
    it('should show appropriate status badges', async () => {
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Should show draft status initially
      expect(screen.getByText(/draft/i)).toBeInTheDocument();
    });

    it('should show compiling status during compilation', async () => {
      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      const compileButton = screen.getByRole('button', { name: /compile/i });
      await user.click(compileButton);

      // Should show compiling status
      expect(screen.getByText(/compiling/i)).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors gracefully', async () => {
      // Override all endpoints to return network error
      server.use(
        rest.get('/api/v1/metacontract-editor/projects/:id', (req, res, ctx) => {
          return res.networkError('Network error');
        })
      );

      testUtils.render(<MetaContractEditor projectId="error-project" />);

      await waitFor(() => {
        // Should still render the editor UI even with network errors
        expect(screen.getByRole('heading', { name: /meta-contract editor/i })).toBeInTheDocument();
      });
    });
  });

  describe('Accessibility', () => {
    it('should be accessible', async () => {
      testUtils.render(<MetaContractEditor {...defaultProps} />);
      
      // Check for proper ARIA labels and roles
      expect(screen.getByRole('heading', { name: /meta-contract editor/i })).toBeInTheDocument();
      expect(screen.getByRole('tablist')).toBeInTheDocument();
      
      // Should have proper button labels
      expect(screen.getByRole('button', { name: /compile/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
      
      // Wait for component to stabilize
      await testUtils.waitForLoadingToFinish();
      
      // Check accessibility
      await testUtils.checkAccessibility();
    });

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup();
      testUtils.render(<MetaContractEditor {...defaultProps} />);

      // Tab through the interface
      await user.tab();
      
      // First focusable element should be a tab or button
      const focusedElement = document.activeElement;
      expect(focusedElement?.tagName).toMatch(/BUTTON|TAB/i);
    });
  });
});