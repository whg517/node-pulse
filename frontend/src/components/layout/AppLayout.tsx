/**
 * AppLayout Component
 *
 * Main wrapper for all authenticated pages.
 * Contains: Sidebar (left) + Header (top) + main content area
 * Manages sidebar collapse state with useState
 * Responsive: sidebar overlay on mobile (<768px), fixed on desktop
 */

import { useState, useEffect, Suspense } from 'react'
import type { ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import { BreadcrumbProvider } from './BreadcrumbContext'
import { Breadcrumb } from './Breadcrumb'

export interface AppLayoutProps {
  /** Page content */
  children: ReactNode
  /** Alert count for sidebar badge */
  alertCount?: number
}

/**
 * AppLayout Component
 */
export function AppLayout({ children, alertCount = 0 }: AppLayoutProps) {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)

  // Handle responsive sidebar
  useEffect(() => {
    const handleResize = () => {
      // Close mobile sidebar on resize to desktop
      if (window.innerWidth >= 768) {
        setIsSidebarOpen(false)
      }
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  // Toggle sidebar collapse (desktop) or open/close (mobile)
  const handleToggleSidebar = () => {
    if (window.innerWidth < 768) {
      setIsSidebarOpen(!isSidebarOpen)
    } else {
      setIsSidebarCollapsed(!isSidebarCollapsed)
    }
  }

  // Close mobile sidebar
  const handleCloseSidebar = () => {
    setIsSidebarOpen(false)
  }

  return (
    <div className="min-h-screen bg-[var(--color-bg-page)]">
      {/* Sidebar */}
      <Sidebar
        isCollapsed={isSidebarCollapsed}
        isOpen={isSidebarOpen}
        onToggle={handleToggleSidebar}
        alertCount={alertCount}
      />

      {/* Main content area */}
      <div
        className={`flex min-h-screen flex-col transition-all duration-300 ease-in-out
          ${isSidebarCollapsed ? 'md:ml-16' : 'md:ml-64'}`}
      >
        {/* Header */}
        <Header onMenuToggle={handleToggleSidebar} />

        {/* Breadcrumb + Page content */}
        <BreadcrumbProvider>
          <div className="flex-1 p-4 md:p-6" onClick={isSidebarOpen ? handleCloseSidebar : undefined}>
            <div className="mb-4">
              <Breadcrumb />
            </div>
            <Suspense
              fallback={
                <div className="flex items-center justify-center py-20">
                  <div
                    className="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--color-brand)]"
                    role="status"
                    aria-label="Loading"
                  />
                </div>
              }
            >
              {children}
            </Suspense>
          </div>
        </BreadcrumbProvider>

        {/* Footer (optional) */}
        <footer className="border-t border-[var(--color-border)] bg-[var(--color-bg-surface)] py-4 text-center text-sm text-[var(--color-text-muted)]">
          NodePulse - Network Monitoring System
        </footer>
      </div>
    </div>
  )
}

export default AppLayout
