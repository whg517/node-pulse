import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { NodeSummaryCard } from '../NodeSummaryCard'
import type { NodeDTO } from '../../../api/types'

// Mock react-router-dom navigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('NodeSummaryCard', () => {
  const mockNode: NodeDTO = {
    id: 'node-1',
    name: 'Test Node',
    ip: '192.168.1.1',
    region: 'us-east',
    tags: ['tag1', 'tag2'],
    status: 'online',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders node name', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })

  it('renders node region', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    expect(screen.getByText('us-east')).toBeInTheDocument()
  })

  it('renders health status label', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    expect(screen.getByText('Healthy')).toBeInTheDocument()
  })

  it('navigates to node detail on click', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    const card = screen.getByRole('button')
    fireEvent.click(card)
    expect(mockNavigate).toHaveBeenCalledWith('/nodes/node-1', { state: { breadcrumbLabel: 'Test Node' } })
  })

  it('navigates on Enter key', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    const card = screen.getByRole('button')
    fireEvent.keyDown(card, { key: 'Enter' })
    expect(mockNavigate).toHaveBeenCalledWith('/nodes/node-1', { state: { breadcrumbLabel: 'Test Node' } })
  })

  it('navigates on Space key', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    const card = screen.getByRole('button')
    fireEvent.keyDown(card, { key: ' ' })
    expect(mockNavigate).toHaveBeenCalledWith('/nodes/node-1', { state: { breadcrumbLabel: 'Test Node' } })
  })

  it('does not navigate on other keys', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    const card = screen.getByRole('button')
    fireEvent.keyDown(card, { key: 'Tab' })
    expect(mockNavigate).not.toHaveBeenCalled()
  })

  it('renders latency when provided', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" latency={42.5} />
      </MemoryRouter>
    )
    // The component uses toFixed(0) so 42.5 → "43", and units.ms is a separate text node
    expect(screen.getByText(/43/)).toBeInTheDocument()
  })

  it('renders packet loss when provided', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" packetLoss={1.5} />
      </MemoryRouter>
    )
    // The component renders packet loss with toFixed(1) so 1.5 → "1.5"
    expect(screen.getByText(/1\.5/)).toBeInTheDocument()
  })

  it('does not render metrics section when neither latency nor packet loss provided', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    expect(screen.queryByText(/units\.ms/)).not.toBeInTheDocument()
  })

  it('renders tags', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" />
      </MemoryRouter>
    )
    expect(screen.getByText('tag1')).toBeInTheDocument()
    expect(screen.getByText('tag2')).toBeInTheDocument()
  })

  it('renders +N for extra tags beyond 3', () => {
    const nodeWithManyTags = { ...mockNode, tags: ['t1', 't2', 't3', 't4', 't5'] }
    render(
      <MemoryRouter>
        <NodeSummaryCard node={nodeWithManyTags} healthStatus="healthy" />
      </MemoryRouter>
    )
    expect(screen.getByText('+2')).toBeInTheDocument()
  })

  it('renders lastSeen when provided', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard
          node={mockNode}
          healthStatus="healthy"
          lastSeen="2024-01-15T10:00:00Z"
        />
      </MemoryRouter>
    )
    expect(screen.getByText('Last Seen:')).toBeInTheDocument()
  })

  it('renders warning status', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="warning" />
      </MemoryRouter>
    )
    expect(screen.getByText('Warning')).toBeInTheDocument()
  })

  it('renders critical status', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="critical" />
      </MemoryRouter>
    )
    expect(screen.getByText('Critical')).toBeInTheDocument()
  })

  it('renders offline status', () => {
    render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="offline" />
      </MemoryRouter>
    )
    expect(screen.getByText('Offline')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <MemoryRouter>
        <NodeSummaryCard node={mockNode} healthStatus="healthy" className="my-custom" />
      </MemoryRouter>
    )
    const card = container.querySelector('.node-summary-card')
    expect(card?.className).toContain('my-custom')
  })

  it('handles node without tags', () => {
    const nodeNoTags = { ...mockNode, tags: [] }
    render(
      <MemoryRouter>
        <NodeSummaryCard node={nodeNoTags} healthStatus="healthy" />
      </MemoryRouter>
    )
    // Should not throw and renders normally
    expect(screen.getByText('Test Node')).toBeInTheDocument()
  })
})
