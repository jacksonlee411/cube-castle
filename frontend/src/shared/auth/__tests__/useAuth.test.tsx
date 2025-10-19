// @vitest-environment jsdom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { AuthProvider } from '@/shared/auth/AuthProvider';
import { useAuth } from '@/shared/auth/hooks';

const AuthConsumer: React.FC = () => {
  const { userPermissions, hasPermission } = useAuth();
  return (
    <div>
      <span data-testid="permissions">{userPermissions.join(',')}</span>
      <span data-testid="position-read">{hasPermission('position:read') ? 'yes' : 'no'}</span>
      <span data-testid="position-write">{hasPermission('position:write') ? 'yes' : 'no'}</span>
      <span data-testid="job-catalog-read">{hasPermission('job-catalog:read') ? 'yes' : 'no'}</span>
    </div>
  );
};

describe('useAuth permissions', () => {
  beforeEach(() => {
    (globalThis as { __SCOPES__?: string[] }).__SCOPES__ = [
      'job-catalog:read',
      'position:read',
    ];
  });

  afterEach(() => {
    delete (globalThis as { __SCOPES__?: string[] }).__SCOPES__;
  });

  it('exposes sorted permissions and permission checks', () => {
    render(
      <MemoryRouter>
        <AuthProvider>
          <AuthConsumer />
        </AuthProvider>
      </MemoryRouter>
    );

    expect(screen.getByTestId('permissions').textContent).toBe('job-catalog:read,position:read');
    expect(screen.getByTestId('position-read').textContent).toBe('yes');
    expect(screen.getByTestId('job-catalog-read').textContent).toBe('yes');
    expect(screen.getByTestId('position-write').textContent).toBe('no');
  });
});
