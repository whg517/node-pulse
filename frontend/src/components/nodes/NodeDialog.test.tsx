import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react'
import '@testing-library/jest-dom'
import { NodeDialog } from './NodeDialog'
import type { NodeDTO } from '../../api/types'

// Helper functions for querying within container
function getInputByLabel(container: HTMLElement, labelText: string) {
  return within(container).getByRole('textbox', { name: new RegExp(labelText, 'i') })
}

function getTextareaByLabel(container: HTMLElement, labelText: string) {
  return within(container).getByRole('textbox', { name: new RegExp(labelText, 'i') })
}

describe('NodeDialog', () => {
  const mockNode: NodeDTO = {
    id: 'node-1',
    name: 'Test Node',
    ip: '192.168.1.100',
    region: 'us-east-1',
    tags: ['production', 'critical'],
    status: 'online',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  describe('Create Mode', () => {
    it('renders form fields', () => {
      render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      expect(screen.getByText('nodes.nodeName', { exact: false })).toBeInTheDocument()
      expect(screen.getByText('nodes.ipAddress', { exact: false })).toBeInTheDocument()
      expect(screen.getByText('nodes.region', { exact: false })).toBeInTheDocument()
      expect(screen.getByText('nodes.tags')).toBeInTheDocument()
      expect(screen.getByText('nodes.addNode')).toBeInTheDocument()
      expect(screen.getByText('nodes.createNode')).toBeInTheDocument()
    })

    it('validates required fields', async () => {
      const onSubmit = vi.fn()
      render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.errorNameRequired')).toBeInTheDocument()
        expect(screen.getByText('nodes.errorIpRequired')).toBeInTheDocument()
        expect(screen.getByText('nodes.errorRegionRequired')).toBeInTheDocument()
      })
      expect(onSubmit).not.toHaveBeenCalled()
    })

    it('validates name length', async () => {
      const onSubmit = vi.fn()
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const nameInput = getInputByLabel(container, 'nodes.nodeName')
      const ipInput = getInputByLabel(container, 'nodes.ipAddress')
      const regionInput = getInputByLabel(container, 'nodes.region')

      // Test too short
      fireEvent.change(nameInput, { target: { value: 'A' } })
      fireEvent.change(ipInput, { target: { value: '192.168.1.1' } })
      fireEvent.change(regionInput, { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.errorNameMin')).toBeInTheDocument()
      })

      // Test too long
      const longName = 'A'.repeat(101)
      fireEvent.change(nameInput, { target: { value: longName } })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.errorNameMax')).toBeInTheDocument()
      })
    })

    it('validates IP address format', async () => {
      const onSubmit = vi.fn()
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const ipInput = getInputByLabel(container, 'nodes.ipAddress')
      const nameInput = getInputByLabel(container, 'nodes.nodeName')
      const regionInput = getInputByLabel(container, 'nodes.region')

      fireEvent.change(nameInput, { target: { value: 'Test Node' } })
      fireEvent.change(ipInput, { target: { value: 'invalid-ip' } })
      fireEvent.change(regionInput, { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.errorIpInvalid')).toBeInTheDocument()
      })
    })

    it('accepts valid IPv4 addresses', async () => {
      const onSubmit = vi.fn()
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const ipInput = getInputByLabel(container, 'nodes.ipAddress')
      const nameInput = getInputByLabel(container, 'nodes.nodeName')
      const regionInput = getInputByLabel(container, 'nodes.region')

      fireEvent.change(nameInput, { target: { value: 'Test Node' } })
      fireEvent.change(ipInput, { target: { value: '192.168.1.100' } })
      fireEvent.change(regionInput, { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.queryByText('nodes.errorIpInvalid')).not.toBeInTheDocument()
      })
    })

    it('accepts valid IPv6 addresses', async () => {
      const onSubmit = vi.fn()
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const ipInput = getInputByLabel(container, 'nodes.ipAddress')
      const nameInput = getInputByLabel(container, 'nodes.nodeName')
      const regionInput = getInputByLabel(container, 'nodes.region')

      fireEvent.change(nameInput, { target: { value: 'Test Node' } })
      fireEvent.change(ipInput, { target: { value: '2001:0db8:85a3:0000:0000:8a2e:0370:7334' } })
      fireEvent.change(regionInput, { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.queryByText('nodes.errorIpInvalid')).not.toBeInTheDocument()
      })
    })

    it('validates region length', async () => {
      const onSubmit = vi.fn()
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const regionInput = getInputByLabel(container, 'nodes.region')
      const nameInput = getInputByLabel(container, 'nodes.nodeName')
      const ipInput = getInputByLabel(container, 'nodes.ipAddress')

      // Test too short
      fireEvent.change(nameInput, { target: { value: 'Test Node' } })
      fireEvent.change(ipInput, { target: { value: '192.168.1.1' } })
      fireEvent.change(regionInput, { target: { value: 'A' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.errorRegionMin')).toBeInTheDocument()
      })

      // Test too long
      const longRegion = 'A'.repeat(51)
      fireEvent.change(regionInput, { target: { value: longRegion } })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.errorRegionMax')).toBeInTheDocument()
      })
    })

    it('validates tags count', async () => {
      const onSubmit = vi.fn()
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const tagsInput = getTextareaByLabel(container, 'nodes.tags')
      const nameInput = getInputByLabel(container, 'nodes.nodeName')
      const ipInput = getInputByLabel(container, 'nodes.ipAddress')
      const regionInput = getInputByLabel(container, 'nodes.region')

      const elevenTags = Array.from({ length: 11 }, (_, i) => `tag${i}`).join(', ')

      fireEvent.change(nameInput, { target: { value: 'Test Node' } })
      fireEvent.change(ipInput, { target: { value: '192.168.1.1' } })
      fireEvent.change(regionInput, { target: { value: 'us-east-1' } })
      fireEvent.change(tagsInput, { target: { value: elevenTags } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.errorTagsMax')).toBeInTheDocument()
      })
    })

    it('validates individual tag length', async () => {
      const onSubmit = vi.fn()
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const tagsInput = getTextareaByLabel(container, 'nodes.tags')
      const nameInput = getInputByLabel(container, 'nodes.nodeName')
      const ipInput = getInputByLabel(container, 'nodes.ipAddress')
      const regionInput = getInputByLabel(container, 'nodes.region')

      const longTag = 'A'.repeat(31)

      fireEvent.change(nameInput, { target: { value: 'Test Node' } })
      fireEvent.change(ipInput, { target: { value: '192.168.1.1' } })
      fireEvent.change(regionInput, { target: { value: 'us-east-1' } })
      fireEvent.change(tagsInput, { target: { value: longTag } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.errorTagLength')).toBeInTheDocument()
      })
    })

    it('submits form with valid data', async () => {
      const onSubmit = vi.fn().mockResolvedValue(undefined)
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      fireEvent.change(getInputByLabel(container, 'nodes.nodeName'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'nodes.ipAddress'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'nodes.region'), { target: { value: 'us-east-1' } })
      fireEvent.change(getTextareaByLabel(container, 'nodes.tags'), { target: { value: 'production, critical' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith({
          name: 'Test Node',
          ip: '192.168.1.100',
          region: 'us-east-1',
          tags: ['production', 'critical'],
        })
      })
    })

    it('calls onCancel when cancel button clicked', () => {
      const onCancel = vi.fn()
      render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={onCancel}
        />
      )

      const cancelButton = screen.getByText('common.cancel')
      fireEvent.click(cancelButton)

      expect(onCancel).toHaveBeenCalled()
    })

    it('shows loading state while submitting', async () => {
      const onSubmit = vi.fn().mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)))
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      fireEvent.change(getInputByLabel(container, 'nodes.nodeName'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'nodes.ipAddress'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'nodes.region'), { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('nodes.creating')).toBeInTheDocument()
        expect(submitButton).toBeDisabled()
      })
    })
  })

  describe('Edit Mode', () => {
    it('pre-fills form with node data', () => {
      const { container } = render(
        <NodeDialog
          mode="edit"
          node={mockNode}
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const nameInput = getInputByLabel(container, 'nodes.nodeName') as HTMLInputElement
      const ipInput = getInputByLabel(container, 'nodes.ipAddress') as HTMLInputElement
      const regionInput = getInputByLabel(container, 'nodes.region') as HTMLInputElement
      const tagsInput = getTextareaByLabel(container, 'nodes.tags') as HTMLInputElement

      expect(nameInput.value).toBe('Test Node')
      expect(ipInput.value).toBe('192.168.1.100')
      expect(regionInput.value).toBe('us-east-1')
      expect(tagsInput.value).toBe('production, critical')
      expect(screen.getByText('nodes.editNode')).toBeInTheDocument()
    })

    it('displays save button in edit mode', () => {
      render(
        <NodeDialog
          mode="edit"
          node={mockNode}
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      expect(screen.getByText('common.saveChanges')).toBeInTheDocument()
      expect(screen.queryByText('nodes.createNode')).not.toBeInTheDocument()
    })

    it('submits updated data', async () => {
      const onSubmit = vi.fn().mockResolvedValue(undefined)
      const { container } = render(
        <NodeDialog
          mode="edit"
          node={mockNode}
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      const nameInput = getInputByLabel(container, 'nodes.nodeName')
      fireEvent.change(nameInput, { target: { value: 'Updated Node' } })

      const submitButton = screen.getByText('common.saveChanges')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith({
          name: 'Updated Node',
          ip: '192.168.1.100',
          region: 'us-east-1',
          tags: ['production', 'critical'],
        })
      })
    })
  })

  describe('Edge Cases', () => {
    it('handles tags with extra spaces', async () => {
      const onSubmit = vi.fn().mockResolvedValue(undefined)
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      fireEvent.change(getInputByLabel(container, 'nodes.nodeName'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'nodes.ipAddress'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'nodes.region'), { target: { value: 'us-east-1' } })
      fireEvent.change(getTextareaByLabel(container, 'nodes.tags'), { target: { value: '  production ,  critical  ' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith({
          name: 'Test Node',
          ip: '192.168.1.100',
          region: 'us-east-1',
          tags: ['production', 'critical'],
        })
      })
    })

    it('handles empty tags', async () => {
      const onSubmit = vi.fn().mockResolvedValue(undefined)
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      fireEvent.change(getInputByLabel(container, 'nodes.nodeName'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'nodes.ipAddress'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'nodes.region'), { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith({
          name: 'Test Node',
          ip: '192.168.1.100',
          region: 'us-east-1',
          tags: [],
        })
      })
    })

    it('handles submission error gracefully', async () => {
      const onSubmit = vi.fn().mockRejectedValue(new Error('Network error'))
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={onSubmit}
          onCancel={vi.fn()}
        />
      )

      fireEvent.change(getInputByLabel(container, 'nodes.nodeName'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'nodes.ipAddress'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'nodes.region'), { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('nodes.createNode')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith('Failed to submit node:', expect.any(Error))
        expect(submitButton).not.toBeDisabled()
      })

      consoleErrorSpy.mockRestore()
    })
  })
})
