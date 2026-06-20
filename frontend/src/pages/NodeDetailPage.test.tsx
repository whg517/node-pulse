import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi } from 'vitest'
import NodeDetailPage from './NodeDetailPage'
import { useNodeDetail } from '../hooks/useNodeDetail'
import { fetchHistory, fetchLatestMTR } from '../api/data'

vi.mock('../i18n', () => ({
  default: {},
  i18nInitPromise: Promise.resolve(),
}))

// Mock useTheme hook
vi.mock('../hooks/useTheme', () => ({
  useTheme: () => ({ isDark: false }),
}))

// Mock the hook
vi.mock('../hooks/useNodeDetail')

vi.mock('../api/data', () => ({
  fetchHistory: vi.fn(() => Promise.resolve({ data: [] })),
  fetchLatestMTR: vi.fn(() => Promise.resolve(null)),
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
const mockFetchLatestMTR = fetchLatestMTR as ReturnType<typeof vi.fn>

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

  await waitFor(() => {
    expect(mockFetchLatestMTR).toHaveBeenCalledWith('node-1')
  })
}

describe('NodeDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetchHistory.mockResolvedValue({ data: [] })
    mockFetchLatestMTR.mockResolvedValue(null)
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

    expect(screen.getByText('Loading...')).toBeInTheDocument()
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

    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
    expect(screen.getByText('Failed to load node')).toBeInTheDocument()
    expect(screen.getByText('Back')).toBeInTheDocument()
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

    expect(screen.getByText('The requested node does not exist.')).toBeInTheDocument()
    expect(screen.getByText('Not found')).toBeInTheDocument()
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

    expect(screen.getByText('Packet Loss')).toBeInTheDocument()
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

    expect(screen.getByText('Online')).toBeInTheDocument()
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

  it('renders latest MTR path data', async () => {
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

    mockFetchLatestMTR.mockResolvedValue({
      target: 'example.com',
      totalHops: 2,
      completedAt: '2024-01-01T12:00:00Z',
      success: true,
      hops: [
        {
          hopNumber: 1,
          ip: '192.0.2.1',
          hostname: 'gateway.local',
          asNumber: 'AS64500',
          sent: 10,
          received: 10,
          lossRate: 0,
          lastRTTMs: 1.2,
          avgRTTMs: 1.4,
          bestRTTMs: 1.1,
          worstRTTMs: 1.8,
          stdDevMs: 0.2,
          location: 'LAN',
        },
        {
          hopNumber: 2,
          ip: '198.51.100.1',
          sent: 10,
          received: 9,
          lossRate: 10,
          lastRTTMs: 32.4,
          avgRTTMs: 35.1,
          bestRTTMs: 31.8,
          worstRTTMs: 42.7,
          stdDevMs: 3.3,
        },
      ],
    })

    await renderNodeDetailPage()

    expect(screen.getByText('example.com')).toBeInTheDocument()
    expect(screen.getByText('192.0.2.1')).toBeInTheDocument()
    expect(screen.getByText('198.51.100.1')).toBeInTheDocument()
    expect(screen.getByText('10.0%')).toBeInTheDocument()
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
