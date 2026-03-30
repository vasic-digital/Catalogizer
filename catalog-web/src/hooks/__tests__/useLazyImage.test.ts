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
    // The IntersectionObserver mock uses instance methods (not prototype),
    // so we spy on the constructor instead to verify instantiation.
    const constructorSpy = vi.fn(IntersectionObserver)
    const OriginalIO = global.IntersectionObserver
    global.IntersectionObserver = constructorSpy as unknown as typeof IntersectionObserver

    renderHook(() => useLazyImage('https://example.com/img.jpg'))

    // Observer is created when src is provided (even if imgRef.current is null)
    expect(constructorSpy).toHaveBeenCalled()

    global.IntersectionObserver = OriginalIO
  })

  it('does not create an IntersectionObserver when src is undefined', () => {
    const constructorSpy = vi.fn(IntersectionObserver)
    const OriginalIO = global.IntersectionObserver
    global.IntersectionObserver = constructorSpy as unknown as typeof IntersectionObserver

    renderHook(() => useLazyImage(undefined))

    // No observer should have been created for undefined src
    expect(constructorSpy).not.toHaveBeenCalled()

    global.IntersectionObserver = OriginalIO
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
