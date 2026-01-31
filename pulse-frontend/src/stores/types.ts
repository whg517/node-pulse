// Shared TypeScript interfaces for Zustand stores

// ============== Auth Types ==============
export interface User {
  id: string
  username: string
  role: 'admin' | 'operator' | 'viewer'
}

// ============== Node Types ==============
export interface Node {
  id: string
  name: string
  ip: string
  region: string
  tags: string[]
  status: 'online' | 'offline' | 'connecting'
}

export type NodeStatus = 'online' | 'offline' | 'connecting'

// ============== Alert Types ==============
export interface AlertRule {
  id: string
  metric: 'latency' | 'packet_loss_rate' | 'jitter'
  threshold: number
  level: 'P0' | 'P1' | 'P2'
  nodeId: string | null
  enabled: boolean
}

export interface AlertRecord {
  id: string
  nodeId: string
  metric: string
  level: string
  status: 'pending' | 'processing' | 'resolved'
  timestamp: string
}

export interface AlertFilter {
  level?: 'P0' | 'P1' | 'P2' | 'all'
  status?: 'pending' | 'processing' | 'resolved' | 'all'
  nodeId?: string | null
  searchQuery?: string
}

// ============== Dashboard Types ==============
export interface DashboardFilter {
  region: string | null
  status: 'online' | 'offline' | 'all'
  searchQuery: string
}

export type TimeRange = '24h' | '7d' | '30d'
