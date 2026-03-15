import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { AppLayout } from '../AppLayout'

// Mock Sidebar and Header to avoid deep dependency chains
vi.mock('../Sidebar', () => ({
  Sidebar: ({ isCollapsed, isOpen, onToggle, alertCount }: {
    isCollapsed: boolean
    isOpen: boolean
    onToggle: () => void
    alertCount?: number
  }) => (
    <aside
      data-testid="sidebar"
      data-collapsed={String(isCollapsed)}
      data-open={String(isOpen)}
      data-alert-count={alertCount}
    >
      <button onClick={onToggle}>Toggle</button>
    </aside>
  ),
}))

vi.mock('../Header', () => ({
  Header: ({ onMenuToggle }: { onMenuToggle: () => void }) => (
    <header data-testid="app-header">
      <button onClick={onMenuToggle}>Menu</button>
    </header>
  ),
}))

describe('AppLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset window.innerWidth to desktop default
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1280,
    })
  })

  it('renders children', () => {
    render(
      <MemoryRouter>
        <AppLayout>
          <div data-testid="page-content">Page Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    expect(screen.getByTestId('page-content')).toBeInTheDocument()
  })

  it('renders sidebar', () => {
    render(
      <MemoryRouter>
        <AppLayout>
          <div>Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    expect(screen.getByTestId('sidebar')).toBeInTheDocument()
  })

  it('renders header', () => {
    render(
      <MemoryRouter>
        <AppLayout>
          <div>Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    expect(screen.getByTestId('app-header')).toBeInTheDocument()
  })

  it('passes alertCount to sidebar', () => {
    render(
      <MemoryRouter>
        <AppLayout alertCount={5}>
          <div>Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    expect(screen.getByTestId('sidebar')).toHaveAttribute('data-alert-count', '5')
  })

  it('toggles sidebar collapse on desktop', () => {
    render(
      <MemoryRouter>
        <AppLayout>
          <div>Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    const sidebar = screen.getByTestId('sidebar')
    expect(sidebar).toHaveAttribute('data-collapsed', 'false')

    // Click header menu button to toggle (desktop)
    const menuBtn = screen.getByTestId('app-header').querySelector('button')!
    fireEvent.click(menuBtn)

    expect(sidebar).toHaveAttribute('data-collapsed', 'true')
  })

  it('toggles mobile sidebar open on mobile', () => {
    Object.defineProperty(window, 'innerWidth', { value: 500, writable: true, configurable: true })

    render(
      <MemoryRouter>
        <AppLayout>
          <div>Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    const sidebar = screen.getByTestId('sidebar')
    expect(sidebar).toHaveAttribute('data-open', 'false')

    const menuBtn = screen.getByTestId('app-header').querySelector('button')!
    fireEvent.click(menuBtn)

    expect(sidebar).toHaveAttribute('data-open', 'true')
  })

  it('closes mobile sidebar on window resize to desktop', async () => {
    Object.defineProperty(window, 'innerWidth', { value: 500, writable: true, configurable: true })

    render(
      <MemoryRouter>
        <AppLayout>
          <div>Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    // Open mobile sidebar
    const menuBtn = screen.getByTestId('app-header').querySelector('button')!
    fireEvent.click(menuBtn)
    expect(screen.getByTestId('sidebar')).toHaveAttribute('data-open', 'true')

    // Resize to desktop
    Object.defineProperty(window, 'innerWidth', { value: 1280, writable: true, configurable: true })
    await act(async () => {
      fireEvent(window, new Event('resize'))
    })

    expect(screen.getByTestId('sidebar')).toHaveAttribute('data-open', 'false')
  })

  it('renders footer', () => {
    render(
      <MemoryRouter>
        <AppLayout>
          <div>Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    expect(screen.getByText(/NodePulse/)).toBeInTheDocument()
  })

  it('closes mobile sidebar when main content is clicked while open', () => {
    Object.defineProperty(window, 'innerWidth', { value: 500, writable: true, configurable: true })

    render(
      <MemoryRouter>
        <AppLayout>
          <div data-testid="content">Content</div>
        </AppLayout>
      </MemoryRouter>
    )
    // Open the sidebar
    const menuBtn = screen.getByTestId('app-header').querySelector('button')!
    fireEvent.click(menuBtn)
    expect(screen.getByTestId('sidebar')).toHaveAttribute('data-open', 'true')

    // Click on main content
    fireEvent.click(screen.getByTestId('content'))
    expect(screen.getByTestId('sidebar')).toHaveAttribute('data-open', 'false')
  })
})
