import api from './api'

export interface SyncEndpoint {
  id: number
  name: string
  url: string
  api_key?: string
  enabled: boolean
  last_sync_at?: string
  created_at: string
  updated_at: string
}

export interface CreateEndpointRequest {
  name: string
  url: string
  api_key?: string
  enabled?: boolean
}

export interface UpdateEndpointRequest {
  name?: string
  url?: string
  api_key?: string
  enabled?: boolean
}

export interface SyncSession {
  id: number
  endpoint_id: number
  status: 'running' | 'completed' | 'failed'
  started_at: string
  completed_at?: string
  items_synced: number
  items_failed: number
  error_message?: string
}

export interface SyncStatistics {
  total_syncs: number
  successful_syncs: number
  failed_syncs: number
  total_items_synced: number
  last_sync_at?: string
}

export const syncApi = {
  listEndpoints: (): Promise<SyncEndpoint[]> =>
    api.get('/sync/endpoints').then((res) => res.data),

  getEndpoint: (id: number): Promise<SyncEndpoint> =>
    api.get(`/sync/endpoints/${id}`).then((res) => res.data),

  createEndpoint: (body: CreateEndpointRequest): Promise<SyncEndpoint> =>
    api.post('/sync/endpoints', body).then((res) => res.data),

  updateEndpoint: (id: number, body: UpdateEndpointRequest): Promise<SyncEndpoint> =>
    api.put(`/sync/endpoints/${id}`, body).then((res) => res.data),

  deleteEndpoint: (id: number): Promise<void> =>
    api.delete(`/sync/endpoints/${id}`).then(() => {/* no content */}),

  startSync: (id: number): Promise<SyncSession> =>
    api.post(`/sync/endpoints/${id}/sync`).then((res) => res.data),

  getSessions: (): Promise<SyncSession[]> =>
    api.get('/sync/sessions').then((res) => res.data),

  getSession: (id: number): Promise<SyncSession> =>
    api.get(`/sync/sessions/${id}`).then((res) => res.data),

  getStatistics: (): Promise<SyncStatistics> =>
    api.get('/sync/statistics').then((res) => res.data),

  scheduleSync: (body: { endpoint_id: number; cron_expression: string }): Promise<{ message: string }> =>
    api.post('/sync/schedules', body).then((res) => res.data),

  cleanup: (body?: { older_than_days?: number }): Promise<{ message: string }> =>
    api.post('/sync/cleanup', body).then((res) => res.data),
}
