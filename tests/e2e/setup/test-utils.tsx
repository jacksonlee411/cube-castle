/**
 * Enhanced test utilities for React Testing Library with meta-contract editor specific helpers
 */
import React, { ReactElement } from 'react';
import { render, RenderOptions, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from 'next-themes';
import { Toaster } from 'sonner';
import { setupMSW } from './msw.setup';

// Custom providers wrapper for testing
interface ProvidersProps {
  children: React.ReactNode;
}

const TestProviders: React.FC<ProvidersProps> = ({ children }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider
        attribute="class"
        defaultTheme="light"
        enableSystem={false}
        disableTransitionOnChange
      >
        {children}
        <Toaster />
      </ThemeProvider>
    </QueryClientProvider>
  );
};

// Custom render function with providers
const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => {
  return render(ui, { wrapper: TestProviders, ...options });
};

// Test utilities
export const testUtils = {
  // Re-export everything from RTL
  ...screen,
  render: customRender,
  waitFor,
  userEvent,

  // Custom utilities for meta-contract editor
  async findByTestId(testId: string, options?: any) {
    return screen.findByTestId(testId, options);
  },

  async waitForLoadingToFinish() {
    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });
  },

  async waitForCompilationToComplete() {
    await waitFor(() => {
      expect(screen.queryByText(/compiling/i)).not.toBeInTheDocument();
    });
  },

  // Monaco Editor specific utilities
  async getMonacoEditor() {
    return screen.findByTestId('monaco-editor');
  },

  async typeInMonacoEditor(content: string) {
    const editor = await this.getMonacoEditor();
    await userEvent.clear(editor);
    await userEvent.type(editor, content);
  },

  // Visual Editor utilities
  async dragAndDrop(sourceTestId: string, targetTestId: string) {
    const source = screen.getByTestId(sourceTestId);
    const target = screen.getByTestId(targetTestId);

    await userEvent.pointer([
      { keys: '[MouseLeft>]', target: source },
      { coords: { x: target.getBoundingClientRect().x, y: target.getBoundingClientRect().y } },
      { keys: '[/MouseLeft]' }
    ]);
  },

  // Template utilities
  async selectTemplate(templateName: string) {
    const templateButton = screen.getByRole('button', { name: new RegExp(templateName, 'i') });
    await userEvent.click(templateButton);
  },

  async searchTemplates(query: string) {
    const searchInput = screen.getByPlaceholderText(/search templates/i);
    await userEvent.clear(searchInput);
    await userEvent.type(searchInput, query);
  },

  // Form utilities
  async fillForm(formData: Record<string, string>) {
    for (const [label, value] of Object.entries(formData)) {
      const field = screen.getByLabelText(new RegExp(label, 'i'));
      await userEvent.clear(field);
      await userEvent.type(field, value);
    }
  },

  async clickButton(buttonName: string | RegExp) {
    const button = screen.getByRole('button', { name: buttonName });
    await userEvent.click(button);
  },

  // Assertion utilities
  expectElementToBeVisible(testId: string) {
    expect(screen.getByTestId(testId)).toBeVisible();
  },

  expectElementToHaveText(testId: string, text: string | RegExp) {
    expect(screen.getByTestId(testId)).toHaveTextContent(text);
  },

  async expectToast(message: string | RegExp) {
    await waitFor(() => {
      expect(screen.getByText(message)).toBeInTheDocument();
    });
  },

  expectFormValidation(fieldLabel: string, errorMessage: string | RegExp) {
    const field = screen.getByLabelText(new RegExp(fieldLabel, 'i'));
    expect(field).toBeInvalid();
    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  },

  // WebSocket utilities
  simulateWebSocketMessage(message: any) {
    // This will be used with the mocked WebSocket
    const messageEvent = new MessageEvent('message', {
      data: JSON.stringify(message)
    });
    window.dispatchEvent(messageEvent);
  },

  // Accessibility utilities
  async checkAccessibility(element?: HTMLElement) {
    const { axe } = await import('@axe-core/react');
    const results = await axe(element || document.body);
    expect(results).toHaveNoViolations();
  },

  // Performance utilities
  measureRenderTime(componentName: string, renderFn: () => void) {
    const start = performance.now();
    renderFn();
    const end = performance.now();
    console.log(`${componentName} render time: ${end - start}ms`);
    return end - start;
  }
};

// Setup function to be called in test files
export const setupTests = () => {
  setupMSW();
  
  // Setup global mocks
  beforeEach(() => {
    // Clear all mocks
    jest.clearAllMocks();
    
    // Mock window.matchMedia
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      })),
    });

    // Mock IntersectionObserver
    global.IntersectionObserver = jest.fn().mockImplementation(() => ({
      observe: jest.fn(),
      unobserve: jest.fn(),
      disconnect: jest.fn(),
    }));

    // Mock ResizeObserver
    global.ResizeObserver = jest.fn().mockImplementation(() => ({
      observe: jest.fn(),
      unobserve: jest.fn(),
      disconnect: jest.fn(),
    }));

    // Mock getComputedStyle
    global.getComputedStyle = jest.fn().mockImplementation(() => ({
      getPropertyValue: jest.fn().mockReturnValue(''),
    }));
  });
};

// Custom matchers
expect.extend({
  toBeAccessible: async function(received: HTMLElement) {
    const { axe } = await import('@axe-core/react');
    const results = await axe(received);
    
    if (results.violations.length === 0) {
      return {
        message: () => `expected element to have accessibility violations`,
        pass: true,
      };
    } else {
      return {
        message: () => 
          `expected element to be accessible, but found ${results.violations.length} violations:\n` +
          results.violations.map(v => `- ${v.description}`).join('\n'),
        pass: false,
      };
    }
  },
});

// Export everything
export * from '@testing-library/react';
export { testUtils as render };
export { userEvent };
export default testUtils;