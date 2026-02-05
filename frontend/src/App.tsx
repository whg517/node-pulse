import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
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
import PerformanceDashboard from './pages/PerformanceDashboard'
import ProtectedRoute from './components/common/ProtectedRoute'
import { useAuthStore } from './stores/authStore'

function App() {
  const restoreSession = useAuthStore((state) => state.restoreSession)

  // Restore session on app startup
  useEffect(() => {
    restoreSession()
  }, [restoreSession])

  return (
    <BrowserRouter>
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<LoginPage />} />

        {/* Protected routes - require authentication */}
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <DashboardPage />
            </ProtectedRoute>
          }
        />

        {/* Future protected routes (will be implemented in later stories) */}
        <Route
          path="/nodes"
          element={
            <ProtectedRoute>
              <NodeManagementPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/nodes/:id"
          element={
            <ProtectedRoute>
              <NodeDetailPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/comparison"
          element={
            <ProtectedRoute>
              <NodeComparisonPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/alerts/rules"
          element={
            <ProtectedRoute>
              <AlertRulesPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/alerts/records"
          element={
            <ProtectedRoute>
              <AlertRecordsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/alerts/history"
          element={
            <ProtectedRoute>
              <AlertHistoryPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/webhooks"
          element={
            <ProtectedRoute>
              <WebhooksPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/export"
          element={
            <ProtectedRoute>
              <DataExportPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/performance"
          element={
            <ProtectedRoute>
              <PerformanceDashboard />
            </ProtectedRoute>
          }
        />

        {/* Default redirect */}
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
