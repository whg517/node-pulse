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
  // Honor user notification preferences (F4): a master switch and a minimum
  // severity filter. We read lazily so the service stays usable before the
  // settings store is wired (tests, partial init).
  const prefs = readNotificationPrefs()
  if (!prefs.enabled) return null
  if (!meetsMinLevel(alertLevel, prefs.minLevel)) return null

  const titles: Record<string, string> = { P0: '🚨 Critical Alert', P1: '⚠️ Warning Alert', P2: '📋 Notice Alert' }
  const title = titles[alertLevel] || 'Alert Notification'
  const body = `${nodeName}: ${metric} exceeded threshold (${threshold})`
  return showNotification(title, { body, requireInteraction: alertLevel === 'P0', tag: `alert-${alertId}` }, alertId)
}

// --- F4 preference plumbing ---

interface NotificationPrefsLite {
  enabled: boolean
  minLevel: 'P0' | 'P1' | 'P2'
}

// Decoupled from the settings store so this service file has no React/zustand
// import cycle; useGlobalRealtime wires the real store via setNotificationPrefsSource.
let prefsSource: (() => NotificationPrefsLite | null) | null = null

export function setNotificationPrefsSource(fn: () => NotificationPrefsLite | null) {
  prefsSource = fn
}

function readNotificationPrefs(): NotificationPrefsLite {
  if (prefsSource) {
    const p = prefsSource()
    if (p) return p
  }
  return { enabled: true, minLevel: 'P1' }
}

const LEVEL_RANK: Record<string, number> = { P0: 0, P1: 1, P2: 2 }

function meetsMinLevel(level: string, min: string): boolean {
  const lv = LEVEL_RANK[level]
  const mn = LEVEL_RANK[min]
  if (lv === undefined || mn === undefined) return true
  // Lower rank = more severe (P0=0). Notify when level is at least as severe
  // as the minimum — i.e. level rank <= min rank.
  return lv <= mn
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
