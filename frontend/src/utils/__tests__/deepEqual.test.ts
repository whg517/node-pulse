import { describe, it, expect } from 'vitest'
import { deepEqual, memoCompare } from '../deepEqual'

describe('deepEqual', () => {
  it('returns true for identical primitives', () => {
    expect(deepEqual(1, 1)).toBe(true)
    expect(deepEqual('a', 'a')).toBe(true)
    expect(deepEqual(true, true)).toBe(true)
  })

  it('returns true for same reference', () => {
    const obj = { a: 1 }
    expect(deepEqual(obj, obj)).toBe(true)
  })

  it('returns false for null vs undefined', () => {
    expect(deepEqual(null, undefined)).toBe(false)
  })

  it('returns true for both null', () => {
    expect(deepEqual(null, null)).toBe(true)
  })

  it('returns false for different primitive types', () => {
    expect(deepEqual(1, '1')).toBe(false)
  })

  it('returns false for arrays of different lengths', () => {
    expect(deepEqual([1, 2], [1, 2, 3])).toBe(false)
  })

  it('returns true for equal arrays', () => {
    expect(deepEqual([1, 2, 3], [1, 2, 3])).toBe(true)
  })

  it('returns false for arrays with different elements', () => {
    expect(deepEqual([1, 2], [1, 3])).toBe(false)
  })

  it('returns false for objects with different number of keys', () => {
    expect(deepEqual({ a: 1, b: 2 }, { a: 1 })).toBe(false)
  })

  it('returns true for equal nested objects', () => {
    expect(deepEqual({ a: { b: 1 } }, { a: { b: 1 } })).toBe(true)
  })

  it('returns false for objects with different values', () => {
    expect(deepEqual({ a: 1 }, { a: 2 })).toBe(false)
  })

  it('returns false for value vs null', () => {
    expect(deepEqual(1, null)).toBe(false)
  })
})

describe('memoCompare', () => {
  it('returns true when all props are deeply equal', () => {
    const prev = { count: 1, data: [1, 2, 3] }
    const next = { count: 1, data: [1, 2, 3] }
    expect(memoCompare(prev, next)).toBe(true)
  })

  it('returns false when a prop differs', () => {
    const prev = { count: 1, data: [1, 2, 3] }
    const next = { count: 2, data: [1, 2, 3] }
    expect(memoCompare(prev, next)).toBe(false)
  })

  it('returns false when nested prop differs', () => {
    const prev = { info: { name: 'Alice' } }
    const next = { info: { name: 'Bob' } }
    expect(memoCompare(prev, next)).toBe(false)
  })
})
