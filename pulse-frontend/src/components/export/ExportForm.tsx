/**
 * ExportForm Component
 *
 * Form component for configuring export parameters.
 * Handles node selection, time range, metrics, and format.
 */

import { useState } from 'react'
import type { NodeDTO } from '../../api/types'
import type { CreateExportRequest, ExportMetric } from '../../types/export'

interface ExportFormProps {
  nodes: NodeDTO[]
  onSubmit: (request: CreateExportRequest) => Promise<void>
  loading?: boolean
}

interface FormErrors {
  nodeIds?: string
  timeRange?: string
  metrics?: string
}

export function ExportForm({ nodes, onSubmit, loading = false }: ExportFormProps) {
  // Form state
  const [selectedNodeIds, setSelectedNodeIds] = useState<string[]>([])
  const [timeRange, setTimeRange] = useState<'7d' | '30d' | 'custom'>('7d')
  const [customStartDate, setCustomStartDate] = useState('')
  const [customEndDate, setCustomEndDate] = useState('')
  const [selectedMetrics, setSelectedMetrics] = useState<ExportMetric[]>(['latency'])
  const [format, setFormat] = useState<'csv' | 'excel'>('csv')
  const [errors, setErrors] = useState<FormErrors>({})

  /**
   * Calculate time range based on selection
   */
  const getTimeRange = (): { start: string; end: string } => {
    const now = new Date()
    const end = now.toISOString()

    let start: Date
    if (timeRange === '7d') {
      start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
    } else if (timeRange === '30d') {
      start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
    } else {
      // Custom range
      start = new Date(customStartDate)
    }

    return {
      start: start.toISOString(),
      end: customEndDate ? new Date(customEndDate).toISOString() : end,
    }
  }

  /**
   * Validate form inputs
   */
  const validate = (): boolean => {
    const newErrors: FormErrors = {}

    // Validate nodes
    if (selectedNodeIds.length === 0) {
      newErrors.nodeIds = 'Select at least one node'
    } else if (selectedNodeIds.length > 50) {
      newErrors.nodeIds = 'Maximum 50 nodes allowed'
    }

    // Validate time range
    if (timeRange === 'custom') {
      if (!customStartDate || !customEndDate) {
        newErrors.timeRange = 'Select both start and end dates'
      } else {
        const start = new Date(customStartDate)
        const end = new Date(customEndDate)
        if (start >= end) {
          newErrors.timeRange = 'End date must be after start date'
        }
      }
    }

    // Validate metrics
    if (selectedMetrics.length === 0) {
      newErrors.metrics = 'Select at least one metric'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  /**
   * Handle form submission
   */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) {
      return
    }

    const { start, end } = getTimeRange()

    // Let errors propagate to parent for handling
    await onSubmit({
      node_ids: selectedNodeIds,
      start_time: start,
      end_time: end,
      metrics: selectedMetrics,
      format,
    })
  }

  /**
   * Toggle node selection
   */
  const toggleNode = (nodeId: string) => {
    setSelectedNodeIds((prev) =>
      prev.includes(nodeId) ? prev.filter((id) => id !== nodeId) : [...prev, nodeId]
    )
    // Clear node error when nodes are selected
    if (errors.nodeIds) {
      setErrors((prev) => ({ ...prev, nodeIds: undefined }))
    }
  }

  /**
   * Toggle metric selection
   */
  const toggleMetric = (metric: ExportMetric) => {
    setSelectedMetrics((prev) =>
      prev.includes(metric) ? prev.filter((m) => m !== metric) : [...prev, metric]
    )
    // Clear metric error when metrics are selected
    if (errors.metrics) {
      setErrors((prev) => ({ ...prev, metrics: undefined }))
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Node Selection */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Nodes <span className="text-red-500">*</span>
        </label>
        <div className="text-xs text-gray-500 mb-2">
          Selected: {selectedNodeIds.length} / 50 nodes
        </div>
        <div className="max-h-48 overflow-y-auto border border-gray-300 rounded-md p-3 space-y-2">
          {nodes.map((node) => (
            <label
              key={node.id}
              className="flex items-center space-x-2 cursor-pointer hover:bg-gray-50 p-1 rounded"
            >
              <input
                type="checkbox"
                checked={selectedNodeIds.includes(node.id)}
                onChange={() => toggleNode(node.id)}
                disabled={loading}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <span className="text-sm text-gray-900">
                {node.name} ({node.ip})
              </span>
            </label>
          ))}
        </div>
        {errors.nodeIds && (
          <p className="mt-1 text-sm text-red-600" data-testid="node-error">
            {errors.nodeIds}
          </p>
        )}
      </div>

      {/* Time Range Selection */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Time Range <span className="text-red-500">*</span>
        </label>
        <div className="space-y-2">
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="radio"
              name="timeRange"
              value="7d"
              checked={timeRange === '7d'}
              onChange={() => setTimeRange('7d')}
              disabled={loading}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
            />
            <span className="text-sm text-gray-900">Last 7 days</span>
          </label>
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="radio"
              name="timeRange"
              value="30d"
              checked={timeRange === '30d'}
              onChange={() => setTimeRange('30d')}
              disabled={loading}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
            />
            <span className="text-sm text-gray-900">Last 30 days</span>
          </label>
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="radio"
              name="timeRange"
              value="custom"
              checked={timeRange === 'custom'}
              onChange={() => setTimeRange('custom')}
              disabled={loading}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
            />
            <span className="text-sm text-gray-900">Custom Range</span>
          </label>
        </div>

        {/* Custom Date Pickers */}
        {timeRange === 'custom' && (
          <div className="mt-3 grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="startDate" className="block text-sm font-medium text-gray-700 mb-1">
                Start Date
              </label>
              <input
                id="startDate"
                type="date"
                value={customStartDate}
                onChange={(e) => setCustomStartDate(e.target.value)}
                max={customEndDate || new Date().toISOString().split('T')[0]}
                disabled={loading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            <div>
              <label htmlFor="endDate" className="block text-sm font-medium text-gray-700 mb-1">
                End Date
              </label>
              <input
                id="endDate"
                type="date"
                value={customEndDate}
                onChange={(e) => setCustomEndDate(e.target.value)}
                min={customStartDate}
                max={new Date().toISOString().split('T')[0]}
                disabled={loading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>
        )}
        {errors.timeRange && (
          <p className="mt-1 text-sm text-red-600">{errors.timeRange}</p>
        )}
      </div>

      {/* Metric Selection */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Metrics <span className="text-red-500">*</span>
        </label>
        <div className="space-y-2">
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="checkbox"
              checked={selectedMetrics.includes('latency')}
              onChange={() => toggleMetric('latency')}
              disabled={loading}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <span className="text-sm text-gray-900">Latency (时延)</span>
          </label>
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="checkbox"
              checked={selectedMetrics.includes('packet_loss_rate')}
              onChange={() => toggleMetric('packet_loss_rate')}
              disabled={loading}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <span className="text-sm text-gray-900">Packet Loss Rate (丢包率)</span>
          </label>
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="checkbox"
              checked={selectedMetrics.includes('jitter')}
              onChange={() => toggleMetric('jitter')}
              disabled={loading}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <span className="text-sm text-gray-900">Jitter (抖动)</span>
          </label>
        </div>
        {errors.metrics && (
          <p className="mt-1 text-sm text-red-600">{errors.metrics}</p>
        )}
      </div>

      {/* Format Selection */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Format <span className="text-red-500">*</span>
        </label>
        <div className="space-y-2">
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="radio"
              name="format"
              value="csv"
              checked={format === 'csv'}
              onChange={() => setFormat('csv')}
              disabled={loading}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
            />
            <span className="text-sm text-gray-900">CSV</span>
          </label>
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="radio"
              name="format"
              value="excel"
              checked={format === 'excel'}
              onChange={() => setFormat('excel')}
              disabled={loading}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
            />
            <span className="text-sm text-gray-900">Excel</span>
          </label>
        </div>
      </div>

      {/* Submit Button */}
      <div className="flex justify-end space-x-3">
        <button
          type="submit"
          disabled={loading}
          className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white font-medium py-2 px-6 rounded-md transition-colors duration-150"
        >
          {loading ? 'Exporting...' : 'Export'}
        </button>
      </div>
    </form>
  )
}
