/**
 * WebSocket Service for Real-time Alert Updates
 * 
 * Manages WebSocket connection to Pulse server for real-time
 * alert notifications and updates.
 * 
 * Features:
 * - Auto-reconnect with exponential backoff
 * - JWT token authentication
 * - Connection state management
 * - Event subscription/unsubscription
 * - Heartbeat/ping-pong for connection health
 * 
 * @example
 * ```typescript
 * // Initialize and connect
 * WebSocketService.initialize((event) => {
 *   console.log('Received:', event)
 * })
 * 
 * WebSocketService.connect()
 * 
 * // Subscribe to alert events
 * WebSocketService.subscribe('alert:new', handler)
 * 
 * // Disconnect
 * WebSocketService.disconnect()
 * ```
 */

import { useAuthStore } from '../stores/authStore'
import { API_BASE_URL } from '../config/constants'

// ============== Types ==============

export type WebSocketEvent = 
  | 'alert:new'
  | 'alert:updated'
  | 'alert:resolved'
  | 'alert:acknowledged'
  | 'alert:note_created'
  | 'node:online'
  | 'node:offline'
  | 'system:heartbeat'
  | 'system:error'
  | 'pong'

export interface WebSocketMessage<T = unknown> {
  type: WebSocketEvent
  payload: T
  timestamp: string
}

export type MessageHandler<T = unknown> = (message: WebSocketMessage<T>) => void

export type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'error'

export interface ConnectionStats {
  connectedAt: number | null
  reconnectCount: number
  lastMessageAt: number | null
  messagesReceived: number
}

// ============== Configuration ==============

const CONFIG = {
  // Reconnect settings
  maxReconnectDelay: 30000, // 30 seconds
  initialReconnectDelay: 1000, // 1 second
  reconnectDelayMultiplier: 2,
  
  // Heartbeat settings
  heartbeatInterval: 30000, // 30 seconds
  heartbeatTimeout: 10000, // 10 seconds
  
  // Connection timeout
  connectionTimeout: 10000, // 10 seconds
}

// ============== Service State ==============

let ws: WebSocket | null = null
let connectionState: ConnectionState = 'disconnected'
let reconnectDelay = CONFIG.initialReconnectDelay
let reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null
let heartbeatIntervalId: ReturnType<typeof setInterval> | null = null
let heartbeatTimeoutId: ReturnType<typeof setTimeout> | null = null
let connectionStats: ConnectionStats = {
  connectedAt: null,
  reconnectCount: 0,
  lastMessageAt: null,
  messagesReceived: 0,
}

const eventHandlers = new Map<WebSocketEvent, Set<MessageHandler>>()
let globalHandler: ((message: WebSocketMessage<unknown>) => void) | null = null

// ============== Connection Management ==============

/**
 * Get current connection state
 */
export function getConnectionState(): ConnectionState {
  return connectionState
}

/**
 * Get connection statistics
 */
export function getConnectionStats(): ConnectionStats {
  return { ...connectionStats }
}

/**
 * Build WebSocket URL with authentication
 */
function buildWebSocketUrl(): string {
  const state = useAuthStore.getState()
  const token = state.accessToken
  
  const apiUrl = new URL(API_BASE_URL || window.location.origin, window.location.origin)
  apiUrl.protocol = apiUrl.protocol === 'https:' ? 'wss:' : 'ws:'
  apiUrl.pathname = '/ws'
  apiUrl.search = ''
  let url = apiUrl.toString()
  
  // Add token if available
  if (token) {
    url += `?token=${encodeURIComponent(token)}`
  }
  
  return url
}

/**
 * Send heartbeat to keep connection alive
 */
function sendHeartbeat(): void {
  if (ws?.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'ping', timestamp: Date.now() }))
    
    // Set timeout for pong response
    heartbeatTimeoutId = setTimeout(() => {
      console.warn('[WebSocket] Heartbeat timeout, closing connection')
      handleClose({ code: 4000, reason: 'Heartbeat timeout' } as CloseEvent)
    }, CONFIG.heartbeatTimeout)
  }
}

/**
 * Handle incoming messages
 */
