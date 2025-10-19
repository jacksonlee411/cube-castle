// @vitest-environment jsdom
import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter, useLocation } from 'react-router-dom'
import { AuthProvider } from '@/shared/auth/AuthProvider'
import { Sidebar } from '../Sidebar'

const LocationDisplay: React.FC = () => {
  const location = useLocation()
  return <div data-testid="location">{location.pathname}</div>
}

const renderSidebar = (initialPath = '/organizations') => {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <AuthProvider>
        <Sidebar />
        <LocationDisplay />
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('Sidebar', () => {
  beforeEach(() => {
    ;(globalThis as { __SCOPES__?: string[] }).__SCOPES__ = [
      'org:read',
      'position:read',
      'job-catalog:read',
    ]
  })

  afterEach(() => {
    delete (globalThis as { __SCOPES__?: string[] }).__SCOPES__
  })

  it('renders primary and secondary navigation entries', () => {
    renderSidebar()

    expect(screen.getByText('仪表板')).toBeInTheDocument()
    expect(screen.getByText('组织架构')).toBeInTheDocument()
    expect(screen.getByText('职位管理')).toBeInTheDocument()
    expect(screen.getByText('职位列表')).toBeInTheDocument()
    expect(screen.getByText('职类管理')).toBeInTheDocument()
  })

  it('navigates to target route on click', async () => {
    renderSidebar()

    fireEvent.click(screen.getByText('仪表板'))
    expect(screen.getByTestId('location').textContent).toBe('/dashboard')
  })

  it('hides job catalog navigation when permission absent', () => {
    ;(globalThis as { __SCOPES__?: string[] }).__SCOPES__ = ['org:read']
    renderSidebar()

    expect(screen.queryByText('职位管理')).not.toBeInTheDocument()
  })
})
