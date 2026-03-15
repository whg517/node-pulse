import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Sidebar } from '../Sidebar'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

describe('Sidebar', () => {
  const defaultProps = {
    isCollapsed: false,
    isOpen: false,
    onToggle: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders navigation items', () => {
    render(
      <MemoryRouter>
        <Sidebar {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.dashboard')).toBeInTheDocument()
    expect(screen.getByText('nav.nodes')).toBeInTheDocument()
    expect(screen.getByText('nav.alerts')).toBeInTheDocument()
  })

  it('shows mobile backdrop when open', () => {
    render(
      <MemoryRouter>
        <Sidebar {...defaultProps} isOpen={true} />
      </MemoryRouter>
    )
    const backdrop = document.querySelector('[aria-hidden="true"]')
    expect(backdrop).toBeInTheDocument()
  })

  it('does not show mobile backdrop when closed', () => {
    render(
      <MemoryRouter>
        <Sidebar {...defaultProps} isOpen={false} />
      </MemoryRouter>
    )
    const backdrop = document.querySelector('[aria-hidden="true"]')
    expect(backdrop).not.toBeInTheDocument()
  })

  it('calls onToggle when backdrop is clicked', () => {
    render(
      <MemoryRouter>
        <Sidebar {...defaultProps} isOpen={true} />
      </MemoryRouter>
    )
    const backdrop = document.querySelector('[aria-hidden="true"]')
    fireEvent.click(backdrop!)
    expect(defaultProps.onToggle).toHaveBeenCalledTimes(1)
  })

  it('renders NodePulse branding', () => {
    render(
      <MemoryRouter>
        <Sidebar {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByText('NodePulse')).toBeInTheDocument()
  })

  it('renders with collapsed state', () => {
    const { container } = render(
      <MemoryRouter>
        <Sidebar {...defaultProps} isCollapsed={true} />
      </MemoryRouter>
    )
    const aside = container.querySelector('aside')
    expect(aside?.className).toContain('w-16')
  })

  it('renders all navigation links', () => {
    render(
      <MemoryRouter>
        <Sidebar {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.reports')).toBeInTheDocument()
    expect(screen.getByText('nav.settings')).toBeInTheDocument()
  })

  it('passes alertCount badge to alerts nav item', () => {
    render(
      <MemoryRouter>
        <Sidebar {...defaultProps} alertCount={3} />
      </MemoryRouter>
    )
    expect(screen.getByText('3')).toBeInTheDocument()
  })
})
