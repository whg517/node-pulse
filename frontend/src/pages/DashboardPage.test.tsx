import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DashboardPage from './DashboardPage'
import { useAuthStore } from '../stores/authStore'
import { useDashboardData } from '../hooks/useDashboardData'
import { useDashboard } from '../hooks/useDashboard'
import { useNavigate } from 'react-router-dom'

// Mock matchMedia for useTheme hook
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock i18n
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: Record<string, unknown>) => {
      // Handle interpolation for auto-refresh text
      if (key === 'dashboard.autoRefresh' && options?.interval) {
        return `Auto-refreshing every ${options.interval} seconds`
      }
      // Simple translation mock - return the key as readable text
      const translations: Record<string, string> = {
        'dashboard.title': 'Dashboard',
        'dashboard.realTimeMetrics': 'Real-time Metrics',
        'dashboard.refreshData': 'Refresh Data',
        'dashboard.viewAllNodes': 'View All Nodes',
        'dashboard.autoRefresh': 'Auto-refreshing every 5 seconds',
        'dashboard.nodeHealthOverview': 'Node Health Overview',
        'dashboard.averageMetrics': 'Average Metrics',
        'dashboard.latencyTrendChart': 'Network Latency Trend',
        'dashboard.packetLossChart': 'Packet Loss Rate',
        'dashboard.probeSuccessRate': 'Probe Success Rate',
        'dashboard.nodesRequiringAttention': 'nodes requiring attention',
        'metrics.onlineRate': 'Online Rate',
        'metrics.anomalyRate': 'Anomaly Rate',
        'metrics.avgLatency': 'Avg Latency',
        'metrics.avgPacketLoss': 'Avg Packet Loss',
        'metrics.avgJitter': 'Avg Jitter',
        'metrics.totalNodes': 'Total Nodes',
        'metrics.latency': 'Latency',
        'metrics.packetLoss': 'Packet Loss',
        'metrics.jitter': 'Jitter',
        'units.ms': 'ms',
        'units.percent': '%',
        'common.welcome': 'Welcome',
        'common.logout': 'Logout',
        'common.loading': 'Loading...',
        'nodes.live': 'Live',
        'nodes.noNodes': 'No nodes found',
        'nodes.region': 'Region',
        'nodes.lastSeen': 'Last Seen',
        'status.online': 'Online',
        'status.healthy': 'Healthy',
        'status.warning': 'Warning',
        'status.critical': 'Critical',
        'status.offline': 'Offline',
        'status.unknown': 'Unknown',
      }
      return translations[key] || key
    },
    i18n: {
      language: 'en',
      changeLanguage: vi.fn(),
    },
  }),
}))

// Mock auth store, hooks, and router
vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('../hooks/useDashboardData', () => ({
  useDashboardData: vi.fn(),
}))

vi.mock('../hooks/useDashboard', () => ({
  useDashboard: vi.fn(),
}))

vi.mock('react-router-dom', async () => {
  const mod = await vi.importActual('react-router-dom')
  return {
    ...mod,
    useNavigate: vi.fn(),
  }
})

describe('DashboardPage', () => {
  let mockUseAuthStore: any
  let mockUseDashboardData: any
  let mockUseDashboard: any

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuthStore = vi.mocked(useAuthStore)
    mockUseDashboardData = vi.mocked(useDashboardData)
    mockUseDashboard = vi.mocked(useDashboard)
  })

  const setupDefaultMocks = () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    mockUseDashboard.mockReturnValue({
      stats: {
        totalNodes: 0,
        onlineNodes: 0,
        offlineNodes: 0,
        healthyNodes: 0,
        warningNodes: 0,
        criticalNodes: 0,
        unknownNodes: 0,
        onlineRate: 0,
        anomalyRate: 0,
        averageLatency: 0,
        averagePacketLoss: 0,
        averageJitter: 0,
      },
      nodeHealthSummaries: [],
      sortedByAnomaly: [],
    })
  }

  it('renders dashboard with username', () => {
    setupDefaultMocks()

    render(<DashboardPage />)

    expect(screen.getByText(/welcome, testuser/i)).toBeInTheDocument()
    expect(screen.getByText('NodePulse')).toBeInTheDocument()
  })

  it('renders dashboard content with components', () => {
    setupDefaultMocks()

    render(<DashboardPage />)

    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText(/Real-time Metrics/i)).toBeInTheDocument()
  })

  it('renders metrics summary cards', () => {
    setupDefaultMocks()

    render(<DashboardPage />)

    // Multiple elements may contain these texts (header stats + summary cards)
    expect(screen.getAllByText('Avg Latency').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Avg Packet Loss').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Avg Jitter').length).toBeGreaterThan(0)
  })

  it('displays error state when API fails', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    const mockError = new Error('Failed to fetch data')
    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: mockError,
      refetch: vi.fn(),
    })

    mockUseDashboard.mockReturnValue({
      stats: {
        totalNodes: 0,
        onlineNodes: 0,
        offlineNodes: 0,
        healthyNodes: 0,
        warningNodes: 0,
        criticalNodes: 0,
        unknownNodes: 0,
        onlineRate: 0,
        anomalyRate: 0,
        averageLatency: 0,
        averagePacketLoss: 0,
        averageJitter: 0,
      },
      nodeHealthSummaries: [],
      sortedByAnomaly: [],
    })

    render(<DashboardPage />)

    expect(screen.getByText('Failed to fetch data')).toBeInTheDocument()
  })

  it('handles logout button click', () => {
    const mockLogout = vi.fn()
    const mockClearAuth = vi.fn()
    const mockNavigate = vi.fn()

    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: mockLogout,
      clearAuth: mockClearAuth,
    })

    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    mockUseDashboard.mockReturnValue({
      stats: {
        totalNodes: 0,
        onlineNodes: 0,
        offlineNodes: 0,
        healthyNodes: 0,
        warningNodes: 0,
        criticalNodes: 0,
        unknownNodes: 0,
        onlineRate: 0,
        anomalyRate: 0,
        averageLatency: 0,
        averagePacketLoss: 0,
        averageJitter: 0,
      },
      nodeHealthSummaries: [],
      sortedByAnomaly: [],
    })

    vi.mocked(useNavigate).mockReturnValue(mockNavigate)

    render(<DashboardPage />)

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    logoutButton.click()

    expect(mockLogout).toHaveBeenCalled()
  })

  it('shows auto-refresh indicator when data is loaded', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    mockUseDashboardData.mockReturnValue({
      nodes: [{ id: '1', name: 'Node-1' }], // Non-empty array
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    mockUseDashboard.mockReturnValue({
      stats: {
        totalNodes: 1,
        onlineNodes: 1,
        offlineNodes: 0,
        healthyNodes: 1,
        warningNodes: 0,
        criticalNodes: 0,
        unknownNodes: 0,
        onlineRate: 100,
        anomalyRate: 0,
        averageLatency: 50,
        averagePacketLoss: 0,
        averageJitter: 10,
      },
      nodeHealthSummaries: [],
      sortedByAnomaly: [],
    })

    render(<DashboardPage />)

    expect(screen.getByText(/Auto-refreshing every 5 seconds/i)).toBeInTheDocument()
  })

  it('has correct accessibility attributes', () => {
    setupDefaultMocks()

    render(<DashboardPage />)

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    expect(logoutButton).toHaveAttribute('type', 'button')
  })
})
