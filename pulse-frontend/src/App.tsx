import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/DashboardPage'
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
              <div>Node Detail (Coming Soon)</div>
            </ProtectedRoute>
          }
        />
        <Route
          path="/comparison"
          element={
            <ProtectedRoute>
              <div>Node Comparison (Coming Soon)</div>
            </ProtectedRoute>
          }
        />
        <Route
          path="/alerts/rules"
          element={
            <ProtectedRoute>
              <div>Alert Rules (Coming Soon)</div>
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
          path="/export"
          element={
            <ProtectedRoute>
              <div>Data Export (Coming Soon)</div>
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
