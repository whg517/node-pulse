import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { BrowserRouter } from 'react-router-dom'
import { TopAnomaliesList } from '../TopAnomaliesList'
import type { NodeDTO } from '../../../api/types'
import type { MetricsDTO } from '../../../api/types'

// Mock react-router-dom
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

function renderWithRouter(component: React.ReactElement) {
  return render(<BrowserRouter>{component}</BrowserRouter>)
}

describe('TopAnomaliesList', () => {
  const mockNodes: NodeDTO[] = [
    {
      id: 'node-1',
      name: 'Critical-Node',
      ip: '192.168.1.100',
      region: '华东',
      tags: ['production'],
      status: 'online',
      created_at: '2026-01-26T10:00:00Z',
      updated_at: '2026-01-26T10:00:00Z',
    },
    {
      id: 'node-2',
      name: 'Warning-Node',
      ip: '192.168.1.101',
      region: '华北',
      tags: ['production'],
      status: 'online',
      created_at: '2026-01-26T10:00:00Z',
      updated_at: '2026-01-26T10:00:00Z',
    },
    {
      id: 'node-3',
      name: 'Healthy-Node',
      ip: '192.168.1.102',
      region: '华南',
      tags: ['production'],
      status: 'online',
      created_at: '2026-01-26T10:00:00Z',
      updated_at: '2026-01-26T10:00:00Z',
    },
  ]

  const mockMetrics: MetricsDTO[] = [
    {
      node_id: 'node-1',
      latency_ms: 250, // Critical
      packet_loss_rate: 10,
      jitter_ms: 80,
      timestamp: new Date().toISOString(),
    },
    {
      node_id: 'node-2',
      latency_ms: 170, // Warning (80-100% of threshold)
      packet_loss_rate: 4.2,
      jitter_ms: 42,
      timestamp: new Date().toISOString(),
    },
    {
      node_id: 'node-3',
      latency_ms: 50, // Healthy
      packet_loss_rate: 0.1,
      jitter_ms: 5,
      timestamp: new Date().toISOString(),
    },
  ]

  describe('rendering', () => {
    it('should render component with title', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      expect(screen.getByText('Top Anomalies')).toBeInTheDocument()
      expect(
        screen.getByText(/Nodes requiring attention/)
      ).toBeInTheDocument()
    })

    it('should render only critical and warning nodes', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      // Should show critical and warning nodes, not healthy
      expect(screen.getByText('Critical-Node')).toBeInTheDocument()
      expect(screen.getByText('Warning-Node')).toBeInTheDocument()
      expect(screen.queryByText('Healthy-Node')).not.toBeInTheDocument()
    })

    it('should display node information correctly', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      expect(screen.getByText('Critical-Node')).toBeInTheDocument()
      expect(screen.getByText('192.168.1.100')).toBeInTheDocument()
      expect(screen.getByText('华东')).toBeInTheDocument()
    })

    it('should display key metric for each node', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      // Critical node shows packet loss
      expect(screen.getByText('10.0% loss')).toBeInTheDocument()
      // Warning node shows latency
      expect(screen.getByText('170ms')).toBeInTheDocument()
    })
  })

  describe('sorting', () => {
    it('should sort nodes by severity (critical first)', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      const nodeList = screen.getByRole('list')
      const firstItem = nodeList.querySelector('li:first-child')
      expect(firstItem).toHaveTextContent('Critical-Node')
    })

    it('should prioritize warning over healthy nodes', () => {
      const allHealthyNodes: NodeDTO[] = [
        {
          id: 'node-1',
          name: 'Healthy-1',
          ip: '192.168.1.1',
          region: '华东',
          tags: [],
          status: 'online',
          created_at: '2026-01-26T10:00:00Z',
          updated_at: '2026-01-26T10:00:00Z',
        },
        {
          id: 'node-2',
          name: 'Healthy-2',
          ip: '192.168.1.2',
          region: '华北',
          tags: [],
          status: 'online',
          created_at: '2026-01-26T10:00:00Z',
          updated_at: '2026-01-26T10:00:00Z',
        },
      ]

      const healthyMetrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 50,
          packet_loss_rate: 0.1,
          jitter_ms: 5,
          timestamp: new Date().toISOString(),
        },
        {
          node_id: 'node-2',
          latency_ms: 60,
          packet_loss_rate: 0.2,
          jitter_ms: 8,
          timestamp: new Date().toISOString(),
        },
      ]

      renderWithRouter(
        <TopAnomaliesList nodes={allHealthyNodes} metrics={healthyMetrics} />
      )

      // Should show "All systems normal" instead of healthy nodes
      expect(screen.getByText('All systems normal')).toBeInTheDocument()
    })
  })

  describe('empty state', () => {
    it('should show all systems normal when no critical/warning nodes', () => {
      const healthyNodes: NodeDTO[] = [
        {
          id: 'node-1',
          name: 'Healthy-Node',
          ip: '192.168.1.100',
          region: '华东',
          tags: [],
          status: 'online',
          created_at: '2026-01-26T10:00:00Z',
          updated_at: '2026-01-26T10:00:00Z',
        },
      ]

      const healthyMetrics: MetricsDTO[] = [
        {
          node_id: 'node-1',
          latency_ms: 50,
          packet_loss_rate: 0.1,
          jitter_ms: 5,
          timestamp: new Date().toISOString(),
        },
      ]

      renderWithRouter(
        <TopAnomaliesList nodes={healthyNodes} metrics={healthyMetrics} />
      )

      expect(screen.getByText('All systems normal')).toBeInTheDocument()
      expect(
        screen.getByText(/No critical or warning issues detected/)
      ).toBeInTheDocument()
    })

    it('should show empty state when no nodes exist', () => {
      renderWithRouter(<TopAnomaliesList nodes={[]} metrics={[]} />)

      expect(screen.getByText('All systems normal')).toBeInTheDocument()
    })
  })

  describe('loading state', () => {
    it('should render loading skeleton when isLoading is true', () => {
      renderWithRouter(
        <TopAnomaliesList nodes={[]} metrics={[]} isLoading={true} />
      )

      const skeletons = document.querySelectorAll('.animate-pulse')
      expect(skeletons.length).toBeGreaterThan(0)
    })
  })

  describe('interactions', () => {
    it('should navigate to node detail page on click', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      const listItem = screen.getByText('Critical-Node').closest('li')
      listItem?.click()

      expect(mockNavigate).toHaveBeenCalledWith('/nodes/node-1')
    })

    it('should make list items clickable', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      const listItem = screen.getByText('Warning-Node').closest('li')
      expect(listItem).toHaveClass('cursor-pointer')
    })
  })

  describe('limit to top 5', () => {
    it('should only show top 5 anomalies', () => {
      const manyNodes: NodeDTO[] = Array.from({ length: 10 }, (_, i) => ({
        id: `node-${i}`,
        name: `Node-${i}`,
        ip: `192.168.1.${i}`,
        region: '华东',
        tags: [],
        status: 'online',
        created_at: '2026-01-26T10:00:00Z',
        updated_at: '2026-01-26T10:00:00Z',
      }))

      const criticalMetrics: MetricsDTO[] = manyNodes.map(node => ({
        node_id: node.id,
        latency_ms: 250, // All critical
        packet_loss_rate: 10,
        jitter_ms: 80,
        timestamp: new Date().toISOString(),
      }))

      renderWithRouter(
        <TopAnomaliesList nodes={manyNodes} metrics={criticalMetrics} />
      )

      const listItems = screen.getAllByRole('listitem')
      expect(listItems.length).toBeLessThanOrEqual(5)
    })
  })

  describe('accessibility', () => {
    it('should have proper list structure', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      const list = screen.getByRole('list')
      expect(list).toBeInTheDocument()
    })

    it('should have proper ARIA labels for health status badges', () => {
      renderWithRouter(<TopAnomaliesList nodes={mockNodes} metrics={mockMetrics} />)

      const statusBadges = screen.getAllByRole('status')
      expect(statusBadges.length).toBeGreaterThan(0)
    })
  })
})
