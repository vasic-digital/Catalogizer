import {
  cn,
  formatDate,
  formatRelativeTime,
  truncateText,
  capitalizeFirst,
  debounce,
} from '../utils'

describe('utils', () => {
  describe('cn', () => {
    it('merges class names correctly', () => {
      expect(cn('class1', 'class2')).toBe('class1 class2')
      expect(cn('class1', { class2: true })).toBe('class1 class2')
      expect(cn('class1', { class2: false })).toBe('class1')
    })

    it('handles conflicting Tailwind classes', () => {
      expect(cn('px-2', 'px-4')).toBe('px-4')
      expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500')
    })

    it('handles empty inputs', () => {
      expect(cn()).toBe('')
      expect(cn('', 'class1')).toBe('class1')
    })
  })

  describe('formatDate', () => {
    it('formats date string correctly', () => {
      const date = '2024-01-15'
      const result = formatDate(date)
      expect(result).toMatch(/January 15, 2024/)
    })

    it('formats Date object correctly', () => {
      const date = new Date('2024-01-15')
      const result = formatDate(date)
      expect(result).toMatch(/January 15, 2024/)
    })
  })

  describe('formatRelativeTime', () => {
    beforeEach(() => {
      vi.useFakeTimers()
      vi.setSystemTime(new Date('2024-01-15T12:00:00Z'))
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('returns "Just now" for very recent dates', () => {
      const recent = new Date('2024-01-15T11:59:30Z')
      expect(formatRelativeTime(recent)).toBe('Just now')
    })

    it('returns minutes ago for recent dates', () => {
      const recent = new Date('2024-01-15T11:45:00Z')
      expect(formatRelativeTime(recent)).toBe('15m ago')
    })

    it('returns hours ago for dates within 24 hours', () => {
      const recent = new Date('2024-01-15T08:00:00Z')
      expect(formatRelativeTime(recent)).toBe('4h ago')
    })

    it('returns days ago for dates within a week', () => {
      const recent = new Date('2024-01-13T12:00:00Z')
      expect(formatRelativeTime(recent)).toBe('2d ago')
    })

    it('returns formatted date for older dates', () => {
      const old = new Date('2024-01-01T12:00:00Z')
      const result = formatRelativeTime(old)
      expect(result).toMatch(/January 1, 2024/)
    })
  })

  describe('truncateText', () => {
    it('returns original text if shorter than maxLength', () => {
      expect(truncateText('hello', 10)).toBe('hello')
    })

    it('truncates text and adds ellipsis if longer than maxLength', () => {
      expect(truncateText('hello world', 8)).toBe('hello wo...')
    })

    it('handles exact maxLength', () => {
      expect(truncateText('hello', 5)).toBe('hello')
    })

    it('handles empty string', () => {
      expect(truncateText('', 5)).toBe('')
    })
  })

  describe('capitalizeFirst', () => {
    it('capitalizes first letter of string', () => {
      expect(capitalizeFirst('hello')).toBe('Hello')
      expect(capitalizeFirst('HELLO')).toBe('HELLO')
    })

    it('handles empty string', () => {
      expect(capitalizeFirst('')).toBe('')
    })

    it('handles single character', () => {
      expect(capitalizeFirst('a')).toBe('A')
    })
  })

  describe('debounce', () => {
    beforeEach(() => {
      vi.useFakeTimers()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('delays function execution', () => {
      const mockFn = vi.fn()
      const debouncedFn = debounce(mockFn, 100)

      debouncedFn()
      expect(mockFn).not.toHaveBeenCalled()

      vi.advanceTimersByTime(100)
      expect(mockFn).toHaveBeenCalledTimes(1)
    })

    it('resets delay on subsequent calls', () => {
      const mockFn = vi.fn()
      const debouncedFn = debounce(mockFn, 100)

      debouncedFn()
      vi.advanceTimersByTime(50)
      debouncedFn()
      vi.advanceTimersByTime(50)
      expect(mockFn).not.toHaveBeenCalled()

      vi.advanceTimersByTime(50)
      expect(mockFn).toHaveBeenCalledTimes(1)
    })

    it('calls function with correct arguments', () => {
      const mockFn = vi.fn()
      const debouncedFn = debounce(mockFn, 100)

      debouncedFn('arg1', 'arg2')
      vi.advanceTimersByTime(100)

      expect(mockFn).toHaveBeenCalledWith('arg1', 'arg2')
    })

    it('handles multiple rapid calls', () => {
      const mockFn = vi.fn()
      const debouncedFn = debounce(mockFn, 100)

      debouncedFn()
      debouncedFn()
      debouncedFn()

      vi.advanceTimersByTime(100)

      expect(mockFn).toHaveBeenCalledTimes(1)
    })
  })
})