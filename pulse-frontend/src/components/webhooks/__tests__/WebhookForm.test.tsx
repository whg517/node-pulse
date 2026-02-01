import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { WebhookForm } from '../WebhookForm'
import type { Webhook } from '../../../stores/webhooksStore'

describe('WebhookForm', () => {
  it('renders form fields', () => {
    render(
      <WebhookForm
        mode="create"
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.getByLabelText(/Webhook URL/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Event Format \(JSON\)/i)).toBeInTheDocument()
    expect(screen.getByLabelText('Enabled')).toBeInTheDocument()
  })

  it('validates HTTPS URL - rejects HTTP', async () => {
    const onSubmit = vi.fn()
    render(
      <WebhookForm
        mode="create"
        onSubmit={onSubmit}
        onCancel={vi.fn()}
      />
    )

    const urlInput = screen.getByLabelText(/Webhook URL/i)
    fireEvent.change(urlInput, { target: { value: 'http://example.com/webhook' } })

    const submitButton = screen.getByText('Add Webhook')
    fireEvent.click(submitButton)

    expect(await screen.findByText(/URL must use HTTPS protocol/i)).toBeInTheDocument()
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('validates JSON format - rejects invalid JSON', async () => {
    const onSubmit = vi.fn()
    render(
      <WebhookForm
        mode="create"
        onSubmit={onSubmit}
        onCancel={vi.fn()}
      />
    )

    const eventFormatTextarea = screen.getByLabelText(/Event Format \(JSON\)/i)
    fireEvent.change(eventFormatTextarea, { target: { value: '{ invalid json' } })

    const submitButton = screen.getByText('Add Webhook')
    fireEvent.click(submitButton)

    expect(await screen.findByText(/Invalid JSON format/i)).toBeInTheDocument()
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('submits form with valid HTTPS URL and JSON', async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined)
    render(
      <WebhookForm
        mode="create"
        onSubmit={onSubmit}
        onCancel={vi.fn()}
      />
    )

    const urlInput = screen.getByLabelText(/Webhook URL/i)
    fireEvent.change(urlInput, { target: { value: 'https://example.com/webhook' } })

    const submitButton = screen.getByText('Add Webhook')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        url: 'https://example.com/webhook',
        event_format: expect.any(Object),
        enabled: true,
      })
    })
  })

  it('pre-fills form with initial data in edit mode', () => {
    const initialWebhook: Webhook = {
      id: 'webhook-1',
      url: 'https://example.com/webhook',
      eventFormat: { version: '2.0' },
      enabled: false,
    }

    render(
      <WebhookForm
        mode="edit"
        initialData={initialWebhook}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    const urlInput = screen.getByLabelText(/Webhook URL/i) as HTMLInputElement
    const enabledCheckbox = screen.getByLabelText('Enabled') as HTMLInputElement

    expect(urlInput.value).toBe('https://example.com/webhook')
    expect(enabledCheckbox.checked).toBe(false)
  })

  it('calls onCancel when cancel button clicked', () => {
    const onCancel = vi.fn()
    render(
      <WebhookForm
        mode="create"
        onSubmit={vi.fn()}
        onCancel={onCancel}
      />
    )

    const cancelButton = screen.getByText('Cancel')
    fireEvent.click(cancelButton)

    expect(onCancel).toHaveBeenCalled()
  })

  it('resets to default event format when reset button clicked', () => {
    render(
      <WebhookForm
        mode="create"
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    const resetButton = screen.getByText('Reset to Default')
    fireEvent.click(resetButton)

    const eventFormatTextarea = screen.getByLabelText(/Event Format \(JSON\)/i)
    const value = eventFormatTextarea.value
    expect(value).toContain('"version"')
    expect(value).toContain('"1.0"')
  })
})
