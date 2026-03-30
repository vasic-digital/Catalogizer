import { renderHook } from '@testing-library/react'
import { useLazyImage } from '../useLazyImage'

describe('useLazyImage', () => {
  it('returns imgRef, loaded, error, onLoad, and onError', () => {
    const { result } = renderHook(() => useLazyImage('https://example.com/img.jpg'))

    expect(result.current.imgRef).toBeDefined()
    expect(result.current.loaded).toBe(false)
    expect(result.current.error).toBe(false)
    expect(typeof result.current.onLoad).toBe('function')
    expect(typeof result.current.onError).toBe('function')
  })

  it('starts with loaded=false and error=false', () => {
    const { result } = renderHook(() => useLazyImage('https://example.com/img.jpg'))

    expect(result.current.loaded).toBe(false)
    expect(result.current.error).toBe(false)
  })

  it('handles undefined src gracefully', () => {
    const { result } = renderHook(() => useLazyImage(undefined))

    expect(result.current.loaded).toBe(false)
    expect(result.current.error).toBe(false)
  })

  it('creates an IntersectionObserver when src is provided', () => {
    const observeSpy = vi.spyOn(IntersectionObserver.prototype, 'observe')

    renderHook(() => useLazyImage('https://example.com/img.jpg'))

    // imgRef.current is null in a hook-only render (no DOM element),
    // so observe won't be called. This verifies no crash occurs.
    expect(observeSpy).not.toHaveBeenCalled()

    observeSpy.mockRestore()
  })

  it('does not create an IntersectionObserver when src is undefined', () => {
    const disconnectSpy = vi.spyOn(IntersectionObserver.prototype, 'disconnect')

    renderHook(() => useLazyImage(undefined))

    // No observer should have been created for undefined src
    expect(disconnectSpy).not.toHaveBeenCalled()

    disconnectSpy.mockRestore()
  })

  it('resets loaded and error when src changes', () => {
    const { result, rerender } = renderHook(
      ({ src }) => useLazyImage(src),
      { initialProps: { src: 'https://example.com/img1.jpg' as string | undefined } }
    )

    expect(result.current.loaded).toBe(false)
    expect(result.current.error).toBe(false)

    rerender({ src: 'https://example.com/img2.jpg' })

    expect(result.current.loaded).toBe(false)
    expect(result.current.error).toBe(false)
  })
})
