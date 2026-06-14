import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { BrowserRouter } from 'react-router-dom'
import { NodeListTable } from '../NodeListTable'
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

describe('NodeListTable', () => {
  const mockNodes: NodeDTO[] = [
    {
      id: 'node-1',
      name: 'Beacon-CN-East-1',
      ip: '192.168.1.100',
      region: '华东',
      tags: ['production'],
      status: 'online',
      created_at: '2026-01-26T10:00:00Z',
      updated_at: '2026-01-26T10:00:00Z',
    },
    {
      id: 'node-2',
      name: 'Beacon-CN-North-1',
      ip: '192.168.1.101',
      region: '华北',
      tags: ['production'],
      status: 'offline',
      created_at: '2026-01-26T10:00:00Z',
      updated_at: '2026-01-26T10:00:00Z',
    },
  ]

  const mockMetrics: MetricsDTO[] = [
    {
      node_id: 'node-1',
      latency_ms: 50,
      packet_loss_rate: 0.1,
      jitter_ms: 5,
      timestamp: new Date().toISOString(),
    },
  ]

  describe('rendering', () => {
    it('should render table with nodes', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      expect(screen.getByText('Node List')).toBeInTheDocument()
      expect(screen.getByText('Beacon-CN-East-1')).toBeInTheDocument()
      expect(screen.getByText('Beacon-CN-North-1')).toBeInTheDocument()
      expect(screen.getByText('192.168.1.100')).toBeInTheDocument()
      expect(screen.getByText('192.168.1.101')).toBeInTheDocument()
    })

    it('should render all column headers', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      expect(screen.getByText('Node Name')).toBeInTheDocument()
      expect(screen.getByText('IP Address')).toBeInTheDocument()
      expect(screen.getByText('Region')).toBeInTheDocument()
      expect(screen.getByText('Status')).toBeInTheDocument()
      expect(screen.getByText('Health')).toBeInTheDocument()
    })

    it('should render region badges', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      expect(screen.getByText('华东')).toBeInTheDocument()
      expect(screen.getByText('华北')).toBeInTheDocument()
    })

    it('should render online/offline status', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      const onlineBadges = screen.getAllByText('Online')
      const offlineBadges = screen.getAllByText('Offline')
      expect(onlineBadges.length).toBeGreaterThan(0)
      expect(offlineBadges.length).toBeGreaterThan(0)
    })
  })

  describe('health status display', () => {
    it('should display health status badge for each node', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      // node-1 has metrics, should show health status
      // node-2 has no metrics, should show offline
      const healthStatuses = screen.getAllByText(/Healthy|Offline|Warning|Critical/)
      expect(healthStatuses.length).toBeGreaterThan(0)
    })

    it('should show offline status for nodes without metrics', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      // node-2 has no metrics
      expect(screen.getAllByText('Offline').length).toBeGreaterThan(0)
    })
  })

  describe('loading state', () => {
    it('should render loading skeleton when isLoading is true', () => {
      renderWithRouter(
        <NodeListTable nodes={[]} metrics={[]} isLoading={true} />
      )

      // Check for skeleton elements (they have specific classes)
      const skeletons = document.querySelectorAll('.animate-pulse')
      expect(skeletons.length).toBeGreaterThan(0)
    })
  })

  describe('empty state', () => {
    it('should render empty state when no nodes exist', () => {
      renderWithRouter(<NodeListTable nodes={[]} metrics={[]} />)

      expect(screen.getByText('No nodes')).toBeInTheDocument()
      expect(
        screen.getByText(/No monitoring nodes configured yet/)
      ).toBeInTheDocument()
    })

    it('should show empty state message and call to action', () => {
      renderWithRouter(<NodeListTable nodes={[]} metrics={[]} />)

      expect(
        screen.getByText(/Get started by adding your first node/)
      ).toBeInTheDocument()
    })
  })

  describe('interactions', () => {
    it('should navigate to node detail page on row click', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      const rows = screen.getAllByRole('row')
      // Click second data row (skip header)
      rows[1].click()

      expect(mockNavigate).toHaveBeenCalledWith('/nodes/node-1', { state: { breadcrumbLabel: 'Beacon-CN-East-1' } })
    })

    it('should make rows clickable', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      const rows = screen.getAllByRole('row')
      // Data rows should have cursor-pointer class
      expect(rows[1]).toHaveClass('cursor-pointer')
    })
  })

  describe('accessibility', () => {
    it('should have proper table structure', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      const table = screen.getByRole('table')
      expect(table).toBeInTheDocument()

      const headers = screen.getAllByRole('columnheader')
      expect(headers.length).toBe(5) // Node Name, IP, Region, Status, Health
    })

    it('should have proper ARIA labels for status badges', () => {
      renderWithRouter(<NodeListTable nodes={mockNodes} metrics={mockMetrics} />)

      const statusBadges = screen.getAllByRole('status')
      expect(statusBadges.length).toBeGreaterThan(0)
    })
  })

  describe('responsive design', () => {
    it('should render overflow-x-auto for table container', () => {
      const { container } = renderWithRouter(
        <NodeListTable nodes={mockNodes} metrics={mockMetrics} />
      )

      const overflowContainer = container.querySelector('.overflow-x-auto')
      expect(overflowContainer).toBeInTheDocument()
    })
  })
})
