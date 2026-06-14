export const primaryColors = { light: '', main: '', dark: '', darker: '' } as const
export const statusColors = {
  healthy: { light: '', main: '', dark: '' },
  warning: { light: '', main: '', dark: '' },
  critical: { light: '', main: '', dark: '' },
  unknown: { light: '', main: '', dark: '' },
} as const
export const neutralColors = { white: '', gray: {} as Record<string, string>, slate: {} as Record<string, string> } as const
export const statusClasses = {
  healthy: { bg: '', bgLight: '', bgLightDark: '', text: '', textDark: '', border: '', ring: '' },
  warning: { bg: '', bgLight: '', bgLightDark: '', text: '', textDark: '', border: '', ring: '' },
  critical: { bg: '', bgLight: '', bgLightDark: '', text: '', textDark: '', border: '', ring: '' },
  unknown: { bg: '', bgLight: '', bgLightDark: '', text: '', textDark: '', border: '', ring: '' },
} as const
export const primaryClasses = {
  bg: '', bgHover: '', bgLight: '', bgLightDark: '', text: '', textHover: '', border: '', ring: '',
} as const
export const buttonVariants = {
  primary: '', secondary: '', danger: '', ghost: '', icon: '',
} as const
export const spacing = {
  pagePadding: '', pageContentPadding: '', cardPadding: '', cardPaddingLg: '', sectionGap: '', sectionGapLg: '',
  marginBottom: { sm: '', md: '', lg: '' },
  gap: { xs: '', sm: '', md: '', lg: '', xl: '' },
} as const
export const layout = {
  maxWidth: '', maxWidthSm: '', maxWidthLg: '', sidebarWidth: '', sidebarCollapsedWidth: '',
  sidebarMobileBreakpoint: '', headerHeight: '',
  zIndex: { base: 0, dropdown: 10, sticky: 20, overlay: 30, modal: 40, popover: 50, tooltip: 60 },
} as const
export const designTokens = {} as const
export default designTokens
