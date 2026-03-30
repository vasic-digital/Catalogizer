import type {
  ActivityItem,
  DashboardStats,
  MediaStats,
  UserStats,
  SystemStatus,
} from '../dashboard'

describe('dashboard types', () => {
  describe('ActivityItem interface', () => {
    it('has all required fields', () => {
      const item: ActivityItem = {
        id: 'act-1',
        type: 'media_played',
        title: 'Played The Matrix',
        user: 'admin',
        timestamp: new Date('2024-06-01T12:00:00Z'),
        metadata: {},
      }
      expect(item.id).toBe('act-1')
      expect(item.type).toBe('media_played')
      expect(item.title).toBe('Played The Matrix')
      expect(item.user).toBe('admin')
      expect(item.timestamp).toBeInstanceOf(Date)
      expect(item.metadata).toEqual({})
    })

    it('accepts all type values', () => {
      const types: ActivityItem['type'][] = [
        'media_played', 'media_uploaded', 'media_downloaded',
        'user_login', 'system_event', 'media_viewed',
      ]
      expect(types).toHaveLength(6)
    })

    it('accepts metadata with arbitrary fields', () => {
      const item: ActivityItem = {
        id: 'act-2',
        type: 'media_uploaded',
        title: 'Uploaded file.mp4',
        user: 'user1',
        timestamp: new Date(),
        metadata: {
          file_size: 1500000000,
          format: 'mp4',
          quality: '1080p',
        },
      }
      expect(item.metadata.file_size).toBe(1500000000)
      expect(item.metadata.format).toBe('mp4')
    })
  })

  describe('DashboardStats interface', () => {
    it('has all required fields', () => {
      const stats: DashboardStats = {
        total_media: 5000,
        total_users: 25,
        active_users: 10,
        storage_used: 500000000000,
        recent_additions: 42,
        avg_quality_score: '8.5',
        sessions_today: 15,
        avg_session_duration: 1800,
      }
      expect(stats.total_media).toBe(5000)
      expect(stats.total_users).toBe(25)
      expect(stats.active_users).toBe(10)
      expect(stats.storage_used).toBe(500000000000)
      expect(stats.recent_additions).toBe(42)
      expect(stats.avg_quality_score).toBe('8.5')
      expect(stats.sessions_today).toBe(15)
      expect(stats.avg_session_duration).toBe(1800)
    })
  })

  describe('MediaStats interface', () => {
    it('has all required fields', () => {
      const stats: MediaStats = {
        total_items: 1234,
        total_size: 5368709120,
        recent_additions: 25,
        by_type: { movie: 500, tv_show: 300 },
        by_quality: { '1080p': 400, '720p': 350 },
      }
      expect(stats.total_items).toBe(1234)
      expect(stats.total_size).toBe(5368709120)
      expect(stats.by_type.movie).toBe(500)
      expect(stats.by_quality['1080p']).toBe(400)
    })

    it('accepts empty record maps', () => {
      const stats: MediaStats = {
        total_items: 0,
        total_size: 0,
        recent_additions: 0,
        by_type: {},
        by_quality: {},
      }
      expect(Object.keys(stats.by_type)).toHaveLength(0)
      expect(Object.keys(stats.by_quality)).toHaveLength(0)
    })
  })

  describe('UserStats interface', () => {
    it('has all required fields', () => {
      const stats: UserStats = {
        active_users: 10,
        total_users: 25,
        sessions_today: 15,
        avg_session_duration: 1800,
      }
      expect(stats.active_users).toBe(10)
      expect(stats.total_users).toBe(25)
      expect(stats.sessions_today).toBe(15)
      expect(stats.avg_session_duration).toBe(1800)
    })
  })

  describe('SystemStatus interface', () => {
    it('has all required fields', () => {
      const status: SystemStatus = {
        status: 'healthy',
        cpu_usage: 25.5,
        memory_usage: 60.2,
        disk_usage: 45.0,
        network_status: 'online',
        last_backup: new Date('2024-06-01T00:00:00Z'),
        uptime: '30 days 5 hours',
      }
      expect(status.status).toBe('healthy')
      expect(status.cpu_usage).toBe(25.5)
      expect(status.memory_usage).toBe(60.2)
      expect(status.disk_usage).toBe(45.0)
      expect(status.network_status).toBe('online')
      expect(status.last_backup).toBeInstanceOf(Date)
      expect(status.uptime).toBe('30 days 5 hours')
    })

    it('accepts all status values', () => {
      const statuses: SystemStatus['status'][] = ['healthy', 'warning', 'error']
      expect(statuses).toHaveLength(3)
    })

    it('accepts all network_status values', () => {
      const netStatuses: SystemStatus['network_status'][] = ['online', 'offline']
      expect(netStatuses).toHaveLength(2)
    })

    it('accepts null for last_backup', () => {
      const status: SystemStatus = {
        status: 'warning',
        cpu_usage: 80,
        memory_usage: 90,
        disk_usage: 95,
        network_status: 'offline',
        last_backup: null,
        uptime: '1 hour',
      }
      expect(status.last_backup).toBeNull()
    })
  })
})
