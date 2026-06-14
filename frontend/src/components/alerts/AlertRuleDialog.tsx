import { useTranslation } from 'react-i18next'
import { AlertRuleForm } from './AlertRuleForm'
import type { AlertRule } from '@/stores/types'
import type { NodeDTO, CreateAlertRuleRequest } from '@/api/types'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface AlertRuleDialogProps {
  mode: 'create' | 'edit'
  initialData?: AlertRule
  nodes: NodeDTO[]
  onSubmit: (data: CreateAlertRuleRequest) => Promise<void>
  onCancel: () => void
  open: boolean
}

export function AlertRuleDialog({ mode, initialData, nodes, onSubmit, onCancel, open }: AlertRuleDialogProps) {
  const { t } = useTranslation()
  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onCancel() }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{mode === 'create' ? t('alerts.createRule') : t('alerts.editRule')}</DialogTitle>
        </DialogHeader>
        <AlertRuleForm mode={mode} initialData={initialData} nodes={nodes} onSubmit={onSubmit} onCancel={onCancel} />
      </DialogContent>
    </Dialog>
  )
}
