import React from 'react'
import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { RequireScopes } from '../RequireScopes'

describe('RequireScopes', () => {
  beforeEach(() => {
    ;(global as any).__SCOPES__ = []
  })

  it('renders fallback when missing required scopes', () => {
    render(
      <RequireScopes allOf={["org:read"]} fallback={<div data-testid="fallback">no-access</div>}>
        <div data-testid="content">content</div>
      </RequireScopes>
    )
    expect(screen.getByTestId('fallback')).toBeInTheDocument()
  })

  it('renders children when scopes are satisfied', () => {
    ;(global as any).__SCOPES__ = ['org:read', 'org:validate']
    render(
      <RequireScopes allOf={["org:read"]} anyOf={["org:validate", "org:read:hierarchy"]} fallback={<div>no</div>}>
        <div data-testid="content">content</div>
      </RequireScopes>
    )
    expect(screen.getByTestId('content')).toBeInTheDocument()
  })
})

