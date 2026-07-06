import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAlertsStore } from '@/stores/alertsStore'
import { useNodesStore } from '@/stores/nodesStore'
import { useSettingsStore } from '@/stores/settingsStore'
import type { NodeStatus } from '@/stores/types'
import * as NotificationService from '@/services/NotificationService'
import * as WebSocketService from '@/services/WebSocketService'

type AlertEventPayload = {
  id: string
  node_id: string
  metric: string
  level: string
  status?: 'pending' | 'in_progress' | 'resolved'
  created_at?: string
  updated_at?: string
  threshold?: string | number
}

type NodeStatusPayload = {
  node_id: string
  status?: string
}

/**
 * useGlobalRealtime keeps the WebSocket + browser-notification connection alive
 * for the whole authenticated session (mounted once in AppLayout), instead of
 * only on the Dashboard. Consumed event types:
 * - alert:new/updated/resolved (alertsStore upsert)
 * - alert:new (browser notification)
 * - node:online/node:offline (nodesStore.setNodeStatus — realtime node tile)
 */
export function useGlobalRealtime() {
  const navigate = useNavigate()
  const upsertAlertRecord = useAlertsStore((state) => state.upsertAlertRecord)
  const setNodeStatus = useNodesStore((state) => state.setNodeStatus)

  // Browser notifications: request permission + wire click handler app-wide.
  useEffect(() => {
    // Wire the user's notification preferences into NotificationService so the
    // severity filter (F4) applies to every alert:new event globally.
    NotificationService.setNotificationPrefsSource(() => {
      const prefs = useSettingsStore.getState().notificationPrefs
      return { enabled: prefs.enabled, minLevel: prefs.minLevel }
    })

    const handleNotificationClick = (alertId: string) =>
      navigate(`/alerts/records?highlight=${alertId}`)
    NotificationService.initialize(handleNotificationClick)
    return () => NotificationService.destroy()
  }, [navigate])

  // WebSocket: one persistent connection for the session.
  useEffect(() => {
    const handleMessage = (message: WebSocketService.WebSocketMessage<unknown>) => {
      if (
        message.type === 'alert:new' ||
        message.type === 'alert:updated' ||
        message.type === 'alert:resolved'
      ) {
        const payload = message.payload as AlertEventPayload
        upsertAlertRecord({
          id: payload.id,
          nodeId: payload.node_id,
          metric: payload.metric,
          level: payload.level,
          status: payload.status ?? 'pending',
          timestamp: payload.created_at ?? payload.updated_at ?? message.timestamp,
        })
      }

      if (message.type === 'alert:new') {
        const payload = message.payload as AlertEventPayload
        const nodeName = `Node ${payload.node_id.slice(0, 8)}`
        NotificationService.showAlertNotification(
          payload.id,
          payload.level,
          nodeName,
          payload.metric,
          String(payload.threshold)
        )
      }

      // node:online/offline — update the node tile in real time. The backend
      // emits these on heartbeat-arrival transitions (offline/connecting -> online)
      // and from the node-status sweeper (online -> offline after timeout).
      if (message.type === 'node:online' || message.type === 'node:offline') {
        const payload = message.payload as NodeStatusPayload
        const status = (message.type === 'node:online' ? 'online' : 'offline') as NodeStatus
        setNodeStatus(payload.node_id, status)
      }
    }
    WebSocketService.initialize(handleMessage)
    WebSocketService.connect()
    return () => WebSocketService.disconnect()
  }, [upsertAlertRecord, setNodeStatus])
}
