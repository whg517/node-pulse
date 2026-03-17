import '@testing-library/jest-dom'
import { vi } from 'vitest'

const noisyConsolePatterns = [
  'i18next is maintained with support from Locize',
  '[ECharts] Can\'t get DOM width or height.',
  '[ECharts] Specified `grid.containLabel` but no `use(LegacyGridContainLabel)`',
  '[ECharts] Component toolbox is used but not imported.',
]

function shouldSuppressConsoleOutput(args: unknown[]): boolean {
  const message = args
    .map((value) => (typeof value === 'string' ? value : value instanceof Error ? value.message : String(value)))
    .join(' ')

  return noisyConsolePatterns.some((pattern) => message.includes(pattern))
}

const originalConsoleError = console.error.bind(console)
const originalConsoleWarn = console.warn.bind(console)
const originalConsoleLog = console.log.bind(console)
const originalConsoleInfo = console.info.bind(console)
const originalConsoleDebug = console.debug.bind(console)

console.error = (...args: unknown[]) => {
  if (shouldSuppressConsoleOutput(args)) {
    return
  }
  originalConsoleError(...args)
}

console.warn = (...args: unknown[]) => {
  if (shouldSuppressConsoleOutput(args)) {
    return
  }
  originalConsoleWarn(...args)
}

console.log = (...args: unknown[]) => {
  if (shouldSuppressConsoleOutput(args)) {
    return
  }
  originalConsoleLog(...args)
}

console.info = (...args: unknown[]) => {
  if (shouldSuppressConsoleOutput(args)) {
    return
  }
  originalConsoleInfo(...args)
}

console.debug = (...args: unknown[]) => {
  if (shouldSuppressConsoleOutput(args)) {
    return
  }
  originalConsoleDebug(...args)
}

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next')

  return {
    ...actual,
    useTranslation: () => ({
      t: (key: string, options?: Record<string, unknown>) => {
        if (options && 'defaultValue' in options && typeof options.defaultValue === 'string') {
          return options.defaultValue
        }
        return key
      },
      i18n: {
        language: 'en',
        changeLanguage: vi.fn(),
      },
    }),
  }
})

// Mock localStorage for i18n and settings
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
    get length() {
      return Object.keys(store).length
    },
    key: vi.fn((index: number) => Object.keys(store)[index] || null),
  }
})()

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
})

// Mock window.matchMedia for theme detection
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock HTMLCanvasElement.getContext for ECharts
HTMLCanvasElement.prototype.getContext = vi.fn(function (this: HTMLCanvasElement) {
  return {
    fillRect: vi.fn(),
    clearRect: vi.fn(),
    getImageData: vi.fn(),
    putImageData: vi.fn(),
    createImageData: vi.fn(),
    setTransform: vi.fn(),
    drawImage: vi.fn(),
    save: vi.fn(),
    fillText: vi.fn(),
    restore: vi.fn(),
    beginPath: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    closePath: vi.fn(),
    stroke: vi.fn(),
    translate: vi.fn(),
    scale: vi.fn(),
    rotate: vi.fn(),
    arc: vi.fn(),
    fill: vi.fn(),
    measureText: vi.fn(function () {
      return { width: 0 }
    }),
    transform: vi.fn(),
    rect: vi.fn(),
    clip: vi.fn(),
    bezierCurveTo: vi.fn(),
    quadraticCurveTo: vi.fn(),
    ellipse: vi.fn(),
    arcTo: vi.fn(),
  }
}) as any

// Suppress unhandled rejections from error testing
window.addEventListener('unhandledrejection', (event) => {
  // Prevent test failures from expected errors in error-handling tests
  if (event.reason instanceof Error) {
    event.preventDefault()
  }
})
