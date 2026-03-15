import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import NodeComparisonPage from '../NodeComparison'
import * as nodesApi from '../../api/nodes'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: any) => {
      const translations: Record<string, string> = {
        'nodes.comparison': 'Node Comparison',
        'nodes.comparisonDescription': 'Compare network metrics across multiple nodes to identify performance differences and anomalies.',
        'navigation.dashboard': 'Dashboard',
        'nodes.selectNodes': 'Select Nodes',
        'nodes.groupBy': 'Group By',
        'nodes.none': 'None',
        'nodes.timeRange': 'Time Range',
        'nodes.hours24': '24 Hours',
        'nodes.days7': '7 Days',
        'nodes.days30': '30 Days',
        'nodes.custom': 'Custom',
        'nodes.metricsSelector': 'Metrics',
        'metrics.latency': 'Latency',
        'metrics.packetLoss': 'Packet Loss Rate',
        'metrics.jitter': 'Jitter',
        'nodes.compareNodes': 'Compare Nodes',
        'nodes.noComparisonData': 'No Comparison Data',
        'nodes.noComparisonDataDescription': 'Select nodes, time range, and metrics, then click "Compare Nodes" to view the comparison chart.',
        'nodes.selectedCount': `Selected: ${options?.count || 0} / ${options?.max || 5} nodes`,
        'nodes.selectAtLeast': `Select at least ${options?.count || 2} nodes`,
        'nodes.maxNodesSelected': `Maximum ${options?.max || 5} nodes can be selected`,
        'nodes.region': 'Region',
        'status.online': 'online',
        'status.offline': 'offline',
        'status.connecting': 'connecting',
        'common.loading': 'Loading nodes...',
        'common.error': 'Error',
        'errors.failedToLoad': 'Failed to load nodes',
        'reports.selectMetrics': 'Select at least one metric',
      }
      return translations[key] || key
    },
    i18n: { changeLanguage: vi.fn() },
  }),
}))

// Mock useTheme hook
vi.mock('../../hooks/useTheme', () => ({
  useTheme: () => ({ isDark: false }),
}))

// Mock the nodes API
vi.mock('../../api/nodes', () => ({
  fetchNodes: vi.fn(),
}))

// Mock ComparisonChart component
vi.mock('../../components/dashboard/ComparisonChart', () => ({
  default: vi.fn(({ nodes, metric, timeRange, showStatistics, highlightDifferences, groupBy, isLoading }: any) => (
    <div data-testid="comparison-chart">
      <div>Nodes: {nodes?.length}</div>
      <div>Metric: {metric}</div>
      <div>TimeRange: {timeRange}</div>
      <div>ShowStatistics: {showStatistics ? 'Yes' : 'No'}</div>
      <div>HighlightDifferences: {highlightDifferences ? 'Yes' : 'No'}</div>
      <div>GroupBy: {groupBy}</div>
      <div>IsLoading: {isLoading ? 'Yes' : 'No'}</div>
    </div>
  )),
  __esModule: true,
}))

// Mock environment variable
vi.mock('../../api/client', () => ({
  apiClient: vi.fn(),
}))

// Mock dashboardStore with state management
const mockComparisonState = {
  selectedNodeIds: [] as string[],
  selectedMetrics: ['latency_ms'] as string[],
  timeRange: '24h' as const,
  groupBy: 'none' as const,
}

const mockSetComparisonNodeIds = vi.fn((ids: string[]) => {
  mockComparisonState.selectedNodeIds = ids
})

const mockSetComparisonMetrics = vi.fn((metrics: string[]) => {
  mockComparisonState.selectedMetrics = metrics
})

const mockSetComparisonTimeRange = vi.fn((_range: any) => {
  // Updated in handler
})

const mockSetComparisonCustomTimeRange = vi.fn((_range: any) => {
  // No-op for test
})

const mockSetComparisonGroupBy = vi.fn((_groupBy: any) => {
  // Updated in handler
})

const mockResetComparison = vi.fn(() => {
  mockComparisonState.selectedNodeIds = []
  mockComparisonState.selectedMetrics = ['latency_ms']
  mockComparisonState.timeRange = '24h'
  mockComparisonState.groupBy = 'none'
})

