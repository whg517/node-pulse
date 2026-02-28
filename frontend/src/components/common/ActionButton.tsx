/**
 * ActionButton Component
 *
 * Standardized action button with consistent styling.
 * Supports primary, secondary, danger, and ghost variants.
 */

import type { ReactNode } from 'react'
import { buttonVariants } from '../../config/designTokens'

export interface ActionButtonProps {
  /** Button content */
  children: ReactNode
  /** Button variant */
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
  /** Button size */
  size?: 'sm' | 'md' | 'lg'
  /** Whether button is disabled */
  disabled?: boolean
  /** Whether button is in loading state */
  loading?: boolean
  /** Click handler */
  onClick?: () => void
  /** Type attribute */
  type?: 'button' | 'submit' | 'reset'
  /** Additional CSS classes */
  className?: string
  /** Left icon */
  leftIcon?: ReactNode
  /** Right icon */
  rightIcon?: ReactNode
}

export function ActionButton({
  children,
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
  onClick,
  type = 'button',
  className = '',
  leftIcon,
  rightIcon,
}: ActionButtonProps) {
  const baseStyles = buttonVariants[variant]
  
  // Size adjustments
  const sizeStyles = {
    sm: 'text-xs px-3 py-1.5',
    md: 'text-sm px-4 py-2',
    lg: 'text-base px-6 py-3',
  }

  const combinedStyles = `${baseStyles} ${sizeStyles[size]} ${className}`

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled || loading}
      className={combinedStyles}
    >
      {/* Left icon */}
      {leftIcon && (
        <span className={`${children ? '-ml-1 mr-2' : ''}`}>
          {leftIcon}
        </span>
      )}
      
      {/* Button content or loading spinner */}
      {loading ? (
        <svg
          className="mx-auto h-4 w-4 animate-spin"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      ) : (
        children
      )}
      
      {/* Right icon */}
      {rightIcon && !loading && (
        <span className={`${children ? 'ml-2 -mr-1' : ''}`}>
          {rightIcon}
        </span>
      )}
    </button>
  )
}

export default ActionButton
