import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { AlertRecordsFilter } from '../AlertRecordsFilter'
import type { AlertRecordFilters } from '../../../api/alertRecords'
import type { NodeDTO } from '../../../api/types'

describe('AlertRecordsFilter', () => {
  const mockNodes: NodeDTO[] = [
    { id: 'node-1', name: 'Node 1', ip: '192.168.1.1', region: 'us-east', tags: [], status: 'online', created_at: '2024-01-01', updated_at: '2024-01-01' },
    { id: 'node-2', name: 'Node 2', ip: '192.168.1.2', region: 'us-west', tags: [], status: 'online', created_at: '2024-01-01', updated_at: '2024-01-01' },
  ]

  const defaultProps = {
    filters: {},
    nodes: mockNodes,
    searchQuery: '',
    onFilterChange: vi.fn(),
    onSearchChange: vi.fn(),
    onReset: vi.fn(),
  }

  it('renders filter controls', () => {
    render(<AlertRecordsFilter {...defaultProps} />)

    expect(screen.getByLabelText(/搜索/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/节点/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/开始时间/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/结束时间/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/告警级别/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/处理状态/i)).toBeInTheDocument()
    expect(screen.getByText('重置筛选')).toBeInTheDocument()
    expect(screen.getByText('应用筛选')).toBeInTheDocument()
  })

  it('renders node options in dropdown', () => {
    render(<AlertRecordsFilter {...defaultProps} />)

    const nodeSelect = screen.getByLabelText(/节点/i)
    expect(screen.getByText('全部节点')).toBeInTheDocument()
    expect(screen.getByText('Node 1 (192.168.1.1)')).toBeInTheDocument()
    expect(screen.getByText('Node 2 (192.168.1.2)')).toBeInTheDocument()
  })

  it('calls onFilterChange with node filter when applied', () => {
    const onFilterChange = vi.fn()
    render(<AlertRecordsFilter {...defaultProps} onFilterChange={onFilterChange} />)

    const nodeSelect = screen.getByLabelText(/节点/i)
    fireEvent.change(nodeSelect, { target: { value: 'node-1' } })

    const applyButton = screen.getByText('应用筛选')
    fireEvent.click(applyButton)

    expect(onFilterChange).toHaveBeenCalledWith({ node_id: 'node-1' })
  })

  it('calls onFilterChange with level filter when applied', () => {
    const onFilterChange = vi.fn()
    render(<AlertRecordsFilter {...defaultProps} onFilterChange={onFilterChange} />)

    const levelSelect = screen.getByLabelText(/告警级别/i)
    fireEvent.change(levelSelect, { target: { value: 'P0' } })

    const applyButton = screen.getByText('应用筛选')
    fireEvent.click(applyButton)

    expect(onFilterChange).toHaveBeenCalledWith({ level: 'P0' })
  })

  it('calls onFilterChange with status filter when applied', () => {
    const onFilterChange = vi.fn()
    render(<AlertRecordsFilter {...defaultProps} onFilterChange={onFilterChange} />)

    const statusSelect = screen.getByLabelText(/处理状态/i)
    fireEvent.change(statusSelect, { target: { value: 'pending' } })

    const applyButton = screen.getByText('应用筛选')
    fireEvent.click(applyButton)

    expect(onFilterChange).toHaveBeenCalledWith({ status: 'pending' })
  })

  it('calls onFilterChange with time range filters when applied', () => {
    const onFilterChange = vi.fn()
    render(<AlertRecordsFilter {...defaultProps} onFilterChange={onFilterChange} />)

    const startTimeInput = screen.getByLabelText(/开始时间/i)
    const endTimeInput = screen.getByLabelText(/结束时间/i)

    fireEvent.change(startTimeInput, { target: { value: '2024-01-01T00:00' } })
    fireEvent.change(endTimeInput, { target: { value: '2024-12-31T23:59' } })

    const applyButton = screen.getByText('应用筛选')
    fireEvent.click(applyButton)

    expect(onFilterChange).toHaveBeenCalledWith({
      start_time: '2024-01-01T00:00',
      end_time: '2024-12-31T23:59',
    })
  })

  it('calls onFilterChange with multiple filters when applied', () => {
    const onFilterChange = vi.fn()
    render(<AlertRecordsFilter {...defaultProps} onFilterChange={onFilterChange} />)

    const nodeSelect = screen.getByLabelText(/节点/i)
    const levelSelect = screen.getByLabelText(/告警级别/i)

    fireEvent.change(nodeSelect, { target: { value: 'node-1' } })
    fireEvent.change(levelSelect, { target: { value: 'P1' } })

    const applyButton = screen.getByText('应用筛选')
    fireEvent.click(applyButton)

    expect(onFilterChange).toHaveBeenCalledWith({
      node_id: 'node-1',
      level: 'P1',
    })
  })

  it('calls onSearchChange when search input applied', () => {
    const onSearchChange = vi.fn()
    render(<AlertRecordsFilter {...defaultProps} onSearchChange={onSearchChange} />)

    const searchInput = screen.getByLabelText(/搜索/i)
    fireEvent.change(searchInput, { target: { value: 'test search' } })

    const applyButton = screen.getByText('应用筛选')
    fireEvent.click(applyButton)

    expect(onSearchChange).toHaveBeenCalledWith('test search')
  })

  it('calls onReset when reset button clicked', () => {
    const onReset = vi.fn()
    const filters: AlertRecordFilters = { node_id: 'node-1', level: 'P0' }
    render(<AlertRecordsFilter {...defaultProps} filters={filters} onReset={onReset} />)

    const resetButton = screen.getByText('重置筛选')
    fireEvent.click(resetButton)

    expect(onReset).toHaveBeenCalled()
  })

  it('resets all filter inputs when onReset called', () => {
    const onReset = vi.fn()
    const filters: AlertRecordFilters = { node_id: 'node-1', level: 'P0' }
    const { rerender } = render(
      <AlertRecordsFilter {...defaultProps} filters={filters} onReset={onReset} />
    )

    const resetButton = screen.getByText('重置筛选')
    fireEvent.click(resetButton)

    expect(onReset).toHaveBeenCalled()

    // After reset, component should be re-rendered with empty filters
    rerender(
      <AlertRecordsFilter {...defaultProps} filters={{}} onReset={onReset} searchQuery="" />
    )
  })
})
