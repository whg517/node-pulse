import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { useEffect } from 'react'
import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/DashboardPage'
import NodeDetailPage from './pages/NodeDetailPage'
import NodeComparisonPage from './pages/NodeComparison'
import NodeManagementPage from './pages/NodeManagementPage'
import AlertRulesPage from './pages/AlertRulesPage'
import AlertRecordsPage from './pages/AlertRecordsPage'
import AlertHistoryPage from './pages/AlertHistoryPage'
import WebhooksPage from './pages/WebhooksPage'
import DataExportPage from './pages/DataExportPage'
import SessionsPage from './pages/SessionsPage'
import ReportsPage from './pages/Reports'
import PreferencesPage from './pages/PreferencesPage'
import UsersPage from './pages/UsersPage'
import SystemHealthPage from './pages/SystemHealthPage'
import NotFoundPage from './pages/NotFoundPage'
import ProtectedRoute from './components/common/ProtectedRoute'
import { AppLayout } from './components/layout'
import { useAuthStore, setupCrossTabLogoutSync, setupVisibilityHandler } from './stores/authStore'
import { useAlertsStore } from './stores/alertsStore'

/**
 * Layout wrapper for protected routes
 * Combines ProtectedRoute with AppLayout
 */
function ProtectedLayout() {
  const alertRecords = useAlertsStore((state) => state.alertRecords)
  // Count unresolved alerts for badge
  const alertCount = alertRecords.filter((r) => r.status !== 'resolved').length

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

  // Restore session on app startup
  useEffect(() => {
    restoreSession()
  }, [restoreSession])

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
    <BrowserRouter>
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

          {/* Alerts */}
          <Route path="/alerts" element={<Navigate to="rules" replace />} />
          <Route path="/alerts/rules" element={<AlertRulesPage />} />
          <Route path="/alerts/records" element={<AlertRecordsPage />} />
          <Route path="/alerts/history" element={<AlertHistoryPage />} />

          {/* Reports */}
          <Route path="/reports" element={<ReportsPage />} />
          <Route path="/reports/history" element={<DataExportPage />} />

          {/* Integrations */}
          <Route path="/integrations" element={<Navigate to="webhooks" replace />} />
          <Route path="/integrations/webhooks" element={<WebhooksPage />} />
          <Route path="/integrations/health" element={<SystemHealthPage />} />

          {/* Settings */}
          <Route path="/settings" element={<Navigate to="preferences" replace />} />
          <Route path="/settings/preferences" element={<PreferencesPage />} />
          <Route path="/settings/sessions" element={<SessionsPage />} />
          <Route path="/settings/users" element={<UsersPage />} />
        </Route>

        {/* 404 - Not Found */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
