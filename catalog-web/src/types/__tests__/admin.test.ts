import type {
  SystemInfo,
  User,
  StorageInfo,
  Backup,
} from '../admin'

describe('admin types', () => {
  describe('SystemInfo interface', () => {
    it('has all required fields', () => {
      const info: SystemInfo = {
        version: '1.0.0',
        uptime: 86400,
        cpuUsage: 45.2,
        memoryUsage: 67.8,
        diskUsage: {
          total: 1000000000000,
          used: 500000000000,
          free: 500000000000,
        },
        activeConnections: 12,
        totalRequests: 150000,
      }
      expect(info.version).toBe('1.0.0')
      expect(info.uptime).toBe(86400)
      expect(info.cpuUsage).toBe(45.2)
      expect(info.memoryUsage).toBe(67.8)
      expect(info.diskUsage.total).toBe(1000000000000)
      expect(info.diskUsage.used).toBe(500000000000)
      expect(info.diskUsage.free).toBe(500000000000)
      expect(info.activeConnections).toBe(12)
      expect(info.totalRequests).toBe(150000)
    })

    it('diskUsage contains total, used, and free', () => {
      const info: SystemInfo = {
        version: '2.0.0',
        uptime: 0,
        cpuUsage: 0,
        memoryUsage: 0,
        diskUsage: { total: 0, used: 0, free: 0 },
        activeConnections: 0,
        totalRequests: 0,
      }
      expect(info.diskUsage).toHaveProperty('total')
      expect(info.diskUsage).toHaveProperty('used')
      expect(info.diskUsage).toHaveProperty('free')
    })
  })

  describe('User interface', () => {
    it('has all required fields', () => {
      const user: User = {
        id: 'user-1',
        username: 'admin',
        email: 'admin@example.com',
        role: 'admin',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
      }
      expect(user.id).toBe('user-1')
      expect(user.username).toBe('admin')
      expect(user.email).toBe('admin@example.com')
      expect(user.role).toBe('admin')
      expect(user.status).toBe('active')
      expect(user.lastLogin).toBeUndefined()
    })

    it('accepts all role values', () => {
      const admin: User = { id: '1', username: 'a', email: 'a@b.c', role: 'admin', status: 'active', createdAt: '' }
      const regular: User = { id: '2', username: 'b', email: 'b@b.c', role: 'user', status: 'active', createdAt: '' }
      const viewer: User = { id: '3', username: 'c', email: 'c@b.c', role: 'viewer', status: 'active', createdAt: '' }
      expect(admin.role).toBe('admin')
      expect(regular.role).toBe('user')
      expect(viewer.role).toBe('viewer')
    })

    it('accepts all status values', () => {
      const active: User = { id: '1', username: 'a', email: 'a@b.c', role: 'admin', status: 'active', createdAt: '' }
      const inactive: User = { id: '2', username: 'b', email: 'b@b.c', role: 'user', status: 'inactive', createdAt: '' }
      const suspended: User = { id: '3', username: 'c', email: 'c@b.c', role: 'viewer', status: 'suspended', createdAt: '' }
      expect(active.status).toBe('active')
      expect(inactive.status).toBe('inactive')
      expect(suspended.status).toBe('suspended')
    })

    it('accepts optional lastLogin', () => {
      const user: User = {
        id: 'user-1',
        username: 'admin',
        email: 'admin@example.com',
        role: 'admin',
        status: 'active',
        lastLogin: '2024-06-01T12:00:00Z',
        createdAt: '2024-01-01T00:00:00Z',
      }
      expect(user.lastLogin).toBe('2024-06-01T12:00:00Z')
    })
  })

  describe('StorageInfo interface', () => {
    it('has all required fields', () => {
      const storage: StorageInfo = {
        path: '/media/nas',
        totalSpace: 8000000000000,
        usedSpace: 5000000000000,
        availableSpace: 3000000000000,
        mediaCount: 85000,
      }
      expect(storage.path).toBe('/media/nas')
      expect(storage.totalSpace).toBe(8000000000000)
      expect(storage.mediaCount).toBe(85000)
      expect(storage.lastScan).toBeUndefined()
    })

    it('accepts optional lastScan', () => {
      const storage: StorageInfo = {
        path: '/media/local',
        totalSpace: 1000000000000,
        usedSpace: 500000000000,
        availableSpace: 500000000000,
        mediaCount: 1000,
        lastScan: '2024-06-01T00:00:00Z',
      }
      expect(storage.lastScan).toBe('2024-06-01T00:00:00Z')
    })
  })

  describe('Backup interface', () => {
    it('has all required fields', () => {
      const backup: Backup = {
        id: 'backup-1',
        filename: 'catalogizer-backup-2024-06-01.tar.gz',
        size: 2500000000,
        createdAt: '2024-06-01T00:00:00Z',
        type: 'full',
        status: 'completed',
      }
      expect(backup.id).toBe('backup-1')
      expect(backup.filename).toContain('backup')
      expect(backup.size).toBe(2500000000)
      expect(backup.type).toBe('full')
      expect(backup.status).toBe('completed')
    })

    it('accepts all type values', () => {
      const full: Backup = { id: '1', filename: 'f', size: 0, createdAt: '', type: 'full', status: 'completed' }
      const incremental: Backup = { id: '2', filename: 'f', size: 0, createdAt: '', type: 'incremental', status: 'completed' }
      expect(full.type).toBe('full')
      expect(incremental.type).toBe('incremental')
    })

    it('accepts all status values', () => {
      const completed: Backup = { id: '1', filename: 'f', size: 0, createdAt: '', type: 'full', status: 'completed' }
      const inProgress: Backup = { id: '2', filename: 'f', size: 0, createdAt: '', type: 'full', status: 'in-progress' }
      const failed: Backup = { id: '3', filename: 'f', size: 0, createdAt: '', type: 'full', status: 'failed' }
      expect(completed.status).toBe('completed')
      expect(inProgress.status).toBe('in-progress')
      expect(failed.status).toBe('failed')
    })
  })
})
