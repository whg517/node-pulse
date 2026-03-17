import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { SidebarItem } from '../SidebarItem'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

describe('SidebarItem', () => {
  const defaultProps = {
    icon: <span data-testid="icon">🏠</span>,
    label: 'nav.dashboard',
    path: '/dashboard',
    isCollapsed: false,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders with label when not collapsed', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByText('nav.dashboard')).toBeInTheDocument()
  })

  it('hides label when collapsed', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} isCollapsed={true} />
      </MemoryRouter>
    )
    expect(screen.queryByText('nav.dashboard')).not.toBeInTheDocument()
  })

  it('renders icon', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} />
      </MemoryRouter>
    )
    expect(screen.getByTestId('icon')).toBeInTheDocument()
  })

  it('renders badge when count > 0 and not collapsed', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} badge={5} />
      </MemoryRouter>
    )
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('shows 99+ for large badge counts', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} badge={150} />
      </MemoryRouter>
    )
    expect(screen.getByText('99+')).toBeInTheDocument()
  })

  it('does not show badge when count is 0', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} badge={0} />
      </MemoryRouter>
    )
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  it('shows tooltip on hover when collapsed', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} isCollapsed={true} />
      </MemoryRouter>
    )
    const link = screen.getByRole('link')
    fireEvent.mouseEnter(link)
    expect(screen.getByText('nav.dashboard')).toBeInTheDocument()
  })

  it('hides tooltip on mouse leave when collapsed', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} isCollapsed={true} />
      </MemoryRouter>
    )
    const link = screen.getByRole('link')
    fireEvent.mouseEnter(link)
    fireEvent.mouseLeave(link)
    expect(screen.queryByText('nav.dashboard')).not.toBeInTheDocument()
  })

  it('does not show tooltip on hover when not collapsed', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} isCollapsed={false} />
      </MemoryRouter>
    )
    const link = screen.getByRole('link')
    fireEvent.mouseEnter(link)
    // Label is already visible, tooltip should not appear as extra element
    const labels = screen.getAllByText('nav.dashboard')
    // Only 1 label visible (in the link, not tooltip)
    expect(labels.length).toBe(1)
  })

  it('shows badge in tooltip when collapsed with badge', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} isCollapsed={true} badge={5} />
      </MemoryRouter>
    )
    const link = screen.getByRole('link')
    fireEvent.mouseEnter(link)
    expect(screen.getByText('(5)')).toBeInTheDocument()
  })

  it('shows 99+ in tooltip badge when count > 99', () => {
    render(
      <MemoryRouter>
        <SidebarItem {...defaultProps} isCollapsed={true} badge={200} />
      </MemoryRouter>
    )
    const link = screen.getByRole('link')
    fireEvent.mouseEnter(link)
    expect(screen.getByText('(99+)')).toBeInTheDocument()
  })
})
