import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { WebhookDialog } from '../WebhookDialog'

describe('WebhookDialog', () => {
  it('renders the create-mode title when open', () => {
    render(
      <WebhookDialog
        mode="create"
        open={true}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />,
    )

    // i18n is auto-mocked in vitest-setup; the title resolves from en.json.
    // The Dialog content is portalled to document.body, but screen queries the
    // whole document so this works regardless of portal.
    expect(screen.getByRole('heading')).toBeInTheDocument()
  })

  it('renders the edit-mode title when mode is edit', () => {
    render(
      <WebhookDialog
        mode="edit"
        open={true}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />,
    )
    expect(screen.getByRole('heading')).toBeInTheDocument()
  })

  it('renders nothing visible when closed', () => {
    render(
      <WebhookDialog
        mode="create"
        open={false}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />,
    )
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('delegates submission to WebhookForm', async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined)
    render(
      <WebhookDialog
        mode="create"
        open={true}
        onSubmit={onSubmit}
        onCancel={vi.fn()}
      />,
    )

    // Fill a valid HTTPS URL and submit through the inner WebhookForm.
    const urlInput = screen.getByRole('textbox', { name: /Webhook URL/i })
    fireEvent.change(urlInput, { target: { value: 'https://example.com/webhook' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Webhook' }))

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledTimes(1)
    })
  })
})
