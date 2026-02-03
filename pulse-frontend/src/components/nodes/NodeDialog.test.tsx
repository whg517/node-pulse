import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { NodeDialog } from './NodeDialog'
import type { NodeDTO } from '../../api/types'

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

      expect(screen.getByText('Name')).toBeInTheDocument()
      expect(screen.getByText('IP Address')).toBeInTheDocument()
      expect(screen.getByText('Region')).toBeInTheDocument()
      expect(screen.getByText('Tags')).toBeInTheDocument()
      expect(screen.getByText('Add New Node')).toBeInTheDocument()
      expect(screen.getByText('Create Node')).toBeInTheDocument()
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

      const submitButton = screen.getByText('Create Node')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Name is required')).toBeInTheDocument()
        expect(screen.getByText('IP address is required')).toBeInTheDocument()
        expect(screen.getByText('Region is required')).toBeInTheDocument()
      })
      expect(onSubmit).not.toHaveBeenCalled()
    })

    it('validates name length', async () => {
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const nameInput = getInputByLabel(container, 'Name')

      // Test too short
      fireEvent.change(nameInput, { target: { value: 'A' } })
      fireEvent.blur(nameInput)

      await waitFor(() => {
        expect(screen.getByText('Name must be at least 2 characters')).toBeInTheDocument()
      })

      // Test too long
      const longName = 'A'.repeat(101)
      fireEvent.change(nameInput, { target: { value: longName } })
      fireEvent.blur(nameInput)

      await waitFor(() => {
        expect(screen.getByText('Name must be less than 100 characters')).toBeInTheDocument()
      })
    })

    it('validates IP address format', async () => {
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const ipInput = getInputByLabel(container, 'IP Address')
      fireEvent.change(ipInput, { target: { value: 'invalid-ip' } })
      fireEvent.blur(ipInput)

      await waitFor(() => {
        expect(screen.getByText('Invalid IP address format')).toBeInTheDocument()
      })
    })

    it('accepts valid IPv4 addresses', async () => {
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const ipInput = getInputByLabel(container, 'IP Address')
      fireEvent.change(ipInput, { target: { value: '192.168.1.100' } })
      fireEvent.blur(ipInput)

      await waitFor(() => {
        expect(screen.queryByText('Invalid IP address format')).not.toBeInTheDocument()
      })
    })

    it('accepts valid IPv6 addresses', async () => {
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const ipInput = getInputByLabel(container, 'IP Address')
      fireEvent.change(ipInput, { target: { value: '2001:0db8:85a3:0000:0000:8a2e:0370:7334' } })
      fireEvent.blur(ipInput)

      await waitFor(() => {
        expect(screen.queryByText('Invalid IP address format')).not.toBeInTheDocument()
      })
    })

    it('validates region length', async () => {
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const regionInput = getInputByLabel(container, 'Region')

      // Test too short
      fireEvent.change(regionInput, { target: { value: 'A' } })
      fireEvent.blur(regionInput)

      await waitFor(() => {
        expect(screen.getByText('Region must be at least 2 characters')).toBeInTheDocument()
      })

      // Test too long
      const longRegion = 'A'.repeat(51)
      fireEvent.change(regionInput, { target: { value: longRegion } })
      fireEvent.blur(regionInput)

      await waitFor(() => {
        expect(screen.getByText('Region must be less than 50 characters')).toBeInTheDocument()
      })
    })

    it('validates tags count', async () => {
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const tagsInput = getTextareaByLabel(container, 'Tags')
      const elevenTags = Array.from({ length: 11 }, (_, i) => `tag${i}`).join(', ')

      fireEvent.change(tagsInput, { target: { value: elevenTags } })
      fireEvent.blur(tagsInput)

      await waitFor(() => {
        expect(screen.getByText('Maximum 10 tags allowed')).toBeInTheDocument()
      })
    })

    it('validates individual tag length', async () => {
      const { container } = render(
        <NodeDialog
          mode="create"
          onSubmit={vi.fn()}
          onCancel={vi.fn()}
        />
      )

      const tagsInput = getTextareaByLabel(container, 'Tags')
      const longTag = 'A'.repeat(31)

      fireEvent.change(tagsInput, { target: { value: longTag } })
      fireEvent.blur(tagsInput)

      await waitFor(() => {
        expect(screen.getByText('Each tag must be less than 30 characters')).toBeInTheDocument()
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

      fireEvent.change(getInputByLabel(container, 'Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'Region'), { target: { value: 'us-east-1' } })
      fireEvent.change(getTextareaByLabel(container, 'Tags'), { target: { value: 'production, critical' } })

      const submitButton = screen.getByText('Create Node')
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

      const cancelButton = screen.getByText('Cancel')
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

      fireEvent.change(getInputByLabel(container, 'Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'Region'), { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('Create Node')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Creating...')).toBeInTheDocument()
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

      const nameInput = getInputByLabel(container, 'Name')
      const ipInput = getInputByLabel(container, 'IP Address')
      const regionInput = getInputByLabel(container, 'Region')
      const tagsInput = getTextareaByLabel(container, 'Tags')

      expect(nameInput.value).toBe('Test Node')
      expect(ipInput.value).toBe('192.168.1.100')
      expect(regionInput.value).toBe('us-east-1')
      expect(tagsInput.value).toBe('production, critical')
      expect(screen.getByText('Edit Node')).toBeInTheDocument()
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

      expect(screen.getByText('Save Changes')).toBeInTheDocument()
      expect(screen.queryByText('Create Node')).not.toBeInTheDocument()
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

      const nameInput = getInputByLabel(container, 'Name')
      fireEvent.change(nameInput, { target: { value: 'Updated Node' } })

      const submitButton = screen.getByText('Save Changes')
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

      fireEvent.change(getInputByLabel(container, 'Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'Region'), { target: { value: 'us-east-1' } })
      fireEvent.change(getTextareaByLabel(container, 'Tags'), { target: { value: '  production ,  critical  ' } })

      const submitButton = screen.getByText('Create Node')
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

      fireEvent.change(getInputByLabel(container, 'Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'Region'), { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('Create Node')
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

      fireEvent.change(getInputByLabel(container, 'Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByLabel(container, 'IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByLabel(container, 'Region'), { target: { value: 'us-east-1' } })

      const submitButton = screen.getByText('Create Node')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith('Failed to submit node:', expect.any(Error))
        expect(submitButton).not.toBeDisabled()
      })

      consoleErrorSpy.mockRestore()
    })
  })
})
