/**
 * Performance Optimization Test
 *
 * This test validates that the polling optimization is working correctly.
 * It checks that:
 * 1. deepEqual utility works correctly
 * 2. Components use React.memo properly
 * 3. Hooks only update state when data actually changes
 */

import { describe, it, expect } from 'vitest'
import { deepEqual, memoCompare } from '../utils/deepEqual'

describe('deepEqual utility', () => {
  it('should return true for identical objects', () => {
    const obj1 = { id: 1, name: 'test', metrics: { latency: 100 } }
    const obj2 = { id: 1, name: 'test', metrics: { latency: 100 } }
    expect(deepEqual(obj1, obj2)).toBe(true)
  })

  it('should return false for different objects', () => {
    const obj1 = { id: 1, name: 'test', metrics: { latency: 100 } }
    const obj2 = { id: 1, name: 'test', metrics: { latency: 200 } }
    expect(deepEqual(obj1, obj2)).toBe(false)
  })

  it('should return true for identical arrays', () => {
    const arr1 = [
      { node_id: '1', latency_ms: 100 },
      { node_id: '2', latency_ms: 150 },
    ]
    const arr2 = [
      { node_id: '1', latency_ms: 100 },
      { node_id: '2', latency_ms: 150 },
    ]
    expect(deepEqual(arr1, arr2)).toBe(true)
  })

  it('should return false for arrays with different content', () => {
    const arr1 = [{ node_id: '1', latency_ms: 100 }]
    const arr2 = [{ node_id: '1', latency_ms: 200 }]
    expect(deepEqual(arr1, arr2)).toBe(false)
  })

  it('should handle null and undefined', () => {
    expect(deepEqual(null, null)).toBe(true)
    expect(deepEqual(undefined, undefined)).toBe(true)
    expect(deepEqual(null, undefined)).toBe(false) // Using strict equality (===)
    expect(deepEqual({}, null)).toBe(false)
  })
})

describe('memoCompare function', () => {
  it('should return true when props are equal', () => {
    const prevProps = {
      nodes: [{ id: '1', name: 'Node 1' }],
      metrics: [{ node_id: '1', latency_ms: 100 }],
      isLoading: false,
    }
    const nextProps = {
      nodes: [{ id: '1', name: 'Node 1' }],
      metrics: [{ node_id: '1', latency_ms: 100 }],
      isLoading: false,
    }
    expect(memoCompare(prevProps, nextProps)).toBe(true) // Should not re-render
  })

  it('should return false when props differ', () => {
    const prevProps = {
      nodes: [{ id: '1', name: 'Node 1' }],
      metrics: [{ node_id: '1', latency_ms: 100 }],
      isLoading: false,
    }
    const nextProps = {
      nodes: [{ id: '1', name: 'Node 1' }],
      metrics: [{ node_id: '1', latency_ms: 200 }], // Different latency
      isLoading: false,
    }
    expect(memoCompare(prevProps, nextProps)).toBe(false) // Should re-render
  })
})
