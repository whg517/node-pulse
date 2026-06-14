import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { AlertRulesTable } from '../AlertRulesTable'
import type { AlertRule } from '../../../stores/types'
import type { NodeDTO } from '../../../api/types'

describe('AlertRulesTable', () => {
  const mockNodes: NodeDTO[] = [
    { id: 'node-1', name: 'Node 1', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online', created_at: '2024-01-01', updated_at: '2024-01-01' },
    { id: 'node-2', name: 'Node 2', ip: '192.168.1.2', region: 'us-west', tags: [], status: 'online', created_at: '2024-01-01', updated_at: '2024-01-01' },
  ]

  const mockRules: AlertRule[] = [
    {
      id: 'rule-1',
      metric: 'latency',
      threshold: 100,
      level: 'P1',
      nodeId: 'node-1',
      enabled: true,
    },
    {
      id: 'rule-2',
      metric: 'packet_loss_rate',
      threshold: 5,
      level: 'P0',
      nodeId: null,
      enabled: false,
    },
  ]

  it('renders alert rules table', () => {
    render(
      <AlertRulesTable
        rules={mockRules}
        nodes={mockNodes}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    expect(screen.getByText('Latency')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
    expect(screen.getByText('P1')).toBeInTheDocument()
    expect(screen.getByText('Node 1')).toBeInTheDocument()
    expect(screen.getByText('Enabled')).toBeInTheDocument()
  })

  it('renders empty state when no rules', () => {
    render(
      <AlertRulesTable
        rules={[]}
        nodes={mockNodes}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    expect(screen.getByText('No alert rules')).toBeInTheDocument()
    expect(screen.getByText('Get started by creating a new alert rule.')).toBeInTheDocument()
  })

  it('hides action buttons when user cannot edit', () => {
    render(
      <AlertRulesTable
        rules={mockRules}
        nodes={mockNodes}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={false}
      />
    )

    expect(screen.queryByText('Edit')).not.toBeInTheDocument()
    expect(screen.queryByText('Delete')).not.toBeInTheDocument()
  })

  it('calls onEdit when edit button clicked', () => {
    const onEdit = vi.fn()
    render(
      <AlertRulesTable
        rules={mockRules}
        nodes={mockNodes}
        onEdit={onEdit}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    const editButtons = screen.getAllByText('Edit')
    fireEvent.click(editButtons[0])

    expect(onEdit).toHaveBeenCalledWith('rule-1')
  })

  it('calls onDelete when delete button clicked', () => {
    const onDelete = vi.fn()
    render(
      <AlertRulesTable
        rules={mockRules}
        nodes={mockNodes}
        onEdit={vi.fn()}
        onDelete={onDelete}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    expect(onDelete).toHaveBeenCalledWith('rule-1')
  })

  it('displays global rules correctly', () => {
    render(
      <AlertRulesTable
        rules={mockRules}
        nodes={mockNodes}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    // Rule 2 is global (nodeId is null)
    expect(screen.getByText('Global')).toBeInTheDocument()
  })
})
