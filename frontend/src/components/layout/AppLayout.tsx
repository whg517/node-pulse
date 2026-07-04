import { Suspense } from 'react'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'
import { AppSidebar } from './Sidebar'
import { Header } from './Header'
import { BreadcrumbProvider } from './BreadcrumbContext'
import { Breadcrumb } from './Breadcrumb'
import { useGlobalRealtime } from '@/hooks/useGlobalRealtime'

export interface AppLayoutProps {
  children: ReactNode
  alertCount?: number
}

export function AppLayout({ children, alertCount = 0 }: AppLayoutProps) {
  const { t } = useTranslation()
  // G7: keep the WS + browser-notification connection alive app-wide (one
  // persistent session connection), instead of only on the Dashboard.
  useGlobalRealtime()

  return (
    <SidebarProvider>
      <AppSidebar alertCount={alertCount} />
      <SidebarInset>
        <Header />
        <BreadcrumbProvider>
          <div className="flex-1 p-4 md:p-6">
            <div className="mb-4">
              <Breadcrumb />
            </div>
            <Suspense
              fallback={
                <div className="flex items-center justify-center py-20">
                  <div
                    className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
                    role="status"
                    aria-label={t('common.loading')}
                  />
                </div>
              }
            >
              {children}
            </Suspense>
          </div>
        </BreadcrumbProvider>
        <footer className="border-t bg-background py-4 text-center text-sm text-muted-foreground">
          NodePulse - {t('dashboard.systemStatus')}
        </footer>
      </SidebarInset>
    </SidebarProvider>
  )
}

export default AppLayout
