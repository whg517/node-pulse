/**
 * ExportStatusCard Component Tests
 *
 * Tests the export status tracking card component
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ExportStatusCard } from './ExportStatusCard'
import type { ExportTask } from '../../types/export'

describe('ExportStatusCard', () => {
  const mockExport: ExportTask = {
    id: 'export-123',
    user_id: 'user-1',
    node_ids: ['node-1', 'node-2'],
    start_time: '2024-01-01T00:00:00Z',
    end_time: '2024-01-07T23:59:59Z',
    metrics: ['latency', 'packet_loss_rate'],
    format: 'csv',
    status: 'pending',
    created_at: '2024-01-01T00:00:00Z',
  }

  describe('Rendering', () => {
    it('renders export task details', () => {
      render(<ExportStatusCard exportTask={mockExport} />)

      expect(screen.getByText(/export task/i)).toBeInTheDocument()
      expect(screen.getByText(/2 selected/i)).toBeInTheDocument()
    })

    it('shows CSV format indicator', () => {
      render(<ExportStatusCard exportTask={mockExport} />)

      expect(screen.getByText(/csv/i)).toBeInTheDocument()
    })
  })

  describe('Status Display', () => {
    it('shows pending status with spinner', () => {
      render(<ExportStatusCard exportTask={{ ...mockExport, status: 'pending' }} />)

      expect(screen.getAllByText(/pending/i).length).toBeGreaterThan(0)
    })

    it('shows processing status with spinner', () => {
      render(<ExportStatusCard exportTask={{ ...mockExport, status: 'processing' }} />)

      expect(screen.getAllByText(/processing/i).length).toBeGreaterThan(0)
    })

    it('shows completed status without spinner', () => {
      render(
        <ExportStatusCard
          exportTask={{
            ...mockExport,
            status: 'completed',
            file_size: 1024000,
            record_count: 1000,
          }}
        />
      )

      expect(screen.getAllByText(/completed/i).length).toBeGreaterThan(0)
      expect(screen.getByText(/1000 KB/i)).toBeInTheDocument()
      expect(screen.getByText(/1,000/i)).toBeInTheDocument()
    })

    it('shows failed status with error message', () => {
      render(
        <ExportStatusCard
          exportTask={{
            ...mockExport,
            status: 'failed',
            error: 'Export failed: No data found',
          }}
        />
      )

      expect(screen.getAllByText(/failed/i).length).toBeGreaterThan(0)
      expect(screen.getByText(/export failed: no data found/i)).toBeInTheDocument()
    })
  })

  describe('Actions', () => {
    it('shows download button when completed', () => {
      const onDownload = vi.fn()
      render(
        <ExportStatusCard
          exportTask={{ ...mockExport, status: 'completed' }}
          onDownload={onDownload}
        />
      )

      const downloadButton = screen.getByRole('button', { name: /download/i })
      expect(downloadButton).toBeInTheDocument()
    })

    it('calls onDownload when download button clicked', async () => {
      const onDownload = vi.fn()
      const user = userEvent.setup()

      render(
        <ExportStatusCard
          exportTask={{ ...mockExport, status: 'completed' }}
          onDownload={onDownload}
        />
      )

      await user.click(screen.getByRole('button', { name: /download/i }))
      expect(onDownload).toHaveBeenCalledWith('export-123')
    })

    it('does not show download button when not completed', () => {
      const onDownload = vi.fn()
      render(
        <ExportStatusCard
          exportTask={{ ...mockExport, status: 'pending' }}
          onDownload={onDownload}
        />
      )

      expect(screen.queryByRole('button', { name: /download/i })).not.toBeInTheDocument()
    })

    it('shows dismiss button', async () => {
      const onDismiss = vi.fn()
      const user = userEvent.setup()

      render(<ExportStatusCard exportTask={mockExport} onDismiss={onDismiss} />)

      await user.click(screen.getByRole('button', { name: /dismiss/i }))
      expect(onDismiss).toHaveBeenCalled()
    })
  })

  describe('Time Information', () => {
    it('shows time range for export', () => {
      render(<ExportStatusCard exportTask={mockExport} />)

      expect(screen.getByText(/2024/i)).toBeInTheDocument()
      expect(screen.getAllByText(/jan/i).length).toBeGreaterThan(0)
    })

    it('shows estimated completion time for pending/processing', () => {
      render(
        <ExportStatusCard
          exportTask={{
            ...mockExport,
            status: 'processing',
            created_at: new Date(Date.now() - 60000).toISOString(), // 1 min ago
          }}
        />
      )

      // Estimated completion is calculated based on time elapsed
      expect(screen.getByText(/min remaining/i)).toBeInTheDocument()
    })

    it('shows actual completion time when completed', () => {
      render(
        <ExportStatusCard
          exportTask={{
            ...mockExport,
            status: 'completed',
            created_at: '2024-01-01T00:00:00Z',
            completed_at: '2024-01-01T00:05:00Z',
          }}
        />
      )

      expect(screen.getByText(/5 min/i)).toBeInTheDocument()
    })
  })
})
