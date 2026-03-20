/**
 * Accessibility Utilities
 * 
 * Provides accessibility helpers and constants for WCAG 2.1 AA compliance.
 * 
 * Features:
 * - Focus management
 * - Keyboard navigation
 * - Screen reader announcements
 * - ARIA attributes helpers
 * - Color contrast checking
 * - Skip links
 * 
 * @packageDocumentation
 */

// ============== Constants ==============

/**
 * Minimum contrast ratio for WCAG 2.1 AA
 * - Normal text: 4.5:1
 * - Large text: 3:1
 * - UI components: 3:1
 */
export const CONTRAST_RATIOS = {
  NORMAL_TEXT: 4.5,
  LARGE_TEXT: 3,
  UI_COMPONENTS: 3,
}

/**
 * Minimum touch target size in pixels
 * Recommended by WCAG and Apple HIG
 */
export const MIN_TOUCH_TARGET_SIZE = 44

/**
 * Focus ring styles for accessibility
 */
export const FOCUS_RING_STYLES = {
  default: 'focus:outline-none focus:ring-2 focus:ring-[var(--color-brand)] focus:ring-offset-2',
  dark: 'focus:ring-offset-slate-800',
}

// ============== Type Definitions ==============

export interface Color {
  r: number
  g: number
  b: number
}

export interface ContrastResult {
  ratio: number
  passesAA: boolean
  passesAAA: boolean
  passesLargeAA: boolean
  passesLargeAAA: boolean
}

// ============== Color Utilities ==============

/**
 * Parse hex color to RGB
 */
export function hexToRgb(hex: string): Color | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : null
}

/**
 * Calculate relative luminance of a color
 * @see https://www.w3.org/WAI/GL/wiki/Relative_luminance
 */
export function getLuminance(color: Color): number {
  const [rs, gs, bs] = [color.r / 255, color.g / 255, color.b / 255]

  const r = rs <= 0.03928 ? rs / 12.92 : Math.pow((rs + 0.055) / 1.055, 2.4)
  const g = gs <= 0.03928 ? gs / 12.92 : Math.pow((gs + 0.055) / 1.055, 2.4)
  const b = bs <= 0.03928 ? bs / 12.92 : Math.pow((bs + 0.055) / 1.055, 2.4)

  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

/**
 * Calculate contrast ratio between two colors
 * @see https://www.w3.org/TR/WCAG20-TECHS/G17.html
 */
export function getContrastRatio(color1: Color, color2: Color): number {
  const lum1 = getLuminance(color1)
  const lum2 = getLuminance(color2)
  const brightest = Math.max(lum1, lum2)
  const darkest = Math.min(lum1, lum2)
  return (brightest + 0.05) / (darkest + 0.05)
}

/**
 * Check if color contrast passes WCAG 2.1 AA
 */
export function checkContrast(hex1: string, hex2: string): ContrastResult {
  const color1 = hexToRgb(hex1)
  const color2 = hexToRgb(hex2)

  if (!color1 || !color2) {
    return {
      ratio: 0,
      passesAA: false,
      passesAAA: false,
      passesLargeAA: false,
      passesLargeAAA: false,
    }
  }

  const ratio = getContrastRatio(color1, color2)

  return {
    ratio,
    passesAA: ratio >= CONTRAST_RATIOS.NORMAL_TEXT,
    passesAAA: ratio >= 7, // AAA level for normal text
    passesLargeAA: ratio >= CONTRAST_RATIOS.LARGE_TEXT,
    passesLargeAAA: ratio >= 4.5, // AAA level for large text
  }
}

// ============== Focus Management ==============

/**
 * Get all focusable elements in a container
 */
export function getFocusableElements(container: HTMLElement): HTMLElement[] {
  const focusableSelectors = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
    'audio[controls]',
    'video[controls]',
    '[contenteditable]:not([contenteditable="false"])',
    'details>summary:first-of-type',
    'details',
  ]

  return Array.from(
    container.querySelectorAll<HTMLElement>(focusableSelectors.join(','))
  )
}

/**
 * Trap focus within a container (for modals/dialogs)
 */
export function trapFocus(container: HTMLElement): () => void {
  const focusableElements = getFocusableElements(container)
  const firstElement = focusableElements[0]
  const lastElement = focusableElements[focusableElements.length - 1]

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key !== 'Tab') return

    if (e.shiftKey) {
      if (document.activeElement === firstElement) {
        e.preventDefault()
        lastElement.focus()
      }
    } else {
      if (document.activeElement === lastElement) {
        e.preventDefault()
        firstElement.focus()
      }
    }
  }

  container.addEventListener('keydown', handleKeyDown)

  // Focus first element
  firstElement?.focus()

  // Return cleanup function
  return () => {
    container.removeEventListener('keydown', handleKeyDown)
  }
}

