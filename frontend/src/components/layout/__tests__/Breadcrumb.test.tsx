import React from 'react'
import { describe, it, expect, vi } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { Breadcrumb } from '../Breadcrumb'

// ---------------------------------------------------------------------------
// Mock BreadcrumbContext — test Breadcrumb rendering in isolation
// ---------------------------------------------------------------------------
// The Breadcrumb component only consumes useBreadcrumb() from context.
// We mock it here so we don't need a React Router data router context.

const mockItems: Array<{ path: string; label: string }> = []
vi.mock('../useBreadcrumb', () => ({
  useBreadcrumb: () => ({
    items: mockItems,
    setDynamicLabel: vi.fn(),
    clearDynamicLabels: vi.fn(),
  }),
}))

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

afterEach(() => {
  mockItems.length = 0
  cleanup()
})

/** Simulate breadcrumb items for a single-level path (Home + page) */
function mockSingleLevel(pagePath: string, pageLabel: string) {
  mockItems.length = 0
  mockItems.push(
    { path: '/dashboard', label: 'Home' },
    { path: pagePath, label: pageLabel },
  )
}

/** Simulate breadcrumb items for a two-level path (Home + section + page) */
function mockTwoLevel(sectionPath: string, sectionLabel: string, pagePath: string, pageLabel: string) {
  mockItems.length = 0
  mockItems.push(
    { path: '/dashboard', label: 'Home' },
    { path: sectionPath, label: sectionLabel },
    { path: pagePath, label: pageLabel },
  )
}

/** Render Breadcrumb. Link from react-router-dom needs a router context,
    so we mock it as a plain <a> tag. */
vi.mock('react-router-dom', () => ({
  Link: (props: { to: string; children: React.ReactNode; className?: string }) =>
    React.createElement('a', { href: props.to, className: props.className }, props.children),
}))

function renderBreadcrumb() {
  return render(<Breadcrumb />)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('Breadcrumb', () => {
  it('renders navigation element with aria-label', () => {
    mockSingleLevel('/dashboard', 'Dashboard')
    renderBreadcrumb()
    expect(screen.getByRole('navigation', { name: 'Breadcrumb' })).toBeInTheDocument()
  })

  it('renders Home icon (SVG) for the first item', () => {
    mockSingleLevel('/dashboard', 'Dashboard')
    renderBreadcrumb()
    const nav = screen.getByRole('navigation', { name: 'Breadcrumb' })
    const svgs = nav.querySelectorAll('svg')
    expect(svgs.length).toBeGreaterThanOrEqual(1)
  })

  it('renders breadcrumb for /dashboard', () => {
    mockSingleLevel('/dashboard', 'Dashboard')
    renderBreadcrumb()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
  })

  it('renders breadcrumb for /nodes', () => {
    mockSingleLevel('/nodes', 'Nodes')
    renderBreadcrumb()
    expect(screen.getByText('Nodes')).toBeInTheDocument()
  })

  it('renders breadcrumb for /alerts/rules (two levels)', () => {
    mockTwoLevel('/alerts', 'Alerts', '/alerts/rules', 'Alert Rules')
    renderBreadcrumb()
    expect(screen.getByText('Alert Rules')).toBeInTheDocument()
  })

  it('renders chevron separators for multi-level paths', () => {
    mockTwoLevel('/alerts', 'Alerts', '/alerts/rules', 'Alert Rules')
    renderBreadcrumb()
    const nav = screen.getByRole('navigation', { name: 'Breadcrumb' })
    const svgs = nav.querySelectorAll('svg')
    // 1 HomeIcon + 1 ChevronRight = at least 2
    expect(svgs.length).toBeGreaterThan(1)
  })

  it('renders last item as text, not a link', () => {
    mockSingleLevel('/dashboard', 'Dashboard')
    renderBreadcrumb()
    const links = screen.queryAllByRole('link')
    const dashboardLink = links.find((l) => l.textContent?.includes('Dashboard'))
    expect(dashboardLink).toBeUndefined()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
  })

  it('renders Home as a link pointing to /dashboard', () => {
    mockSingleLevel('/dashboard', 'Dashboard')
    renderBreadcrumb()
    const links = screen.queryAllByRole('link')
    const homeLink = links.find((l) => l.getAttribute('href') === '/dashboard')
    expect(homeLink).toBeInTheDocument()
  })

  it('renders correct breadcrumb for /reports/history (export history, not alert history)', () => {
    mockTwoLevel('/reports', 'Reports', '/reports/history', 'Export History')
    renderBreadcrumb()
    expect(screen.getByText('Export History')).toBeInTheDocument()
    expect(screen.queryByText('Alert History')).not.toBeInTheDocument()
  })

  it('renders correct breadcrumb for /alerts/history (alert history)', () => {
    mockTwoLevel('/alerts', 'Alerts', '/alerts/history', 'Alert History')
    renderBreadcrumb()
    expect(screen.getByText('Alert History')).toBeInTheDocument()
  })

  it('renders intermediate items as links', () => {
    mockTwoLevel('/alerts', 'Alerts', '/alerts/rules', 'Alert Rules')
    renderBreadcrumb()
    const links = screen.queryAllByRole('link')
    const alertsLink = links.find((l) => l.textContent?.includes('Alerts'))
    expect(alertsLink).toBeInTheDocument()
    // Last item should NOT be a link
    const alertRulesLink = links.find((l) => l.textContent?.includes('Alert Rules'))
    expect(alertRulesLink).toBeUndefined()
  })
})
