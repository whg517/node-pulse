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
  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border shadow-lg rounded-md bg-white max-w-lg w-full">
        {/* Dialog Header */}
        <div className="flex justify-between items-center pb-4 border-b border-gray-200">
          <h3 className="text-lg leading-6 font-medium text-gray-900">
            {mode === 'create' ? 'Create Alert Rule' : 'Edit Alert Rule'}
          </h3>
          <button
            type="button"
            onClick={onCancel}
            className="text-gray-400 hover:text-gray-500 focus:outline-none"
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
