import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { lazy, useEffect, useRef, Suspense } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from './lib/query-client'
import ProtectedRoute from './components/common/ProtectedRoute'
import { AppLayout } from './components/layout'
import { useAuthStore, setupCrossTabLogoutSync, setupVisibilityHandler } from './stores/authStore'
import { useAlertsStore } from './stores/alertsStore'
import { initializeTheme } from './stores/settingsStore'

// Apply persisted theme immediately (before first render) so there is no
// flash of unstyled/wrong-theme content on page load.
initializeTheme()

const LoginPage = lazy(() => import('./pages/LoginPage'))
const DashboardPage = lazy(() => import('./pages/DashboardPage'))
const NodeDetailPage = lazy(() => import('./pages/NodeDetailPage'))
const NodeComparisonPage = lazy(() => import('./pages/NodeComparison'))
const NodeManagementPage = lazy(() => import('./pages/NodeManagementPage'))
const ProbeManagementPage = lazy(() => import('./pages/ProbeManagementPage'))
const BeaconConfigPage = lazy(() => import('./pages/BeaconConfigPage'))
const AlertRulesPage = lazy(() => import('./pages/AlertRulesPage'))
const AlertRecordsPage = lazy(() => import('./pages/AlertRecordsPage'))
const AlertHistoryPage = lazy(() => import('./pages/AlertHistoryPage'))
const WebhooksPage = lazy(() => import('./pages/WebhooksPage'))
const DataExportPage = lazy(() => import('./pages/DataExportPage'))
const SessionsPage = lazy(() => import('./pages/SessionsPage'))
const ReportsPage = lazy(() => import('./pages/Reports'))
const PreferencesPage = lazy(() => import('./pages/PreferencesPage'))
const UsersPage = lazy(() => import('./pages/UsersPage'))
const SystemHealthPage = lazy(() => import('./pages/SystemHealthPage'))
const PerformanceDashboard = lazy(() => import('./pages/PerformanceDashboard'))
const NotFoundPage = lazy(() => import('./pages/NotFoundPage'))

/**
 * Layout wrapper for protected routes
 * Combines ProtectedRoute with AppLayout
 */
function ProtectedLayout() {
  const alertRecords = useAlertsStore((state) => state.alertRecords)
  const fetchAlertRecords = useAlertsStore((state) => state.fetchAlertRecords)
  // Count unresolved alerts for badge
  const alertCount = alertRecords.filter((r) => r.status !== 'resolved').length

  // Fetch alert records once on mount so the sidebar badge is populated
  useEffect(() => {
    fetchAlertRecords().catch(() => {})
  }, [fetchAlertRecords])

  return (
    <ProtectedRoute>
      <AppLayout alertCount={alertCount}>
        <Outlet />
      </AppLayout>
    </ProtectedRoute>
  )
}

function App() {
  const restoreSession = useAuthStore((state) => state.restoreSession)
  const sessionRestoredRef = useRef(false)

  // Restore session on app startup only once.
  // useRef guard prevents double-invocation in React StrictMode (dev), which would
  // cause isLoading to flip true→false twice, making ProtectedRoute unmount/remount
  // child pages and re-trigger their useEffect hooks (e.g. loadWebhooks).
  useEffect(() => {
    if (sessionRestoredRef.current) return
    sessionRestoredRef.current = true
    restoreSession()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Setup cross-tab logout sync and visibility handler
  useEffect(() => {
    const cleanupCrossTab = setupCrossTabLogoutSync()
    const cleanupVisibility = setupVisibilityHandler()

    return () => {
      cleanupCrossTab()
      cleanupVisibility()
    }
  }, [])

  return (
    <QueryClientProvider client={queryClient}>
    <BrowserRouter>
      <Suspense fallback={null}>
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<LoginPage />} />

        {/* Protected routes with AppLayout */}
        <Route element={<ProtectedLayout />}>
          {/* Dashboard */}
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/" element={<Navigate to="/dashboard" replace />} />

          {/* Nodes */}
          <Route path="/nodes" element={<NodeManagementPage />} />
          <Route path="/nodes/:id" element={<NodeDetailPage />} />
          <Route path="/nodes/comparison" element={<NodeComparisonPage />} />
          <Route path="/nodes/probes" element={<ProbeManagementPage />} />
          <Route path="/beacons/config" element={<BeaconConfigPage />} />

          {/* Alerts */}
          <Route path="/alerts" element={<Navigate to="/alerts/rules" replace />} />
          <Route path="/alerts/rules" element={<AlertRulesPage />} />
          <Route path="/alerts/records" element={<AlertRecordsPage />} />
          <Route path="/alerts/history" element={<AlertHistoryPage />} />

          {/* Performance */}
          <Route path="/performance" element={<PerformanceDashboard />} />

          {/* Reports */}
          <Route path="/reports" element={<ReportsPage />} />
          <Route path="/reports/history" element={<DataExportPage />} />

          {/* Integrations */}
          <Route path="/integrations" element={<Navigate to="/integrations/webhooks" replace />} />
          <Route path="/integrations/webhooks" element={<WebhooksPage />} />
          <Route path="/integrations/health" element={<SystemHealthPage />} />

          {/* Settings */}
          <Route path="/settings" element={<Navigate to="/settings/preferences" replace />} />
          <Route path="/settings/preferences" element={<PreferencesPage />} />
          <Route path="/settings/sessions" element={<SessionsPage />} />
          <Route path="/settings/users" element={<UsersPage />} />

          {/* Short aliases for E2E and legacy navigation */}
          <Route path="/webhooks" element={<Navigate to="/integrations/webhooks" replace />} />
          <Route path="/sessions" element={<Navigate to="/settings/sessions" replace />} />
          <Route path="/comparison" element={<Navigate to="/nodes/comparison" replace />} />
        </Route>

        {/* 404 - Not Found */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
      </Suspense>
    </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App
