/**
 * NodeDialog Component
 *
 * Modal dialog for creating or editing a node.
 * Handles form validation and submission.
 */

import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import type { NodeDTO, CreateNodeRequest, UpdateNodeRequest } from '../../api/types'

interface NodeDialogProps {
  mode: 'create' | 'edit'
  node?: NodeDTO
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

export function NodeDialog({ mode, node, onSubmit, onCancel }: NodeDialogProps) {
  const { t } = useTranslation()
  const [formData, setFormData] = useState<FormData>({
    name: '',
    ip: '',
    region: '',
    tags: '',
  })
  const [errors, setErrors] = useState<FormErrors>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Initialize form data when node changes
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

    // Name validation
    if (!formData.name.trim()) {
      newErrors.name = t('nodes.errorNameRequired')
    } else if (formData.name.length < 2) {
      newErrors.name = t('nodes.errorNameMin')
    } else if (formData.name.length > 100) {
      newErrors.name = t('nodes.errorNameMax')
    }

    // IP validation
    if (!formData.ip.trim()) {
      newErrors.ip = t('nodes.errorIpRequired')
    } else {
      // Simple IPv4 validation
      const ipv4Regex =
        /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
      // Simple IPv6 validation (simplified)
      const ipv6Regex = /^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$/

      if (!ipv4Regex.test(formData.ip) && !ipv6Regex.test(formData.ip)) {
        newErrors.ip = t('nodes.errorIpInvalid')
      }
    }

    // Region validation
    if (!formData.region.trim()) {
      newErrors.region = t('nodes.errorRegionRequired')
    } else if (formData.region.length < 2) {
      newErrors.region = t('nodes.errorRegionMin')
    } else if (formData.region.length > 50) {
      newErrors.region = t('nodes.errorRegionMax')
    }

    // Tags validation
    const tagArray = formData.tags
      .split(',')
      .map((t) => t.trim())
      .filter((t) => t.length > 0)

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

    if (!validate()) {
      return
    }

    setIsSubmitting(true)
    try {
      const tagArray = formData.tags
        .split(',')
        .map((t) => t.trim())
        .filter((t) => t.length > 0)

      const data: CreateNodeRequest | UpdateNodeRequest =
        mode === 'create'
          ? ({
              name: formData.name.trim(),
              ip: formData.ip.trim(),
              region: formData.region.trim(),
              tags: tagArray,
            } as CreateNodeRequest)
          : ({
              name: formData.name.trim(),
              ip: formData.ip.trim(),
              region: formData.region.trim(),
              tags: tagArray,
            } as UpdateNodeRequest)

      await onSubmit(data)
    } catch (error) {
      console.error('Failed to submit node:', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target
    setFormData((prev) => ({ ...prev, [name]: value }))
    // Clear error for this field when user starts typing
    if (errors[name as keyof FormErrors]) {
      setErrors((prev) => ({ ...prev, [name]: undefined }))
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-slate-800 rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-slate-700">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-slate-100">
            {mode === 'create' ? t('nodes.addNode') : t('nodes.editNode')}
          </h3>
        </div>

        <form onSubmit={handleSubmit} className="px-6 py-4 space-y-4">
          {/* Name Field */}
          <div>
            <label
              htmlFor="name"
              className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1"
            >
              {t('nodes.nodeName')} <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              disabled={isSubmitting}
              className="w-full px-3 py-2 border border-gray-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-gray-900 dark:text-slate-100 placeholder-gray-400 dark:placeholder-slate-500 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 dark:disabled:bg-slate-600"
              placeholder="e.g., Production Server 1"
            />
            {errors.name && (
              <p className="mt-1 text-sm text-red-600">{errors.name}</p>
            )}
          </div>

          {/* IP Address Field */}
          <div>
            <label
              htmlFor="ip"
              className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1"
            >
              {t('nodes.ipAddress')} <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="ip"
              name="ip"
              value={formData.ip}
              onChange={handleChange}
              disabled={isSubmitting}
              className="w-full px-3 py-2 border border-gray-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-gray-900 dark:text-slate-100 placeholder-gray-400 dark:placeholder-slate-500 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 dark:disabled:bg-slate-600 font-mono"
              placeholder="e.g., 192.168.1.100"
            />
            {errors.ip && (
              <p className="mt-1 text-sm text-red-600">{errors.ip}</p>
            )}
          </div>

          {/* Region Field */}
          <div>
            <label
              htmlFor="region"
              className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1"
            >
              {t('nodes.region')} <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="region"
              name="region"
              value={formData.region}
              onChange={handleChange}
              disabled={isSubmitting}
              className="w-full px-3 py-2 border border-gray-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-gray-900 dark:text-slate-100 placeholder-gray-400 dark:placeholder-slate-500 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 dark:disabled:bg-slate-600"
              placeholder="e.g., us-east-1"
            />
            {errors.region && (
              <p className="mt-1 text-sm text-red-600">{errors.region}</p>
            )}
          </div>

          {/* Tags Field */}
          <div>
            <label
              htmlFor="tags"
              className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1"
            >
              {t('nodes.tags')}
            </label>
            <textarea
              id="tags"
              name="tags"
              value={formData.tags}
              onChange={handleChange}
              disabled={isSubmitting}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-gray-900 dark:text-slate-100 placeholder-gray-400 dark:placeholder-slate-500 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-100 dark:disabled:bg-slate-600"
              placeholder="e.g., production, critical, backend (comma-separated)"
            />
            <p className="mt-1 text-xs text-gray-500 dark:text-slate-400">
              {t('nodes.tagsHint')}
            </p>
            {errors.tags && (
              <p className="mt-1 text-sm text-red-600">{errors.tags}</p>
            )}
          </div>

          {/* Actions */}
          <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200 dark:border-slate-700">
            <button
              type="button"
              onClick={onCancel}
              disabled={isSubmitting}
              className="px-4 py-2 bg-gray-200 dark:bg-slate-600 text-gray-800 dark:text-slate-200 rounded-md hover:bg-gray-300 dark:hover:bg-slate-500 transition-colors disabled:bg-gray-100 dark:disabled:bg-slate-700 disabled:cursor-not-allowed"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:bg-blue-300 disabled:cursor-not-allowed"
            >
              {isSubmitting
                ? mode === 'create'
                  ? t('nodes.creating')
                  : t('common.saving')
                : mode === 'create'
                ? t('nodes.createNode')
                : t('common.saveChanges')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
