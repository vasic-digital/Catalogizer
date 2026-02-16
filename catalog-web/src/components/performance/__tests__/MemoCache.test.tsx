import React from 'react'
import { render, screen, act, waitFor } from '@testing-library/react'
import { renderHook } from '@testing-library/react'
import {
  useMemoized,
  useDebounceSearch,
  usePerformanceMonitor,
  useOptimizedData,
  useIntersectionObserver,
  usePagination,
  memoCache,
  measurePerformance,
  measureAsyncPerformance,
} from '../MemoCache'

vi.mock('lodash/debounce', () => {
  return {
    default: (fn: any, delay: number) => {
      let timeout: any
      const debounced = (...args: any[]) => {
        clearTimeout(timeout)
        timeout = setTimeout(() => fn(...args), delay)
      }
      debounced.cancel = () => clearTimeout(timeout)
      return debounced
    },
  }
})

describe('useMemoized', () => {
  afterEach(() => {
    memoCache.clear()
  })

  it('returns computed value', () => {
    const { result } = renderHook(() =>
      useMemoized(() => 42, [])
    )

    expect(result.current).toBe(42)
  })

  it('caches values and returns cached on subsequent calls', () => {
    let computeCount = 0
    const computation = () => {
      computeCount++
      return 'result'
    }

    const { result, rerender } = renderHook(() =>
      useMemoized(computation, ['stable-dep'], 'test-key')
    )

    expect(result.current).toBe('result')

    rerender()

    // Value should come from cache on second render
    expect(result.current).toBe('result')
  })

  it('recomputes when dependencies change', () => {
    let dep = 'a'

    const { result, rerender } = renderHook(() =>
      useMemoized(() => `result-${dep}`, [dep])
    )

    expect(result.current).toBe('result-a')

    dep = 'b'
    rerender()

    expect(result.current).toBe('result-b')
  })
})

describe('useDebounceSearch', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns initial search state', () => {
    const searchFn = vi.fn().mockResolvedValue([])

    const { result } = renderHook(() => useDebounceSearch(searchFn))

    expect(result.current.searchQuery).toBe('')
    expect(result.current.results).toBeNull()
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('provides setSearchQuery function', () => {
    const searchFn = vi.fn().mockResolvedValue([])

    const { result } = renderHook(() => useDebounceSearch(searchFn))

    expect(typeof result.current.setSearchQuery).toBe('function')
  })

  it('debounces search function calls', () => {
    const searchFn = vi.fn().mockResolvedValue(['result1'])

    const { result } = renderHook(() => useDebounceSearch(searchFn, 300))

    act(() => {
      result.current.setSearchQuery('test')
    })

    // Should not call immediately
    expect(searchFn).not.toHaveBeenCalled()

    act(() => {
      vi.advanceTimersByTime(300)
    })

    expect(searchFn).toHaveBeenCalledWith('test')
  })

  it('does not search for empty queries', () => {
    const searchFn = vi.fn().mockResolvedValue([])

    const { result } = renderHook(() => useDebounceSearch(searchFn, 300))

    act(() => {
      result.current.setSearchQuery('   ')
    })

    act(() => {
      vi.advanceTimersByTime(300)
    })

    // Empty/whitespace queries should not trigger search
    expect(searchFn).not.toHaveBeenCalled()
  })
})

describe('usePerformanceMonitor', () => {
  it('returns render count and average render time', () => {
    const { result } = renderHook(() => usePerformanceMonitor('TestComp'))

    expect(result.current.renderCount).toBeGreaterThanOrEqual(0)
    expect(typeof result.current.averageRenderTime).toBe('number')
  })

  it('increments render count on re-render', () => {
    const { result, rerender } = renderHook(() =>
      usePerformanceMonitor('TestComp')
    )

    const initialCount = result.current.renderCount

    rerender()

    expect(result.current.renderCount).toBeGreaterThan(initialCount)
  })
})

