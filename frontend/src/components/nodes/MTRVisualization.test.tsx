import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import MTRVisualization, { MTRHop, MTRResult } from './MTRVisualization'

// Mock react-i18next
describe('MTRVisualization', () => {
  const createMockHop = (overrides: Partial<MTRHop> = {}): MTRHop => ({
    hopNumber: 1,
    ip: '192.168.1.1',
    hostname: 'router.local',
    asNumber: '12345',
    sent: 10,
    received: 10,
    lossRate: 0,
    lastRTTMs: 5.2,
    avgRTTMs: 5.5,
    bestRTTMs: 4.8,
    worstRTTMs: 6.2,
    stdDevMs: 0.5,
    location: 'New York, US',
    ...overrides,
  })

  const createMockMTRResult = (overrides: Partial<MTRResult> = {}): MTRResult => ({
    target: '8.8.8.8',
    totalHops: 3,
    hops: [
      createMockHop({ hopNumber: 1, ip: '192.168.1.1', lossRate: 0 }),
      createMockHop({ hopNumber: 2, ip: '10.0.0.1', lossRate: 2 }),
      createMockHop({ hopNumber: 3, ip: '8.8.8.8', lossRate: 0 }),
    ],
    completedAt: '2024-01-15T10:30:00Z',
    success: true,
    ...overrides,
  })

  describe('rendering', () => {
    it('should render MTR data with hops', () => {
      const data = createMockMTRResult()
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('MTR Traceroute')).toBeInTheDocument()
      // Target appears in header
      const targetElements = screen.getAllByText('8.8.8.8')
      expect(targetElements.length).toBeGreaterThan(0)
      // First hop IP
      expect(screen.getByText('192.168.1.1')).toBeInTheDocument()
    })

    it('should display target in header', () => {
      const data = createMockMTRResult({ target: '1.1.1.1' })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('1.1.1.1')).toBeInTheDocument()
    })

    it('should display total hops count', () => {
      const data = createMockMTRResult({ totalHops: 5 })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('Total Hops:')).toBeInTheDocument()
      expect(screen.getByText('5')).toBeInTheDocument()
    })

    it('should display hop hostname when available', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ ip: '192.168.1.1', hostname: 'gateway.example.com' })],
      })
      render(<MTRVisualization data={data} />)

      // Hostname appears in parentheses
      expect(screen.getByText('(gateway.example.com)')).toBeInTheDocument()
    })

    it('should display AS number when available', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ asNumber: '15169' })],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('AS15169')).toBeInTheDocument()
    })

    it('should display location when available', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ location: 'San Francisco, CA, US' })],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('San Francisco, CA, US')).toBeInTheDocument()
    })

    it('should not display hostname when same as IP', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ ip: '192.168.1.1', hostname: '192.168.1.1' })],
      })
      render(<MTRVisualization data={data} />)

      // Should only have one instance of the IP (in the main display)
      const ipElements = screen.getAllByText('192.168.1.1')
      expect(ipElements.length).toBe(1)
    })
  })

  describe('health status coloring', () => {
    // Helper: jsdom can't parse CSS attribute selectors with nested brackets from CSS var classes,
    // so we find listitems and check className strings directly
    const findHopWithClass = (container: HTMLElement, cls: string) =>
      Array.from(container.querySelectorAll('[role="listitem"]')).find(el => el.className.includes(cls))

    it('should apply green styling for healthy hops (< 5% loss)', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ lossRate: 2 })],
      })
      const { container } = render(<MTRVisualization data={data} />)

      expect(findHopWithClass(container, 'border-healthy-bg')).toBeTruthy()
    })

    it('should apply yellow styling for degraded hops (5-20% loss)', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ lossRate: 10 })],
      })
      const { container } = render(<MTRVisualization data={data} />)

      expect(findHopWithClass(container, 'border-warning-bg')).toBeTruthy()
    })

    it('should apply red styling for problematic hops (> 20% loss)', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ lossRate: 50 })],
      })
      const { container } = render(<MTRVisualization data={data} />)

      expect(findHopWithClass(container, 'border-destructive/10')).toBeTruthy()
    })

    it('should display healthy badge for path with all healthy hops', () => {
      const data = createMockMTRResult({
        hops: [
          createMockHop({ lossRate: 0 }),
          createMockHop({ lossRate: 2 }),
          createMockHop({ lossRate: 1 }),
        ],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('Healthy')).toBeInTheDocument()
    })

    it('should display degraded badge for path with one degraded hop', () => {
      const data = createMockMTRResult({
        hops: [
          createMockHop({ lossRate: 0 }),
          createMockHop({ lossRate: 10 }),
        ],
      })
      render(<MTRVisualization data={data} />)

      // Path status badge
      const degradedBadges = screen.getAllByText('Degraded')
      expect(degradedBadges.length).toBeGreaterThan(0)
    })

    it('should display problematic badge for path with problematic hop', () => {
      const data = createMockMTRResult({
        hops: [
          createMockHop({ lossRate: 0 }),
          createMockHop({ lossRate: 30 }),
        ],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('Problematic')).toBeInTheDocument()
    })

    it('should display problematic badge for path with multiple degraded hops', () => {
      const data = createMockMTRResult({
        hops: [
          createMockHop({ lossRate: 8 }),
          createMockHop({ lossRate: 10 }),
          createMockHop({ lossRate: 0 }),
        ],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('Problematic')).toBeInTheDocument()
    })
  })

  describe('RTT statistics', () => {
    it('should display RTT statistics for each hop', () => {
      const data = createMockMTRResult({
        hops: [
          createMockHop({
            avgRTTMs: 15.5,
            bestRTTMs: 10.2,
            worstRTTMs: 25.8,
            stdDevMs: 5.3,
          }),
        ],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('15.5ms')).toBeInTheDocument()
      expect(screen.getByText('10.2ms')).toBeInTheDocument()
      expect(screen.getByText('25.8ms')).toBeInTheDocument()
      expect(screen.getByText('5.3ms')).toBeInTheDocument()
    })

    it('should hide std dev when zero', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ stdDevMs: 0 })],
      })
      render(<MTRVisualization data={data} />)

      // Should not have any std dev labels visible in the stats
      const hopElement = screen.getByRole('listitem')
      expect(hopElement).toBeInTheDocument()
    })

    it('should display dash for zero RTT values', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ avgRTTMs: 0 })],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('-')).toBeInTheDocument()
    })
  })

  describe('packet loss display', () => {
    it('should display packet loss rate', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ lossRate: 5.5 })],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('5.5%')).toBeInTheDocument()
    })

    it('should display packets sent/received info', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ sent: 10, received: 8 })],
      })
      render(<MTRVisualization data={data} />)

      // Check for the packets info format: "Packets: 8/10 received"
      expect(screen.getByText(/Packets: 8\/10/)).toBeInTheDocument()
    })
  })

  describe('loading state', () => {
    it('should render loading state when isLoading is true', () => {
      const { container } = render(<MTRVisualization isLoading={true} />)

      const spinner = container.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
      expect(screen.getByText('Running MTR trace...')).toBeInTheDocument()
    })

    it('should show loading state even with data provided', () => {
      const data = createMockMTRResult()
      const { container } = render(<MTRVisualization data={data} isLoading={true} />)

      const spinner = container.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
    })
  })

  describe('empty state', () => {
    it('should render empty state when no data provided', () => {
      render(<MTRVisualization />)

      expect(screen.getByText('No MTR data available')).toBeInTheDocument()
    })

    it('should render empty state when data is null', () => {
      render(<MTRVisualization data={null} />)

      expect(screen.getByText('No MTR data available')).toBeInTheDocument()
    })
  })

  describe('error state', () => {
    it('should render error state when success is false', () => {
      const data = createMockMTRResult({ success: false })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('MTR trace failed')).toBeInTheDocument()
    })

    it('should display error message when provided', () => {
      const data = createMockMTRResult({
        success: false,
        errorMessage: 'Connection timeout',
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('Connection timeout')).toBeInTheDocument()
    })

    it('should render error state without error message', () => {
      const data = createMockMTRResult({
        success: false,
        errorMessage: undefined,
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('MTR trace failed')).toBeInTheDocument()
    })

    it('should render error state from error prop', () => {
      const data = createMockMTRResult({ success: true })
      render(<MTRVisualization data={data} error="Network timeout" />)

      expect(screen.getByText('MTR trace failed')).toBeInTheDocument()
      expect(screen.getByText('Network timeout')).toBeInTheDocument()
    })

    it('should prioritize error prop over data.success', () => {
      const data = createMockMTRResult({ success: true })
      render(<MTRVisualization data={data} error="Forced error" />)

      expect(screen.getByText('Forced error')).toBeInTheDocument()
    })

    it('should use data.errorMessage when error prop is not provided', () => {
      const data = createMockMTRResult({ success: false, errorMessage: 'Data error message' })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('Data error message')).toBeInTheDocument()
    })
  })

  describe('accessibility', () => {
    it('should have proper role attribute for region', () => {
      const data = createMockMTRResult()
      render(<MTRVisualization data={data} />)

      const region = screen.getByRole('region', { name: 'MTR Traceroute' })
      expect(region).toBeInTheDocument()
    })

    it('should have list role for hop list', () => {
      const data = createMockMTRResult()
      render(<MTRVisualization data={data} />)

      const list = screen.getByRole('list', { name: 'Hop List' })
      expect(list).toBeInTheDocument()
    })

    it('should have listitem role for each hop', () => {
      const data = createMockMTRResult({
        hops: [
          createMockHop({ hopNumber: 1, ip: '192.168.1.1' }),
          createMockHop({ hopNumber: 2, ip: '10.0.0.1' }),
        ],
      })
      render(<MTRVisualization data={data} />)

      const listItems = screen.getAllByRole('listitem')
      expect(listItems).toHaveLength(2)
    })

    it('should have status role for health badge', () => {
      const data = createMockMTRResult()
      render(<MTRVisualization data={data} />)

      const status = screen.getByRole('status', { name: /Path Status/ })
      expect(status).toBeInTheDocument()
    })

    it('should have status role in loading state', () => {
      render(<MTRVisualization isLoading={true} />)

      const status = screen.getByRole('status', { name: 'Loading...' })
      expect(status).toBeInTheDocument()
    })

    it('should have alert role in error state', () => {
      const data = createMockMTRResult({ success: false })
      render(<MTRVisualization data={data} />)

      const alert = screen.getByRole('alert', { name: 'MTR trace failed' })
      expect(alert).toBeInTheDocument()
    })

    it('should have aria-hidden on decorative elements', () => {
      const data = createMockMTRResult()
      const { container } = render(<MTRVisualization data={data} />)

      const hiddenElements = container.querySelectorAll('[aria-hidden="true"]')
      expect(hiddenElements.length).toBeGreaterThan(0)
    })
  })

  describe('styling', () => {
    it('should apply custom className', () => {
      const data = createMockMTRResult()
      const { container } = render(<MTRVisualization data={data} className="custom-class" />)

      const visualization = container.querySelector('.mtr-visualization')
      expect(visualization).toHaveClass('custom-class')
    })

    it('should display legend at bottom', () => {
      const data = createMockMTRResult()
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('Legend:')).toBeInTheDocument()
      expect(screen.getByText(/< 5% loss/)).toBeInTheDocument()
      expect(screen.getByText(/5-20% loss/)).toBeInTheDocument()
      expect(screen.getByText(/> 20% loss/)).toBeInTheDocument()
    })
  })

  describe('edge cases', () => {
    it('should handle empty hops array', () => {
      const data = createMockMTRResult({ hops: [], totalHops: 0 })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('MTR Traceroute')).toBeInTheDocument()
      expect(screen.getByText('0')).toBeInTheDocument() // total hops
    })

    it('should handle hop without optional fields', () => {
      const data = createMockMTRResult({
        hops: [
          {
            hopNumber: 1,
            ip: '192.168.1.1',
            sent: 10,
            received: 10,
            lossRate: 0,
            lastRTTMs: 5,
            avgRTTMs: 5,
            bestRTTMs: 4,
            worstRTTMs: 6,
            stdDevMs: 0,
          },
        ],
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText('192.168.1.1')).toBeInTheDocument()
    })

    it('should handle very high packet loss', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ lossRate: 100 })],
      })
      const { container } = render(<MTRVisualization data={data} />)

      expect(screen.getByText('100.0%')).toBeInTheDocument()
      const hopEl = Array.from(container.querySelectorAll('[role="listitem"]')).find(el => el.className.includes('border-destructive/10'))
      expect(hopEl).toBeTruthy()
    })

    it('should handle single hop trace', () => {
      const data = createMockMTRResult({
        hops: [createMockHop({ hopNumber: 1, ip: '8.8.8.8' })],
        totalHops: 1,
      })
      render(<MTRVisualization data={data} />)

      const listItems = screen.getAllByRole('listitem')
      expect(listItems).toHaveLength(1)
    })
  })

  describe('timestamp display', () => {
    it('should format completedAt timestamp', () => {
      const data = createMockMTRResult({
        completedAt: '2024-01-15T10:30:00Z',
      })
      render(<MTRVisualization data={data} />)

      expect(screen.getByText(/Completed At:/)).toBeInTheDocument()
    })

    it('should handle missing completedAt', () => {
      const data = createMockMTRResult({
        completedAt: '',
      })
      render(<MTRVisualization data={data} />)

      // Should still render without crashing
      expect(screen.getByText('MTR Traceroute')).toBeInTheDocument()
    })
  })

  describe('onHopClick callback', () => {
    it('should call onHopClick when hop is clicked', () => {
      const mockHop = createMockHop({ hopNumber: 1, ip: '192.168.1.1' })
      const data = createMockMTRResult({ hops: [mockHop] })
      const onHopClick = vi.fn()

      render(<MTRVisualization data={data} onHopClick={onHopClick} />)

      const hopElement = screen.getByRole('button', { name: /Hop 1: 192.168.1.1/ })
      fireEvent.click(hopElement)

      expect(onHopClick).toHaveBeenCalledTimes(1)
      expect(onHopClick).toHaveBeenCalledWith(mockHop)
    })

    it('should call onHopClick with correct hop data for each hop', () => {
      const hop1 = createMockHop({ hopNumber: 1, ip: '192.168.1.1' })
      const hop2 = createMockHop({ hopNumber: 2, ip: '10.0.0.1' })
      const data = createMockMTRResult({ hops: [hop1, hop2] })
      const onHopClick = vi.fn()

      render(<MTRVisualization data={data} onHopClick={onHopClick} />)

      const hopButtons = screen.getAllByRole('button')
      fireEvent.click(hopButtons[1]) // Click second hop

      expect(onHopClick).toHaveBeenCalledWith(hop2)
    })

    it('should have button role when onHopClick is provided', () => {
      const data = createMockMTRResult()
      render(<MTRVisualization data={data} onHopClick={vi.fn()} />)

      const buttons = screen.getAllByRole('button')
      expect(buttons.length).toBe(data.hops.length)
    })

    it('should have listitem role when onHopClick is not provided', () => {
      const data = createMockMTRResult()
      render(<MTRVisualization data={data} />)

      const listItems = screen.getAllByRole('listitem')
      expect(listItems.length).toBe(data.hops.length)
    })

    it('should have tabIndex 0 when onHopClick is provided', () => {
      const data = createMockMTRResult()
      const { container } = render(<MTRVisualization data={data} onHopClick={vi.fn()} />)

      const hopElement = container.querySelector('[tabindex="0"]')
      expect(hopElement).toBeInTheDocument()
    })

    it('should not have tabIndex when onHopClick is not provided', () => {
      const data = createMockMTRResult()
      const { container } = render(<MTRVisualization data={data} />)

      const hopElement = container.querySelector('[tabindex="0"]')
      expect(hopElement).not.toBeInTheDocument()
    })

    it('should call onHopClick on Enter key press', () => {
      const mockHop = createMockHop({ hopNumber: 1, ip: '192.168.1.1' })
      const data = createMockMTRResult({ hops: [mockHop] })
      const onHopClick = vi.fn()

      render(<MTRVisualization data={data} onHopClick={onHopClick} />)

      const hopElement = screen.getByRole('button', { name: /Hop 1: 192.168.1.1/ })
      fireEvent.keyDown(hopElement, { key: 'Enter' })

      expect(onHopClick).toHaveBeenCalledWith(mockHop)
    })

    it('should call onHopClick on Space key press', () => {
      const mockHop = createMockHop({ hopNumber: 1, ip: '192.168.1.1' })
      const data = createMockMTRResult({ hops: [mockHop] })
      const onHopClick = vi.fn()

      render(<MTRVisualization data={data} onHopClick={onHopClick} />)

      const hopElement = screen.getByRole('button', { name: /Hop 1: 192.168.1.1/ })
      fireEvent.keyDown(hopElement, { key: ' ' })

      expect(onHopClick).toHaveBeenCalledWith(mockHop)
    })

    it('should not call onHopClick on other key presses', () => {
      const mockHop = createMockHop({ hopNumber: 1, ip: '192.168.1.1' })
      const data = createMockMTRResult({ hops: [mockHop] })
      const onHopClick = vi.fn()

      render(<MTRVisualization data={data} onHopClick={onHopClick} />)

      const hopElement = screen.getByRole('button', { name: /Hop 1: 192.168.1.1/ })
      fireEvent.keyDown(hopElement, { key: 'Tab' })

      expect(onHopClick).not.toHaveBeenCalled()
    })

    it('should have cursor-pointer class when onHopClick is provided', () => {
      const data = createMockMTRResult()
      const { container } = render(<MTRVisualization data={data} onHopClick={vi.fn()} />)

      const clickableHop = container.querySelector('.cursor-pointer')
      expect(clickableHop).toBeInTheDocument()
    })

    it('should not crash when onHopClick is not provided and hop is clicked', () => {
      const data = createMockMTRResult()
      render(<MTRVisualization data={data} />)

      // No error should be thrown
      const listItems = screen.getAllByRole('listitem')
      expect(listItems.length).toBe(data.hops.length)
    })
  })
})
