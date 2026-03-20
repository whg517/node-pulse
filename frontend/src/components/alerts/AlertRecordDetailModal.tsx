import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { AlertRecordDTO, AlertRecordStatus } from '../../api/alertRecords'
import type { NodeDTO } from '../../api/types'

interface AlertRecordDetailModalProps {
  record: AlertRecordDTO
  nodes: NodeDTO[]
  canEdit: boolean
  onClose: () => void
  onStatusUpdate: (id: string, status: AlertRecordStatus) => Promise<void>
}

export function AlertRecordDetailModal({
  record,
  nodes,
  canEdit,
  onClose,
  onStatusUpdate,
}: AlertRecordDetailModalProps) {
  const navigate = useNavigate()
  const [isUpdating, setIsUpdating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Helper to get node name by ID
  const node = nodes.find((n) => n.id === record.node_id)

  // Helper to get status display name
  const getStatusDisplayName = (status: string) => {
    switch (status) {
      case 'pending':
        return '未处理'
      case 'in_progress':
        return '处理中'
      case 'resolved':
        return '已解决'
      default:
        return status
    }
  }

  // Helper to get status badge color
  const getStatusBadgeColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300'
      case 'in_progress':
        return 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-300'
      case 'resolved':
        return 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-300'
      default:
        return 'bg-slate-100 dark:bg-slate-700/50 text-slate-700 dark:text-slate-300'
    }
  }

  // Helper to get level badge color
  const getLevelBadgeColor = (level: string) => {
    switch (level) {
      case 'P0':
        return 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300'
      case 'P1':
        return 'bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-300'
      case 'P2':
        return 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-300'
      default:
        return 'bg-slate-100 dark:bg-slate-700/50 text-slate-700 dark:text-slate-300'
    }
  }

  // Helper to get metric display name
  const getMetricDisplayName = (metric: string) => {
    switch (metric) {
      case 'latency':
        return '延迟'
      case 'packet_loss_rate':
        return '丢包率'
      case 'jitter':
        return '抖动'
      default:
        return metric
    }
  }

  // Helper to format timestamp
  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }

  const handleStatusUpdate = async (newStatus: AlertRecordStatus) => {
    setIsUpdating(true)
    setError(null)
    try {
      await onStatusUpdate(record.id, newStatus)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update status')
    } finally {
      setIsUpdating(false)
    }
  }

  const handleViewNodeDetails = () => {
    onClose()
    navigate(`/nodes/${record.node_id}`)
  }

  return (
    <div
      className="fixed inset-0 bg-black/50 overflow-y-auto h-full w-full z-50"
      onClick={onClose}
    >
      <div
        className="relative top-20 mx-auto p-5 border border-[var(--color-border)] shadow-lg rounded-md bg-[var(--color-bg-elevated)] max-w-2xl w-full"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex justify-between items-start mb-4">
          <h3 className="text-lg font-medium text-[var(--color-text-primary)]">告警记录详情</h3>
          <button
            type="button"
            onClick={onClose}
            className="text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]"
          >
            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-4 bg-red-50 dark:bg-red-900/20 border-l-4 border-red-400 dark:border-red-600 p-4 rounded-md">
            <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
          </div>
        )}

        {/* Content */}
        <div className="space-y-4">
          {/* Alert ID */}
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)]">告警 ID</label>
            <p className="mt-1 text-sm text-[var(--color-text-primary)] font-mono">{record.id}</p>
          </div>

          {/* Node Information */}
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)]">节点</label>
            <div className="mt-1 flex items-center gap-2">
              <p className="text-sm text-[var(--color-text-primary)]">{node?.name || record.node_id}</p>
              <button
                type="button"
                onClick={handleViewNodeDetails}
                className="text-blue-500 hover:text-blue-400 dark:text-blue-400 dark:hover:text-blue-300 text-sm"
              >
                查看节点详情
              </button>
            </div>
            {node && (
              <p className="text-xs text-[var(--color-text-muted)]">IP: {node.ip}</p>
            )}
          </div>

          {/* Metric Type */}
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)]">指标类型</label>
            <p className="mt-1 text-sm text-[var(--color-text-primary)]">{getMetricDisplayName(record.metric)}</p>
          </div>

          {/* Alert Level */}
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)]">告警级别</label>
            <div className="mt-1">
              <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getLevelBadgeColor(record.level)}`}>
                {record.level}
              </span>
            </div>
          </div>

          {/* Status */}
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)]">状态</label>
            <div className="mt-1">
              <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusBadgeColor(record.status)}`}>
                {getStatusDisplayName(record.status)}
              </span>
            </div>
          </div>

          {/* Timestamps */}
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)]">创建时间</label>
            <p className="mt-1 text-sm text-[var(--color-text-primary)]">{formatTimestamp(record.created_at)}</p>
          </div>

          {record.updated_at !== record.created_at && (
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)]">更新时间</label>
              <p className="mt-1 text-sm text-[var(--color-text-primary)]">{formatTimestamp(record.updated_at)}</p>
            </div>
          )}

          {/* Status Update Actions */}
          {canEdit && record.status !== 'resolved' && (
            <div className="pt-4 border-t border-[var(--color-border)]">
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">更新状态</label>
              <div className="flex gap-3">
                {record.status === 'pending' && (
                  <button
                    type="button"
                    onClick={() => handleStatusUpdate('in_progress')}
                    disabled={isUpdating}
                    className="px-4 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700 transition-colors disabled:opacity-50"
                  >
                    标记为处理中
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => handleStatusUpdate('resolved')}
                  disabled={isUpdating}
                  className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors disabled:opacity-50"
                >
                  标记为已解决
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="mt-6 flex justify-end">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] rounded-md hover:bg-[var(--color-bg-subtle)] transition-colors"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  )
}
