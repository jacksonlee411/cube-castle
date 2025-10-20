// @vitest-environment jsdom
import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter, useLocation } from 'react-router-dom'
import { AuthProvider } from '@/shared/auth/AuthProvider'
import { NavigationItem, type NavigationItemConfig } from '../NavigationItem'

const LocationDisplay: React.FC = () => {
  const location = useLocation()
  return <div data-testid="location">{location.pathname}</div>
}

const renderWithProviders = (item: NavigationItemConfig, initialPath = '/positions') => {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <AuthProvider>
        <NavigationItem {...item} />
        <LocationDisplay />
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('NavigationItem', () => {
  beforeEach(() => {
    ;(globalThis as { __SCOPES__?: string[] }).__SCOPES__ = [
      'position:read',
      'job-catalog:read',
    ]
  })

  afterEach(() => {
    delete (globalThis as { __SCOPES__?: string[] }).__SCOPES__
  })

  it('renders collapsible sub navigation and navigates on click', async () => {
    renderWithProviders({
      label: '职位管理',
      path: '/positions',
      icon: {} as any,
      subItems: [
        { label: '职位列表', path: '/positions', permission: 'position:read' },
        { label: '职类管理', path: '/positions/catalog/family-groups', permission: 'job-catalog:read' },
      ],
    })

    expect(screen.getByText('职位列表')).toBeInTheDocument()
    expect(screen.getByText('职类管理')).toBeInTheDocument()
    expect(screen.getByTestId('location').textContent).toBe('/positions')

    fireEvent.click(screen.getByText('职类管理'))
    expect(screen.getByTestId('location').textContent).toBe('/positions/catalog/family-groups')
  })

  it('hides item when lacking top-level permission', () => {
    ;(globalThis as { __SCOPES__?: string[] }).__SCOPES__ = ['org:read']

    const { queryByText } = renderWithProviders({
      label: '职位管理',
      path: '/positions',
      icon: {} as any,
      permission: 'position:read',
    })

    expect(queryByText('职位管理')).not.toBeInTheDocument()
  })

  it('filters sub navigation items based on permission set', () => {
    ;(globalThis as { __SCOPES__?: string[] }).__SCOPES__ = ['position:read']

    renderWithProviders({
      label: '职位管理',
      path: '/positions',
      icon: {} as any,
      permission: 'position:read',
      subItems: [
        { label: '职位列表', path: '/positions', permission: 'position:read' },
        { label: '职类管理', path: '/positions/catalog/family-groups', permission: 'job-catalog:read' },
      ],
    })

    expect(screen.getByText('职位列表')).toBeInTheDocument()
    expect(screen.queryByText('职类管理')).not.toBeInTheDocument()
  })

  it('exposes aria-expanded via Expandable.Target wrapper', () => {
    renderWithProviders(
      {
        label: '职位管理',
        path: '/positions',
        icon: {} as any,
        subItems: [
          { label: '职位列表', path: '/positions', permission: 'position:read' },
          { label: '职类管理', path: '/positions/catalog/family-groups', permission: 'job-catalog:read' },
        ],
      },
      '/positions/catalog/family-groups',
    )

    const trigger = screen.getByRole('button', { name: '职位管理' })
    expect(trigger).toHaveAttribute('aria-expanded', 'true')
  })
})
