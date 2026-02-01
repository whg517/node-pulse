import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { AlertRecordDetailModal } from '../AlertRecordDetailModal'
import type { AlertRecordDTO } from '../../../api/alertRecords'
import type { NodeDTO } from '../../../api/types'

// Mock react-router
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('AlertRecordDetailModal', () => {
  const mockNodes: NodeDTO[] = [
    { id: 'node-1', name: 'Test Node 1', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online', created_at: '2024-01-01', updated_at: '2024-01-01' },
    { id: 'node-2', name: 'Test Node 2', ip: '192.168.1.2', region: 'us-west', tags: [], status: 'online', created_at: '2024-01-01', updated_at: '2024-01-01' },
  ]

  const mockRecord: AlertRecordDTO = {
    id: 'record-1',
    alert_event_id: 'event-1',
    node_id: 'node-1',
    metric: 'latency',
    level: 'P1',
    status: 'pending',
    created_at: '2024-01-01T10:00:00Z',
    updated_at: '2024-01-01T10:00:00Z',
  }

  it('renders modal with record details', () => {
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.getByText('告警记录详情')).toBeInTheDocument()
    expect(screen.getByText('告警 ID')).toBeInTheDocument()
    expect(screen.getByText('节点')).toBeInTheDocument()
    expect(screen.getByText('指标类型')).toBeInTheDocument()
    expect(screen.getByText('告警级别')).toBeInTheDocument()
    expect(screen.getByText('状态')).toBeInTheDocument()
    expect(screen.getByText('创建时间')).toBeInTheDocument()
  })

  it('displays record information correctly', () => {
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.getByText('record-1')).toBeInTheDocument()
    expect(screen.getByText('Test Node 1')).toBeInTheDocument()
    expect(screen.getByText('延迟')).toBeInTheDocument()
    expect(screen.getByText('P1')).toBeInTheDocument()
    expect(screen.getByText('未处理')).toBeInTheDocument()
  })

  it('shows status update buttons for pending status when user can edit', () => {
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.getByText('更新状态')).toBeInTheDocument()
    expect(screen.getByText('标记为处理中')).toBeInTheDocument()
    expect(screen.getByText('标记为已解决')).toBeInTheDocument()
  })

  it('shows only resolved button for in_progress status', () => {
    const inProgressRecord: AlertRecordDTO = {
      ...mockRecord,
      status: 'in_progress',
    }

    render(
      <AlertRecordDetailModal
        record={inProgressRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.queryByText('标记为处理中')).not.toBeInTheDocument()
    expect(screen.getByText('标记为已解决')).toBeInTheDocument()
  })

  it('hides status update buttons for resolved status', () => {
    const resolvedRecord: AlertRecordDTO = {
      ...mockRecord,
      status: 'resolved',
    }

    render(
      <AlertRecordDetailModal
        record={resolvedRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.queryByText('更新状态')).not.toBeInTheDocument()
    expect(screen.queryByText('标记为处理中')).not.toBeInTheDocument()
    expect(screen.queryByText('标记为已解决')).not.toBeInTheDocument()
  })

  it('hides status update buttons when user cannot edit', () => {
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={false}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.queryByText('更新状态')).not.toBeInTheDocument()
  })

  it('calls onClose when close button clicked', () => {
    const onClose = vi.fn()
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={onClose}
        onStatusUpdate={vi.fn()}
      />
    )

    const closeButton = screen.getByText('关闭')
    fireEvent.click(closeButton)

    expect(onClose).toHaveBeenCalled()
  })

  it('calls onClose when X button clicked', () => {
    const onClose = vi.fn()
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={onClose}
        onStatusUpdate={vi.fn()}
      />
    )

    const xButton = screen.getByRole('button', { name: '' }).querySelector('svg')
    if (xButton) {
      fireEvent.click(xButton.closest('button')!)
      expect(onClose).toHaveBeenCalled()
    }
  })

  it('calls onStatusUpdate when status update button clicked', async () => {
    const onStatusUpdate = vi.fn().mockResolvedValue(undefined)
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={onStatusUpdate}
      />
    )

    const resolveButton = screen.getByText('标记为已解决')
    fireEvent.click(resolveButton)

    await waitFor(() => {
      expect(onStatusUpdate).toHaveBeenCalledWith('record-1', 'resolved')
    })
  })

  it('navigates to node details when view node details clicked', () => {
    const onClose = vi.fn()
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={onClose}
        onStatusUpdate={vi.fn()}
      />
    )

    const viewNodeButton = screen.getByText('查看节点详情')
    fireEvent.click(viewNodeButton)

    expect(mockNavigate).toHaveBeenCalledWith('/nodes/node-1')
    expect(onClose).toHaveBeenCalled()
  })

  it('displays node IP when node found', () => {
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.getByText('IP: 192.168.1.1')).toBeInTheDocument()
  })

  it('shows error message when status update fails', async () => {
    const onStatusUpdate = vi.fn().mockRejectedValue(new Error('Update failed'))
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={onStatusUpdate}
      />
    )

    const resolveButton = screen.getByText('标记为已解决')
    fireEvent.click(resolveButton)

    await waitFor(() => {
      expect(screen.getByText('Update failed')).toBeInTheDocument()
    })
  })

  it('displays updated timestamp when different from created timestamp', () => {
    const recordWithUpdate: AlertRecordDTO = {
      ...mockRecord,
      updated_at: '2024-01-02T12:00:00Z',
    }

    render(
      <AlertRecordDetailModal
        record={recordWithUpdate}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.getByText('更新时间')).toBeInTheDocument()
  })

  it('does not display updated timestamp when same as created timestamp', () => {
    render(
      <AlertRecordDetailModal
        record={mockRecord}
        nodes={mockNodes}
        canEdit={true}
        onClose={vi.fn()}
        onStatusUpdate={vi.fn()}
      />
    )

    expect(screen.queryByText('更新时间')).not.toBeInTheDocument()
  })
})