function handleMessage(event: MessageEvent): void {
  connectionStats.lastMessageAt = Date.now()
  connectionStats.messagesReceived++
  
  try {
    const message: WebSocketMessage<unknown> = JSON.parse(event.data)
    
    if (import.meta.env.DEV) {
      console.log('[WebSocket] Message received:', message)
    }
    
    // Handle pong response to heartbeat
    if (message.type === ('pong' as WebSocketEvent)) {
      if (heartbeatTimeoutId) {
        clearTimeout(heartbeatTimeoutId)
        heartbeatTimeoutId = null
      }
      return
    }
    
    // Call global handler
    if (globalHandler) {
      globalHandler(message)
    }
    
    // Call specific event handlers
    const handlers = eventHandlers.get(message.type)
    if (handlers) {
      handlers.forEach((handler) => handler(message as WebSocketMessage<unknown>))
    }
  } catch (error) {
    console.error('[WebSocket] Failed to parse message:', error)
  }
}

/**
 * Handle connection open
 */
function handleOpen(): void {
  connectionState = 'connected'
  connectionStats.connectedAt = Date.now()
  reconnectDelay = CONFIG.initialReconnectDelay // Reset reconnect delay
  
  console.log('[WebSocket] Connected')
  
  // Start heartbeat
  heartbeatIntervalId = setInterval(sendHeartbeat, CONFIG.heartbeatInterval)
  
  // Notify state change listeners
  notifyStateChange()
}

/**
 * Handle connection close
 */
function handleClose(event: CloseEvent): void {
  connectionState = 'disconnected'
  connectionStats.connectedAt = null
  
  // Clear heartbeat
  if (heartbeatIntervalId) {
    clearInterval(heartbeatIntervalId)
    heartbeatIntervalId = null
  }
  if (heartbeatTimeoutId) {
    clearTimeout(heartbeatTimeoutId)
    heartbeatTimeoutId = null
  }
  
  console.log('[WebSocket] Disconnected:', event.code, event.reason)
  
  // Attempt reconnection if not closed intentionally
  if (event.code !== 1000) { // 1000 = normal closure
    scheduleReconnect()
  }
  
  notifyStateChange()
}

/**
 * Handle connection error
 */
function handleError(): void {
  connectionState = 'error'
  console.error('[WebSocket] Connection error')
  notifyStateChange()
}

/**
 * Schedule automatic reconnection
 */
function scheduleReconnect(): void {
  if (reconnectTimeoutId) {
    clearTimeout(reconnectTimeoutId)
  }
  
  connectionStats.reconnectCount++
  console.log(`[WebSocket] Reconnecting in ${reconnectDelay}ms (attempt ${connectionStats.reconnectCount})`)
  
  reconnectTimeoutId = setTimeout(() => {
    connect()
    // Increase delay for next attempt
    reconnectDelay = Math.min(
      reconnectDelay * CONFIG.reconnectDelayMultiplier,
      CONFIG.maxReconnectDelay
    )
  }, reconnectDelay)
}

// ============== Public API ==============

/**
 * Initialize WebSocket service
 * Sets up event handlers and global message handler
 * 
 * @param onMessage - Global message handler for all events
 */
export function initialize(onMessage: (message: WebSocketMessage<unknown>) => void): void {
  globalHandler = onMessage
  
  // Clear any existing reconnection attempts
  if (reconnectTimeoutId) {
    clearTimeout(reconnectTimeoutId)
    reconnectTimeoutId = null
  }
  
  console.log('[WebSocket] Initialized')
}

/**
 * Connect to WebSocket server
 */
export function connect(): void {
  if (ws?.readyState === WebSocket.OPEN || ws?.readyState === WebSocket.CONNECTING) {
    console.log('[WebSocket] Already connected or connecting')
    return
  }
  
  connectionState = 'connecting'
  notifyStateChange()
  
  const url = buildWebSocketUrl()
  console.log('[WebSocket] Connecting to:', url.replace(/\?.*$/, '?token=***'))
  
  ws = new WebSocket(url)
  
  ws.onopen = handleOpen
  ws.onmessage = handleMessage
  ws.onerror = handleError
  ws.onclose = handleClose
  
  // Connection timeout
  const timeoutId = setTimeout(() => {
    if (ws?.readyState === WebSocket.CONNECTING) {
      console.error('[WebSocket] Connection timeout')
      ws.close()
    }
  }, CONFIG.connectionTimeout)
  
  ws.onopen = () => {
    clearTimeout(timeoutId)
    handleOpen()
  }
}

