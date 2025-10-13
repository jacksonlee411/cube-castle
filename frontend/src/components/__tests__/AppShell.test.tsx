// @vitest-environment jsdom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { CanvasProvider } from '@workday/canvas-kit-react/common';
import { AppShell } from '../../layout/AppShell';
import { vi } from 'vitest';

// Minimal mocks for Canvas Kit components used by AppShell tree
vi.mock('@workday/canvas-kit-react/layout', () => {
  const MockContainer = ({ children, ...rest }: React.HTMLAttributes<HTMLDivElement>) => (
    <div {...rest}>{children}</div>
  );
  return {
    Box: MockContainer,
    Flex: MockContainer,
  };
});
vi.mock('@workday/canvas-kit-react/text', () => ({
  Heading: ({ children, ...rest }: React.HTMLAttributes<HTMLHeadingElement>) => (
    <h1 {...rest}>{children}</h1>
  ),
  Text: ({ children, ...rest }: React.HTMLAttributes<HTMLSpanElement>) => (
    <span {...rest}>{children}</span>
  ),
}));
vi.mock('@workday/canvas-kit-react/icon', () => ({
  SystemIcon: ({ children, ...rest }: React.HTMLAttributes<HTMLSpanElement>) => (
    <span data-testid="icon" {...rest}>{children}</span>
  ),
}));
vi.mock('@workday/canvas-kit-react/button', () => ({
  SecondaryButton: ({ children, onClick, ...rest }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button type="button" onClick={onClick} {...rest}>
      {children}
    </button>
  ),
  PrimaryButton: ({ children, onClick, ...rest }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button type="button" onClick={onClick} {...rest}>
      {children}
    </button>
  ),
}));
vi.mock('@workday/canvas-kit-react/tokens', () => ({
  space: { l: 16, m: 12, s: 8 },
  colors: { blueberry500: '#0875e1', frenchVanilla100: '#fff' },
}));

const TestWrapper = ({ children }: { children: React.ReactNode }) => (
  <CanvasProvider>
    <MemoryRouter initialEntries={['/organizations']}>
      {children}
    </MemoryRouter>
  </CanvasProvider>
);

describe('AppShell Layout', () => {
  it('renders header with brand title', () => {
    render(<AppShell />, { wrapper: TestWrapper });
    
    // 头部品牌文本（图标为SVG，不参与纯文本匹配）
    expect(screen.getByText('Cube Castle')).toBeInTheDocument();
  });

  it('renders sidebar navigation', () => {
    render(<AppShell />, { wrapper: TestWrapper });
    
    expect(screen.getByText(/仪表板/)).toBeInTheDocument();
    expect(screen.getByText(/组织架构/)).toBeInTheDocument();
    expect(screen.getByText(/职位管理/)).toBeInTheDocument();
  });
});
