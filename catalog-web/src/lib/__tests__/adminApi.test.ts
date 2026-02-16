import { adminApi } from '../adminApi'

describe('adminApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getSystemInfo', () => {
    it('returns system information with expected fields', async () => {
      const result = await adminApi.getSystemInfo()

      expect(result).toHaveProperty('version')
      expect(result).toHaveProperty('uptime')
      expect(result).toHaveProperty('cpuUsage')
      expect(result).toHaveProperty('memoryUsage')
      expect(result).toHaveProperty('diskUsage')
      expect(result).toHaveProperty('activeConnections')
      expect(result).toHaveProperty('totalRequests')
    })

    it('returns disk usage with total, used, and free', async () => {
      const result = await adminApi.getSystemInfo()

      expect(result.diskUsage).toHaveProperty('total')
      expect(result.diskUsage).toHaveProperty('used')
      expect(result.diskUsage).toHaveProperty('free')
      expect(result.diskUsage.total).toBe(result.diskUsage.used + result.diskUsage.free)
    })

    it('returns numeric values for CPU and memory usage', async () => {
      const result = await adminApi.getSystemInfo()

      expect(typeof result.cpuUsage).toBe('number')
      expect(typeof result.memoryUsage).toBe('number')
      expect(result.cpuUsage).toBeGreaterThanOrEqual(0)
      expect(result.cpuUsage).toBeLessThanOrEqual(100)
    })
  })

  describe('getUsers', () => {
    it('returns an array of users', async () => {
      const result = await adminApi.getUsers()

      expect(Array.isArray(result)).toBe(true)
      expect(result.length).toBeGreaterThan(0)
    })

    it('returns users with required fields', async () => {
      const result = await adminApi.getUsers()
      const user = result[0]

      expect(user).toHaveProperty('id')
      expect(user).toHaveProperty('username')
      expect(user).toHaveProperty('email')
      expect(user).toHaveProperty('role')
      expect(user).toHaveProperty('status')
      expect(user).toHaveProperty('createdAt')
    })

    it('includes an admin user', async () => {
      const result = await adminApi.getUsers()
      const admin = result.find(u => u.role === 'admin')

      expect(admin).toBeDefined()
      expect(admin!.username).toBe('admin')
    })
  })

  describe('getStorageInfo', () => {
    it('returns an array of storage info objects', async () => {
      const result = await adminApi.getStorageInfo()

      expect(Array.isArray(result)).toBe(true)
      expect(result.length).toBeGreaterThan(0)
    })

    it('returns storage info with space metrics', async () => {
      const result = await adminApi.getStorageInfo()
      const storage = result[0]

      expect(storage).toHaveProperty('path')
      expect(storage).toHaveProperty('totalSpace')
      expect(storage).toHaveProperty('usedSpace')
      expect(storage).toHaveProperty('availableSpace')
      expect(storage).toHaveProperty('mediaCount')
      expect(typeof storage.totalSpace).toBe('number')
    })
  })

  describe('getBackups', () => {
    it('returns an array of backup objects', async () => {
      const result = await adminApi.getBackups()

      expect(Array.isArray(result)).toBe(true)
      expect(result.length).toBeGreaterThan(0)
    })

    it('returns backups with required fields', async () => {
      const result = await adminApi.getBackups()
      const backup = result[0]

      expect(backup).toHaveProperty('id')
      expect(backup).toHaveProperty('filename')
      expect(backup).toHaveProperty('size')
      expect(backup).toHaveProperty('createdAt')
      expect(backup).toHaveProperty('type')
      expect(backup).toHaveProperty('status')
    })

    it('returns backups with valid type values', async () => {
      const result = await adminApi.getBackups()
      result.forEach(backup => {
        expect(['full', 'incremental']).toContain(backup.type)
      })
    })
  })

  describe('createBackup', () => {
    it('completes without error for full backup', async () => {
      await expect(adminApi.createBackup('full')).resolves.toBeUndefined()
    })

    it('completes without error for incremental backup', async () => {
      await expect(adminApi.createBackup('incremental')).resolves.toBeUndefined()
    })
  })

  describe('restoreBackup', () => {
    it('completes without error', async () => {
      await expect(adminApi.restoreBackup('1')).resolves.toBeUndefined()
    })
  })

  describe('scanStorage', () => {
    it('completes without error', async () => {
      await expect(adminApi.scanStorage('/media/movies')).resolves.toBeUndefined()
    })
  })

  describe('updateUser', () => {
    it('completes without error', async () => {
      await expect(
        adminApi.updateUser('1', { role: 'admin' } as any)
      ).resolves.toBeUndefined()
    })
  })
})
