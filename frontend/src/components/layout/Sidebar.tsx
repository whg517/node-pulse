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
} from '@phosphor-icons/react'

export interface SidebarProps {
  alertCount?: number
}

const navItems = [
  { icon: SquaresFour, labelKey: 'nav.dashboard', path: '/dashboard' },
  { icon: Desktop, labelKey: 'nav.nodes', path: '/nodes' },
  { icon: Lightning, labelKey: 'nav.probes', path: '/nodes/probes' },
  { icon: GlobeSimple, labelKey: 'nav.beaconConfig', path: '/beacons/config' },
  { icon: Siren, labelKey: 'nav.alerts', path: '/alerts' },
  { icon: ChartBar, labelKey: 'nav.reports', path: '/reports' },
  { icon: Link, labelKey: 'nav.integrations', path: '/integrations/webhooks' },
  { icon: GearSix, labelKey: 'nav.settings', path: '/settings/preferences' },
]

export function AppSidebar({ alertCount = 0 }: SidebarProps) {
  const { t } = useTranslation()

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
              {navItems.map((item) => (
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
