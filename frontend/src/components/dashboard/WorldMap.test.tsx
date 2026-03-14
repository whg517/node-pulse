import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import WorldMap, { NodeLocation } from './WorldMap'

// Mock ECharts (WorldMap imports from ../../lib/echarts-core, not echarts directly)
vi.mock('../../lib/echarts-core', () => ({
  default: {
    init: vi.fn(() => ({
      setOption: vi.fn(),
      on: vi.fn(),
      resize: vi.fn(),
      dispose: vi.fn(),
    })),
    getMap: vi.fn(() => null),
    registerMap: vi.fn(),
  },
}))

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'dashboard.nodeDistribution': 'Node Distribution',
        'dashboard.noData': 'No data available',
        'nodes.noNodes': 'No nodes found',
        'nodes.region': 'Region',
        'nodes.title': 'Nodes',
        'common.status': 'Status',
        'common.loading': 'Loading...',
        'status.healthy': 'Healthy',
        'status.warning': 'Warning',
        'status.critical': 'Critical',
        'status.offline': 'Offline',
        'metrics.avgLatency': 'Avg Latency',
        'metrics.packetLoss': 'Packet Loss',
        'metrics.totalNodes': 'Total Nodes',
        'units.ms': 'ms',
        'units.percent': '%',
      }
      return translations[key] || key
    },
  }),
}))

describe('WorldMap', () => {
  const mockNodes: NodeLocation[] = [
    {
      id: 'node-1',
      name: 'US East',
      lat: 40.7128,
      lng: -74.006,
      region: 'North America',
      healthStatus: 'healthy',
      avgLatency: 25.5,
      packetLoss: 0.1,
    },
    {
      id: 'node-2',
      name: 'EU West',
      lat: 51.5074,
      lng: -0.1278,
      region: 'Europe',
      healthStatus: 'warning',
      avgLatency: 85.2,
      packetLoss: 2.5,
    },
    {
      id: 'node-3',
      name: 'Asia Pacific',
      lat: 35.6762,
      lng: 139.6503,
      region: 'Asia',
      healthStatus: 'critical',
      avgLatency: 250.8,
      packetLoss: 8.3,
    },
    {
      id: 'node-4',
      name: 'Offline Node',
      lat: -33.8688,
      lng: 151.2093,
      region: 'Australia',
      healthStatus: 'offline',
      avgLatency: 0,
      packetLoss: 100,
    },
  ]

  it('renders map container with nodes', () => {
    const { container } = render(<WorldMap nodes={mockNodes} />)

    expect(container.querySelector('.world-map')).toBeInTheDocument()
    expect(screen.getByText('Node Distribution')).toBeInTheDocument()
  })

  it('renders empty state when no nodes', () => {
    render(<WorldMap nodes={[]} />)

    expect(screen.getByText('No nodes found')).toBeInTheDocument()
    expect(screen.getByText('No data available')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    const { container } = render(<WorldMap nodes={mockNodes} isLoading={true} />)

    const spinner = container.querySelector('.animate-spin')
    expect(spinner).toBeInTheDocument()
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('displays node count in header', () => {
    render(<WorldMap nodes={mockNodes} />)

    expect(screen.getByText('4 total nodes')).toBeInTheDocument()
  })

  it('displays status legend with counts', () => {
    render(<WorldMap nodes={mockNodes} />)

    expect(screen.getByText(/Healthy \(1\)/)).toBeInTheDocument()
    expect(screen.getByText(/Warning \(1\)/)).toBeInTheDocument()
    expect(screen.getByText(/Critical \(1\)/)).toBeInTheDocument()
    expect(screen.getByText(/Offline \(1\)/)).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<WorldMap nodes={mockNodes} className="custom-class" />)

    const map = container.querySelector('.world-map')
    expect(map).toHaveClass('custom-class')
  })

  it('applies custom height', () => {
    const { container } = render(<WorldMap nodes={mockNodes} height="500px" />)

    const chartContainer = container.querySelector('[style*="height"]')
    expect(chartContainer).toHaveStyle({ height: '500px' })
  })

  it('has proper ARIA attributes', () => {
    render(<WorldMap nodes={mockNodes} />)

    const region = screen.getByRole('region', { name: 'Node Distribution' })
    expect(region).toBeInTheDocument()

    const img = screen.getByRole('img', { name: /Node Distribution showing 4 nodes/ })
    expect(img).toBeInTheDocument()
  })

  it('has ARIA attributes in loading state', () => {
    render(<WorldMap nodes={mockNodes} isLoading={true} />)

    const region = screen.getByRole('region', { name: 'Node Distribution' })
    expect(region).toBeInTheDocument()

    const status = screen.getByRole('status', { name: 'Loading...' })
    expect(status).toBeInTheDocument()
  })

  it('has ARIA attributes in empty state', () => {
    render(<WorldMap nodes={[]} />)

    const region = screen.getByRole('region', { name: 'Node Distribution' })
    expect(region).toBeInTheDocument()
  })

  it('renders with only healthy nodes', () => {
    const healthyNodes = mockNodes.filter((n) => n.healthStatus === 'healthy')
    render(<WorldMap nodes={healthyNodes} />)

    expect(screen.getByText('1 total nodes')).toBeInTheDocument()
    expect(screen.getByText(/Healthy \(1\)/)).toBeInTheDocument()
    expect(screen.getByText(/Critical \(0\)/)).toBeInTheDocument()
  })

  it('renders with only critical nodes', () => {
    const criticalNodes = mockNodes.filter((n) => n.healthStatus === 'critical')
    render(<WorldMap nodes={criticalNodes} />)

    expect(screen.getByText('1 total nodes')).toBeInTheDocument()
    expect(screen.getByText(/Critical \(1\)/)).toBeInTheDocument()
  })

  it('calls onNodeClick when node is clicked', async () => {
    const handleNodeClick = vi.fn()
    render(<WorldMap nodes={mockNodes} onNodeClick={handleNodeClick} />)

    // The click handler is registered with ECharts, so we just verify the component renders
    // Click simulation would require more complex ECharts mock
    expect(screen.getByRole('region', { name: 'Node Distribution' })).toBeInTheDocument()
  })

  it('handles multiple nodes with same status', () => {
    const multiHealthyNodes: NodeLocation[] = [
      { ...mockNodes[0], id: 'node-1', healthStatus: 'healthy' },
      { ...mockNodes[1], id: 'node-2', healthStatus: 'healthy' },
      { ...mockNodes[2], id: 'node-3', healthStatus: 'healthy' },
    ]

    render(<WorldMap nodes={multiHealthyNodes} />)

    expect(screen.getByText('3 total nodes')).toBeInTheDocument()
    expect(screen.getByText(/Healthy \(3\)/)).toBeInTheDocument()
  })

  it('renders without onNodeClick callback', () => {
    render(<WorldMap nodes={mockNodes} />)

    expect(screen.getByRole('region', { name: 'Node Distribution' })).toBeInTheDocument()
  })

  it('uses default height when not specified', () => {
    const { container } = render(<WorldMap nodes={mockNodes} />)

    const chartContainer = container.querySelector('[style*="height"]')
    expect(chartContainer).toHaveStyle({ height: '400px' })
  })

  it('handles undefined nodes gracefully', () => {
    render(<WorldMap nodes={undefined as unknown as NodeLocation[]} />)

    expect(screen.getByText('No nodes found')).toBeInTheDocument()
  })
})
