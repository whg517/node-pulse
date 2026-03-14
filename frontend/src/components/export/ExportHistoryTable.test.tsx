/**
 * ExportHistoryTable Component Tests
 *
 * Tests the export history table component with pagination,
 * filtering, and action buttons.
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ExportHistoryTable } from './ExportHistoryTable'
import type { ExportTask } from '../../types/export'

describe('ExportHistoryTable', () => {
  const mockExports: ExportTask[] = Array.from({ length: 25 }, (_, i) => ({
    id: `export-${i}`,
    user_id: 'user-1',
    node_ids: [`node-${i}`],
    start_time: '2024-01-01T00:00:00Z',
    end_time: '2024-01-07T23:59:59Z',
    metrics: ['latency'],
    format: 'csv',
    status: i % 3 === 0 ? 'completed' : i % 3 === 1 ? 'failed' : 'pending',
    created_at: new Date(Date.now() - i * 3600000).toISOString(),
    completed_at: i % 3 === 0 ? new Date(Date.now() - i * 3600000 + 300000).toISOString() : undefined,
    file_size: i % 3 === 0 ? 1024000 : undefined,
    record_count: i % 3 === 0 ? 1000 : undefined,
    error: i % 3 === 1 ? 'Export failed' : undefined,
  }))

  const defaultProps = {
    exports: mockExports,
    onDownload: vi.fn(),
    onDelete: vi.fn(),
    loading: false,
  }

  describe('Rendering', () => {
    it('renders export history table', () => {
      render(<ExportHistoryTable {...defaultProps} />)

      expect(screen.getByText(/export history/i)).toBeInTheDocument()
      expect(screen.getByRole('table')).toBeInTheDocument()
    })

    it('shows loading state', () => {
      render(<ExportHistoryTable {...defaultProps} loading={true} />)

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('shows empty state when no exports', () => {
      render(<ExportHistoryTable {...defaultProps} exports={[]} />)

      expect(screen.getByText(/no export history/i)).toBeInTheDocument()
    })

    it('displays export rows', () => {
      render(<ExportHistoryTable {...defaultProps} />)

      // Should show first page (20 items)
      expect(screen.getAllByRole('row').length).toBeGreaterThan(1)
    })
  })

  describe('Pagination', () => {
    it('shows pagination controls when more than 20 items', () => {
      render(<ExportHistoryTable {...defaultProps} exports={mockExports} />)

      expect(screen.getByRole('navigation', { name: /pagination/i })).toBeInTheDocument()
    })

    it('paginates through results', async () => {
      const user = userEvent.setup()
      render(<ExportHistoryTable {...defaultProps} exports={mockExports} />)

      const firstPageRows = screen.getAllByRole('row').length

      // Click next page
      const nextButton = screen.getByRole('button', { name: /next/i })
      await user.click(nextButton)

      await waitFor(() => {
        const secondPageRows = screen.getAllByRole('row').length
        expect(secondPageRows).toBeLessThan(firstPageRows)
      })
    })

    it('does not show pagination for 20 or fewer items', () => {
      const smallList = mockExports.slice(0, 10)
      render(<ExportHistoryTable {...defaultProps} exports={smallList} />)

      expect(screen.queryByRole('navigation', { name: /pagination/i })).not.toBeInTheDocument()
    })
  })

  describe('Filtering', () => {
    it('filters by status', async () => {
      const user = userEvent.setup()
      render(<ExportHistoryTable {...defaultProps} />)

      // Click completed filter
      const completedButton = screen.getByRole('button', { name: /completed/i })
      await user.click(completedButton)

      // Should show fewer rows
      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        expect(rows.length).toBeLessThan(mockExports.length)
      })
    })

    it('shows all exports when "All" filter selected', async () => {
      render(<ExportHistoryTable {...defaultProps} />)

      // Make sure all is selected
      expect(screen.getByRole('button', { name: /all/i })).toHaveClass('bg-blue-600')
    })
  })

  describe('Actions', () => {
    it('shows download button for completed exports', () => {
      render(<ExportHistoryTable {...defaultProps} />)

      const downloadButtons = screen.getAllByRole('button', { name: /download/i })
      expect(downloadButtons.length).toBeGreaterThan(0)
    })

    it('calls onDownload when download button clicked', async () => {
      const user = userEvent.setup()
      render(<ExportHistoryTable {...defaultProps} />)

      const downloadButton = screen.getAllByRole('button', { name: /download/i })[0]
      await user.click(downloadButton)

      expect(defaultProps.onDownload).toHaveBeenCalled()
    })

    it('shows delete button for all exports', () => {
      render(<ExportHistoryTable {...defaultProps} />)

      const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
      expect(deleteButtons.length).toBeGreaterThan(0)
    })

    it('calls onDelete when delete button clicked', async () => {
      const user = userEvent.setup()
      render(<ExportHistoryTable {...defaultProps} />)

      const deleteButton = screen.getAllByRole('button', { name: /delete/i })[0]
      await user.click(deleteButton)

      expect(defaultProps.onDelete).toHaveBeenCalled()
    })

    it('does not show download button for non-completed exports', () => {
      const failedExports = mockExports.filter((e) => e.status === 'failed')
      render(<ExportHistoryTable {...defaultProps} exports={failedExports} />)

      const downloadButtons = screen.queryAllByRole('button', { name: /^download$/i })
      expect(downloadButtons.length).toBe(0)
    })
  })

  describe('Status Badges', () => {
    it('shows correct status badges', () => {
      render(<ExportHistoryTable {...defaultProps} />)

      expect(screen.getAllByText(/completed/i).length).toBeGreaterThan(0)
      expect(screen.getAllByText(/failed/i).length).toBeGreaterThan(0)
      expect(screen.getAllByText(/pending/i).length).toBeGreaterThan(0)
    })
  })
})
