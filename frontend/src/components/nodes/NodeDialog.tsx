import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import type { NodeDTO, CreateNodeRequest, UpdateNodeRequest } from '../../api/types'

interface NodeDialogProps {
  mode: 'create' | 'edit'
  node?: NodeDTO
  open: boolean
  onSubmit: (data: CreateNodeRequest | UpdateNodeRequest) => Promise<void>
  onCancel: () => void
}

interface FormData {
  name: string
  ip: string
  region: string
  tags: string
}

interface FormErrors {
  name?: string
  ip?: string
  region?: string
  tags?: string
}

export function NodeDialog({ mode, node, open, onSubmit, onCancel }: NodeDialogProps) {
  const { t } = useTranslation()
  const [formData, setFormData] = useState<FormData>({
    name: '',
    ip: '',
    region: '',
    tags: '',
  })
  const [errors, setErrors] = useState<FormErrors>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (node && mode === 'edit') {
      setFormData({
        name: node.name,
        ip: node.ip,
        region: node.region,
        tags: node.tags.join(', '),
      })
    }
  }, [node, mode])

  const validate = (): boolean => {
    const newErrors: FormErrors = {}

    if (!formData.name.trim()) {
      newErrors.name = t('nodes.errorNameRequired')
    } else if (formData.name.length < 2) {
      newErrors.name = t('nodes.errorNameMin')
    } else if (formData.name.length > 100) {
      newErrors.name = t('nodes.errorNameMax')
    }

    if (!formData.ip.trim()) {
      newErrors.ip = t('nodes.errorIpRequired')
    } else {
      const ipv4Regex =
        /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
      const ipv6Regex = /^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$/

      if (!ipv4Regex.test(formData.ip) && !ipv6Regex.test(formData.ip)) {
        newErrors.ip = t('nodes.errorIpInvalid')
      }
    }

    if (!formData.region.trim()) {
      newErrors.region = t('nodes.errorRegionRequired')
    } else if (formData.region.length < 2) {
      newErrors.region = t('nodes.errorRegionMin')
    } else if (formData.region.length > 50) {
      newErrors.region = t('nodes.errorRegionMax')
    }

    const tagArray = formData.tags
      .split(',')
      .map((v) => v.trim())
      .filter((v) => v.length > 0)

    if (tagArray.length > 10) {
      newErrors.tags = t('nodes.errorTagsMax')
    }

    tagArray.forEach((tag) => {
      if (tag.length > 30) {
        newErrors.tags = t('nodes.errorTagLength')
      }
    })

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validate()) return

    setIsSubmitting(true)
    try {
      const tagArray = formData.tags
        .split(',')
        .map((v) => v.trim())
        .filter((v) => v.length > 0)

      const data: CreateNodeRequest | UpdateNodeRequest = {
        name: formData.name.trim(),
        ip: formData.ip.trim(),
        region: formData.region.trim(),
        tags: tagArray,
      }

      await onSubmit(data)
    } catch (error) {
      console.error('Failed to submit node:', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target
    setFormData((prev) => ({ ...prev, [name]: value }))
    if (errors[name as keyof FormErrors]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }))
    }
  }

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onCancel() }}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {mode === 'create' ? t('nodes.addNode') : t('nodes.editNode')}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">{t('nodes.nodeName')} <span className="text-destructive">*</span></Label>
            <Input
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              disabled={isSubmitting}
              placeholder="e.g., Production Server 1"
            />
            {errors.name && <p className="text-sm text-destructive">{errors.name}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="ip">{t('nodes.ipAddress')} <span className="text-destructive">*</span></Label>
            <Input
              id="ip"
              name="ip"
              value={formData.ip}
              onChange={handleChange}
              disabled={isSubmitting}
              placeholder="e.g., 192.168.1.100"
              className="font-mono"
            />
            {errors.ip && <p className="text-sm text-destructive">{errors.ip}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="region">{t('nodes.region')} <span className="text-destructive">*</span></Label>
            <Input
              id="region"
              name="region"
              value={formData.region}
              onChange={handleChange}
              disabled={isSubmitting}
              placeholder="e.g., us-east-1"
            />
            {errors.region && <p className="text-sm text-destructive">{errors.region}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="tags">{t('nodes.tags')}</Label>
            <Textarea
              id="tags"
              name="tags"
              value={formData.tags}
              onChange={handleChange}
              disabled={isSubmitting}
              rows={3}
              placeholder="e.g., production, critical, backend (comma-separated)"
            />
            <p className="text-xs text-muted-foreground">{t('nodes.tagsHint')}</p>
            {errors.tags && <p className="text-sm text-destructive">{errors.tags}</p>}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting
                ? mode === 'create' ? t('nodes.creating') : t('common.saving')
                : mode === 'create' ? t('nodes.createNode') : t('common.saveChanges')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
