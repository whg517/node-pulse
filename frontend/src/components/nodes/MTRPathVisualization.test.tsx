import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import MTRPathVisualization, { MTRHop } from './MTRPathVisualization'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}))

describe('MTRPathVisualization', () => {
  const createMockHop = (overrides: Partial<MTRHop> = {}): MTRHop => ({
    hopNumber: 1,
    ip: '192.168.1.1',
    sent: 10,
    received: 10,
    lossRate: 0,
    avgRTTMs: 50,
    bestRTTMs: 30,
    worstRTTMs: 80,
    stdDevMs: 10,
    ...overrides,
  })

  it('renders hops', () => {
    const hops = [createMockHop()]
    render(<MTRPathVisualization hops={hops} />)
    expect(screen.getByText('192.168.1.1')).toBeInTheDocument()
  })

  it('renders empty state', () => {
    render(<MTRPathVisualization hops={[]} />)
    expect(screen.getByText('mtr.noHopData')).toBeInTheDocument()
  })

  // Helper: jsdom can't parse CSS attribute selectors with nested brackets from CSS var classes
  const findHopWithClass = (container: HTMLElement, cls: string) =>
    Array.from(container.querySelectorAll('[role="listitem"]')).find(el => el.className.includes(cls))

  it('applies safe styling', () => {
    const hops = [createMockHop({ lossRate: 0 })]
    const { container } = render(<MTRPathVisualization hops={hops} />)
    expect(findHopWithClass(container, 'border-[var(--color-healthy-bg)]')).toBeTruthy()
  })

  it('applies critical styling for high loss', () => {
    const hops = [createMockHop({ lossRate: 15 })]
    const { container } = render(<MTRPathVisualization hops={hops} />)
    expect(findHopWithClass(container, 'border-[var(--color-critical-bg)]')).toBeTruthy()
  })

  it('applies critical styling for high latency', () => {
    const hops = [createMockHop({ avgRTTMs: 250 })]
    const { container } = render(<MTRPathVisualization hops={hops} />)
    expect(findHopWithClass(container, 'border-[var(--color-critical-bg)]')).toBeTruthy()
  })

  it('applies warning styling for jitter', () => {
    const hops = [createMockHop({ stdDevMs: 60 })]
    const { container } = render(<MTRPathVisualization hops={hops} />)
    expect(findHopWithClass(container, 'border-[var(--color-warning-bg)]')).toBeTruthy()
  })

  it('applies timeout styling', () => {
    const hops = [createMockHop({ lossRate: 100, sent: 10, received: 0 })]
    const { container } = render(<MTRPathVisualization hops={hops} />)
    expect(container.querySelector('[class*="border-[var(--color-border)]"]')).toBeTruthy()
  })

  it('calls onHopClick when clicked', () => {
    const hops = [createMockHop()]
    const onHopClick = vi.fn()
    render(<MTRPathVisualization hops={hops} onHopClick={onHopClick} />)
    fireEvent.click(screen.getByRole('button'))
    expect(onHopClick).toHaveBeenCalledTimes(1)
  })

  it('has button role when interactive', () => {
    const hops = [createMockHop()]
    render(<MTRPathVisualization hops={hops} onHopClick={vi.fn()} />)
    expect(screen.getByRole('button')).toBeInTheDocument()
  })

  it('has listitem role when not interactive', () => {
    const hops = [createMockHop()]
    const { container } = render(<MTRPathVisualization hops={hops} />)
    expect(container.querySelector('[role="listitem"]')).toBeInTheDocument()
  })

  it('has region role', () => {
    const hops = [createMockHop()]
    render(<MTRPathVisualization hops={hops} />)
    expect(screen.getByRole('region')).toBeInTheDocument()
  })

  it('has list role', () => {
    const hops = [createMockHop()]
    render(<MTRPathVisualization hops={hops} />)
    expect(screen.getByRole('list')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const hops = [createMockHop()]
    const { container } = render(<MTRPathVisualization hops={hops} className="test-class" />)
    expect(container.querySelector('.mtr-path-visualization')).toHaveClass('test-class')
  })
})
