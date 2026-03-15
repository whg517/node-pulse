import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Breadcrumb } from '../Breadcrumb'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

describe('Breadcrumb', () => {
  it('renders home icon on root path', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    expect(screen.getByRole('navigation')).toBeInTheDocument()
  })

  it('renders breadcrumb for /dashboard', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.dashboard')).toBeInTheDocument()
  })

  it('renders breadcrumb for /nodes', () => {
    render(
      <MemoryRouter initialEntries={['/nodes']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.nodes')).toBeInTheDocument()
  })

  it('renders breadcrumb for /nodes/:id with ID segment', () => {
    render(
      <MemoryRouter initialEntries={['/nodes/550e8400-e29b-41d4-a716-446655440000']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.nodes')).toBeInTheDocument()
    expect(screen.getByText('nav.details')).toBeInTheDocument()
  })

  it('renders breadcrumb for /nodes/123 with numeric ID segment', () => {
    render(
      <MemoryRouter initialEntries={['/nodes/123']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.details')).toBeInTheDocument()
  })

  it('renders breadcrumb for /alerts/rules', () => {
    render(
      <MemoryRouter initialEntries={['/alerts/rules']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.alerts')).toBeInTheDocument()
    expect(screen.getByText('nav.alertRules')).toBeInTheDocument()
  })

  it('renders breadcrumb for unknown segment', () => {
    render(
      <MemoryRouter initialEntries={['/unknownsection']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.unknownsection')).toBeInTheDocument()
  })

  it('renders separator chevrons for multi-level paths', () => {
    const { container } = render(
      <MemoryRouter initialEntries={['/alerts/rules']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    // Multiple SVGs for separators
    const svgs = container.querySelectorAll('svg')
    expect(svgs.length).toBeGreaterThan(1)
  })

  it('has navigation aria-label', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Breadcrumb />
      </MemoryRouter>
    )
    expect(screen.getByLabelText('Breadcrumb')).toBeInTheDocument()
  })
})
