import { renderHook } from '@testing-library/react'
import { useVirtualScroll } from '../useVirtualScroll'

describe('useVirtualScroll', () => {
  const items = Array.from({ length: 100 }, (_, i) => ({ id: i, name: `Item ${i}` }))

  it('returns containerRef, visibleItems, totalHeight, offsetY, and startIndex', () => {
    const { result } = renderHook(() =>
      useVirtualScroll(items, { itemHeight: 40 })
    )

    expect(result.current.containerRef).toBeDefined()
    expect(result.current.visibleItems).toBeDefined()
    expect(typeof result.current.totalHeight).toBe('number')
    expect(typeof result.current.offsetY).toBe('number')
    expect(typeof result.current.startIndex).toBe('number')
  })

  it('calculates totalHeight correctly', () => {
    const { result } = renderHook(() =>
      useVirtualScroll(items, { itemHeight: 40 })
    )

    expect(result.current.totalHeight).toBe(100 * 40)
  })

  it('starts with startIndex of 0', () => {
    const { result } = renderHook(() =>
      useVirtualScroll(items, { itemHeight: 40 })
    )

    expect(result.current.startIndex).toBe(0)
  })

  it('starts with offsetY of 0', () => {
    const { result } = renderHook(() =>
      useVirtualScroll(items, { itemHeight: 40 })
    )

    expect(result.current.offsetY).toBe(0)
  })

  it('handles empty items array', () => {
    const { result } = renderHook(() =>
      useVirtualScroll([], { itemHeight: 40 })
    )

    expect(result.current.totalHeight).toBe(0)
    expect(result.current.visibleItems).toEqual([])
    expect(result.current.startIndex).toBe(0)
    expect(result.current.offsetY).toBe(0)
  })

  it('uses default overscan of 5', () => {
    const { result } = renderHook(() =>
      useVirtualScroll(items, { itemHeight: 40 })
    )

    // With scrollTop=0, containerHeight=0 (no DOM), overscan=5:
    // startIndex = max(0, floor(0/40) - 5) = 0
    // endIndex = min(100, ceil((0+0)/40) + 5) = 5
    // visibleItems should be items[0..4]
    expect(result.current.visibleItems).toHaveLength(5)
    expect(result.current.visibleItems[0]).toEqual(items[0])
    expect(result.current.visibleItems[4]).toEqual(items[4])
  })

  it('respects custom overscan value', () => {
    const { result } = renderHook(() =>
      useVirtualScroll(items, { itemHeight: 40, overscan: 10 })
    )

    // With scrollTop=0, containerHeight=0, overscan=10:
    // endIndex = min(100, ceil(0/40) + 10) = 10
    expect(result.current.visibleItems).toHaveLength(10)
  })

  it('updates when items change', () => {
    const { result, rerender } = renderHook(
      ({ items: hookItems }) => useVirtualScroll(hookItems, { itemHeight: 40 }),
      { initialProps: { items } }
    )

    expect(result.current.totalHeight).toBe(4000)

    const fewerItems = items.slice(0, 20)
    rerender({ items: fewerItems })

    expect(result.current.totalHeight).toBe(800)
  })

  it('handles single item', () => {
    const singleItem = [{ id: 0, name: 'Only item' }]
    const { result } = renderHook(() =>
      useVirtualScroll(singleItem, { itemHeight: 50 })
    )

    expect(result.current.totalHeight).toBe(50)
    expect(result.current.visibleItems).toHaveLength(1)
    expect(result.current.visibleItems[0]).toEqual(singleItem[0])
  })
})
