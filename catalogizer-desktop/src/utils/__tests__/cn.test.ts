import { describe, it, expect } from 'vitest'
import { cn } from '../cn'

describe('cn utility (extended)', () => {
  it('merges multiple class names', () => {
    expect(cn('foo', 'bar', 'baz')).toBe('foo bar baz')
  })

  it('handles conditional class names with false', () => {
    expect(cn('foo', false && 'bar', 'baz')).toBe('foo baz')
  })

  it('handles conditional class names with true', () => {
    expect(cn('foo', true && 'bar', 'baz')).toBe('foo bar baz')
  })

  it('handles undefined and null values', () => {
    expect(cn('foo', undefined, null, 'bar')).toBe('foo bar')
  })

  it('resolves tailwind padding conflicts', () => {
    expect(cn('p-4', 'p-2')).toBe('p-2')
  })

  it('resolves tailwind margin conflicts', () => {
    expect(cn('m-4', 'm-2')).toBe('m-2')
  })

  it('keeps non-conflicting tailwind classes', () => {
    expect(cn('p-4', 'm-2')).toBe('p-4 m-2')
  })

  it('resolves tailwind background color conflicts', () => {
    expect(cn('bg-red-500', 'bg-blue-500')).toBe('bg-blue-500')
  })

  it('resolves tailwind text color conflicts', () => {
    expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500')
  })

  it('resolves tailwind text size conflicts', () => {
    expect(cn('text-sm', 'text-lg')).toBe('text-lg')
  })

  it('resolves tailwind font weight conflicts', () => {
    expect(cn('font-normal', 'font-bold')).toBe('font-bold')
  })

  it('handles empty input', () => {
    expect(cn()).toBe('')
  })

  it('handles single class name', () => {
    expect(cn('foo')).toBe('foo')
  })

  it('handles arrays of class names', () => {
    expect(cn(['foo', 'bar'])).toBe('foo bar')
  })

  it('handles nested arrays', () => {
    expect(cn(['foo', ['bar', 'baz']])).toBe('foo bar baz')
  })

  it('handles object syntax for conditional classes', () => {
    expect(cn({ foo: true, bar: false, baz: true })).toBe('foo baz')
  })

  it('handles mixed arrays and objects', () => {
    expect(cn('base', ['arr1', 'arr2'], { obj: true })).toBe('base arr1 arr2 obj')
  })

  it('resolves tailwind display conflicts', () => {
    expect(cn('block', 'flex')).toBe('flex')
  })

  it('resolves tailwind width conflicts', () => {
    expect(cn('w-full', 'w-1/2')).toBe('w-1/2')
  })

  it('resolves tailwind rounded conflicts', () => {
    expect(cn('rounded-sm', 'rounded-lg')).toBe('rounded-lg')
  })

  it('handles responsive prefix classes without conflict', () => {
    expect(cn('p-2', 'md:p-4')).toBe('p-2 md:p-4')
  })
})
