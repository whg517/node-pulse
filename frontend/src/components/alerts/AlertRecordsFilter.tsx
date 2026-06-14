import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import type { AlertRecordFilters, AlertLevel, AlertRecordStatus } from '@/api/alertRecords'
import type { NodeDTO } from '@/api/types'

interface AlertRecordsFilterProps {
  filters: AlertRecordFilters
  nodes: NodeDTO[]
  searchQuery: string
  onFilterChange: (filters: AlertRecordFilters) => void
  onSearchChange: (query: string) => void
  onReset: () => void
}

export function AlertRecordsFilter({ filters, nodes, searchQuery, onFilterChange, onSearchChange, onReset }: AlertRecordsFilterProps) {
  const { t } = useTranslation()
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
    setNodeId(''); setStartTime(''); setEndTime(''); setLevel(''); setStatus(''); setSearchInput('')
    onReset()
  }

  const selectClass = 'w-full rounded-md border border-input bg-background px-3 py-2 text-sm'
  const inputClass = 'w-full rounded-md border border-input bg-background px-3 py-2 text-sm'

  return (
    <Card>
      <CardContent className="p-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <div className="space-y-2">
            <Label>{t('alerts.search', 'Search')}</Label>
            <Input
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder={t('alerts.searchPlaceholder', 'Node name or metric')}
            />
          </div>

          <div className="space-y-2">
            <Label>{t('alerts.node', 'Node')}</Label>
            <select value={nodeId} onChange={(e) => setNodeId(e.target.value)} className={selectClass}>
              <option value="">{t('alerts.allNodes', 'All Nodes')}</option>
              {nodes.map((node) => (
                <option key={node.id} value={node.id}>{node.name} ({node.ip})</option>
              ))}
            </select>
          </div>

          <div className="space-y-2">
            <Label>{t('alerts.startTime', 'Start Time')}</Label>
            <input type="datetime-local" value={startTime} onChange={(e) => setStartTime(e.target.value)} className={inputClass} />
          </div>

          <div className="space-y-2">
            <Label>{t('alerts.endTime', 'End Time')}</Label>
            <input type="datetime-local" value={endTime} onChange={(e) => setEndTime(e.target.value)} className={inputClass} />
          </div>

          <div className="space-y-2">
            <Label>{t('alerts.level', 'Level')}</Label>
            <select value={level} onChange={(e) => setLevel(e.target.value)} className={selectClass}>
              <option value="">{t('alerts.allLevels', 'All Levels')}</option>
              <option value="P0">P0</option>
              <option value="P1">P1</option>
              <option value="P2">P2</option>
            </select>
          </div>
        </div>

        <div className="mt-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <div className="space-y-2">
            <Label>{t('alerts.status', 'Status')}</Label>
            <select value={status} onChange={(e) => setStatus(e.target.value)} className={selectClass}>
              <option value="">{t('alerts.allStatuses', 'All Statuses')}</option>
              <option value="pending">{t('alerts.pending', 'Pending')}</option>
              <option value="in_progress">{t('alerts.inProgress', 'In Progress')}</option>
              <option value="resolved">{t('alerts.resolved', 'Resolved')}</option>
            </select>
          </div>
        </div>

        <div className="mt-4 flex justify-end gap-3">
          <Button variant="outline" onClick={handleReset}>{t('alerts.resetFilters', 'Reset')}</Button>
          <Button onClick={handleApply}>{t('alerts.applyFilters', 'Apply')}</Button>
        </div>
      </CardContent>
    </Card>
  )
}
