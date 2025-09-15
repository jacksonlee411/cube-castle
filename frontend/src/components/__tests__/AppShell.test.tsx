// @vitest-environment jsdom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { CanvasProvider } from '@workday/canvas-kit-react/common';
import { AppShell } from '../../layout/AppShell';
import { vi } from 'vitest';

// Minimal mocks for Canvas Kit components used by AppShell tree
vi.mock('@workday/canvas-kit-react/layout', () => ({
  Box: (p: any) => <div {...p}>{p.children}</div>,
  Flex: (p: any) => <div {...p}>{p.children}</div>,
}));
vi.mock('@workday/canvas-kit-react/text', () => ({
  Heading: (p: any) => <h1 {...p}>{p.children}</h1>,
  Text: (p: any) => <span {...p}>{p.children}</span>,
}));
vi.mock('@workday/canvas-kit-react/icon', () => ({
  SystemIcon: (p: any) => <span data-testid="icon">{p.children}</span>,
}));
vi.mock('@workday/canvas-kit-react/button', () => ({
  SecondaryButton: (p: any) => <button onClick={p.onClick}>{p.children}</button>,
  PrimaryButton: (p: any) => <button onClick={p.onClick}>{p.children}</button>,
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
  });
});