/**
 * Disconnect from WebSocket server
 * @param intentional - Whether this is an intentional disconnection (default: true)
 */
export function disconnect(intentional = true): void {
  if (reconnectTimeoutId) {
    clearTimeout(reconnectTimeoutId)
    reconnectTimeoutId = null
  }
  
  if (heartbeatIntervalId) {
    clearInterval(heartbeatIntervalId)
    heartbeatIntervalId = null
  }
  
  if (heartbeatTimeoutId) {
    clearTimeout(heartbeatTimeoutId)
    heartbeatTimeoutId = null
  }
  
  if (ws) {
    ws.onclose = null // Don't trigger reconnection
    if (intentional) {
      ws.close(1000, 'Intentional disconnect')
    } else {
      ws.close()
    }
    ws = null
  }
  
  connectionState = 'disconnected'
  connectionStats.connectedAt = null
  
  console.log('[WebSocket] Disconnected')
  notifyStateChange()
}

/**
 * Subscribe to specific event type
 * 
 * @param eventType - Event to subscribe to
 * @param handler - Handler function
 * @returns Unsubscribe function
 */
export function subscribe<T = unknown>(
  eventType: WebSocketEvent,
  handler: MessageHandler<T>
): () => void {
  if (!eventHandlers.has(eventType)) {
    eventHandlers.set(eventType, new Set())
  }
  
  const handlers = eventHandlers.get(eventType)!
  handlers.add(handler as MessageHandler<unknown>)
  
  // Return unsubscribe function
  return () => {
    handlers.delete(handler as MessageHandler<unknown>)
    if (handlers.size === 0) {
      eventHandlers.delete(eventType)
    }
  }
}

/**
 * Subscribe to multiple event types
 * 
 * @param eventTypes - Events to subscribe to
 * @param handler - Handler function
 * @returns Unsubscribe function for all subscriptions
 */
export function subscribeMultiple<T = unknown>(
  eventTypes: WebSocketEvent[],
  handler: MessageHandler<T>
): () => void {
  const unsubscribers = eventTypes.map((type) => subscribe(type, handler))
  
  return () => {
    unsubscribers.forEach((unsub) => unsub())
  }
}

/**
 * Send message to server
 */
export function send<T = unknown>(message: Omit<WebSocketMessage<T>, 'timestamp'>): void {
  if (ws?.readyState !== WebSocket.OPEN) {
    console.warn('[WebSocket] Cannot send message, not connected')
    return
  }
  
  const payload = {
    ...message,
    timestamp: new Date().toISOString(),
  }
  
  ws.send(JSON.stringify(payload))
  
  if (import.meta.env.DEV) {
    console.log('[WebSocket] Sent:', payload)
  }
}

// ============== State Change Listeners ==============

type StateChangeListener = (state: ConnectionState) => void
const stateChangeListeners = new Set<StateChangeListener>()

/**
 * Subscribe to connection state changes
 */
export function onStateChange(listener: StateChangeListener): () => void {
  stateChangeListeners.add(listener)
  return () => stateChangeListeners.delete(listener)
}

function notifyStateChange(): void {
  stateChangeListeners.forEach((listener) => listener(connectionState))
}

// ============== Utility Functions ==============

/**
 * Reset connection statistics
 */
export function resetStats(): void {
  connectionStats = {
    connectedAt: null,
    reconnectCount: 0,
    lastMessageAt: null,
    messagesReceived: 0,
  }
}

/**
 * Check if currently connected
 */
export function isConnected(): boolean {
  return connectionState === 'connected' && ws?.readyState === WebSocket.OPEN
}

/**
 * Get number of active subscriptions
 */
export function getSubscriptionCount(): number {
  let count = 0
  eventHandlers.forEach((handlers) => {
    count += handlers.size
  })
  return count
}

// Export service object
export const WebSocketService = {
  // Connection
  connect,
  disconnect,
  getConnectionState,
  getConnectionStats,
  isConnected,
  onStateChange,
  
  // Messaging
  send,
  subscribe,
  subscribeMultiple,
  
  // Lifecycle
  initialize,
  resetStats,
  
  // Stats
  getSubscriptionCount,
}
