// @vitest-environment jsdom
import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { render, screen, waitFor } from '@testing-library/react';

// Minimal Canvas Kit mocks
vi.mock('@workday/canvas-kit-react/layout', () => ({
  Box: (p: any) => <div {...p}>{p.children}</div>,
  Flex: (p: any) => <div {...p}>{p.children}</div>,
}));
vi.mock('@workday/canvas-kit-react/text', () => ({
  Text: (p: any) => <span {...p}>{p.children}</span>,
  Heading: (p: any) => <h2 {...p}>{p.children}</h2>,
}));
vi.mock('@workday/canvas-kit-react/card', () => ({
  Card: (p: any) => <div {...p}>{p.children}</div>,
}));
vi.mock('@workday/canvas-kit-react/button', () => ({
  PrimaryButton: (p: any) => <button onClick={p.onClick}>{p.children}</button>,
  SecondaryButton: (p: any) => <button onClick={p.onClick}>{p.children}</button>,
}));
vi.mock('@workday/canvas-kit-react/icon', () => ({
  SystemIcon: (p: any) => <span data-testid="icon">{p.children}</span>,
}));
vi.mock('@workday/canvas-kit-react/tokens', () => ({
  colors: { cinnamon600: '#933', blueberry100: '#ace' },
  borderRadius: { m: 4, s: 2 },
}));

// Mock heavy temporal master-detail view to avoid deep Canvas deps
vi.mock('../../src/features/temporal/components/TemporalMasterDetailView', () => ({
  TemporalMasterDetailView: (props: any) => (
    <div data-testid="temporal-view">Temporal View for {props.organizationCode}</div>
  ),
}));

import { OrganizationTemporalPage } from '../../src/features/organizations/OrganizationTemporalPage';

describe('OrganizationTemporalPage', () => {
  beforeEach(() => {
    // Provide a fake token to bypass RequireAuth redirect
    localStorage.setItem(
      'cube_castle_oauth_token',
      JSON.stringify({ accessToken: 'x', tokenType: 'Bearer', expiresIn: 3600, issuedAt: Date.now() })
    );
  });

  it('renders detail page with breadcrumb and title', async () => {
    render(
      <MemoryRouter initialEntries={[`/organizations/1000001/temporal`]}>
        <Routes>
          <Route path="/organizations/:code/temporal" element={<OrganizationTemporalPage />} />
        </Routes>
      </MemoryRouter>
    );

    // 顶部导航与子视图渲染
    await waitFor(() => {
      expect(screen.getByText('← 组织列表')).toBeInTheDocument();
      expect(screen.getByText(/组织详情/)).toBeInTheDocument();
      expect(screen.getByTestId('temporal-view')).toBeInTheDocument();
      expect(screen.getByText(/1000001/)).toBeInTheDocument();
    });
  });
});
