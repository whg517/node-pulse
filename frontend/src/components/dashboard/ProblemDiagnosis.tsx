import { useState } from 'react'

export type ProblemType = 'node_local' | 'cross_border_link' | 'carrier_routing' | 'none'
export type ConfidenceLevel = 'high' | 'medium' | 'low'

export interface ProblemDiagnosisProps {
  problemType: ProblemType
  confidence: ConfidenceLevel
  details?: string
  className?: string
  isExpanded?: boolean
}

/**
 * ProblemDiagnosis component for displaying problem type detection
 *
 * Shows the detected problem type with confidence level and detailed information.
 * Supports expandable card interaction pattern.
 *
 * @param props - ProblemDiagnosis props
 * @returns ProblemDiagnosis component
 *
 * @example
 * <ProblemDiagnosis
 *   problemType="node_local"
 *   confidence="high"
 *   details="High latency detected on this node only"
 *   isExpanded={false}
 * />
 */
export default function ProblemDiagnosis({
  problemType,
  confidence,
  details = '',
  className = '',
  isExpanded: initiallyExpanded = false,
}: ProblemDiagnosisProps) {
  const [isExpanded, setIsExpanded] = useState(initiallyExpanded)

  const problemConfig = {
    node_local: {
      label: 'Node Local Fault',
      labelZh: '节点本地故障',
      description: 'Issue detected on this specific node only',
      descriptionZh: '仅在此节点检测到问题',
      color: 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-800 dark:text-red-300',
      icon: '⚠️',
    },
    cross_border_link: {
      label: 'Cross-Border Link Issue',
      labelZh: '跨境链路问题',
      description: 'Issue affecting multiple nodes across regions',
      descriptionZh: '影响多个跨区域节点的问题',
      color: 'bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800 text-orange-800 dark:text-orange-300',
      icon: '🌍',
    },
    carrier_routing: {
      label: 'Carrier Routing Issue',
      labelZh: '运营商路由问题',
      description: 'Issue related to ISP routing changes',
      descriptionZh: '与ISP路由变更相关的问题',
      color: 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-300',
      icon: '🔀',
    },
    none: {
      label: 'No Issues Detected',
      labelZh: '未检测到问题',
      description: 'All metrics are within normal ranges',
      descriptionZh: '所有指标均在正常范围内',
      color: 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-800 dark:text-green-300',
      icon: '✓',
    },
  }

  const config = problemConfig[problemType]

  const confidenceConfig = {
    high: { label: 'High', labelZh: '高', color: 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-300' },
    medium: { label: 'Medium', labelZh: '中', color: 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-300' },
    low: { label: 'Low', labelZh: '低', color: 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300' },
  }

  const conf = confidenceConfig[confidence]

  return (
    <div
      className={`problem-diagnosis bg-[var(--color-bg-surface)] rounded-lg border-2 p-4 shadow-sm ${config.color} ${className}`}
      role="region"
      aria-label="Problem diagnosis"
    >
      {/* Header - always visible */}
      <div
        className="flex items-center justify-between cursor-pointer"
        onClick={() => setIsExpanded(!isExpanded)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            setIsExpanded(!isExpanded)
          }
        }}
        role="button"
        tabIndex={0}
        aria-expanded={isExpanded}
      >
        <div className="flex items-center space-x-3">
          <span className="text-2xl" aria-hidden="true">
            {config.icon}
          </span>
          <div>
            <h3 className="text-lg font-semibold">{config.label}</h3>
            <p className="text-sm opacity-75">{config.description}</p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <span
            className={`px-3 py-1 rounded-full text-sm font-medium ${conf.color}`}
            aria-label={`Confidence level: ${conf.label}`}
          >
            {conf.label} Confidence
          </span>
          <button
            className="text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] transition-colors"
            aria-label={isExpanded ? 'Collapse details' : 'Expand details'}
          >
            <svg
              className={`w-5 h-5 transform transition-transform ${isExpanded ? 'rotate-180' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 9l-7 7-7-7"
              />
            </svg>
          </button>
        </div>
      </div>

      {/* Expandable details */}
      {isExpanded && (
        <div className="mt-4 pt-4 border-t border-current border-opacity-20">
          <div className="space-y-3">
            <div>
              <h4 className="font-semibold mb-1">Chinese (中文)</h4>
              <p className="text-sm opacity-90">{config.labelZh}</p>
              <p className="text-sm opacity-75 mt-1">{config.descriptionZh}</p>
            </div>

            {details && (
              <div>
                <h4 className="font-semibold mb-1">Diagnostic Details</h4>
                <p className="text-sm opacity-90">{details}</p>
              </div>
            )}

            <div className="text-sm opacity-75 italic">
              Note: This is an automated assessment based on multi-node comparison.
              Actual root cause may require further investigation.
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