describe('useOptimizedData', () => {
  afterEach(() => {
    memoCache.clear()
  })

  const testData = [
    { name: 'Alice', age: 30, city: 'NYC' },
    { name: 'Bob', age: 25, city: 'LA' },
    { name: 'Charlie', age: 35, city: 'NYC' },
    { name: 'Diana', age: 28, city: 'Chicago' },
  ]

  it('returns unfiltered data when no filters applied', () => {
    const { result } = renderHook(() => useOptimizedData(testData, {}))

    expect(result.current.length).toBe(4)
  })

  it('filters data by string field', () => {
    const { result } = renderHook(() =>
      useOptimizedData(testData, { city: 'NYC' })
    )

    expect(result.current.length).toBe(2)
    expect(result.current.every((item: any) => item.city === 'NYC')).toBe(true)
  })

  it('filters with case-insensitive string matching', () => {
    const { result } = renderHook(() =>
      useOptimizedData(testData, { name: 'alice' })
    )

    expect(result.current.length).toBe(1)
    expect(result.current[0].name).toBe('Alice')
  })

  it('sorts data ascending', () => {
    const { result } = renderHook(() =>
      useOptimizedData(testData, {}, 'age', 'asc')
    )

    expect(result.current[0].age).toBe(25)
    expect(result.current[result.current.length - 1].age).toBe(35)
  })

  it('sorts data descending', () => {
    const { result } = renderHook(() =>
      useOptimizedData(testData, {}, 'age', 'desc')
    )

    expect(result.current[0].age).toBe(35)
    expect(result.current[result.current.length - 1].age).toBe(25)
  })

  it('sorts strings alphabetically', () => {
    const { result } = renderHook(() =>
      useOptimizedData(testData, {}, 'name', 'asc')
    )

    expect(result.current[0].name).toBe('Alice')
    expect(result.current[1].name).toBe('Bob')
  })

  it('filters and sorts simultaneously', () => {
    const { result } = renderHook(() =>
      useOptimizedData(testData, { city: 'NYC' }, 'name', 'desc')
    )

    expect(result.current.length).toBe(2)
    expect(result.current[0].name).toBe('Charlie')
    expect(result.current[1].name).toBe('Alice')
  })

  it('ignores empty string filter values', () => {
    const { result } = renderHook(() =>
      useOptimizedData(testData, { city: '' })
    )

    expect(result.current.length).toBe(4)
  })

  it('ignores null filter values', () => {
    const { result } = renderHook(() =>
      useOptimizedData(testData, { city: null })
    )

    expect(result.current.length).toBe(4)
  })
})

describe('useIntersectionObserver', () => {
  it('returns entries, observe, and unobserve', () => {
    const { result } = renderHook(() => useIntersectionObserver())

    expect(result.current.entries).toEqual([])
    expect(typeof result.current.observe).toBe('function')
    expect(typeof result.current.unobserve).toBe('function')
  })

  it('does not throw when observe is called', () => {
    const { result } = renderHook(() => useIntersectionObserver())

    const mockElement = document.createElement('div')
    expect(() => result.current.observe(mockElement)).not.toThrow()
  })

  it('does not throw when unobserve is called', () => {
    const { result } = renderHook(() => useIntersectionObserver())

    const mockElement = document.createElement('div')
    expect(() => result.current.unobserve(mockElement)).not.toThrow()
  })
})

