import { useEffect } from 'react'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface ToastProps {
  id: string
  type: ToastType
  title: string
  message?: string
  onClose: (id: string) => void
}

const toastStyles = {
  success: 'bg-healthy-bg text-healthy-text border-healthy',
  error: 'bg-destructive/10 text-destructive border-destructive',
  warning: 'bg-warning-bg text-warning-text border-warning',
  info: 'bg-primary/10 text-primary border-primary',
}

export function ToastNotification({ id, type, title, message, onClose }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose(id)
    }, 3000)

    return () => clearTimeout(timer)
  }, [id, onClose])

  return (
    <div
      className={`fixed top-4 right-4 z-50 p-4 rounded-lg border shadow-lg transition-all duration-300 ease-in-out ${toastStyles[type]}`}
      role="alert"
      aria-live="assertive"
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h3 className="font-semibold text-sm">{title}</h3>
          {message && <p className="text-sm mt-1">{message}</p>}
        </div>
        <button
          onClick={() => onClose(id)}
          className="ml-4 text-current opacity-60 hover:opacity-100"
          aria-label="Close notification"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>
    </div>
  )
}
