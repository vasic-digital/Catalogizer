export interface ActivityItem {
  id: string
  type: 'media_played' | 'media_uploaded' | 'media_downloaded' | 'user_login' | 'system_event' | 'media_viewed'
  title: string
  user: string
  timestamp: Date
  metadata: Record<string, any>
}

export interface DashboardStats {
  total_media: number
  total_users: number
  active_users: number
  storage_used: number
  recent_additions: number
  avg_quality_score: string
  sessions_today: number
  avg_session_duration: number
}

export interface MediaStats {
  total_items: number
  total_size: number
  recent_additions: number
  by_type: Record<string, number>
  by_quality: Record<string, number>
}

export interface UserStats {
  active_users: number
  total_users: number
  sessions_today: number
  avg_session_duration: number
}

export interface SystemStatus {
  status: 'healthy' | 'warning' | 'error'
  cpu_usage: number
  memory_usage: number
  disk_usage: number
  network_status: 'online' | 'offline'
  last_backup: Date | null
  uptime: string
}

export interface QuickAction {
  id: string
  title: string
  description: string
  icon: React.ComponentType<any>
  action: () => void
  variant?: 'default' | 'secondary' | 'outline' | 'ghost'
}