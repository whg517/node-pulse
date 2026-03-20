/**
 * ConfirmDialog Component
 *
 * Standardized confirmation dialog for destructive actions.
 * Consistent styling across all delete/confirm operations.
 */

import { useEffect } from 'react'
import { buttonVariants } from '../../config/designTokens'

export interface ConfirmDialogProps {
  /** Whether dialog is open */
  open: boolean
  /** Dialog title */
  title: string
  /** Confirmation message */
  message: string | React.ReactNode
  /** Confirm button text */
  confirmText?: string
  /** Cancel button text */
  cancelText?: string
  /** Callback when confirm is clicked */
  onConfirm: () => void
  /** Callback when cancel is clicked or dialog is closed */
  onCancel: () => void
  /** Whether the confirm action is loading */
  loading?: boolean
  /** Danger variant (red confirm button) */
  variant?: 'default' | 'danger'
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmText = 'common.confirm',
  cancelText = 'common.cancel',
  onConfirm,
  onCancel,
  loading = false,
  variant = 'danger',
}: ConfirmDialogProps) {

  // Close on Escape key
  useEffect(() => {
    if (!open) return

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onCancel()
      }
    }

    document.addEventListener('keydown', handleEscape)
    return () => document.removeEventListener('keydown', handleEscape)
  }, [open, onCancel])

  if (!open) return null

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 z-40 bg-black/50 transition-opacity"
        onClick={onCancel}
        aria-hidden="true"
      />

      {/* Dialog */}
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div
          className="w-full max-w-md transform overflow-hidden rounded-lg bg-[var(--color-bg-elevated)] p-6 shadow-xl transition-all"
          role="dialog"
          aria-modal="true"
          aria-labelledby="dialog-title"
        >
          {/* Icon and title */}
          <div className="flex items-center gap-3">
            <div
              className={`flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full ${
                variant === 'danger'
                  ? 'bg-[var(--color-critical-bg)]'
                  : 'bg-[var(--color-brand-muted)]'
              }`}
            >
              {variant === 'danger' ? (
                <svg
                  className="h-6 w-6 text-[var(--color-critical)]"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                  />
                </svg>
              ) : (
                <svg
                  className="h-6 w-6 text-[var(--color-brand)]"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              )}
            </div>
            <div>
              <h3
                id="dialog-title"
                className="text-lg font-semibold text-[var(--color-text-primary)]"
              >
                {title}
              </h3>
            </div>
          </div>

          {/* Message */}
          <div className="mt-4">
            <p className="text-sm text-[var(--color-text-secondary)]">
              {message}
            </p>
          </div>

          {/* Actions */}
          <div className="mt-6 flex justify-end gap-3">
            <button
              type="button"
              onClick={onCancel}
              disabled={loading}
              className={`${buttonVariants.secondary} disabled:opacity-50 disabled:cursor-not-allowed`}
            >
              {cancelText}
            </button>
            <button
              type="button"
              onClick={onConfirm}
              disabled={loading}
              className={`${
                variant === 'danger' ? buttonVariants.danger : buttonVariants.primary
              } disabled:opacity-50 disabled:cursor-not-allowed`}
            >
              {loading ? 'common.loading...' : confirmText}
            </button>
          </div>
        </div>
      </div>
    </>
  )
}

export default ConfirmDialog
