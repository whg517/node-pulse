/**
 * Report Schedules API (ADR-001).
 * Server-side recurring reports. Backend: /api/v1/reports/schedules.
 */
import { apiClient } from './client'

export interface ReportScheduleDTO {
  id: string
  owner_user_id: string
  name: string
  frequency: 'daily' | 'weekly' | 'monthly'
  time_of_day: string
  node_ids: string[]
  metrics: string[]
  format: 'csv' | 'pdf'
  recipient_email?: string
  enabled: boolean
  last_run_at?: string | null
  next_run_at?: string | null
  created_at: string
  updated_at: string
}

export interface CreateScheduleRequest {
  name: string
  frequency: 'daily' | 'weekly' | 'monthly'
  time_of_day?: string
  node_ids: string[]
  metrics?: string[]
  format?: 'csv' | 'pdf'
  recipient_email?: string
  enabled?: boolean
}

export async function listReportSchedules(): Promise<{ data: { schedules: ReportScheduleDTO[] }; message: string; timestamp: string }> {
  return apiClient('/api/v1/reports/schedules')
}

export async function createReportSchedule(req: CreateScheduleRequest): Promise<{ data: ReportScheduleDTO; message: string; timestamp: string }> {
  return apiClient('/api/v1/reports/schedules', { method: 'POST', body: JSON.stringify(req) })
}

export async function updateReportSchedule(id: string, req: CreateScheduleRequest): Promise<{ data: ReportScheduleDTO; message: string; timestamp: string }> {
  return apiClient(`/api/v1/reports/schedules/${id}`, { method: 'PUT', body: JSON.stringify(req) })
}

export async function deleteReportSchedule(id: string): Promise<{ message: string; timestamp: string }> {
  return apiClient(`/api/v1/reports/schedules/${id}`, { method: 'DELETE' })
}
