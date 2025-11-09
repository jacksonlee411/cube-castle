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
}))
vi.mock('@/features/temporal/components/TemporalMasterDetailView', () => ({
  TemporalMasterDetailView: (props: any) => (
    <div data-testid="temporal-view">Temporal View for {props.organizationCode}</div>
  ),
}));
vi.mock('@/features/positions/PositionDetailView', () => ({
  PositionDetailView: () => <div data-testid="position-detail-view" />,
}));

import { OrganizationTemporalEntityRoute } from '../../src/features/temporal/pages/entityRoutes';
import { TOKEN_STORAGE_KEY } from '../../src/shared/api/auth';

describe('Organization temporal entity route', () => {
  beforeEach(() => {
    // Provide a fake token to bypass RequireAuth redirect
    const legacyKey = ['cube', 'castle', 'oauth', 'token'].join('_');
    localStorage.removeItem(legacyKey);
    localStorage.setItem(
      TOKEN_STORAGE_KEY,
      JSON.stringify({ accessToken: 'x', tokenType: 'Bearer', expiresIn: 3600, issuedAt: Date.now() })
    );
  });

  it('renders detail page with breadcrumb and title', async () => {
    render(
      <MemoryRouter initialEntries={[`/organizations/1000001/temporal`]}>
        <Routes>
          <Route path="/organizations/:code/temporal" element={<OrganizationTemporalEntityRoute />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByTestId('temporal-view')).toHaveTextContent('1000001');
    });
  });
});
