// @vitest-environment jsdom
import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Mocks for Canvas Kit components used in OrganizationTree
vi.mock('@workday/canvas-kit-react/layout', () => ({
  Box: (p: any) => <div {...p}>{p.children}</div>,
  Flex: (p: any) => <div {...p}>{p.children}</div>,
}));
vi.mock('@workday/canvas-kit-react/text', () => ({
  Text: (p: any) => <span {...p}>{p.children}</span>,
  Heading: (p: any) => <h2 {...p}>{p.children}</h2>,
}));
vi.mock('@workday/canvas-kit-react/card', () => ({
  Card: (p: any) => <div onClick={p.onClick} {...p}>{p.children}</div>,
}));
vi.mock('@workday/canvas-kit-react/loading-dots', () => ({
  LoadingDots: () => <div>loading...</div>,
}));
vi.mock('@workday/canvas-kit-react/button', () => ({
  SecondaryButton: (p: any) => <button onClick={p.onClick}>{p.children}</button>,
}));
vi.mock('@workday/canvas-kit-react/tokens', () => ({
  colors: { blueberry100: '#ace', cinnamon100: '#fcc', cinnamon600: '#933' },
  borderRadius: { m: 4, s: 2 },
  space: { xs: 4, s: 8, m: 12, l: 16 },
}));
vi.mock('../../src/shared/components/StatusBadge', () => ({
  StatusBadge: (p: any) => <span data-testid="status-badge">{p.status}</span>,
}));

// Mock react-router navigate
const navigateMock = vi.fn();
vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...actual,
    useNavigate: () => navigateMock,
  };
});

// Test subject import after mocks
import { OrganizationTree } from '../../src/features/organizations/components/OrganizationTree';

describe('OrganizationTree', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    navigateMock.mockReset();
  });

  it('navigates to detail when clicking node (default behavior)', async () => {
    // Mock fetch for GraphQL client
    const fetchMock = vi.spyOn(global, 'fetch' as any).mockImplementation(async (input: RequestInfo, init?: RequestInit) => {
      const url = typeof input === 'string' ? input : (input as any)?.url || '';
      if (url.includes('/auth/dev-token')) {
        return { ok: true, json: async () => ({ accessToken: 'test-token', expiresIn: 3600 }) } as any;
      }
      const body = init?.body ? JSON.parse(init.body as string) : {};
      const query: string = body.query || '';
      if (query.includes('GetRootOrganizations')) {
        return {
          ok: true,
          json: async () => ({ data: { organizations: { data: [
            { code: '100', name: '集团', unitType: 'DEPARTMENT', status: 'ACTIVE', level: 1, parentCode: null, codePath: '/100', namePath: '/集团' }
          ] } } }),
        } as any;
      }
      if (query.includes('GetRootChildrenCount')) {
        return { ok: true, json: async () => ({ data: { organizationSubtree: { childrenCount: 0 } } }) } as any;
      }
      return { ok: true, json: async () => ({ data: {} }) } as any;
    });

    render(
      <MemoryRouter>
        <OrganizationTree />
      </MemoryRouter>
    );

    const node = await screen.findByText('集团');
    fireEvent.click(node);

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/organizations/100/temporal');
    });
    fetchMock.mockRestore();
  });

  it('copy buttons copy deep link and paths', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });

    // Same fetch mock as above
    vi.spyOn(global, 'fetch' as any).mockImplementation(async (input: RequestInfo, init?: RequestInit) => {
      const url = typeof input === 'string' ? input : (input as any)?.url || '';
      if (url.includes('/auth/dev-token')) {
        return { ok: true, json: async () => ({ accessToken: 'test-token', expiresIn: 3600 }) } as any;
      }
      const body = init?.body ? JSON.parse(init.body as string) : {};
      const query: string = body.query || '';
      if (query.includes('GetRootOrganizations')) {
        return { ok: true, json: async () => ({ data: { organizations: { data: [
          { code: '100', name: '集团', unitType: 'DEPARTMENT', status: 'ACTIVE', level: 1, parentCode: null, codePath: '/100', namePath: '/集团' }
        ] } } }) } as any;
      }
      if (query.includes('GetRootChildrenCount')) {
        return { ok: true, json: async () => ({ data: { organizationSubtree: { childrenCount: 0 } } }) } as any;
      }
      return { ok: true, json: async () => ({ data: {} }) } as any;
    });

    render(
      <MemoryRouter>
        <OrganizationTree />
      </MemoryRouter>
    );

    // Select node to show selected area
    const node = await screen.findByText('集团');
    fireEvent.click(node);

    const copyLinkBtn = await screen.findByText('复制链接');
    fireEvent.click(copyLinkBtn);
    expect(writeText).toHaveBeenCalled();
  });
});
