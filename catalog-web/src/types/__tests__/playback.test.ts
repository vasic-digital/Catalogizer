import type { UiPlaybackProgress, UiPlaybackSession } from '../playback'

describe('playback types', () => {
  describe('UiPlaybackProgress', () => {
    it('accepts a fully populated progress object', () => {
      const progress: UiPlaybackProgress = {
        mediaItemId: 1,
        positionUnit: 'seconds',
        durationTotal: 7200,
        lastPosition: 3600,
        lastSessionAmount: 1800,
        totalReproductions: 5,
        aggregateAmount: 9000,
        lastSessionEndedAt: '2026-04-10T15:30:00Z',
      }
      expect(progress.mediaItemId).toBe(1)
      expect(progress.positionUnit).toBe('seconds')
      expect(progress.durationTotal).toBe(7200)
      expect(progress.lastPosition).toBe(3600)
      expect(progress.lastSessionAmount).toBe(1800)
      expect(progress.totalReproductions).toBe(5)
      expect(progress.aggregateAmount).toBe(9000)
      expect(progress.lastSessionEndedAt).toBe('2026-04-10T15:30:00Z')
    })

    it('accepts null for durationTotal', () => {
      const progress: UiPlaybackProgress = {
        mediaItemId: 2,
        positionUnit: 'pages',
        durationTotal: null,
        lastPosition: 50,
        lastSessionAmount: 10,
        totalReproductions: 1,
        aggregateAmount: 50,
        lastSessionEndedAt: null,
      }
      expect(progress.durationTotal).toBeNull()
      expect(progress.lastSessionEndedAt).toBeNull()
    })

    it('accepts different position units', () => {
      const units = ['seconds', 'pages', 'events']
      for (const unit of units) {
        const progress: UiPlaybackProgress = {
          mediaItemId: 3,
          positionUnit: unit,
          durationTotal: 100,
          lastPosition: 50,
          lastSessionAmount: 10,
          totalReproductions: 1,
          aggregateAmount: 50,
          lastSessionEndedAt: null,
        }
        expect(progress.positionUnit).toBe(unit)
      }
    })
  })

  describe('UiPlaybackSession', () => {
    it('accepts a fully populated session object', () => {
      const session: UiPlaybackSession = {
        id: 1,
        positionUnit: 'seconds',
        startPosition: 0,
        endPosition: 1800,
        totalAmount: 1800,
        startedAt: '2026-04-10T14:00:00Z',
        endedAt: '2026-04-10T14:30:00Z',
        completed: true,
      }
      expect(session.id).toBe(1)
      expect(session.positionUnit).toBe('seconds')
      expect(session.startPosition).toBe(0)
      expect(session.endPosition).toBe(1800)
      expect(session.totalAmount).toBe(1800)
      expect(session.startedAt).toBe('2026-04-10T14:00:00Z')
      expect(session.endedAt).toBe('2026-04-10T14:30:00Z')
      expect(session.completed).toBe(true)
    })

    it('accepts null for endedAt when session is in progress', () => {
      const session: UiPlaybackSession = {
        id: 2,
        positionUnit: 'seconds',
        startPosition: 1800,
        endPosition: 3600,
        totalAmount: 1800,
        startedAt: '2026-04-10T15:00:00Z',
        endedAt: null,
        completed: false,
      }
      expect(session.endedAt).toBeNull()
      expect(session.completed).toBe(false)
    })

    it('supports page-based sessions', () => {
      const session: UiPlaybackSession = {
        id: 3,
        positionUnit: 'pages',
        startPosition: 100,
        endPosition: 150,
        totalAmount: 50,
        startedAt: '2026-01-01T00:00:00Z',
        endedAt: '2026-01-01T01:00:00Z',
        completed: true,
      }
      expect(session.positionUnit).toBe('pages')
      expect(session.totalAmount).toBe(50)
    })

    it('supports zero-length sessions', () => {
      const session: UiPlaybackSession = {
        id: 4,
        positionUnit: 'seconds',
        startPosition: 0,
        endPosition: 0,
        totalAmount: 0,
        startedAt: '2026-01-01T00:00:00Z',
        endedAt: '2026-01-01T00:00:00Z',
        completed: false,
      }
      expect(session.startPosition).toBe(0)
      expect(session.endPosition).toBe(0)
      expect(session.totalAmount).toBe(0)
    })
  })
})
