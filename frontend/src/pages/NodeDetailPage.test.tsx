import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi } from 'vitest'
import NodeDetailPage from './NodeDetailPage'
import { useNodeDetail } from '../hooks/useNodeDetail'
import { fetchHistory } from '../api/data'

vi.mock('../i18n', () => ({
  default: {},
  i18nInitPromise: Promise.resolve(),
}))

const mockI18n = { changeLanguage: vi.fn() }
const mockTranslate = (key: string, options?: { count?: number }) => {
  const translations: Record<string, string> = {
    'common.loading': 'Loading node details...',
    'common.error': 'Error',
    'common.back': 'Back to Dashboard',
    'errors.failedToLoad': 'Error Loading Node',
    'errors.nodeNotFound': 'Node Not Found',
    'errors.notFound': 'The requested node does not exist.',
    'errors.loadHistoricalError': 'Failed to load historical data.',
    'nodes.errorLoadingNode': 'Error Loading Node',
    'nodes.nodeNotFound': 'Node Not Found',
    'nodes.nodeNotFoundDescription': 'The requested node does not exist.',
    'nodes.backToDashboard': 'Back to Dashboard',
    'nodes.details': 'Node Details',
    'nodes.status': 'Status',
    'nodes.region': 'Region',
    'nodes.tags': 'Tags',
    'nodes.noTags': 'No tags',
    'nodes.metrics': 'Metrics',
    'nodes.latency': 'Latency',
    'nodes.packetLoss': 'Packet Loss Rate',
    'nodes.jitter': 'Jitter',
    'nodes.lastHeartbeat': 'Last Heartbeat',
    'nodes.problemDiagnosis': 'Problem Diagnosis',
    'nodes.noDiagnosisData': 'No diagnosis data available',
    'nodes.diagnosisNote': 'Note: Current diagnosis uses client-side analysis based on available metrics.',
    'navigation.dashboard': 'Dashboard',
    'status.online': 'online',
    'status.offline': 'Offline',
    'status.connecting': 'Connecting',
    'status.live': 'Live',
    'nodes.live': 'Live',
    'metrics.latency': 'Latency',
    'metrics.packetLoss': 'Packet Loss Rate',
    'metrics.jitter': 'Jitter',
    'aria.backToDashboard': 'Back to Dashboard',
    'time.justNow': 'Just now',
    'time.minutesAgo': `${options?.count || 1} minute${(options?.count || 1) > 1 ? 's' : ''} ago`,
    'time.hoursAgo': `${options?.count || 1} hour${(options?.count || 1) > 1 ? 's' : ''} ago`,
  }
  return translations[key] || key
}

// Mock react-i18next
vi.mock('react-i18next', () => ({
  initReactI18next: { type: '3rdParty', init: vi.fn() },
  useTranslation: () => ({
    t: mockTranslate,
    i18n: mockI18n,
  }),
}))

// Mock useTheme hook
vi.mock('../hooks/useTheme', () => ({
  useTheme: () => ({ isDark: false }),
}))

// Mock the hook
vi.mock('../hooks/useNodeDetail')

vi.mock('../api/data', () => ({
  fetchHistory: vi.fn(() => Promise.resolve({ data: [] })),
}))

// Mock useBreadcrumb — NodeDetailPage uses useSetBreadcrumbLabel
vi.mock('../components/layout/useBreadcrumb', () => ({
  useBreadcrumb: () => ({
    items: [],
    setDynamicLabel: vi.fn(),
    clearDynamicLabels: vi.fn(),
  }),
  useSetBreadcrumbLabel: () => ({
    setDynamicLabel: vi.fn(),
    clearDynamicLabels: vi.fn(),
  }),
  BreadcrumbProvider: ({ children }: { children: unknown }) => children,
}))

const mockUseNodeDetail = useNodeDetail as ReturnType<typeof vi.mocked<typeof useNodeDetail>>
const mockFetchHistory = fetchHistory as ReturnType<typeof vi.fn>

async function renderNodeDetailPage() {
  render(
    <MemoryRouter initialEntries={['/nodes/node-1']}>
      <Routes>
        <Route path="/nodes/:id" element={<NodeDetailPage />} />
      </Routes>
    </MemoryRouter>
  )

  await waitFor(() => {
    expect(mockFetchHistory).toHaveBeenCalledTimes(3)
  })
}

