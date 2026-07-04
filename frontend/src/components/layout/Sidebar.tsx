import { useTranslation } from 'react-i18next'
import { NavLink } from 'react-router-dom'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '@/components/ui/sidebar'
import { Badge } from '@/components/ui/badge'
import {
  SquaresFour,
  Desktop,
  Siren,
  ChartBar,
  Link,
  GearSix,
  Lightning,
  GlobeSimple,
  Key,
  ClipboardText,
  Wrench,
  type Icon,
} from '@phosphor-icons/react'
import { useAuthStore } from '@/stores/authStore'

export interface SidebarProps {
  alertCount?: number
}

interface NavItem {
  icon: Icon
  labelKey: string
  path: string
  adminOnly?: boolean
}

const navItems: NavItem[] = [
  { icon: SquaresFour, labelKey: 'nav.dashboard', path: '/dashboard' },
  { icon: Desktop, labelKey: 'nav.nodes', path: '/nodes' },
  { icon: Lightning, labelKey: 'nav.probes', path: '/nodes/probes' },
  { icon: GlobeSimple, labelKey: 'nav.beaconConfig', path: '/beacons/config' },
  { icon: Siren, labelKey: 'nav.alerts', path: '/alerts' },
  { icon: ChartBar, labelKey: 'nav.reports', path: '/reports' },
  { icon: Link, labelKey: 'nav.integrations', path: '/integrations/webhooks' },
  { icon: GearSix, labelKey: 'nav.settings', path: '/settings/preferences' },
  { icon: Key, labelKey: 'nav.apiKeys', path: '/settings/api-keys', adminOnly: true },
  { icon: ClipboardText, labelKey: 'nav.auditLogs', path: '/settings/audit-logs', adminOnly: true },
  { icon: Wrench, labelKey: 'nav.systemConfig', path: '/settings/system-config', adminOnly: true },
]

export function AppSidebar({ alertCount = 0 }: SidebarProps) {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const isAdmin = user?.role === 'admin'

  const visibleNavItems = navItems.filter((item) => !item.adminOnly || isAdmin)

  return (
    <Sidebar>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <NavLink to="/dashboard">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                  <Lightning className="size-4" weight="bold" />
                </div>
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="font-semibold">NodePulse</span>
                  <span className="text-xs text-muted-foreground">Monitoring</span>
                </div>
              </NavLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {visibleNavItems.map((item) => (
                <SidebarMenuItem key={item.path}>
                  <SidebarMenuButton asChild>
                    <NavLink to={item.path}>
                      <item.icon className="size-4" />
                      <span>{t(item.labelKey)}</span>
                    </NavLink>
                  </SidebarMenuButton>
                  {item.path === '/alerts' && alertCount > 0 && (
                    <Badge variant="destructive" className="ml-auto">
                      {alertCount}
                    </Badge>
                  )}
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton>
              <span className="text-xs text-muted-foreground">v2.0</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  )
}

export default AppSidebar