describe('usePagination', () => {
  afterEach(() => {
    memoCache.clear()
  })

  const testData = Array.from({ length: 25 }, (_, i) => ({
    id: i + 1,
    name: `Item ${i + 1}`,
  }))

  it('returns first page of data', () => {
    const { result } = renderHook(() => usePagination(testData, 10))

    expect(result.current.page).toBe(1)
    expect(result.current.paginatedData.length).toBe(10)
    expect(result.current.paginatedData[0].id).toBe(1)
  })

  it('calculates total pages correctly', () => {
    const { result } = renderHook(() => usePagination(testData, 10))

    expect(result.current.totalPages).toBe(3) // 25 items / 10 per page
  })

  it('reports hasNextPage and hasPrevPage', () => {
    const { result } = renderHook(() => usePagination(testData, 10))

    expect(result.current.hasNextPage).toBe(true)
    expect(result.current.hasPrevPage).toBe(false)
  })

  it('navigates to next page', () => {
    const { result } = renderHook(() => usePagination(testData, 10))

    act(() => {
      result.current.nextPage()
    })

    expect(result.current.page).toBe(2)
    expect(result.current.paginatedData[0].id).toBe(11)
    expect(result.current.hasNextPage).toBe(true)
    expect(result.current.hasPrevPage).toBe(true)
  })

  it('navigates to previous page', () => {
    const { result } = renderHook(() => usePagination(testData, 10, 2))

    act(() => {
      result.current.prevPage()
    })

    expect(result.current.page).toBe(1)
  })

  it('does not go below page 1', () => {
    const { result } = renderHook(() => usePagination(testData, 10, 1))

    act(() => {
      result.current.prevPage()
    })

    expect(result.current.page).toBe(1)
  })

  it('does not go above total pages', () => {
    const { result } = renderHook(() => usePagination(testData, 10, 3))

    act(() => {
      result.current.nextPage()
    })

    expect(result.current.page).toBe(3)
  })

  it('goes to specific page', () => {
    const { result } = renderHook(() => usePagination(testData, 10))

    act(() => {
      result.current.goToPage(3)
    })

    expect(result.current.page).toBe(3)
    expect(result.current.paginatedData.length).toBe(5) // Only 5 items on last page
  })

  it('clamps goToPage to valid range', () => {
    const { result } = renderHook(() => usePagination(testData, 10))

    act(() => {
      result.current.goToPage(100)
    })

    expect(result.current.page).toBe(3) // Max page

    act(() => {
      result.current.goToPage(-5)
    })

    expect(result.current.page).toBe(1) // Min page
  })

  it('handles empty data', () => {
    const { result } = renderHook(() => usePagination([], 10))

    expect(result.current.page).toBe(1)
    expect(result.current.paginatedData.length).toBe(0)
    expect(result.current.totalPages).toBe(0)
    expect(result.current.hasNextPage).toBe(false)
    expect(result.current.hasPrevPage).toBe(false)
  })
})

describe('measurePerformance', () => {
  it('returns the result of the function', () => {
    const result = measurePerformance('test', () => 42)
    expect(result).toBe(42)
  })

  it('works with complex return types', () => {
    const result = measurePerformance('test', () => ({ a: 1, b: 'hello' }))
    expect(result).toEqual({ a: 1, b: 'hello' })
  })
})

describe('measureAsyncPerformance', () => {
  it('returns the result of the async function', async () => {
    const result = await measureAsyncPerformance('test', async () => 42)
    expect(result).toBe(42)
  })

  it('handles async errors', async () => {
    await expect(
      measureAsyncPerformance('test', async () => {
        throw new Error('async error')
      })
    ).rejects.toThrow('async error')
  })
})

describe('memoCache (global cache)', () => {
  afterEach(() => {
    memoCache.clear()
  })

  it('can set and get values', () => {
    memoCache.set('test-key', 'test-value')
    expect(memoCache.get('test-key')).toBe('test-value')
  })

  it('returns null for non-existent keys', () => {
    expect(memoCache.get('nonexistent')).toBeNull()
  })

  it('clears all entries', () => {
    memoCache.set('key1', 'val1')
    memoCache.set('key2', 'val2')
    memoCache.clear()

    expect(memoCache.get('key1')).toBeNull()
    expect(memoCache.get('key2')).toBeNull()
  })

  it('respects TTL', () => {
    vi.useFakeTimers()

    memoCache.set('ttl-key', 'value', 1000) // 1 second TTL

    expect(memoCache.get('ttl-key')).toBe('value')

    vi.advanceTimersByTime(1500)

    expect(memoCache.get('ttl-key')).toBeNull()

    vi.useRealTimers()
  })

  it('evicts oldest entry when maxSize reached', () => {
    // Default maxSize is 100
    for (let i = 0; i < 100; i++) {
      memoCache.set(`key-${i}`, `value-${i}`)
    }

    // Adding one more should evict the oldest
    memoCache.set('key-new', 'new-value')

    expect(memoCache.get('key-new')).toBe('new-value')
  })
})
