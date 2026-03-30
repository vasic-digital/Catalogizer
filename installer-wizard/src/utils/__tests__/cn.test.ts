import { describe, it, expect } from 'vitest'
import { cn } from '../cn'

describe('cn utility (installer-wizard)', () => {
  it('merges basic class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('handles conditional false values', () => {
    expect(cn('base', false && 'hidden', 'visible')).toBe('base visible')
  })

  it('handles conditional true values', () => {
    expect(cn('base', true && 'active')).toBe('base active')
  })

  it('handles undefined values', () => {
    expect(cn('foo', undefined, 'bar')).toBe('foo bar')
  })

  it('handles null values', () => {
    expect(cn('foo', null, 'bar')).toBe('foo bar')
  })

  it('handles empty input', () => {
    expect(cn()).toBe('')
  })

  it('handles single class', () => {
    expect(cn('only')).toBe('only')
  })

  it('resolves tailwind padding conflicts', () => {
    expect(cn('p-4', 'p-6')).toBe('p-6')
  })

  it('resolves tailwind margin conflicts', () => {
    expect(cn('m-2', 'm-4')).toBe('m-4')
  })

  it('resolves tailwind background color conflicts', () => {
    expect(cn('bg-blue-500', 'bg-red-500')).toBe('bg-red-500')
  })

  it('resolves tailwind text color conflicts', () => {
    expect(cn('text-gray-500', 'text-blue-600')).toBe('text-blue-600')
  })

  it('resolves tailwind text size conflicts', () => {
    expect(cn('text-sm', 'text-xl')).toBe('text-xl')
  })

  it('resolves tailwind font weight conflicts', () => {
    expect(cn('font-normal', 'font-semibold')).toBe('font-semibold')
  })

  it('resolves tailwind rounded conflicts', () => {
    expect(cn('rounded-md', 'rounded-lg')).toBe('rounded-lg')
  })

  it('resolves tailwind width conflicts', () => {
    expect(cn('w-full', 'w-64')).toBe('w-64')
  })

  it('resolves tailwind display conflicts', () => {
    expect(cn('block', 'inline-flex')).toBe('inline-flex')
  })

  it('preserves non-conflicting classes', () => {
    expect(cn('p-4', 'm-2', 'text-sm')).toBe('p-4 m-2 text-sm')
  })

  it('handles arrays', () => {
    expect(cn(['foo', 'bar'])).toBe('foo bar')
  })

  it('handles object syntax', () => {
    expect(cn({ active: true, hidden: false })).toBe('active')
  })

  it('handles mixed input types', () => {
    expect(cn('base', ['arr'], { obj: true })).toBe('base arr obj')
  })

  it('preserves responsive prefixes without conflict', () => {
    expect(cn('p-2', 'md:p-4', 'lg:p-6')).toBe('p-2 md:p-4 lg:p-6')
  })

  it('resolves same responsive prefix conflicts', () => {
    expect(cn('md:p-2', 'md:p-4')).toBe('md:p-4')
  })

  it('handles variant styling pattern used in Button component', () => {
    const baseClasses = 'inline-flex items-center justify-center rounded-md text-sm font-medium'
    const variantClasses = 'bg-primary text-primary-foreground hover:bg-primary/90'
    const sizeClasses = 'h-10 px-4 py-2'
    const customClasses = 'w-full'

    const result = cn(baseClasses, variantClasses, sizeClasses, customClasses)

    expect(result).toContain('inline-flex')
    expect(result).toContain('bg-primary')
    expect(result).toContain('w-full')
  })

  it('handles Card component className override pattern', () => {
    const baseClasses = 'rounded-lg border bg-card text-card-foreground shadow-sm'
    const overrideClasses = 'border-yellow-200 bg-yellow-50'

    const result = cn(baseClasses, overrideClasses)

    expect(result).toContain('rounded-lg')
    expect(result).toContain('border-yellow-200')
    expect(result).toContain('bg-yellow-50')
    expect(result).not.toContain('bg-card')
  })
})
