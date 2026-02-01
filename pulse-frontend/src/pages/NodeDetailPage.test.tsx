import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi } from 'vitest'
import NodeDetailPage from './NodeDetailPage'
import { useNodeDetail } from '../hooks/useNodeDetail'

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
    expect(screen.getByText('192.168.1.1')).toBeInTheDocument()
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
    expect(screen.getByText('ms')).toBeInTheDocument()

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

    expect(screen.getByText('Online')).toBeInTheDocument()
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
    expect(screen.getByText(/Note: Automated problem diagnosis/)).toBeInTheDocument()
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
    const oneMinuteAgo = new Date(now.getTime() - 60000).toISOString()

    const mockNodeStatus = {
      status: 'online',
      last_heartbeat: oneMinuteAgo,
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

    const backLink = screen.getByLabelText('Back to dashboard')
    expect(backLink).toBeInTheDocument()
    expect(backLink.closest('a')).toHaveAttribute('href', '/dashboard')
  })
})
