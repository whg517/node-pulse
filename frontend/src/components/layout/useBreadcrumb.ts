/**
 * Breadcrumb hooks and context creation.
 *
 * Separated from BreadcrumbProvider so that BreadcrumbContext.tsx
 * only exports components (react-refresh compliance).
 */

import { createContext, useContext } from 'react'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface BreadcrumbItem {
  path: string
  label: string
}

export interface BreadcrumbContextValue {
  items: BreadcrumbItem[]
  setDynamicLabel: (offset: number, label: string) => void
  clearDynamicLabels: () => void
}

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

export const BreadcrumbContext = createContext<BreadcrumbContextValue | null>(null)

// ---------------------------------------------------------------------------
// Hooks
// ---------------------------------------------------------------------------

export function useBreadcrumb(): BreadcrumbContextValue {
  const ctx = useContext(BreadcrumbContext)
  if (!ctx) {
    throw new Error('useBreadcrumb must be used within a <BreadcrumbProvider>')
  }
  return ctx
}

/** Convenience hook for pages that only need to set dynamic labels. */
export function useSetBreadcrumbLabel() {
  const { setDynamicLabel, clearDynamicLabels } = useBreadcrumb()
  return { setDynamicLabel, clearDynamicLabels }
}
