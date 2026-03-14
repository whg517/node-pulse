import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi } from 'vitest'
import NodeDetailPage from './NodeDetailPage'
import { useNodeDetail } from '../hooks/useNodeDetail'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  initReactI18next: { type: '3rdParty', init: vi.fn() },
  useTranslation: () => ({
    t: (key: string, options?: any) => {
      const translations: Record<string, string> = {
        'common.loading': 'Loading node details...',
        'common.error': 'Error',
        'common.back': 'Back to Dashboard',
        'errors.failedToLoad': 'Error Loading Node',
        'errors.nodeNotFound': 'Node Not Found',
        'errors.notFound': 'The requested node does not exist.',
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
        'aria.backToDashboard': 'Back to dashboard',
        'time.justNow': 'Just now',
        'time.minutesAgo': `${options?.count || 1} minute${(options?.count || 1) > 1 ? 's' : ''} ago`,
        'time.hoursAgo': `${options?.count || 1} hour${(options?.count || 1) > 1 ? 's' : ''} ago`,
      }
      return translations[key] || key
    },
    i18n: { changeLanguage: vi.fn() },
  }),
}))

// Mock useTheme hook
vi.mock('../hooks/useTheme', () => ({
  useTheme: () => ({ isDark: false }),
}))

// Mock the hook
vi.mock('../hooks/useNodeDetail')

const mockUseNodeDetail = useNodeDetail as ReturnType<typeof vi.mocked<typeof useNodeDetail>>

describe('NodeDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders loading state', () => {
    mockUseNodeDetail.mockReturnValue({
      node: null,
      nodeStatus: null,
      metrics: null,
      isLoading: true,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    render(
      <MemoryRouter initialEntries={['/nodes/node-1']}>
        <Routes>
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Loading node details...')).toBeInTheDocument()
  })

  it('renders error state', () => {
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

    render(
      <MemoryRouter initialEntries={['/nodes/node-1']}>
        <Routes>
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Error Loading Node')).toBeInTheDocument()
    expect(screen.getByText('Failed to load node')).toBeInTheDocument()
    expect(screen.getByText('Back to Dashboard')).toBeInTheDocument()
  })

  it('renders node not found state', () => {
    mockUseNodeDetail.mockReturnValue({
      node: null,
      nodeStatus: null,
      metrics: null,
      isLoading: false,
      error: null,
      isPolling: false,
      refetch: vi.fn(),
    })

    render(
      <MemoryRouter initialEntries={['/nodes/node-1']}>
        <Routes>
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Node Not Found')).toBeInTheDocument()
    expect(screen.getByText('The requested node does not exist.')).toBeInTheDocument()
  })

  it('renders node details successfully', () => {
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

    render(
      <MemoryRouter initialEntries={['/nodes/node-1']}>
        <Routes>
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Test Node')).toBeInTheDocument()
    expect(screen.getAllByText('192.168.1.1')).toHaveLength(2)
    expect(screen.getByText('us-east')).toBeInTheDocument()
    expect(screen.getByText('production')).toBeInTheDocument()
    expect(screen.getByText('critical')).toBeInTheDocument()
  })

  it('renders metrics cards', () => {
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

    render(
      <MemoryRouter initialEntries={['/nodes/node-1']}>
        <Routes>
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Latency')).toBeInTheDocument()
    expect(screen.getByText('45')).toBeInTheDocument()
    const msLabels = screen.getAllByText('ms')
    expect(msLabels.length).toBeGreaterThan(0)

    expect(screen.getByText('Packet Loss Rate')).toBeInTheDocument()
    expect(screen.getByText('0.5')).toBeInTheDocument()
    expect(screen.getByText('%')).toBeInTheDocument()

    expect(screen.getByText('Jitter')).toBeInTheDocument()
  })

  it('renders node status badge', () => {
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

    expect(screen.getByText('online')).toBeInTheDocument()
    expect(screen.getByText('Live')).toBeInTheDocument()
  })

  it('shows N/A for metrics when not available', () => {
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

    const naValues = screen.getAllByText('N/A')
    expect(naValues.length).toBeGreaterThan(0)
  })

  it('shows no tags message when tags array is empty', () => {
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

    expect(screen.getByText('No tags')).toBeInTheDocument()
  })

  it('renders problem diagnosis section', () => {
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

    render(
      <MemoryRouter initialEntries={['/nodes/node-1']}>
        <Routes>
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Problem Diagnosis')).toBeInTheDocument()
    expect(screen.getByText(/Note: Current diagnosis uses client-side analysis/)).toBeInTheDocument()
  })

  it('formats last heartbeat timestamp correctly', () => {
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

    render(
      <MemoryRouter initialEntries={['/nodes/node-1']}>
        <Routes>
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('1 minute ago')).toBeInTheDocument()
  })

  it('navigates back to dashboard', () => {
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
          <Route path="/dashboard" element={<div>Dashboard Page</div>} />
        </Routes>
      </MemoryRouter>
    )

    const backLink = screen.getByLabelText('Back to Dashboard')
    expect(backLink).toBeInTheDocument()
    expect(backLink.closest('a')).toHaveAttribute('href', '/dashboard')
  })
})
