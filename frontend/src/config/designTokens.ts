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
 * Primary brand colors (Teal — network/signal identity)
 */
export const primaryColors = {
  light: '#2DD4BF',   // teal-400
  main: '#0F766E',    // teal-700 (WCAG AA on white: 5.47:1)
  dark: '#115E59',    // teal-800
  darker: '#134E4A',  // teal-900
} as const

/**
 * Status/semantic colors
 */
export const statusColors = {
  healthy: {
    light: '#34D399',  // emerald-400
    main: '#059669',   // emerald-600
    dark: '#065F46',   // emerald-800
  },
  warning: {
    light: '#FBBF24',  // amber-400
    main: '#D97706',   // amber-600
    dark: '#92400E',   // amber-800
  },
  critical: {
    light: '#F87171',  // red-400
    main: '#DC2626',   // red-600
    dark: '#991B1B',   // red-800
  },
  unknown: {
    light: '#94A3B8',  // slate-400
    main: '#64748B',   // slate-500
    dark: '#475569',   // slate-600
  },
} as const

/**
 * Neutral colors (slate-based — matches CSS variable palette)
 */
export const neutralColors = {
  white: '#FFFFFF',
  gray: {
    50: '#F8FAFB',
    100: '#F1F5F9',
    200: '#E2E8F0',
    300: '#CBD5E1',
    400: '#94A3B8',
    500: '#64748B',
    600: '#475569',
    700: '#334155',
    800: '#1E293B',
    900: '#0F172A',
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
 * Uses CSS variables so dark mode is handled by :root/.dark, not dark: prefixes.
 */
export const statusClasses = {
  healthy: {
    bg: 'bg-[var(--color-healthy)]',
    bgLight: 'bg-[var(--color-healthy-bg)]',
    bgLightDark: '',  // CSS vars handle dark — no dark: prefix needed
    text: 'text-[var(--color-healthy)]',
    textDark: 'text-[var(--color-healthy-text)]',
    border: 'border-[var(--color-healthy-bg)]',
    ring: 'ring-[var(--color-healthy)]',
  },
  warning: {
    bg: 'bg-[var(--color-warning)]',
    bgLight: 'bg-[var(--color-warning-bg)]',
    bgLightDark: '',
    text: 'text-[var(--color-warning)]',
    textDark: 'text-[var(--color-warning-text)]',
    border: 'border-[var(--color-warning-bg)]',
    ring: 'ring-[var(--color-warning)]',
  },
  critical: {
    bg: 'bg-[var(--color-critical)]',
    bgLight: 'bg-[var(--color-critical-bg)]',
    bgLightDark: '',
    text: 'text-[var(--color-critical)]',
    textDark: 'text-[var(--color-critical-text)]',
    border: 'border-[var(--color-critical-bg)]',
    ring: 'ring-[var(--color-critical)]',
  },
  unknown: {
    bg: 'bg-[var(--color-unknown)]',
    bgLight: 'bg-[var(--color-unknown-bg)]',
    bgLightDark: '',
    text: 'text-[var(--color-unknown)]',
    textDark: 'text-[var(--color-unknown-text)]',
    border: 'border-[var(--color-unknown-bg)]',
    ring: 'ring-[var(--color-unknown)]',
  },
} as const

/**
 * Primary color classes for Tailwind CSS
 * Uses CSS variables so dark mode is handled by :root/.dark, not dark: prefixes.
 */
export const primaryClasses = {
  bg: 'bg-[var(--color-brand)]',
  bgHover: 'hover:bg-[var(--color-brand-hover)]',
  bgLight: 'bg-[var(--color-brand-muted)]',
  bgLightDark: 'bg-[var(--color-brand-muted)]',  // CSS vars handle dark
  text: 'text-[var(--color-brand)]',
  textHover: 'hover:text-[var(--color-brand-hover)]',
  border: 'border-[var(--color-brand-muted)]',
  ring: 'ring-[var(--color-brand)]',
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
 * Uses semantic CSS variables from index.css so dark mode is handled by CSS, not JS.
 */
export const buttonVariants = {
  // Primary action button
  primary: 'inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] rounded-md transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[var(--color-brand)] disabled:opacity-50 disabled:cursor-not-allowed',

  // Secondary action button
  secondary: 'inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] bg-[var(--color-bg-surface)] border border-[var(--color-border-strong)] rounded-md hover:bg-[var(--color-hover-overlay)] transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[var(--color-brand)] disabled:opacity-50 disabled:cursor-not-allowed',

  // Danger action button
  danger: 'inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-[var(--color-critical)] hover:bg-[var(--color-critical)] hover:opacity-90 rounded-md transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[var(--color-critical)] disabled:opacity-50 disabled:cursor-not-allowed',

  // Ghost button
  ghost: 'inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] rounded-md hover:bg-[var(--color-hover-overlay)] transition-colors duration-150 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed',

  // Icon button
  icon: 'inline-flex items-center justify-center p-2 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] rounded-lg hover:bg-[var(--color-hover-overlay)] transition-colors duration-150 focus:outline-none',
} as const

/**
 * Card variants
 * Uses semantic CSS variables so gray vs slate inconsistency is resolved in one place.
 */
export const cardVariants = {
  // Standard card
  default: 'rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)] shadow-sm',

  // Elevated card
  elevated: 'rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)] shadow-md',

  // Interactive card
  interactive: 'rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)] shadow-sm hover:shadow-md transition-shadow duration-200 cursor-pointer',
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
