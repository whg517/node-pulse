import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/DashboardPage'
import NodeDetailPage from './pages/NodeDetailPage'
import NodeComparisonPage from './pages/NodeComparison'
import AlertRulesPage from './pages/AlertRulesPage'
import AlertRecordsPage from './pages/AlertRecordsPage'
import WebhooksPage from './pages/WebhooksPage'
import DataExportPage from './pages/DataExportPage'
import ProtectedRoute from './components/common/ProtectedRoute'

function App() {
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
              <div>Node Management (Coming Soon)</div>
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
              <div>Alert History (Coming Soon)</div>
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

        {/* Default redirect */}
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
