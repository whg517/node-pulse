/**
 * Breadcrumb Provider
 *
 * Route-config-driven breadcrumb items for the layout.
 * Reads current pathname and maps segments to breadcrumb labels
 * using a static route config. Supports dynamic label overrides.
 */

import { useState, useMemo, useCallback, type ReactNode } from 'react'
import { useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { BreadcrumbContext, type BreadcrumbItem } from './useBreadcrumb'

// Route segment → i18n key mapping.
// Each route path defines its own label, avoiding global key collisions
// (e.g., /alerts/history vs /reports/history).
const routeLabels: Record<string, string> = {
  dashboard: 'nav.dashboard',
  nodes: 'nav.nodes',
  comparison: 'nav.comparison',
  alerts: 'nav.alerts',
  rules: 'nav.alertRules',
  records: 'nav.alertRecords',
  history: 'nav.alertHistory',       // default, overridden per-path below
  performance: 'nav.performance',
  reports: 'nav.reports',
  webhooks: 'nav.webhooks',
  health: 'nav.systemHealth',
  integrations: 'nav.integrations',
  settings: 'nav.settings',
  preferences: 'nav.preferences',
  sessions: 'nav.sessions',
  users: 'nav.users',
  export: 'nav.export',
}

// Per-path overrides for segments that have different meanings
// depending on their parent path context.
const pathOverrides: Record<string, Record<string, string>> = {
  '/reports/history': { history: 'nav.exportHistory' },
}

function getLabel(segment: string, parentPath: string): string {
  const overrides = pathOverrides[parentPath]
  if (overrides?.[segment]) return overrides[segment]
  return routeLabels[segment] || `nav.${segment}`
}

function isIdSegment(segment: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(segment) || /^\d+$/.test(segment)
}

export function BreadcrumbProvider({ children }: { children: ReactNode }) {
  const { t } = useTranslation()
  const location = useLocation()
  // Dynamic labels keyed by pathname — each route stores its own overrides
  // so no cleanup is needed on navigation.
  const [dynamicLabels, setDynamicLabels] = useState<Record<string, Record<number, string>>>({})

  const currentLabels = dynamicLabels[location.pathname]

  const items = useMemo(() => {
    const segments = location.pathname.split('/').filter(Boolean)
    const breadcrumbItems: BreadcrumbItem[] = []

    // Always add home
    breadcrumbItems.push({ path: '/dashboard', label: t('nav.home') })

    let currentPath = ''
    for (const segment of segments) {
      currentPath += `/${segment}`
      if (isIdSegment(segment)) {
        breadcrumbItems.push({ path: currentPath, label: t('nav.details') })
      } else {
        breadcrumbItems.push({ path: currentPath, label: t(getLabel(segment, currentPath)) })
      }
    }

    // Apply dynamic label overrides (offset 0 = last item)
    if (currentLabels) {
      for (const [offsetStr, label] of Object.entries(currentLabels)) {
        const offset = Number(offsetStr)
        const idx = breadcrumbItems.length + offset - 1
        if (idx >= 0 && idx < breadcrumbItems.length) {
          breadcrumbItems[idx] = { ...breadcrumbItems[idx], label }
        }
      }
    }

    return breadcrumbItems
  }, [location.pathname, t, currentLabels])

  const setDynamicLabel = useCallback((offset: number, label: string) => {
    setDynamicLabels((prev) => ({
      ...prev,
      [location.pathname]: { ...(prev[location.pathname] || {}), [offset]: label },
    }))
  }, [location.pathname])

  const clearDynamicLabels = useCallback(() => {
    setDynamicLabels((prev) => {
      const { [location.pathname]: _, ...rest } = prev
      return rest
    })
  }, [location.pathname])

  return (
    <BreadcrumbContext.Provider value={{ items, setDynamicLabel, clearDynamicLabels }}>
      {children}
    </BreadcrumbContext.Provider>
  )
}