/**
 * Restore focus to previously focused element
 */
let previouslyFocusedElement: HTMLElement | null = null

export function saveFocus(): void {
  previouslyFocusedElement = document.activeElement as HTMLElement
}

export function restoreFocus(): void {
  previouslyFocusedElement?.focus()
  previouslyFocusedElement = null
}

// ============== Screen Reader Utilities ==============

/**
 * Announce message to screen readers
 * Uses aria-live region
 */
export function announceToScreenReader(message: string, priority: 'polite' | 'assertive' = 'polite'): void {
  const announcement = document.createElement('div')
  announcement.setAttribute('role', 'status')
  announcement.setAttribute('aria-live', priority)
  announcement.setAttribute('aria-atomic', 'true')
  announcement.className = 'sr-only'
  announcement.textContent = message
  document.body.appendChild(announcement)

  // Remove after announcement
  setTimeout(() => {
    document.body.removeChild(announcement)
  }, 1000)
}

// ============== Keyboard Navigation ==============

/**
 * Check if key press should activate element
 */
export function isActivationKey(key: string): boolean {
  return key === 'Enter' || key === ' '
}

/**
 * Handle keyboard interaction for custom button
 */
export function handleButtonKeyDown(
  e: React.KeyboardEvent,
  onClick: () => void
): void {
  if (isActivationKey(e.key)) {
    e.preventDefault()
    onClick()
  }
}

/**
 * Handle keyboard interaction for custom link
 */
export function handleLinkKeyDown(
  e: React.KeyboardEvent,
  href: string
): void {
  if (e.key === 'Enter') {
    e.preventDefault()
    window.location.href = href
  }
}

// ============== ARIA Helpers ==============

/**
 * Generate unique ID for aria-labelledby/aria-describedby
 */
let ariaIdCounter = 0
export function generateAriaId(prefix: string): string {
  return `${prefix}-${++ariaIdCounter}`
}

/**
 * Get aria-label for icon-only button
 */
export function getIconOnlyLabel(action: string): string {
  return `${action} (button)`
}

/**
 * Get aria-label for status indicator
 */
export function getStatusLabel(status: string, isActive: boolean): string {
  return `${status} (${isActive ? 'active' : 'inactive'})`
}

// ============== React Hooks ==============

import { useEffect, useCallback } from 'react'

/**
 * Hook for managing focus trap in modals
 */
export function useFocusTrap(containerRef: React.RefObject<HTMLElement>): void {
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const cleanup = trapFocus(container)
    return cleanup
  }, [containerRef])
}

/**
 * Hook for keyboard shortcuts
 */
export function useKeyboardShortcut(
  key: string,
  handler: (e: KeyboardEvent) => void,
  options: { ctrl?: boolean; shift?: boolean; alt?: boolean } = {}
): void {
  const handleKey = useCallback(
    (e: KeyboardEvent) => {
      if (
        e.key.toLowerCase() === key.toLowerCase() &&
        (!!options.ctrl === e.ctrlKey || e.metaKey) &&
        (!!options.shift === e.shiftKey) &&
        (!!options.alt === e.altKey)
      ) {
        handler(e)
      }
    },
    [key, handler, options.ctrl, options.shift, options.alt]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [handleKey])
}

// ============== Accessibility Classes ==============

/**
 * Screen reader only (visually hidden but accessible)
 */
export const SR_ONLY = 'sr-only'

/**
 * Skip link classes
 */
export const SKIP_LINK_CLASSES = `
  absolute left-4 top-0 z-50
  px-4 py-2 bg-[var(--color-brand)] text-white
  transform -translate-y-full
  focus:translate-y-0
  transition-transform
`

// Export all utilities
export const AccessibilityUtils = {
  // Constants
  CONTRAST_RATIOS,
  MIN_TOUCH_TARGET_SIZE,
  FOCUS_RING_STYLES,

  // Color
  hexToRgb,
  getLuminance,
  getContrastRatio,
  checkContrast,

  // Focus
  getFocusableElements,
  trapFocus,
  saveFocus,
  restoreFocus,

  // Screen Reader
  announceToScreenReader,

  // Keyboard
  isActivationKey,
  handleButtonKeyDown,
  handleLinkKeyDown,

  // ARIA
  generateAriaId,
  getIconOnlyLabel,
  getStatusLabel,

  // Hooks
  useFocusTrap,
  useKeyboardShortcut,

  // Classes
  SR_ONLY,
  SKIP_LINK_CLASSES,
}
