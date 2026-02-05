/**
 * Deep comparison utilities for preventing unnecessary re-renders
 */

/**
 * Deeply compares two objects or arrays for equality
 * Handles nested objects, arrays, and primitive values
 */
export function deepEqual(a: any, b: any): boolean {
  // Primitive values or same reference
  if (a === b) return true

  // Null or undefined checks
  if (a == null || b == null) return a === b

  // Different types
  if (typeof a !== typeof b) return false

  // Arrays
  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) return false
    return a.every((item, index) => deepEqual(item, b[index]))
  }

  // Objects
  if (typeof a === 'object' && typeof b === 'object') {
    const keysA = Object.keys(a)
    const keysB = Object.keys(b)

    if (keysA.length !== keysB.length) return false

    return keysA.every(key => deepEqual(a[key], b[key]))
  }

  return false
}

/**
 * Custom comparison function for React.memo that deeply compares props
 * Use this for complex props (objects/arrays) instead of shallow comparison
 */
export function memoCompare<T extends Record<string, any>>(
  prevProps: T,
  nextProps: T
): boolean {
  const keys = Object.keys(nextProps) as Array<keyof T>

  for (const key of keys) {
    if (!deepEqual(prevProps[key], nextProps[key])) {
      return false; // Props are different, should re-render
    }
  }

  return true; // All props are equal, skip re-render
}
