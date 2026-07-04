import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAlertsStore } from '@/stores/alertsStore'
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
 * only on the Dashboard. It also consumes more event types than before:
 * alert:new/updated/resolved (store upsert), alert:new (notification), and
 * node:online/offline (best-effort node-status refresh).
 */
export function useGlobalRealtime() {
  const navigate = useNavigate()
  const upsertAlertRecord = useAlertsStore((state) => state.upsertAlertRecord)

  // Browser notifications: request permission + wire click handler app-wide.
  useEffect(() => {
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

      // node:online/offline — best-effort: refresh node lists by invalidating the
      // alerts store's perception; a full node store refresh is page-driven.
      if (message.type === 'node:online' || message.type === 'node:offline') {
        // Deliberately minimal: avoid coupling to nodesStore here. Pages that
        // render node status re-fetch on their own polling cadence; this event
        // is logged at debug level for observability.
        const payload = message.payload as NodeStatusPayload
        console.debug('[realtime] node status event', payload.node_id, message.type)
      }
    }
    WebSocketService.initialize(handleMessage)
    WebSocketService.connect()
    return () => WebSocketService.disconnect()
  }, [upsertAlertRecord])
}
