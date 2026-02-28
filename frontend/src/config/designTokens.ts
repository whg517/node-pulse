/**
 * Design Tokens
 *
 * Centralized design system tokens for consistent UI across the application.
 * Includes colors, spacing, typography, and layout constants.
 */

// ============================================================================
// Color Palette
// ============================================================================

/**
 * Primary brand colors
 */
export const primaryColors = {
  light: '#3B82F6',   // blue-500
  main: '#2563EB',    // blue-600
  dark: '#1D4ED8',    // blue-700
  darker: '#1E40AF',  // blue-800
} as const

/**
 * Status/semantic colors
 */
export const statusColors = {
  healthy: {
    light: '#86EFAC',  // green-400
    main: '#22C55E',   // green-500
    dark: '#16A34A',   // green-600
  },
  warning: {
    light: '#FCD34D',  // amber-400
    main: '#F59E0B',   // amber-500
    dark: '#D97706',   // amber-600
  },
  critical: {
    light: '#F87171',  // red-400
    main: '#EF4444',   // red-500
    dark: '#DC2626',   // red-600
  },
  unknown: {
    light: '#9CA3AF',  // gray-400
    main: '#6B7280',   // gray-500
    dark: '#4B5563',   // gray-600
  },
} as const

/**
 * Neutral colors (grayscale)
 */
export const neutralColors = {
  white: '#FFFFFF',
  gray: {
    50: '#F9FAFB',
    100: '#F3F4F6',
    200: '#E5E7EB',
    300: '#D1D5DB',
    400: '#9CA3AF',
    500: '#6B7280',
    600: '#4B5563',
    700: '#374151',
    800: '#1F2937',
    900: '#111827',
  },
  slate: {
    50: '#F8FAFC',
    100: '#F1F5F9',
    200: '#E2E8F0',
    300: '#CBD5E1',
    400: '#94A3B8',
    500: '#64748B',
    600: '#475569',
    700: '#334155',
    800: '#1E293B',
    900: '#0F172A',
    950: '#020617',
  },
} as const

// ============================================================================
// Tailwind CSS Class Constants
// ============================================================================

/**
 * Status color classes for Tailwind CSS
 */
export const statusClasses = {
  healthy: {
    bg: 'bg-green-500',
    bgLight: 'bg-green-100',
    bgLightDark: 'dark:bg-green-900/20',
    text: 'text-green-500',
    textDark: 'text-green-800',
    border: 'border-green-200',
    ring: 'ring-green-500',
  },
  warning: {
    bg: 'bg-amber-500',
    bgLight: 'bg-amber-100',
    bgLightDark: 'dark:bg-amber-900/20',
    text: 'text-amber-500',
    textDark: 'text-amber-800',
    border: 'border-amber-200',
    ring: 'ring-amber-500',
  },
  critical: {
    bg: 'bg-red-500',
    bgLight: 'bg-red-100',
    bgLightDark: 'dark:bg-red-900/20',
    text: 'text-red-500',
    textDark: 'text-red-800',
    border: 'border-red-200',
    ring: 'ring-red-500',
  },
  unknown: {
    bg: 'bg-gray-500',
    bgLight: 'bg-gray-100',
    bgLightDark: 'dark:bg-gray-900/20',
    text: 'text-gray-500',
    textDark: 'text-gray-800',
    border: 'border-gray-200',
    ring: 'ring-gray-500',
  },
} as const

/**
 * Primary color classes for Tailwind CSS
 */
export const primaryClasses = {
  bg: 'bg-blue-600',
  bgHover: 'hover:bg-blue-700',
  bgLight: 'bg-blue-50',
  bgLightDark: 'dark:bg-blue-900/20',
  text: 'text-blue-600',
  textHover: 'hover:text-blue-700',
  border: 'border-blue-200',
  ring: 'ring-blue-500',
} as const

// ============================================================================
// Spacing Tokens
// ============================================================================

/**
 * Page layout spacing
 */
