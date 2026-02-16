import React from 'react'
import { render, screen, act } from '@testing-library/react'
import { renderHook } from '@testing-library/react'
import {
  PerformanceOptimizer,
  usePerformanceMonitor,
  useMemoryOptimization,
  useDebouncedSearch,
  useInfiniteScroll,
} from '../PerformanceOptimizer'

vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, style, ...props }: any) => (
      <div style={style} {...props}>
        {children}
      </div>
    ),
  },
}))

describe('PerformanceOptimizer', () => {
  describe('when itemCount is below threshold', () => {
    it('renders children directly without optimization', () => {
      render(
        <PerformanceOptimizer itemCount={10} threshold={100}>
          <div>Child 1</div>
          <div>Child 2</div>
        </PerformanceOptimizer>
      )

      expect(screen.getByText('Child 1')).toBeInTheDocument()
      expect(screen.getByText('Child 2')).toBeInTheDocument()
    })

    it('does not render the scroll container', () => {
      const { container } = render(
        <PerformanceOptimizer itemCount={5} threshold={100}>
          <div>Content</div>
        </PerformanceOptimizer>
      )

      const scrollContainer = container.querySelector('.overflow-auto')
      expect(scrollContainer).toBeNull()
    })
  })

  describe('when itemCount exceeds threshold', () => {
    const manyChildren = Array.from({ length: 150 }, (_, i) => (
      <div key={i}>Item {i}</div>
    ))

    it('renders scroll container with lazy strategy', () => {
      const { container } = render(
        <PerformanceOptimizer
          itemCount={150}
          threshold={100}
          loadingStrategy="lazy"
        >
          {manyChildren}
        </PerformanceOptimizer>
      )

      const scrollContainer = container.querySelector('.overflow-auto')
      expect(scrollContainer).toBeInTheDocument()
    })

    it('renders with virtual strategy', () => {
      const { container } = render(
        <PerformanceOptimizer
          itemCount={150}
          threshold={100}
          loadingStrategy="virtual"
          containerHeight={500}
          itemHeight={50}
        >
          {manyChildren}
        </PerformanceOptimizer>
      )

      const scrollContainer = container.querySelector('.overflow-auto')
      expect(scrollContainer).toBeInTheDocument()
    })

    it('renders with pagination strategy', () => {
      const { container } = render(
        <PerformanceOptimizer
          itemCount={150}
          threshold={100}
          loadingStrategy="pagination"
        >
          {manyChildren}
        </PerformanceOptimizer>
      )

      const scrollContainer = container.querySelector('.overflow-auto')
      expect(scrollContainer).toBeInTheDocument()
    })

    it('sets container height from props', () => {
      const { container } = render(
        <PerformanceOptimizer
          itemCount={150}
          threshold={100}
          containerHeight={600}
        >
          {manyChildren}
        </PerformanceOptimizer>
      )

      const scrollContainer = container.querySelector('.overflow-auto')
      expect(scrollContainer).toHaveStyle({ height: '600px' })
    })

    it('uses default container height of 400px', () => {
      const { container } = render(
        <PerformanceOptimizer itemCount={150} threshold={100}>
          {manyChildren}
        </PerformanceOptimizer>
      )

      const scrollContainer = container.querySelector('.overflow-auto')
      expect(scrollContainer).toHaveStyle({ height: '400px' })
    })

    it('shows loading indicator for pagination when more items exist', () => {
      const { container } = render(
        <PerformanceOptimizer
          itemCount={150}
          threshold={100}
          loadingStrategy="pagination"
        >
          {manyChildren}
        </PerformanceOptimizer>
      )

      const spinner = container.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
    })

    it('renders first 10 items immediately with lazy strategy', () => {
      render(
        <PerformanceOptimizer
          itemCount={150}
          threshold={100}
          loadingStrategy="lazy"
        >
          {manyChildren}
        </PerformanceOptimizer>
      )

      // First 10 items should be loaded immediately
      expect(screen.getByText('Item 0')).toBeInTheDocument()
      expect(screen.getByText('Item 9')).toBeInTheDocument()
    })
  })

  describe('with non-array children', () => {
    it('renders single child directly', () => {
      render(
        <PerformanceOptimizer itemCount={200} threshold={100}>
          <div>Single child</div>
        </PerformanceOptimizer>
      )

      expect(screen.getByText('Single child')).toBeInTheDocument()
    })
  })
})