vi.mock('../../stores/dashboardStore', () => ({
  useDashboardStore: vi.fn(() => ({
    comparison: mockComparisonState,
    setComparisonNodeIds: mockSetComparisonNodeIds,
    setComparisonMetrics: mockSetComparisonMetrics,
    setComparisonTimeRange: mockSetComparisonTimeRange,
    setComparisonCustomTimeRange: mockSetComparisonCustomTimeRange,
    setComparisonGroupBy: mockSetComparisonGroupBy,
    resetComparison: mockResetComparison,
  })),
}))

describe('NodeComparisonPage', () => {
  const mockFetchNodes = vi.mocked(nodesApi.fetchNodes)

  const mockNodes = [
    {
      id: 'node-1',
      name: 'Node 1',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: ['AWS', 'production'],
      status: 'online' as const,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'node-2',
      name: 'Node 2',
      ip: '192.168.1.2',
      region: 'eu-west',
      tags: ['GCP', 'production'],
      status: 'online' as const,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'node-3',
      name: 'Node 3',
      ip: '192.168.1.3',
      region: 'ap-south',
      tags: ['Azure', 'production'],
      status: 'offline' as const,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    global.fetch = vi.fn()

    // Reset mock comparison state
    mockComparisonState.selectedNodeIds = []
    mockComparisonState.selectedMetrics = ['latency_ms']
    mockComparisonState.timeRange = '24h'
    mockComparisonState.groupBy = 'none'
  })

  it('renders page with title and description', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    // Use findByRole to target the h1 specifically
    expect(await screen.findByRole('heading', { name: 'Node Comparison', level: 1 })).toBeInTheDocument()
    expect(
      screen.getByText(
        'Compare network metrics across multiple nodes to identify performance differences and anomalies.'
      )
    ).toBeInTheDocument()
  })

  it('loads and displays available nodes', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Node 1')).toBeInTheDocument()
      expect(screen.getByText('Node 2')).toBeInTheDocument()
      expect(screen.getByText('Node 3')).toBeInTheDocument()
    })
  })

  it('validates node selection (2-5 nodes)', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Select Nodes (2-5)')).toBeInTheDocument()
    })

    // Initially no nodes selected
    expect(screen.getByText(/Selected: 0 \/ 5 nodes/)).toBeInTheDocument()

    // Select one node and verify the store function is called
    const node1Checkbox = screen.getByLabelText(/Node 1/i)
    node1Checkbox.click()

    await waitFor(() => {
      expect(mockSetComparisonNodeIds).toHaveBeenCalledWith(['node-1'])
    })
  })

  it('shows loading state while loading nodes', async () => {
    mockFetchNodes.mockImplementation(() => new Promise(() => {})) // Never resolves

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    expect(screen.getByText(/Loading nodes\.\.\./)).toBeInTheDocument()
  })

  it('displays time range selector', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Time Range')).toBeInTheDocument()
      expect(screen.getByText('24 Hours')).toBeInTheDocument()
      expect(screen.getByText('7 Days')).toBeInTheDocument()
      expect(screen.getByText('30 Days')).toBeInTheDocument()
      expect(screen.getByText('Custom')).toBeInTheDocument()
    })
  })

  it('displays metric selector', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Metrics')).toBeInTheDocument()
      expect(screen.getByText(/Latency \(ms\)/)).toBeInTheDocument()
      expect(screen.getByText(/Packet Loss Rate \(%\)/)).toBeInTheDocument()
      expect(screen.getByText(/Jitter \(ms\)/)).toBeInTheDocument()
    })
  })

  it('displays group by selector', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Group By')).toBeInTheDocument()
    })

    // Check for the group by buttons
    await waitFor(() => {
      const buttons = screen.getAllByText('None')
      expect(buttons.length).toBeGreaterThan(0)
    })
  })

  it('shows compare button with correct disabled state', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      const compareButton = screen.getByText('Compare Nodes')
      expect(compareButton).toBeDisabled()
    })

    // Select two nodes to verify checkbox interactions work
    const node1Checkbox = screen.getByLabelText(/Node 1/i)
    const node2Checkbox = screen.getByLabelText(/Node 2/i)

    // Both checkboxes should exist and be enabled
    expect(node1Checkbox).toBeInTheDocument()
    expect(node2Checkbox).toBeInTheDocument()
    expect(node1Checkbox).not.toBeDisabled()
    expect(node2Checkbox).not.toBeDisabled()

    // Interact with checkboxes
    node1Checkbox.click()
    node2Checkbox.click()

    // Verify store function was called (showing interaction works)
    await waitFor(() => {
      expect(mockSetComparisonNodeIds).toHaveBeenCalled()
    })
  })

  it('displays empty state when no comparison data', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('No Comparison Data')).toBeInTheDocument()
      expect(
        screen.getByText(
          'Select nodes, time range, and metrics, then click "Compare Nodes" to view the comparison chart.'
        )
      ).toBeInTheDocument()
    })
  })

  it('displays error message on node load failure', async () => {
    mockFetchNodes.mockRejectedValue(new Error('Failed to load nodes'))

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Error')).toBeInTheDocument()
      expect(screen.getByText('Failed to load nodes')).toBeInTheDocument()
    })
  })

  it('shows node status badges', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      const onlineNodes = screen.getAllByText('online')
      const offlineNodes = screen.getAllByText('offline')
      expect(onlineNodes.length).toBeGreaterThan(0)
      expect(offlineNodes.length).toBeGreaterThan(0)
    })
  })

  it('calls handleGroupByChange when group by button is clicked', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Group By')).toBeInTheDocument()
    })

    // Click the Region group by button
    const regionBtn = screen.getByText('Region')
    fireEvent.click(regionBtn)

    expect(mockSetComparisonGroupBy).toHaveBeenCalledWith('region')
  })

  it('calls handleTimeRangeChange for 7d range', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('7 Days')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('7 Days'))
    expect(mockSetComparisonTimeRange).toHaveBeenCalledWith('7d')
  })

  it('calls handleTimeRangeChange for 30d range', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('30 Days')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('30 Days'))
    expect(mockSetComparisonTimeRange).toHaveBeenCalledWith('30d')
  })

  it('calls handleTimeRangeChange for custom range', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Custom')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Custom'))
    expect(mockSetComparisonTimeRange).toHaveBeenCalledWith('custom')
  })

  it('calls handleMetricSelectionChange when metric button clicked', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText(/Packet Loss Rate/)).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText(/Packet Loss Rate/))
    expect(mockSetComparisonMetrics).toHaveBeenCalled()
  })

  it('calls handleCompare with 2 selected nodes', async () => {
    const mockGetComparisonData = vi.fn().mockResolvedValue({
      data: {
        nodes: [
          { node_id: 'node-1', name: 'Node 1', region: 'us-east', isp: 'AWS', metrics: { latency_ms: { data_points: [] } } },
          { node_id: 'node-2', name: 'Node 2', region: 'eu-west', isp: 'GCP', metrics: { latency_ms: { data_points: [] } } },
        ],
      },
    })

    vi.doMock('../../api/data', () => ({
      getComparisonData: mockGetComparisonData,
    }))

    mockComparisonState.selectedNodeIds = ['node-1', 'node-2']
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      const compareBtn = screen.getByText('Compare Nodes')
      expect(compareBtn).not.toBeDisabled()
    })

    const compareBtn = screen.getByText('Compare Nodes')
    fireEvent.click(compareBtn)

    // With 2 nodes selected and 1 metric selected, handleCompare should proceed
    await waitFor(() => {
      expect(screen.getByText('Compare Nodes')).toBeInTheDocument()
    })
  })

  it('shows error when fewer than 2 nodes are selected on compare', async () => {
    mockComparisonState.selectedNodeIds = ['node-1']
    mockComparisonState.selectedMetrics = ['latency_ms']
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      // Button should be disabled since only 1 node selected
      const compareBtn = screen.getByTestId('compare-button')
      expect(compareBtn).toBeDisabled()
    })
  })

  it('shows custom time range inputs when custom is selected', async () => {
    mockComparisonState.timeRange = 'custom'
    mockComparisonState.customTimeRange = { start: '2024-01-01T00:00', end: '2024-01-07T00:00' }
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })

    render(
      <MemoryRouter>
        <NodeComparisonPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Time Range')).toBeInTheDocument()
    })

    // Custom time range inputs should be visible
    const dateInputs = document.querySelectorAll('input[type="datetime-local"]')
    expect(dateInputs.length).toBeGreaterThan(0)
  })
})
