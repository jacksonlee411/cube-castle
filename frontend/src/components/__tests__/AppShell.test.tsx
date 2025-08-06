import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { CanvasProvider } from '@workday/canvas-kit-react/common';
import { AppShell } from '../../layout/AppShell';

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
    
    expect(screen.getByText('🏰 Cube Castle')).toBeInTheDocument();
    expect(screen.getByText('设置')).toBeInTheDocument();
    expect(screen.getByText('通知')).toBeInTheDocument();
    expect(screen.getByText('用户')).toBeInTheDocument();
  });

  it('renders sidebar navigation without logo', () => {
    render(<AppShell />, { wrapper: TestWrapper });
    
    expect(screen.getByText(/仪表板/)).toBeInTheDocument();
    expect(screen.getByText(/组织架构/)).toBeInTheDocument();
    expect(screen.getByText(/员工管理/)).toBeInTheDocument();
    expect(screen.getByText(/职位管理/)).toBeInTheDocument();
  });
});