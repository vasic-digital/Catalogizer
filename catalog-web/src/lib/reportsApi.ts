import api from './api'

export interface UsageReport {
  period: string
  total_users: number
  active_users: number
  total_media_accessed: number
  total_downloads: number
  storage_growth_bytes: number
  top_media: { id: number; title: string; access_count: number }[]
  user_activity: { user_id: number; username: string; actions: number }[]
}

export interface PerformanceReport {
  period: string
  avg_response_time_ms: number
  p95_response_time_ms: number
  p99_response_time_ms: number
  total_requests: number
  error_rate: number
  uptime_percentage: number
  slowest_endpoints: { path: string; avg_ms: number; count: number }[]
}

export const reportsApi = {
  getUsageReport: (params?: { from?: string; to?: string }): Promise<UsageReport> =>
    api.get('/reports/usage', { params }).then((res) => res.data),

  getPerformanceReport: (params?: { from?: string; to?: string }): Promise<PerformanceReport> =>
    api.get('/reports/performance', { params }).then((res) => res.data),
}