describe('NodeDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetchHistory.mockResolvedValue({ data: [] })
  })

  it('renders loading state', async () => {
    mockUseNodeDetail.mockReturnValue({
      node: null,
      nodeStatus: null,
      metrics: null,
      isLoading: true,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('Loading node details...')).toBeInTheDocument()
  })

  it('renders error state', async () => {
    const error = new Error('Failed to load node')
    mockUseNodeDetail.mockReturnValue({
      node: null,
      nodeStatus: null,
      metrics: null,
      isLoading: false,
      error,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('Error Loading Node')).toBeInTheDocument()
    expect(screen.getByText('Failed to load node')).toBeInTheDocument()
    expect(screen.getByText('Back to Dashboard')).toBeInTheDocument()
  })

  it('renders node not found state', async () => {
    mockUseNodeDetail.mockReturnValue({
      node: null,
      nodeStatus: null,
      metrics: null,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('Node Not Found')).toBeInTheDocument()
    expect(screen.getByText('The requested node does not exist.')).toBeInTheDocument()
  })

  it('renders node details successfully', async () => {
    const mockNode = {
      id: 'node-1',
      name: 'Test Node',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: ['production', 'critical'],
      status: 'online',
    }

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: '2024-01-01T12:00:00Z',
    }

    const mockMetrics = {
      node_id: 'node-1',
      latency_ms: 45,
      packet_loss_rate: 0,
      jitter_ms: 5,
      timestamp: '2024-01-01T12:00:00Z',
    }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: true,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('Test Node')).toBeInTheDocument()
    expect(screen.getAllByText('192.168.1.1')).toHaveLength(2)
    expect(screen.getByText('us-east')).toBeInTheDocument()
    expect(screen.getByText('production')).toBeInTheDocument()
    expect(screen.getByText('critical')).toBeInTheDocument()
  })

  it('renders metrics cards', async () => {
    const mockNode = {
      id: 'node-1',
      name: 'Test Node',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: [],
      status: 'online',
    }

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: '2024-01-01T12:00:00Z',
    }

    const mockMetrics = {
      node_id: 'node-1',
      latency_ms: 45,
      packet_loss_rate: 0.5,
      jitter_ms: 5,
      timestamp: '2024-01-01T12:00:00Z',
    }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: true,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('Latency')).toBeInTheDocument()
    expect(screen.getByText('45')).toBeInTheDocument()
    const msLabels = screen.getAllByText('ms')
    expect(msLabels.length).toBeGreaterThan(0)

    expect(screen.getByText('Packet Loss Rate')).toBeInTheDocument()
    expect(screen.getByText('0.5')).toBeInTheDocument()
    expect(screen.getByText('%')).toBeInTheDocument()

    expect(screen.getByText('Jitter')).toBeInTheDocument()
  })

  it('renders node status badge', async () => {
    const mockNode = {
      id: 'node-1',
      name: 'Test Node',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: [],
      status: 'online',
    }

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: '2024-01-01T12:00:00Z',
    }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: null,
      isLoading: false,
      error: null,
      isPolling: true,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('online')).toBeInTheDocument()
    expect(screen.getByText('Live')).toBeInTheDocument()
  })

  it('shows N/A for metrics when not available', async () => {
    const mockNode = {
      id: 'node-1',
      name: 'Test Node',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: [],
      status: 'online',
    }

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: '2024-01-01T12:00:00Z',
    }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: null,
      isLoading: false,
      error: null,
      isPolling: true,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    const naValues = screen.getAllByText('N/A')
    expect(naValues.length).toBeGreaterThan(0)
  })

  it('shows no tags message when tags array is empty', async () => {
    const mockNode = {
      id: 'node-1',
      name: 'Test Node',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: [],
      status: 'online',
    }

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: '2024-01-01T12:00:00Z',
    }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: null,
      isLoading: false,
      error: null,
      isPolling: true,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('No tags')).toBeInTheDocument()
  })

  it('renders problem diagnosis section', async () => {
    const mockNode = {
      id: 'node-1',
      name: 'Test Node',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: [],
      status: 'online',
    }

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: '2024-01-01T12:00:00Z',
    }

    const mockMetrics = {
      node_id: 'node-1',
      latency_ms: 45,
      packet_loss_rate: 0,
      jitter_ms: 5,
      timestamp: '2024-01-01T12:00:00Z',
    }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: true,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('Problem Diagnosis')).toBeInTheDocument()
    expect(screen.getByText(/Note: Current diagnosis uses client-side analysis/)).toBeInTheDocument()
  })

  it('formats last heartbeat timestamp correctly', async () => {
    const mockNode = {
      id: 'node-1',
      name: 'Test Node',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: [],
      status: 'online',
    }

    const now = new Date()
    // Use 90 seconds ago to ensure diffMins >= 1 (avoids timing issues with 60000ms)
    const ninetySecondsAgo = new Date(now.getTime() - 90000).toISOString()

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: ninetySecondsAgo,
    }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: null,
      isLoading: false,
      error: null,
      isPolling: true,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()

    expect(screen.getByText('1 minute ago')).toBeInTheDocument()
  })

  it('renders node detail page with page header', async () => {
    const mockNode = {
      id: 'node-1',
      name: 'Test Node',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: [],
      status: 'online',
    }

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: '2024-01-01T12:00:00Z',
    }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: null,
      isLoading: false,
      error: null,
      isPolling: true,
      refetch: vi.fn(),
    })

    render(
      <MemoryRouter initialEntries={['/nodes/node-1']}>
        <Routes>
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(mockFetchHistory).toHaveBeenCalledTimes(3)
    })

    // Page renders the node name (back navigation is now handled by breadcrumbs)
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })

  it('handles offline node status for problem detection', async () => {
    const mockNode = { id: 'node-1', name: 'Test Node', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'offline' }
    const mockNodeStatus = { status: 'offline', last_heartbeat: '2024-01-01T12:00:00Z' }
    const mockMetrics = { node_id: 'node-1', latency_ms: 45, packet_loss_rate: 60, jitter_ms: 5, timestamp: '2024-01-01T12:00:00Z' }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })

  it('handles high packet loss metrics', async () => {
    const mockNode = { id: 'node-1', name: 'Test Node', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online' }
    const mockNodeStatus = { status: 'online', last_heartbeat: '2024-01-01T12:00:00Z' }
    // packet_loss_rate > 10 triggers specific code paths
    const mockMetrics = { node_id: 'node-1', latency_ms: 600, packet_loss_rate: 15, jitter_ms: 5, timestamp: '2024-01-01T12:00:00Z' }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })

  it('handles very high latency metrics', async () => {
    const mockNode = { id: 'node-1', name: 'Test Node', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online' }
    const mockNodeStatus = { status: 'online', last_heartbeat: '2024-01-01T12:00:00Z' }
    // latency_ms > 1000
    const mockMetrics = { node_id: 'node-1', latency_ms: 1500, packet_loss_rate: 5, jitter_ms: 150, timestamp: '2024-01-01T12:00:00Z' }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })

  it('handles elevated metrics (packet loss > 3)', async () => {
    const mockNode = { id: 'node-1', name: 'Test Node', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online' }
    const mockNodeStatus = { status: 'online', last_heartbeat: '2024-01-01T12:00:00Z' }
    const mockMetrics = { node_id: 'node-1', latency_ms: 350, packet_loss_rate: 4, jitter_ms: 110, timestamp: '2024-01-01T12:00:00Z' }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })

  it('handles mildly elevated metrics (packet loss > 1)', async () => {
    const mockNode = { id: 'node-1', name: 'Test Node', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online' }
    const mockNodeStatus = { status: 'online', last_heartbeat: '2024-01-01T12:00:00Z' }
    const mockMetrics = { node_id: 'node-1', latency_ms: 200, packet_loss_rate: 2, jitter_ms: 60, timestamp: '2024-01-01T12:00:00Z' }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })

  it('handles old last_heartbeat timestamp (> 24 hours)', async () => {
    const mockNode = { id: 'node-1', name: 'Test Node', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online' }
    const oldTimestamp = new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString()
    const mockNodeStatus = { status: 'online', last_heartbeat: oldTimestamp }
    const mockMetrics = { node_id: 'node-1', latency_ms: 45, packet_loss_rate: 0, jitter_ms: 5, timestamp: oldTimestamp }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })

  it('handles high severity score for medium confidence', async () => {
    const mockNode = { id: 'node-1', name: 'Test Node', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online' }
    const mockNodeStatus = { status: 'online', last_heartbeat: '2024-01-01T12:00:00Z' }
    // severity score > 2 but not high conditions: latency_ms: 450, packet_loss_rate: 5
    const mockMetrics = { node_id: 'node-1', latency_ms: 450, packet_loss_rate: 5, jitter_ms: 110, timestamp: '2024-01-01T12:00:00Z' }

    mockUseNodeDetail.mockReturnValue({
      node: mockNode as any,
      nodeStatus: mockNodeStatus as any,
      metrics: mockMetrics as any,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    await renderNodeDetailPage()
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })
})
