import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { ToastNotification, type ToastProps } from './ToastNotification'

describe('ToastNotification', () => {
  it('renders success toast with title and message', () => {
    const props: ToastProps = {
      id: '1',
      type: 'success',
      title: 'Success',
      message: 'Operation completed',
      onClose: vi.fn(),
    }

    render(<ToastNotification {...props} />)

    expect(screen.getByText('Success')).toBeInTheDocument()
    expect(screen.getByText('Operation completed')).toBeInTheDocument()
  })

  it('renders error toast', () => {
    const props: ToastProps = {
      id: '1',
      type: 'error',
      title: 'Error',
      onClose: vi.fn(),
    }

    render(<ToastNotification {...props} />)

    expect(screen.getByText('Error')).toBeInTheDocument()
  })

  it('renders warning toast', () => {
    const props: ToastProps = {
      id: '1',
      type: 'warning',
      title: 'Warning',
      onClose: vi.fn(),
    }

    render(<ToastNotification {...props} />)

    expect(screen.getByText('Warning')).toBeInTheDocument()
  })

  it('renders info toast', () => {
    const props: ToastProps = {
      id: '1',
      type: 'info',
      title: 'Info',
      onClose: vi.fn(),
    }

    render(<ToastNotification {...props} />)

    expect(screen.getByText('Info')).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', () => {
    const handleClose = vi.fn()
    const props: ToastProps = {
      id: '1',
      type: 'success',
      title: 'Success',
      onClose: handleClose,
    }

    render(<ToastNotification {...props} />)

    const closeButton = screen.getByLabelText('Close notification')
    fireEvent.click(closeButton)

    expect(handleClose).toHaveBeenCalledWith('1')
  })

  it('auto-dismisses after 3 seconds', async () => {
    const handleClose = vi.fn()
    const props: ToastProps = {
      id: '1',
      type: 'success',
      title: 'Success',
      onClose: handleClose,
    }

    render(<ToastNotification {...props} />)

    await waitFor(
      () => {
        expect(handleClose).toHaveBeenCalledWith('1')
      },
      { timeout: 4000 }
    )
  })
})
