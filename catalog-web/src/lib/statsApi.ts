import api from './api'

export interface OverallStats {
  total_files: number
  total_size: number
  total_directories: number
  total_storage_roots: number
  last_scan_at?: string
}

export interface FileTypeStat {
  extension: string
  count: number
  total_size: number
}

export interface SizeDistribution {
  range: string
  count: number
  total_size: number
}

export interface DuplicateStats {
  total_duplicate_groups: number
  total_duplicate_files: number
  wasted_space: number
}

export interface DuplicateGroup {
  hash: string
  count: number
  file_size: number
  paths: string[]
}

export interface GrowthTrend {
  date: string
  files_added: number
  size_added: number
  cumulative_files: number
  cumulative_size: number
}

export interface ScanHistoryEntry {
  id: number
  storage_root_id: number
  storage_root_name: string
  started_at: string
  completed_at?: string
  status: string
  files_found: number
  files_new: number
  files_updated: number
  errors: number
  duration_ms: number
}

export interface AccessPattern {
  hour: number
  count: number
}

export const statsApi = {
  getOverallStats: (): Promise<OverallStats> =>
    api.get('/stats/overall').then((res) => {
      const payload = res.data?.data || res.data
      return {
        total_files: payload.total_files ?? 0,
        total_size: payload.total_size ?? 0,
        total_directories: payload.total_directories ?? 0,
        total_storage_roots: payload.storage_roots_count ?? payload.total_storage_roots ?? 0,
        last_scan_at: payload.last_scan_time
          ? new Date(payload.last_scan_time * 1000).toISOString()
          : payload.last_scan_at,
      }
    }),

  getFileTypeStats: (): Promise<FileTypeStat[]> =>
    api.get('/stats/filetypes').then((res) => {
      const payload = res.data?.data || res.data
      return Array.isArray(payload) ? payload : []
    }),

  getSizeDistribution: (): Promise<SizeDistribution[]> =>
    api.get('/stats/sizes').then((res) => {
      const payload = res.data?.data || res.data
      return Array.isArray(payload) ? payload : []
    }),

  getDuplicateStats: (): Promise<DuplicateStats> =>
    api.get('/stats/duplicates').then((res) => {
      const payload = res.data?.data || res.data
      return {
        total_duplicate_groups: payload.duplicate_groups ?? payload.total_duplicate_groups ?? 0,
        total_duplicate_files: payload.total_duplicates ?? payload.total_duplicate_files ?? 0,
        wasted_space: payload.wasted_space ?? 0,
      }
    }),

  getTopDuplicateGroups: (limit = 10): Promise<DuplicateGroup[]> =>
    api.get('/stats/duplicates/groups', { params: { limit } }).then((res) => {
      const payload = res.data?.data || res.data
      return Array.isArray(payload) ? payload : []
    }),

  getAccessPatterns: (): Promise<AccessPattern[]> =>
    api.get('/stats/access').then((res) => {
      const payload = res.data?.data || res.data
      return Array.isArray(payload) ? payload : []
    }),

  getGrowthTrends: (days = 30): Promise<GrowthTrend[]> =>
    api.get('/stats/growth', { params: { days } }).then((res) => {
      const payload = res.data?.data || res.data
      // API returns { monthly_growth: [...] } instead of a flat array
      const trends = payload.monthly_growth || payload
      return Array.isArray(trends) ? trends : []
    }),

  getScanHistory: (limit = 20): Promise<ScanHistoryEntry[]> =>
    api.get('/stats/scans', { params: { limit } }).then((res) => {
      const payload = res.data?.data || res.data
      // API returns { scans: [...], total_count, limit, offset }
      const scans = payload.scans || payload
      return Array.isArray(scans) ? scans : []
    }),
}
