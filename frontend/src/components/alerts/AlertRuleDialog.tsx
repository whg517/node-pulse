import { useTranslation } from 'react-i18next'
import { AlertRuleForm } from './AlertRuleForm'
import type { AlertRule } from '../../stores/types'
import type { NodeDTO, CreateAlertRuleRequest } from '../../api/types'

interface AlertRuleDialogProps {
  mode: 'create' | 'edit'
  initialData?: AlertRule
  nodes: NodeDTO[]
  onSubmit: (data: CreateAlertRuleRequest) => Promise<void>
  onCancel: () => void
}

export function AlertRuleDialog({ mode, initialData, nodes, onSubmit, onCancel }: AlertRuleDialogProps) {
  const { t } = useTranslation()
  return (
    <div className="fixed inset-0 bg-black/50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border border-[var(--color-border)] shadow-lg rounded-md bg-[var(--color-bg-elevated)] max-w-lg w-full">
        {/* Dialog Header */}
        <div className="flex justify-between items-center pb-4 border-b border-[var(--color-border)]">
          <h3 className="text-lg leading-6 font-medium text-[var(--color-text-primary)]">
            {mode === 'create' ? t('alerts.createRule') : t('alerts.editRule')}
          </h3>
          <button
            type="button"
            onClick={onCancel}
            className="text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] focus:outline-none"
          >
            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Dialog Body */}
        <div className="mt-4">
          <AlertRuleForm
            mode={mode}
            initialData={initialData}
            nodes={nodes}
            onSubmit={onSubmit}
            onCancel={onCancel}
          />
        </div>
      </div>
    </div>
  )
}
