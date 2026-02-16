import { describe, it, expect } from 'vitest'
import { cn } from '../../utils/cn'

describe('cn utility', () => {
  it('merges class names', () => {
    const result = cn('foo', 'bar')
    expect(result).toBe('foo bar')
  })

  it('handles conditional class names with false', () => {
    const result = cn('foo', false && 'bar', 'baz')
    expect(result).toBe('foo baz')
  })

  it('handles undefined and null values', () => {
    const result = cn('foo', undefined, null, 'bar')
    expect(result).toBe('foo bar')
  })

  it('merges tailwind conflicting classes (padding)', () => {
    const result = cn('p-4', 'p-2')
    expect(result).toBe('p-2')
  })

  it('handles empty input', () => {
    const result = cn()
    expect(result).toBe('')
  })

  it('handles arrays of class names', () => {
    const result = cn(['foo', 'bar'])
    expect(result).toBe('foo bar')
  })

  it('merges tailwind conflicting background colors', () => {
    const result = cn('bg-red-500', 'bg-blue-500')
    expect(result).toBe('bg-blue-500')
  })

  it('handles object syntax for conditional classes', () => {
    const result = cn({ foo: true, bar: false, baz: true })
    expect(result).toBe('foo baz')
  })

  it('merges tailwind font size classes', () => {
    const result = cn('text-sm', 'text-lg')
    expect(result).toBe('text-lg')
  })

  it('preserves non-conflicting tailwind classes', () => {
    const result = cn('text-red-500', 'bg-blue-500', 'p-4')
    expect(result).toBe('text-red-500 bg-blue-500 p-4')
  })

  it('merges margin classes correctly', () => {
    const result = cn('mt-2', 'mt-4')
    expect(result).toBe('mt-4')
  })
})
