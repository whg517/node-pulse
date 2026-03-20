import { useState } from 'react'
import type { AlertRecordFilters, AlertLevel, AlertRecordStatus } from '../../api/alertRecords'
import type { NodeDTO } from '../../api/types'

interface AlertRecordsFilterProps {
  filters: AlertRecordFilters
  nodes: NodeDTO[]
  searchQuery: string
  onFilterChange: (filters: AlertRecordFilters) => void
  onSearchChange: (query: string) => void
  onReset: () => void
}

export function AlertRecordsFilter({
  filters,
  nodes,
  searchQuery,
  onFilterChange,
  onSearchChange,
  onReset,
}: AlertRecordsFilterProps) {
  // Local state for form inputs
  const [nodeId, setNodeId] = useState<string>(filters.node_id || '')
  const [startTime, setStartTime] = useState<string>(filters.start_time || '')
  const [endTime, setEndTime] = useState<string>(filters.end_time || '')
  const [level, setLevel] = useState<string>(filters.level || '')
  const [status, setStatus] = useState<string>(filters.status || '')
  const [searchInput, setSearchInput] = useState<string>(searchQuery)

  const handleApply = () => {
    const newFilters: AlertRecordFilters = {}

    if (nodeId) newFilters.node_id = nodeId
    if (startTime) newFilters.start_time = startTime
    if (endTime) newFilters.end_time = endTime
    if (level) newFilters.level = level as AlertLevel
    if (status) newFilters.status = status as AlertRecordStatus

    onFilterChange(newFilters)
    onSearchChange(searchInput)
  }

  const handleReset = () => {
    setNodeId('')
    setStartTime('')
    setEndTime('')
    setLevel('')
    setStatus('')
    setSearchInput('')
    onReset()
  }

  return (
    <div className="bg-[var(--color-bg-surface)] shadow rounded-lg p-6 mb-6 border border-[var(--color-border)]">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        {/* Search Input */}
        <div>
          <label htmlFor="search-input" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
            搜索
          </label>
          <input
            id="search-input"
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="节点名称或指标类型"
            className="w-full px-3 py-2 border border-[var(--color-input-border)] rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-[var(--color-input-bg)] text-[var(--color-text-primary)] placeholder:text-[var(--color-text-placeholder)]"
          />
        </div>

        {/* Node Selection */}
        <div>
          <label htmlFor="node-filter" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
            节点
          </label>
          <select
            id="node-filter"
            value={nodeId}
            onChange={(e) => setNodeId(e.target.value)}
            className="w-full px-3 py-2 border border-[var(--color-input-border)] rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-[var(--color-input-bg)] text-[var(--color-text-primary)]"
          >
            <option value="">全部节点</option>
            {nodes.map((node) => (
              <option key={node.id} value={node.id}>
                {node.name} ({node.ip})
              </option>
            ))}
          </select>
        </div>

        {/* Time Range - Start */}
        <div>
          <label htmlFor="start-time" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
            开始时间
          </label>
          <input
            id="start-time"
            type="datetime-local"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            className="w-full px-3 py-2 border border-[var(--color-input-border)] rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-[var(--color-input-bg)] text-[var(--color-text-primary)]"
          />
        </div>

        {/* Time Range - End */}
        <div>
          <label htmlFor="end-time" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
            结束时间
          </label>
          <input
            id="end-time"
            type="datetime-local"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            className="w-full px-3 py-2 border border-[var(--color-input-border)] rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-[var(--color-input-bg)] text-[var(--color-text-primary)]"
          />
        </div>

        {/* Alert Level */}
        <div>
          <label htmlFor="level-filter" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
            告警级别
          </label>
          <select
            id="level-filter"
            value={level}
            onChange={(e) => setLevel(e.target.value)}
            className="w-full px-3 py-2 border border-[var(--color-input-border)] rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-[var(--color-input-bg)] text-[var(--color-text-primary)]"
          >
            <option value="">全部级别</option>
            <option value="P0">P0</option>
            <option value="P1">P1</option>
            <option value="P2">P2</option>
          </select>
        </div>

        {/* Status */}
        <div>
          <label htmlFor="status-filter" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
            处理状态
          </label>
          <select
            id="status-filter"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="w-full px-3 py-2 border border-[var(--color-input-border)] rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-[var(--color-input-bg)] text-[var(--color-text-primary)]"
          >
            <option value="">全部状态</option>
            <option value="pending">未处理</option>
            <option value="in_progress">处理中</option>
            <option value="resolved">已解决</option>
          </select>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="mt-4 flex justify-end gap-3">
        <button
          type="button"
          onClick={handleReset}
          className="px-4 py-2 bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] rounded-md hover:bg-[var(--color-bg-subtle)] transition-colors"
        >
          重置筛选
        </button>
        <button
          type="button"
          onClick={handleApply}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
        >
          应用筛选
        </button>
      </div>
    </div>
  )
}
