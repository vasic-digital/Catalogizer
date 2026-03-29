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
    api.get('/stats/overall').then((res) => res.data),

  getFileTypeStats: (): Promise<FileTypeStat[]> =>
    api.get('/stats/filetypes').then((res) => res.data),

  getSizeDistribution: (): Promise<SizeDistribution[]> =>
    api.get('/stats/sizes').then((res) => res.data),

  getDuplicateStats: (): Promise<DuplicateStats> =>
    api.get('/stats/duplicates').then((res) => res.data),

  getTopDuplicateGroups: (limit = 10): Promise<DuplicateGroup[]> =>
    api.get('/stats/duplicates/groups', { params: { limit } }).then((res) => res.data),

  getAccessPatterns: (): Promise<AccessPattern[]> =>
    api.get('/stats/access').then((res) => res.data),

  getGrowthTrends: (days = 30): Promise<GrowthTrend[]> =>
    api.get('/stats/growth', { params: { days } }).then((res) => res.data),

  getScanHistory: (limit = 20): Promise<ScanHistoryEntry[]> =>
    api.get('/stats/scans', { params: { limit } }).then((res) => res.data),
}
