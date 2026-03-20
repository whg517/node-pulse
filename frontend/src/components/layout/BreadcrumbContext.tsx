/**
 * Breadcrumb Context
 *
 * Provides route-config-driven breadcrumb items to the layout.
 * Uses React Router's useMatches() to read breadcrumb metadata
 * from each matched route's handle property.
 */

import { createContext, useContext, useState, useEffect, useMemo, type ReactNode } from 'react'
import { useMatches, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface BreadcrumbItem {
  path: string
  label: string
}

/** Shape of the route handle property for breadcrumb metadata. */
export interface BreadcrumbHandle {
  breadcrumb: string // i18n key
}

export interface BreadcrumbContextValue {
  items: BreadcrumbItem[]
  setDynamicLabel: (offset: number, label: string) => void
  clearDynamicLabels: () => void
}

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

const BreadcrumbContext = createContext<BreadcrumbContextValue | null>(null)

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

export function BreadcrumbProvider({ children }: { children: ReactNode }) {
  const { t } = useTranslation()
  const matches = useMatches()
  const location = useLocation()
  const [dynamicLabels, setDynamicLabels] = useState<Record<number, string>>({})

  // Clear dynamic label overrides on every route change
  useEffect(() => {
    setDynamicLabels({})
  }, [location.pathname])

  const items = useMemo(() => {
    // Filter to only matches that have a breadcrumb handle.
    // The parent layout route (ProtectedLayout) at App.tsx has NO handle
    // and must be excluded to prevent a spurious breadcrumb entry.
    const filtered = matches.filter(
      (m): m is typeof m & { handle: BreadcrumbHandle } =>
        m.handle != null && (m.handle as BreadcrumbHandle).breadcrumb != null
    )

    const homeItem: BreadcrumbItem = {
      path: '/dashboard',
      label: t('nav.home'),
    }

    const routeItems: BreadcrumbItem[] = filtered.map((m) => ({
      path: m.pathname,
      label: t((m.handle as BreadcrumbHandle).breadcrumb),
    }))

    // Apply dynamic label overrides (offset 0 = last item)
    const result = [homeItem, ...routeItems]
    for (const [offsetStr, label] of Object.entries(dynamicLabels)) {
      const offset = Number(offsetStr)
      // offset 0 = last, -1 = second-to-last, etc.
      const idx = result.length + offset - 1
      if (idx >= 0 && idx < result.length) {
        result[idx] = { ...result[idx], label }
      }
    }

    return result
  }, [matches, t, dynamicLabels])

  const setDynamicLabel = (offset: number, label: string) => {
    setDynamicLabels((prev) => ({ ...prev, [offset]: label }))
  }

  const clearDynamicLabels = () => {
    setDynamicLabels({})
  }

  return (
    <BreadcrumbContext.Provider value={{ items, setDynamicLabel, clearDynamicLabels }}>
      {children}
    </BreadcrumbContext.Provider>
  )
}

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
