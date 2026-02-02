import { renderHook, act } from '@testing-library/react'
import { usePlayerState } from '../usePlayerState'

// Mock the PlaylistItem type
const createMockItem = (overrides: Partial<any> = {}): any => ({
  id: `item-${Math.random().toString(36).substr(2, 5)}`,
  playlist_id: 'playlist-1',
  media_id: 1,
  position: 0,
  media_item: {
    id: 1,
    title: 'Test Track',
    media_type: 'music',
  },
  added_at: '2024-01-01T00:00:00Z',
  ...overrides,
})

describe('usePlayerState', () => {
  describe('initial state', () => {
    it('returns correct initial values', () => {
      const { result } = renderHook(() => usePlayerState())

      expect(result.current.currentItem).toBeNull()
      expect(result.current.isPlaying).toBe(false)
      expect(result.current.isPaused).toBe(false)
      expect(result.current.volume).toBe(0.8)
      expect(result.current.isMuted).toBe(false)
      expect(result.current.currentTime).toBe(0)
      expect(result.current.duration).toBe(0)
      expect(result.current.playbackRate).toBe(1)
      expect(result.current.isFullscreen).toBe(false)
      expect(result.current.repeatMode).toBe('off')
      expect(result.current.isShuffled).toBe(false)
      expect(result.current.queue).toEqual([])
      expect(result.current.queueIndex).toBe(0)
    })
  })

  describe('play and pause', () => {
    it('play sets isPlaying to true and isPaused to false', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.play()
      })

      expect(result.current.isPlaying).toBe(true)
      expect(result.current.isPaused).toBe(false)
    })

    it('pause sets isPlaying to false and isPaused to true', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.play()
      })
      act(() => {
        result.current.pause()
      })

      expect(result.current.isPlaying).toBe(false)
      expect(result.current.isPaused).toBe(true)
    })

    it('togglePlayPause toggles between play and pause', () => {
      const { result } = renderHook(() => usePlayerState())

      // Initially not playing, toggle should play
      act(() => {
        result.current.togglePlayPause()
      })
      expect(result.current.isPlaying).toBe(true)
      expect(result.current.isPaused).toBe(false)

      // Now playing, toggle should pause
      act(() => {
        result.current.togglePlayPause()
      })
      expect(result.current.isPlaying).toBe(false)
      expect(result.current.isPaused).toBe(true)
    })
  })

  describe('volume control', () => {
    it('setVolume updates the volume', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.setVolume(0.5)
      })

      expect(result.current.volume).toBe(0.5)
    })

    it('setIsMuted toggles mute state', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.setIsMuted(true)
      })

      expect(result.current.isMuted).toBe(true)
    })
  })

  describe('seekTo', () => {
    it('sets the current time', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.seekTo(120.5)
      })

      expect(result.current.currentTime).toBe(120.5)
    })
  })

  describe('playback rate', () => {
    it('setPlaybackRate updates the rate', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.setPlaybackRate(1.5)
      })

      expect(result.current.playbackRate).toBe(1.5)
    })
  })

  describe('repeat and shuffle', () => {
    it('setRepeatMode changes the repeat mode', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.setRepeatMode('one')
      })
      expect(result.current.repeatMode).toBe('one')

      act(() => {
        result.current.setRepeatMode('all')
      })
      expect(result.current.repeatMode).toBe('all')

      act(() => {
        result.current.setRepeatMode('off')
      })
      expect(result.current.repeatMode).toBe('off')
    })

    it('setIsShuffled changes the shuffle state', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.setIsShuffled(true)
      })

      expect(result.current.isShuffled).toBe(true)
    })
  })

  describe('queue management', () => {
    it('addToQueue appends an item to the queue', () => {
      const { result } = renderHook(() => usePlayerState())
      const item = createMockItem({ id: 'item-1' })

      act(() => {
        result.current.addToQueue(item)
      })

      expect(result.current.queue).toHaveLength(1)
      expect(result.current.queue[0]).toEqual(item)
    })

    it('addToQueue appends multiple items sequentially', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })
      const item2 = createMockItem({ id: 'item-2' })

      act(() => {
        result.current.addToQueue(item1)
      })
      act(() => {
        result.current.addToQueue(item2)
      })

      expect(result.current.queue).toHaveLength(2)
    })

    it('removeFromQueue removes an item by index', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })
      const item2 = createMockItem({ id: 'item-2' })
      const item3 = createMockItem({ id: 'item-3' })

      act(() => {
        result.current.setQueue([item1, item2, item3])
      })
      act(() => {
        result.current.removeFromQueue(1)
      })

      expect(result.current.queue).toHaveLength(2)
      expect(result.current.queue[0].id).toBe('item-1')
      expect(result.current.queue[1].id).toBe('item-3')
    })

    it('clearQueue empties the queue, resets index, nullifies currentItem and pauses', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })
      const item2 = createMockItem({ id: 'item-2' })

      act(() => {
        result.current.setQueue([item1, item2])
        result.current.setCurrentItem(item1)
        result.current.setQueueIndex(1)
        result.current.play()
      })
      act(() => {
        result.current.clearQueue()
      })

      expect(result.current.queue).toEqual([])
      expect(result.current.queueIndex).toBe(0)
      expect(result.current.currentItem).toBeNull()
      expect(result.current.isPaused).toBe(true)
      expect(result.current.isPlaying).toBe(false)
    })
  })

  describe('navigation - playNext', () => {
    it('advances to the next item in the queue', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })
      const item2 = createMockItem({ id: 'item-2' })
      const item3 = createMockItem({ id: 'item-3' })

      act(() => {
        result.current.setQueue([item1, item2, item3])
        result.current.setQueueIndex(0)
        result.current.setCurrentItem(item1)
      })
      act(() => {
        result.current.playNext()
      })

      expect(result.current.queueIndex).toBe(1)
      expect(result.current.currentItem).toEqual(item2)
      expect(result.current.isPlaying).toBe(true)
    })

    it('wraps around to start when repeatMode is "all" and at end of queue', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })
      const item2 = createMockItem({ id: 'item-2' })

      act(() => {
        result.current.setQueue([item1, item2])
        result.current.setQueueIndex(1)
        result.current.setCurrentItem(item2)
        result.current.setRepeatMode('all')
      })
      act(() => {
        result.current.playNext()
      })

      expect(result.current.queueIndex).toBe(0)
      expect(result.current.currentItem).toEqual(item1)
      expect(result.current.isPlaying).toBe(true)
    })

    it('pauses when at end of queue and repeatMode is "off"', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })

      act(() => {
        result.current.setQueue([item1])
        result.current.setQueueIndex(0)
        result.current.setCurrentItem(item1)
        result.current.play()
      })
      act(() => {
        result.current.playNext()
      })

      expect(result.current.isPlaying).toBe(false)
      expect(result.current.isPaused).toBe(true)
    })
  })

  describe('navigation - playPrevious', () => {
    it('goes to the previous item in the queue', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })
      const item2 = createMockItem({ id: 'item-2' })

      act(() => {
        result.current.setQueue([item1, item2])
        result.current.setQueueIndex(1)
        result.current.setCurrentItem(item2)
      })
      act(() => {
        result.current.playPrevious()
      })

      expect(result.current.queueIndex).toBe(0)
      expect(result.current.currentItem).toEqual(item1)
      expect(result.current.isPlaying).toBe(true)
    })

    it('wraps around to end when repeatMode is "all" and at start of queue', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })
      const item2 = createMockItem({ id: 'item-2' })
      const item3 = createMockItem({ id: 'item-3' })

      act(() => {
        result.current.setQueue([item1, item2, item3])
        result.current.setQueueIndex(0)
        result.current.setCurrentItem(item1)
        result.current.setRepeatMode('all')
      })
      act(() => {
        result.current.playPrevious()
      })

      expect(result.current.queueIndex).toBe(2)
      expect(result.current.currentItem).toEqual(item3)
      expect(result.current.isPlaying).toBe(true)
    })

    it('does nothing when at start of queue and repeatMode is "off"', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })

      act(() => {
        result.current.setQueue([item1])
        result.current.setQueueIndex(0)
        result.current.setCurrentItem(item1)
      })
      act(() => {
        result.current.playPrevious()
      })

      expect(result.current.queueIndex).toBe(0)
      expect(result.current.currentItem).toEqual(item1)
    })
  })

  describe('skipTo', () => {
    it('jumps to a specific index in the queue and starts playing', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })
      const item2 = createMockItem({ id: 'item-2' })
      const item3 = createMockItem({ id: 'item-3' })

      act(() => {
        result.current.setQueue([item1, item2, item3])
      })
      act(() => {
        result.current.skipTo(2)
      })

      expect(result.current.queueIndex).toBe(2)
      expect(result.current.currentItem).toEqual(item3)
      expect(result.current.isPlaying).toBe(true)
    })

    it('does nothing for out-of-bounds index (negative)', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })

      act(() => {
        result.current.setQueue([item1])
        result.current.setQueueIndex(0)
      })
      act(() => {
        result.current.skipTo(-1)
      })

      expect(result.current.queueIndex).toBe(0)
    })

    it('does nothing for out-of-bounds index (too large)', () => {
      const { result } = renderHook(() => usePlayerState())
      const item1 = createMockItem({ id: 'item-1' })

      act(() => {
        result.current.setQueue([item1])
        result.current.setQueueIndex(0)
      })
      act(() => {
        result.current.skipTo(5)
      })

      expect(result.current.queueIndex).toBe(0)
    })
  })

  describe('fullscreen', () => {
    it('setIsFullscreen updates fullscreen state', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.setIsFullscreen(true)
      })

      expect(result.current.isFullscreen).toBe(true)
    })
  })

  describe('duration', () => {
    it('setDuration updates the duration', () => {
      const { result } = renderHook(() => usePlayerState())

      act(() => {
        result.current.setDuration(300)
      })

      expect(result.current.duration).toBe(300)
    })
  })
})
