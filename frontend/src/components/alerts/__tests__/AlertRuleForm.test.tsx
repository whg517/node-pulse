import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { AlertRuleForm } from '../AlertRuleForm'
import type { AlertRule } from '../../../stores/types'
import type { NodeDTO } from '../../../api/types'

describe('AlertRuleForm', () => {
  const mockNodes: NodeDTO[] = [
    { id: 'node-1', name: 'Node 1', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online', created_at: '2024-01-01', updated_at: '2024-01-01' },
    { id: 'node-2', name: 'Node 2', ip: '192.168.1.2', region: 'us-west', tags: [], status: 'online', created_at: '2024-01-01', updated_at: '2024-01-01' },
  ]

  it('renders form fields', () => {
    render(
      <AlertRuleForm
        mode="create"
        nodes={mockNodes}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.getByLabelText('alerts.alertType')).toBeInTheDocument()
    expect(screen.getByLabelText('alerts.threshold')).toBeInTheDocument()
    expect(screen.getByLabelText('alerts.severity')).toBeInTheDocument()
    expect(screen.getByLabelText('alerts.scope')).toBeInTheDocument()
    expect(screen.getByLabelText('status.enabled')).toBeInTheDocument()
  })

  it('pre-fills form with initial data in edit mode', () => {
    const initialData: AlertRule = {
      id: 'rule-1',
      metric: 'latency',
      threshold: 150,
      level: 'P0',
      nodeId: 'node-1',
      enabled: false,
    }

    render(
      <AlertRuleForm
        mode="edit"
        initialData={initialData}
        nodes={mockNodes}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    const metricSelect = screen.getByLabelText('alerts.alertType') as HTMLSelectElement
    const thresholdInput = screen.getByLabelText('alerts.threshold') as HTMLInputElement
    const levelSelect = screen.getByLabelText('alerts.severity') as HTMLSelectElement
    const enabledCheckbox = screen.getByLabelText('status.enabled') as HTMLInputElement

    expect(metricSelect.value).toBe('latency')
    expect(thresholdInput.value).toBe('150')
    expect(levelSelect.value).toBe('P0')
    expect(enabledCheckbox.checked).toBe(false)
  })

  it('validates threshold', async () => {
    const onSubmit = vi.fn()
    render(
      <AlertRuleForm
        mode="create"
        nodes={mockNodes}
        onSubmit={onSubmit}
        onCancel={vi.fn()}
      />
    )

    const submitButton = screen.getByText('alerts.createRule')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('alerts.errorThresholdPositive')).toBeInTheDocument()
    })
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('submits form with valid data', async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined)
    render(
      <AlertRuleForm
        mode="create"
        nodes={mockNodes}
        onSubmit={onSubmit}
        onCancel={vi.fn()}
      />
    )

    const thresholdInput = screen.getByLabelText('alerts.threshold')
    fireEvent.change(thresholdInput, { target: { value: '100' } })

    const submitButton = screen.getByText('alerts.createRule')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        metric: 'latency',
        threshold: 100,
        level: 'P1',
        node_id: null,
        enabled: true,
      })
    })
  })

  it('calls onCancel when cancel button clicked', () => {
    const onCancel = vi.fn()
    render(
      <AlertRuleForm
        mode="create"
        nodes={mockNodes}
        onSubmit={vi.fn()}
        onCancel={onCancel}
      />
    )

    const cancelButton = screen.getByText('common.cancel')
    fireEvent.click(cancelButton)

    expect(onCancel).toHaveBeenCalled()
  })
})
