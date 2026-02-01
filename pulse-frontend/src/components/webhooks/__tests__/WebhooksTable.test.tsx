import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { WebhooksTable } from '../WebhooksTable'
import type { Webhook } from '../../../stores/webhooksStore'

describe('WebhooksTable', () => {
  const mockWebhooks: Webhook[] = [
    {
      id: 'webhook-1',
      url: 'https://example.com/webhook',
      eventFormat: { version: '1.0' },
      enabled: true,
    },
    {
      id: 'webhook-2',
      url: 'https://hooks.slack.com/services/XXX/YYY',
      eventFormat: { version: '1.0', alert: {} },
      enabled: false,
    },
  ]

  it('renders webhooks table', () => {
    render(
      <WebhooksTable
        webhooks={mockWebhooks}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    expect(screen.getByText('https://example.com/webhook')).toBeInTheDocument()
    expect(screen.getByText('Enabled')).toBeInTheDocument()
    expect(screen.getByText('Disabled')).toBeInTheDocument()
  })

  it('renders empty state when no webhooks', () => {
    render(
      <WebhooksTable
        webhooks={[]}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    expect(screen.getByText('No webhooks configured')).toBeInTheDocument()
    expect(screen.getByText(/get started by adding a webhook/i)).toBeInTheDocument()
  })

  it('hides action buttons when user cannot edit', () => {
    render(
      <WebhooksTable
        webhooks={mockWebhooks}
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
      <WebhooksTable
        webhooks={mockWebhooks}
        onEdit={onEdit}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    const editButtons = screen.getAllByText('Edit')
    fireEvent.click(editButtons[0])

    expect(onEdit).toHaveBeenCalledWith('webhook-1')
  })

  it('calls onDelete when delete button clicked', () => {
    const onDelete = vi.fn()
    render(
      <WebhooksTable
        webhooks={mockWebhooks}
        onEdit={vi.fn()}
        onDelete={onDelete}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    expect(onDelete).toHaveBeenCalledWith('webhook-1')
  })

  it('displays event format field count', () => {
    render(
      <WebhooksTable
        webhooks={mockWebhooks}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onToggleEnabled={vi.fn()}
        canEdit={true}
      />
    )

    // webhook-1 has 1 field (version)
    // webhook-2 has 2 fields (version, alert)
    const fieldCounts = screen.getAllByText(/fields/)
    expect(fieldCounts.length).toBeGreaterThan(0)
  })
})
