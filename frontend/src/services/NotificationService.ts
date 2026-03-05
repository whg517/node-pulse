/**
 * NotificationService - Browser Push Notifications
 * Uses the native Browser Notification API with permission management.
 */

export interface NotificationOptions {
  body: string
  icon?: string
  badge?: string
  tag?: string
  requireInteraction?: boolean
  silent?: boolean
}

export interface NotificationPermissionState {
  granted: boolean
  denied: boolean
  default: boolean
}

export type NotificationClickHandler = (notificationId: string) => void

let clickHandler: NotificationClickHandler | null = null
let isInitialized = false
let silenceMode = false

export function getPermissionState(): NotificationPermissionState {
  if (typeof Notification === 'undefined') {
    return { granted: false, denied: false, default: false }
  }
  return {
    granted: Notification.permission === 'granted',
    denied: Notification.permission === 'denied',
    default: Notification.permission === 'default',
  }
}

export async function requestPermission(): Promise<boolean> {
  if (typeof Notification === 'undefined') {
    console.warn('[NotificationService] Notifications not supported')
    return false
  }
  if (Notification.permission === 'granted') return true
  try {
    const permission = await Notification.requestPermission()
    return permission === 'granted'
  } catch (error) {
    console.error('[NotificationService] Failed to request permission:', error)
    return false
  }
}

export function canShowNotifications(): boolean {
  return typeof Notification !== 'undefined' && Notification.permission === 'granted' && !silenceMode
}

export function initialize(onNotificationClick: NotificationClickHandler): void {
  if (isInitialized) return
  clickHandler = onNotificationClick
  isInitialized = true
  const storedSilenceMode = localStorage.getItem('notification-silence-mode')
  silenceMode = storedSilenceMode === 'true'
  console.log('[NotificationService] Initialized', { canShow: canShowNotifications(), silenceMode })
}

export function destroy(): void {
  clickHandler = null
  isInitialized = false
}

export function showNotification(title: string, options: NotificationOptions, notificationId: string): Notification | null {
  if (!canShowNotifications()) return null
  try {
    const notification = new Notification(title, {
      body: options.body,
      icon: options.icon || '/favicon.ico',
      badge: options.badge || '/favicon.ico',
      tag: options.tag || notificationId,
      requireInteraction: options.requireInteraction ?? false,
      silent: options.silent ?? silenceMode,
    })
    notification.onclick = (event: Event) => {
      event.preventDefault()
      window.focus()
      if (clickHandler) clickHandler(notificationId)
      notification.close()
    }
    if (!options.requireInteraction) setTimeout(() => notification.close(), 10000)
    return notification
  } catch (error) {
    console.error('[NotificationService] Failed to show notification:', error)
    return null
  }
}

export function showAlertNotification(alertId: string, alertLevel: string, nodeName: string, metric: string, threshold: string): Notification | null {
  const titles: Record<string, string> = { P0: '🚨 Critical Alert', P1: '⚠️ Warning Alert', P2: '📋 Notice Alert' }
  const title = titles[alertLevel] || 'Alert Notification'
  const body = `${nodeName}: ${metric} exceeded threshold (${threshold})`
  return showNotification(title, { body, requireInteraction: alertLevel === 'P0', tag: `alert-${alertId}` }, alertId)
}

export function toggleSilenceMode(): boolean {
  silenceMode = !silenceMode
  localStorage.setItem('notification-silence-mode', String(silenceMode))
  return silenceMode
}

export function isSilenceModeEnabled(): boolean {
  return silenceMode
}

export async function requestAndShow(title: string, options: NotificationOptions, notificationId: string): Promise<Notification | null> {
  const hasPermission = await requestPermission()
  if (!hasPermission) return null
  return showNotification(title, options, notificationId)
}

export function getSettings() {
  return {
    supported: typeof Notification !== 'undefined',
    permission: typeof Notification !== 'undefined' ? Notification.permission : 'unsupported',
    canShow: canShowNotifications(),
    silenceMode,
    initialized: isInitialized,
  }
}

export const NotificationService = {
  getPermissionState, requestPermission, canShowNotifications,
  initialize, destroy,
  showNotification, showAlertNotification, requestAndShow,
  toggleSilenceMode, isSilenceModeEnabled,
  getSettings,
}
