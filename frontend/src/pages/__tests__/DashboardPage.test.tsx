import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import '@testing-library/jest-dom'

// --- Mocks -------------------------------------------------------------

// Capture navigation without real routing.
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

// Mock the three data hooks the page composes. Each is controllable per-test
// via module-level state.
let mockDashboardData: Record<string, unknown> = {}
vi.mock('@/hooks/useDashboardData', () => ({
  useDashboardData: () => mockDashboardData,
}))

let mockDashboard: Record<string, unknown> = {}
vi.mock('@/hooks/useDashboard', () => ({
  useDashboard: () => mockDashboard,
}))

let mockDashboardHistory: Record<string, unknown> = {}
vi.mock('@/hooks/useDashboardHistory', () => ({
  useDashboardHistory: () => mockDashboardHistory,
}))

// Store selectors.
let mockAlertsStoreState: Record<string, unknown> = {}
vi.mock('@/stores/alertsStore', () => ({
  useAlertsStore: vi.fn((selector: (s: unknown) => unknown) => selector(mockAlertsStoreState)),
}))
let mockDashboardStoreState: Record<string, unknown> = {}
vi.mock('@/stores/dashboardStore', () => ({
  useDashboardStore: vi.fn((selector: (s: unknown) => unknown) => selector(mockDashboardStoreState)),
}))
// DashboardPage reads `role` from authStore for the empty-state CTA buttons
// (via `const { role } = useAuthStore()` — no selector). The mock returns the
// full state object when called without a selector, and applies it when called with one.
let mockAuthStoreState: Record<string, unknown> = { role: 'admin' }
vi.mock('@/stores/authStore', () => ({
  useAuthStore: vi.fn((selector?: (s: unknown) => unknown) =>
    typeof selector === 'function' ? selector(mockAuthStoreState) : mockAuthStoreState
  ),
}))

// Import after mocks.
import DashboardPage from '../DashboardPage'

function defaultStores() {
  mockAlertsStoreState = { fetchAlertRecords: vi.fn().mockResolvedValue(undefined) }
  mockDashboardStoreState = {
    refreshInterval: 10,
    setRefreshInterval: vi.fn(),
    autoRefresh: true,
    toggleAutoRefresh: vi.fn(),
  }
  mockAuthStoreState = { role: 'admin' }
}

function defaultHooks() {
  mockDashboardData = {
    nodes: [],
    metrics: [],
    isLoading: false,
    error: null,
    refetch: vi.fn().mockResolvedValue(undefined),
  }
  mockDashboard = {
    stats: {
      onlineRate: 100, anomalyRate: 0, averageLatency: 50, averagePacketLoss: 0,
      onlineNodes: 0, totalNodes: 0, warningNodes: 0, criticalNodes: 0,
    },
    sortedByAnomaly: [],
    nodeHealthSummaries: [],
  }
  mockDashboardHistory = { latencyTrend: [], packetLossTrend: [], isLoading: false }
}

function renderPage() {
  return render(
    <MemoryRouter>
      <DashboardPage />
    </MemoryRouter>,
  )
}

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    defaultStores()
    defaultHooks()
  })

  it('renders the page header, refresh controls, and stat cards', () => {
    renderPage()

    // Refresh-interval select and the refresh / view-all buttons.
    expect(screen.getByLabelText(/auto refresh/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /refresh data/i })).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: /view all nodes/i }).length).toBeGreaterThan(0)

    // The four stat-card headings (use exact match to avoid gauge sub-labels).
    expect(screen.getByText('Online Rate')).toBeInTheDocument()
    expect(screen.getByText('Anomaly Rate')).toBeInTheDocument()
    expect(screen.getByText('Avg Latency')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Probe Success Rate' })).toBeInTheDocument()
  })

  it('shows the empty-state copy when there are no nodes', () => {
    renderPage()
    // Cohort A added a "Getting started" panel + the existing topNodes empty
    // state — both contain "no nodes", so match either.
    expect(screen.getAllByText(/no nodes/i).length).toBeGreaterThan(0)
  })

  it('renders node summary cards when nodes are present', () => {
    mockDashboard.sortedByAnomaly = [
      {
        node: { id: 'n1', name: 'Node-Alpha', region: 'us-east' },
        metrics: { latency_ms: 30, packet_loss_rate: 0.01, timestamp: '' },
        healthStatus: 'healthy',
      },
    ]
    renderPage()
    expect(screen.getByText('Node-Alpha')).toBeInTheDocument()
  })

  it('shows an error banner with retry when useDashboardData errors', () => {
    mockDashboardData.error = new Error('fetch failed')
    renderPage()
    expect(screen.getByText('fetch failed')).toBeInTheDocument()
    expect(screen.getByText('Retry')).toBeInTheDocument()
  })

  it('re-fetches data and alerts when the refresh button is clicked', async () => {
    renderPage()
    fireEvent.click(screen.getByRole('button', { name: /refresh data/i }))
    await waitFor(() => {
      expect(mockDashboardData.refetch).toHaveBeenCalledTimes(1)
      expect(mockAlertsStoreState.fetchAlertRecords).toHaveBeenCalled()
    })
  })

  it('navigates to /nodes when "View All Nodes" is clicked', () => {
    renderPage()
    fireEvent.click(screen.getAllByRole('button', { name: /view all nodes/i })[0])
    expect(mockNavigate).toHaveBeenCalledWith('/nodes')
  })

  it('toggles auto-refresh off when the select is set to 0', () => {
    renderPage()
    fireEvent.change(screen.getByLabelText(/auto refresh/i), { target: { value: '0' } })
    expect(mockDashboardStoreState.toggleAutoRefresh).toHaveBeenCalled()
  })

  it('sets a new refresh interval and enables auto-refresh when a non-zero value is chosen', () => {
    mockDashboardStoreState.autoRefresh = false
    renderPage()
    fireEvent.change(screen.getByLabelText(/auto refresh/i), { target: { value: '30' } })
    expect(mockDashboardStoreState.toggleAutoRefresh).toHaveBeenCalled()
    expect(mockDashboardStoreState.setRefreshInterval).toHaveBeenCalledWith(30)
  })
})
