/**
 * Layout Components
 *
 * Export all layout components for the NodePulse frontend.
 */

export { AppLayout } from './AppLayout'
export type { AppLayoutProps } from './AppLayout'

export { Sidebar } from './Sidebar'
export type { SidebarProps } from './Sidebar'

export { SidebarItem } from './SidebarItem'
export type { SidebarItemProps } from './SidebarItem'

export { Header } from './Header'
export type { HeaderProps } from './Header'

export { Breadcrumb } from './Breadcrumb'

export { BreadcrumbProvider, useBreadcrumb, useSetBreadcrumbLabel } from './BreadcrumbContext'
export type { BreadcrumbItem, BreadcrumbHandle, BreadcrumbContextValue } from './BreadcrumbContext'
