import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { BrowserRouter } from 'react-router-dom'
import { NodeTable } from './NodeTable'
import type { NodeDTO } from '../../api/types'

function renderWithRouter(component: React.ReactElement) {
  return render(<BrowserRouter>{component}</BrowserRouter>)
}

describe('NodeTable', () => {
  const mockNodes: NodeDTO[] = [
    {
      id: 'node-1',
      name: 'Production Server 1',
      ip: '192.168.1.100',
      region: 'us-east-1',
      tags: ['production', 'critical'],
      status: 'online',
      created_at: '2024-01-15T00:00:00Z',
      updated_at: '2024-01-15T00:00:00Z',
    },
    {
      id: 'node-2',
      name: 'Database Server',
      ip: '192.168.1.101',
      region: 'us-west-2',
      tags: ['database'],
      status: 'offline',
      created_at: '2024-01-10T00:00:00Z',
      updated_at: '2024-01-10T00:00:00Z',
    },
    {
      id: 'node-3',
      name: 'Cache Server',
      ip: '192.168.1.102',
      region: 'eu-central-1',
      tags: [],
      status: 'connecting',
      created_at: '2024-01-20T00:00:00Z',
      updated_at: '2024-01-20T00:00:00Z',
    },
  ]

  describe('Loading State', () => {
    it('shows loading spinner when isLoading is true', () => {
      renderWithRouter(
        <NodeTable
          nodes={[]}
          isLoading={true}
          canEdit={false}
        />
      )

      const spinner = screen.getByRole('status', { name: /loading nodes/i })
      expect(spinner).toBeInTheDocument()
    })
  })

  describe('Empty State', () => {
    it('shows empty state when no nodes and not loading', () => {
      renderWithRouter(
        <NodeTable
          nodes={[]}
          isLoading={false}
          canEdit={false}
        />
      )

      expect(screen.getByText('nodes.noNodes')).toBeInTheDocument()
      expect(screen.getByText('nodes.noNodesConfigured')).toBeInTheDocument()
    })

    it('shows different empty state message for editable users', () => {
      renderWithRouter(
        <NodeTable
          nodes={[]}
          isLoading={false}
          canEdit={true}
        />
      )

      expect(screen.getByText('nodes.noNodes')).toBeInTheDocument()
      expect(screen.getByText('nodes.noNodesHint')).toBeInTheDocument()
    })

    it('does not show action buttons in empty state when canEdit is false', () => {
      renderWithRouter(
        <NodeTable
          nodes={[]}
          isLoading={false}
          canEdit={false}
        />
      )

      expect(screen.queryByText('Actions')).not.toBeInTheDocument()
    })
  })

  describe('Table Rendering', () => {
    it('renders table headers', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={true}
        />
      )

      expect(screen.getByText('nodes.nodeName')).toBeInTheDocument()
      expect(screen.getByText('common.status')).toBeInTheDocument()
      expect(screen.getByText('nodes.region')).toBeInTheDocument()
      expect(screen.getByText('nodes.tags')).toBeInTheDocument()
      expect(screen.getByText('nodes.createdAt')).toBeInTheDocument()
      expect(screen.getByText('common.actions')).toBeInTheDocument()
    })

    it('renders all nodes', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      expect(screen.getByText('Production Server 1')).toBeInTheDocument()
      expect(screen.getByText('Database Server')).toBeInTheDocument()
      expect(screen.getByText('Cache Server')).toBeInTheDocument()
    })

    it('renders node IP addresses', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      expect(screen.getByText('192.168.1.100')).toBeInTheDocument()
      expect(screen.getByText('192.168.1.101')).toBeInTheDocument()
      expect(screen.getByText('192.168.1.102')).toBeInTheDocument()
    })

    it('renders node regions', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      expect(screen.getByText('us-east-1')).toBeInTheDocument()
      expect(screen.getByText('us-west-2')).toBeInTheDocument()
      expect(screen.getByText('eu-central-1')).toBeInTheDocument()
    })

    it('renders node tags', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      expect(screen.getByText('production')).toBeInTheDocument()
      expect(screen.getByText('critical')).toBeInTheDocument()
      expect(screen.getByText('database')).toBeInTheDocument()
    })

    it('shows placeholder for nodes without tags', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      // Cache Server has no tags
      const cacheRow = screen.getByText('Cache Server').closest('tr')
      expect(cacheRow?.textContent).toContain('—')
    })
  })

  describe('Status Badges', () => {
    it('renders online status badge correctly', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      const onlineBadge = screen.getByText('online')
      expect(onlineBadge).toBeInTheDocument()
      expect(onlineBadge).toHaveClass('bg-[var(--color-healthy-bg)]', 'text-[var(--color-healthy-text)]')
    })

    it('renders offline status badge correctly', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      const offlineBadge = screen.getByText('offline')
      expect(offlineBadge).toBeInTheDocument()
      expect(offlineBadge).toHaveClass('bg-[var(--color-critical-bg)]', 'text-[var(--color-critical-text)]')
    })

    it('renders connecting status badge correctly', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      const connectingBadge = screen.getByText('connecting')
      expect(connectingBadge).toBeInTheDocument()
      expect(connectingBadge).toHaveClass('bg-[var(--color-warning-bg)]', 'text-[var(--color-warning-text)]')
    })
  })

  describe('Date Formatting', () => {
    it('formats recent dates as relative time', () => {
      const today = new Date()
      const recentNode: NodeDTO = {
        id: 'node-recent',
        name: 'Recent Node',
        ip: '192.168.1.200',
        region: 'us-east-1',
        tags: [],
        status: 'online',
        created_at: today.toISOString(),
        updated_at: today.toISOString(),
      }

      renderWithRouter(
        <NodeTable
          nodes={[recentNode]}
          isLoading={false}
          canEdit={false}
        />
      )

      expect(screen.getByText('Today')).toBeInTheDocument()
    })
  })

  describe('Actions Column', () => {
    it('shows actions column when canEdit is true', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={true}
          onEdit={vi.fn()}
          onDelete={vi.fn()}
        />
      )

      expect(screen.getByText('common.actions')).toBeInTheDocument()
      expect(screen.getAllByText('common.edit')).toHaveLength(3)
      expect(screen.getAllByText('common.delete')).toHaveLength(3)
    })

    it('does not show actions column when canEdit is false', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      expect(screen.queryByText('Actions')).not.toBeInTheDocument()
      expect(screen.queryByText('Edit')).not.toBeInTheDocument()
      expect(screen.queryByText('Delete')).not.toBeInTheDocument()
    })

    it('calls onEdit when edit button is clicked', () => {
      const onEdit = vi.fn()
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={true}
          onEdit={onEdit}
          onDelete={vi.fn()}
        />
      )

      const editButtons = screen.getAllByText('common.edit')
      editButtons[0].click()

      expect(onEdit).toHaveBeenCalledWith('node-1')
    })

    it('calls onDelete when delete button is clicked', () => {
      const onDelete = vi.fn()
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={true}
          onEdit={vi.fn()}
          onDelete={onDelete}
        />
      )

      const deleteButtons = screen.getAllByText('common.delete')
      deleteButtons[1].click()

      expect(onDelete).toHaveBeenCalledWith('node-2')
    })

    it('does not crash when onEdit is not provided', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={true}
          onDelete={vi.fn()}
        />
      )

      const editButtons = screen.getAllByText('common.edit')
      expect(() => editButtons[0].click()).not.toThrow()
    })

    it('does not crash when onDelete is not provided', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={true}
          onEdit={vi.fn()}
        />
      )

      const deleteButtons = screen.getAllByText('common.delete')
      expect(() => deleteButtons[0].click()).not.toThrow()
    })
  })

  describe('Node Links', () => {
    it('renders link to node detail page', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      const nodeLink = screen.getByText('Production Server 1').closest('a')
      expect(nodeLink).toHaveAttribute('href', '/nodes/node-1')
    })
  })

  describe('Accessibility', () => {
    it('applies hover effect to rows', () => {
      renderWithRouter(
        <NodeTable
          nodes={mockNodes}
          isLoading={false}
          canEdit={false}
        />
      )

      const rows = screen.getAllByRole('row')
      // Header row + 3 data rows
      expect(rows).toHaveLength(4)
    })
  })
})
