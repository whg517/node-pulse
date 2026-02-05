import { useState } from 'react'
import type { AlertRule } from '../../stores/types'
import type { NodeDTO } from '../../api/types'

interface AlertRuleFormProps {
  mode: 'create' | 'edit'
  initialData?: AlertRule
  nodes: NodeDTO[]
  onSubmit: (data: any) => Promise<void>
  onCancel: () => void
}

export function AlertRuleForm({ mode, initialData, nodes, onSubmit, onCancel }: AlertRuleFormProps) {
  const [metric, setMetric] = useState<string>(initialData?.metric || 'latency')
  const [threshold, setThreshold] = useState<number>(initialData?.threshold || 0)
  const [level, setLevel] = useState<string>(initialData?.level || 'P1')
  const [nodeId, setNodeId] = useState<string | null>(initialData?.nodeId || null)
  const [enabled, setEnabled] = useState<boolean>(initialData?.enabled ?? true)

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const validate = () => {
    const newErrors: Record<string, string> = {}

    if (!threshold || threshold <= 0) {
      newErrors.threshold = 'Threshold must be greater than 0'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) return

    setIsSubmitting(true)
    try {
      const data: any = {
        metric,
        threshold,
        level,
        enabled,
      }

      // Only include node_id if it's not null
      if (nodeId) {
        data.node_id = nodeId
      } else {
        // For global rules, explicitly set to null
        data.node_id = null
      }

      await onSubmit(data)
    } catch (error) {
      console.error('Failed to submit form:', error)
      throw error
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Metric Type Select */}
      <div>
        <label htmlFor="metric" className="block text-sm font-medium text-gray-700">
          Metric Type
        </label>
        <select
          id="metric"
          value={metric}
          onChange={(e) => setMetric(e.target.value)}
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md border"
        >
          <option value="latency">Latency (ms)</option>
          <option value="packet_loss_rate">Packet Loss Rate (%)</option>
          <option value="jitter">Jitter (ms)</option>
        </select>
      </div>

      {/* Threshold Input */}
      <div>
        <label htmlFor="threshold" className="block text-sm font-medium text-gray-700">
          Threshold
        </label>
        <input
          type="number"
          id="threshold"
          value={threshold}
          onChange={(e) => setThreshold(Number(e.target.value))}
          min="0"
          step="0.01"
          className={`mt-1 block w-full border ${errors.threshold ? 'border-red-300' : 'border-gray-300'} rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm`}
        />
        {errors.threshold && (
          <p className="mt-2 text-sm text-red-600">{errors.threshold}</p>
        )}
      </div>

      {/* Alert Level Select */}
      <div>
        <label htmlFor="level" className="block text-sm font-medium text-gray-700">
          Alert Level
        </label>
        <select
          id="level"
          value={level}
          onChange={(e) => setLevel(e.target.value)}
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md border"
        >
          <option value="P0">P0 - Critical</option>
          <option value="P1">P1 - Warning</option>
          <option value="P2">P2 - Info</option>
        </select>
      </div>

      {/* Node Selection */}
      <div>
        <label htmlFor="node" className="block text-sm font-medium text-gray-700">
          Scope
        </label>
        <select
          id="node"
          value={nodeId || ''}
          onChange={(e) => setNodeId(e.target.value || null)}
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md border"
        >
          <option value="">Global Rule (All Nodes)</option>
          {nodes.map((node) => (
            <option key={node.id} value={node.id}>
              {node.name} ({node.ip})
            </option>
          ))}
        </select>
        <p className="mt-2 text-sm text-gray-500">
          Select "Global Rule" to apply to all nodes, or select a specific node.
        </p>
      </div>

      {/* Enabled Toggle */}
      <div className="flex items-start">
        <div className="flex items-center h-5">
          <input
            id="enabled"
            type="checkbox"
            checked={enabled}
            onChange={(e) => setEnabled(e.target.checked)}
            className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-gray-300 rounded"
          />
        </div>
        <div className="ml-3 text-sm">
          <label htmlFor="enabled" className="font-medium text-gray-700">
            Enabled
          </label>
          <p className="text-gray-500">Uncheck to disable this rule without deleting it</p>
        </div>
      </div>

      {/* Submit and Cancel Buttons */}
      <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? 'Saving...' : mode === 'create' ? 'Create Rule' : 'Update Rule'}
        </button>
      </div>
    </form>
  )
}
