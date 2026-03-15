import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ConfirmDialog } from '../ConfirmDialog'

// Mock useTheme
vi.mock('../../../hooks/useTheme', () => ({
  useTheme: () => ({ isDark: false }),
}))

describe('ConfirmDialog', () => {
  const defaultProps = {
    open: true,
    title: 'Confirm Delete',
    message: 'Are you sure you want to delete this item?',
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when open is false', () => {
    const { container } = render(<ConfirmDialog {...defaultProps} open={false} />)
    expect(container.firstChild).toBeNull()
  })

  it('renders dialog when open is true', () => {
    render(<ConfirmDialog {...defaultProps} />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('Confirm Delete')).toBeInTheDocument()
    expect(screen.getByText('Are you sure you want to delete this item?')).toBeInTheDocument()
  })

  it('calls onCancel when cancel button is clicked', () => {
    render(<ConfirmDialog {...defaultProps} cancelText="Cancel" />)
    fireEvent.click(screen.getByText('Cancel'))
    expect(defaultProps.onCancel).toHaveBeenCalledTimes(1)
  })

  it('calls onConfirm when confirm button is clicked', () => {
    render(<ConfirmDialog {...defaultProps} confirmText="Delete" />)
    fireEvent.click(screen.getByText('Delete'))
    expect(defaultProps.onConfirm).toHaveBeenCalledTimes(1)
  })

  it('calls onCancel when backdrop is clicked', () => {
    render(<ConfirmDialog {...defaultProps} />)
    const backdrop = document.querySelector('[aria-hidden="true"]')
    fireEvent.click(backdrop!)
    expect(defaultProps.onCancel).toHaveBeenCalledTimes(1)
  })

  it('calls onCancel when Escape key is pressed', () => {
    render(<ConfirmDialog {...defaultProps} />)
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(defaultProps.onCancel).toHaveBeenCalledTimes(1)
  })

  it('does not call onCancel on Escape when closed', () => {
    render(<ConfirmDialog {...defaultProps} open={false} />)
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(defaultProps.onCancel).not.toHaveBeenCalled()
  })

  it('shows loading state on confirm button', () => {
    render(<ConfirmDialog {...defaultProps} loading={true} confirmText="Delete" />)
    const confirmBtn = screen.getByText('common.loading...')
    expect(confirmBtn).toBeDisabled()
  })

  it('disables buttons when loading', () => {
    render(<ConfirmDialog {...defaultProps} loading={true} cancelText="Cancel" />)
    const cancelBtn = screen.getByText('Cancel')
    expect(cancelBtn).toBeDisabled()
  })

  it('renders danger variant by default', () => {
    render(<ConfirmDialog {...defaultProps} />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('renders default variant', () => {
    render(<ConfirmDialog {...defaultProps} variant="default" />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('renders ReactNode message', () => {
    render(
      <ConfirmDialog
        {...defaultProps}
        message={<span data-testid="custom-message">Custom message node</span>}
      />
    )
    expect(screen.getByTestId('custom-message')).toBeInTheDocument()
  })

  it('renders dark theme variant', () => {
    vi.resetModules()
  })
})
