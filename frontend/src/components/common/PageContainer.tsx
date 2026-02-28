/**
 * PageContainer Component
 *
 * Standardized page wrapper for consistent layout across the application.
 * Handles max-width, padding, background color, and dark mode.
 */

import type { ReactNode } from 'react'
import { useTheme } from '../../hooks/useTheme'
import { layout, spacing } from '../../config/designTokens'

export interface PageContainerProps {
  /** Page content */
  children: ReactNode
  /** Optional custom className for additional styling */
  className?: string
  /** Custom background color (default: gray-50 / gray-950 for dark mode) */
  background?: 'default' | 'white' | 'transparent'
}

export function PageContainer({
  children,
  className = '',
  background = 'default',
}: PageContainerProps) {
  const { isDark } = useTheme()

  // Background classes based on variant
  const backgroundClasses = {
    default: isDark ? 'bg-gray-950' : 'bg-gray-50',
    white: isDark ? 'bg-gray-900' : 'bg-white',
    transparent: 'bg-transparent',
  }

  return (
    <div className={`min-h-screen ${backgroundClasses[background]} ${className}`}>
      <main className={`${layout.maxWidth} mx-auto ${spacing.pageContentPadding}`}>
        {children}
      </main>
    </div>
  )
}

export default PageContainer