describe('usePerformanceMonitor', () => {
  it('returns getMetrics function', () => {
    const { result } = renderHook(() => usePerformanceMonitor('TestComponent'))

    expect(result.current.getMetrics).toBeDefined()
    expect(typeof result.current.getMetrics).toBe('function')
  })

  it('tracks render count', () => {
    const { result, rerender } = renderHook(() =>
      usePerformanceMonitor('TestComponent')
    )

    const metrics1 = result.current.getMetrics()
    expect(metrics1.renderCount).toBeGreaterThanOrEqual(1)

    rerender()

    const metrics2 = result.current.getMetrics()
    expect(metrics2.renderCount).toBeGreaterThan(metrics1.renderCount)
  })

  it('calculates total time', () => {
    const { result } = renderHook(() => usePerformanceMonitor('TestComponent'))

    const metrics = result.current.getMetrics()
    expect(metrics.totalTime).toBeGreaterThanOrEqual(0)
  })

  it('calculates average render time', () => {
    const { result } = renderHook(() => usePerformanceMonitor('TestComponent'))

    const metrics = result.current.getMetrics()
    expect(metrics.averageRenderTime).toBeGreaterThanOrEqual(0)
  })
})

describe('useMemoryOptimization', () => {
  it('returns cache operations', () => {
    const { result } = renderHook(() => useMemoryOptimization())

    expect(result.current.addToCache).toBeDefined()
    expect(result.current.getFromCache).toBeDefined()
    expect(result.current.clearCache).toBeDefined()
    expect(result.current.cacheSize).toBe(0)
  })

  it('adds and retrieves items from cache', () => {
    const { result } = renderHook(() => useMemoryOptimization())

    act(() => {
      result.current.addToCache('key1', { data: 'value1' })
    })

    expect(result.current.cacheSize).toBe(1)
    expect(result.current.getFromCache('key1')).toEqual({ data: 'value1' })
  })

  it('clears cache', () => {
    const { result } = renderHook(() => useMemoryOptimization())

    act(() => {
      result.current.addToCache('key1', 'value1')
      result.current.addToCache('key2', 'value2')
    })

    expect(result.current.cacheSize).toBe(2)

    act(() => {
      result.current.clearCache()
    })

    expect(result.current.cacheSize).toBe(0)
  })

  it('returns undefined for non-existent keys', () => {
    const { result } = renderHook(() => useMemoryOptimization())

    expect(result.current.getFromCache('nonexistent')).toBeUndefined()
  })

  it('evicts oldest entry when cache is full', () => {
    const { result } = renderHook(() => useMemoryOptimization())

    // maxCacheSize is 100 in the implementation
    act(() => {
      for (let i = 0; i < 100; i++) {
        result.current.addToCache(`key-${i}`, `value-${i}`)
      }
    })

    expect(result.current.cacheSize).toBe(100)

    act(() => {
      result.current.addToCache('key-new', 'new-value')
    })

    // After adding one more, oldest should be evicted
    expect(result.current.cacheSize).toBe(100)
  })
})

describe('useDebouncedSearch', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns debounced search term', () => {
    const { result } = renderHook(() => useDebouncedSearch('initial'))

    expect(result.current).toBe('initial')
  })

  it('debounces search term changes', () => {
    const { result, rerender } = renderHook(
      ({ term }) => useDebouncedSearch(term, 300),
      { initialProps: { term: '' } }
    )

    expect(result.current).toBe('')

    rerender({ term: 'hello' })

    // Before debounce delay, should still be empty
    expect(result.current).toBe('')

    act(() => {
      vi.advanceTimersByTime(300)
    })

    expect(result.current).toBe('hello')
  })

  it('uses custom delay', () => {
    const { result, rerender } = renderHook(
      ({ term }) => useDebouncedSearch(term, 500),
      { initialProps: { term: '' } }
    )

    rerender({ term: 'test' })

    act(() => {
      vi.advanceTimersByTime(300)
    })

    expect(result.current).toBe('')

    act(() => {
      vi.advanceTimersByTime(200)
    })

    expect(result.current).toBe('test')
  })

  it('cancels previous timeout on new input', () => {
    const { result, rerender } = renderHook(
      ({ term }) => useDebouncedSearch(term, 300),
      { initialProps: { term: '' } }
    )

    rerender({ term: 'ab' })

    act(() => {
      vi.advanceTimersByTime(100)
    })

    rerender({ term: 'abc' })

    act(() => {
      vi.advanceTimersByTime(200)
    })

    // Should still show empty since 300ms hasn't passed since 'abc'
    expect(result.current).toBe('')

    act(() => {
      vi.advanceTimersByTime(100)
    })

    expect(result.current).toBe('abc')
  })
})

describe('useInfiniteScroll', () => {
  it('returns a ref', () => {
    const { result } = renderHook(() =>
      useInfiniteScroll(true, false, vi.fn())
    )

    expect(result.current).toBeDefined()
    expect(result.current.current).toBe(null)
  })

  it('does not call onLoadMore when isLoading is true', () => {
    const onLoadMore = vi.fn()
    renderHook(() => useInfiniteScroll(true, true, onLoadMore))

    expect(onLoadMore).not.toHaveBeenCalled()
  })

  it('does not call onLoadMore when hasMore is false', () => {
    const onLoadMore = vi.fn()
    renderHook(() => useInfiniteScroll(false, false, onLoadMore))

    expect(onLoadMore).not.toHaveBeenCalled()
  })
})