export const spacing = {
  // Page containers
  pagePadding: 'p-4 md:p-6 lg:p-8',
  pageContentPadding: 'px-4 sm:px-6 lg:px-8 py-8',
  
  // Cards and sections
  cardPadding: 'p-4',
  cardPaddingLg: 'p-6',
  sectionGap: 'gap-6',
  sectionGapLg: 'gap-8',
  
  // Margins
  marginBottom: {
    sm: 'mb-4',
    md: 'mb-6',
    lg: 'mb-8',
  },
  
  // Gaps
  gap: {
    xs: 'gap-1',
    sm: 'gap-2',
    md: 'gap-3',
    lg: 'gap-4',
    xl: 'gap-6',
  },
} as const

// ============================================================================
// Layout Tokens
// ============================================================================

/**
 * Layout constraints
 */
export const layout = {
  maxWidth: 'max-w-7xl',
  maxWidthSm: 'max-w-5xl',
  maxWidthLg: 'max-w-9xl',
  
  // Sidebar dimensions
  sidebarWidth: 'w-64',
  sidebarCollapsedWidth: 'w-16',
  sidebarMobileBreakpoint: 'md', // 768px
  
  // Header dimensions
  headerHeight: 'h-16',
  
  // Z-index layers
  zIndex: {
    base: 0,
    dropdown: 10,
    sticky: 20,
    overlay: 30,
    modal: 40,
    popover: 50,
    tooltip: 60,
  },
} as const

// ============================================================================
// Typography Tokens
// ============================================================================

/**
 * Typography scale
 */
export const typography = {
  // Font sizes
  fontSize: {
    xs: 'text-xs',
    sm: 'text-sm',
    base: 'text-base',
    lg: 'text-lg',
    xl: 'text-xl',
    '2xl': 'text-2xl',
    '3xl': 'text-3xl',
    '4xl': 'text-4xl',
  },
  
  // Font weights
  fontWeight: {
    normal: 'font-normal',
    medium: 'font-medium',
    semibold: 'font-semibold',
    bold: 'font-bold',
  },
  
  // Line heights
  lineHeight: {
    tight: 'leading-tight',
    normal: 'leading-normal',
    relaxed: 'leading-relaxed',
  },
} as const

// ============================================================================
// Animation Tokens
// ============================================================================

/**
 * Animation and transition tokens
 */
export const animation = {
  // Transition duration
  duration: {
    fast: 'duration-150',
    normal: 'duration-200',
    slow: 'duration-300',
  },
  
  // Timing functions
  easing: {
    linear: 'ease-linear',
    in: 'ease-in',
    out: 'ease-out',
    inOut: 'ease-in-out',
  },
  
  // Common transitions
  transition: {
    colors: 'transition-colors',
    all: 'transition-all',
    transform: 'transition-transform',
  },
} as const

// ============================================================================
// Component Tokens
// ============================================================================

/**
 * Button variants
 */
export const buttonVariants = {
  // Primary action button
  primary: 'inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed',
  
  // Secondary action button
  secondary: 'inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed',
  
  // Danger action button
  danger: 'inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-md transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed',
  
  // Ghost button
  ghost: 'inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors duration-150 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed',
  
  // Icon button
  icon: 'inline-flex items-center justify-center p-2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors duration-150 focus:outline-none',
} as const

/**
 * Card variants
 */
export const cardVariants = {
  // Standard card
  default: 'rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm',
  
  // Elevated card
  elevated: 'rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-md',
  
  // Interactive card
  interactive: 'rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm hover:shadow-md transition-shadow duration-200 cursor-pointer',
} as const

// ============================================================================
// Export All
// ============================================================================

export const designTokens = {
  colors: {
    primary: primaryColors,
    status: statusColors,
    neutral: neutralColors,
  },
  classes: {
    status: statusClasses,
    primary: primaryClasses,
  },
  spacing,
  layout,
  typography,
  animation,
  buttonVariants,
  cardVariants,
} as const

export default designTokens
