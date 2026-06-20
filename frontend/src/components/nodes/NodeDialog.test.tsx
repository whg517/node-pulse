import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { NodeDialog } from './NodeDialog'
import type { NodeDTO } from '../../api/types'

function getInputByName(name: string) {
  return screen.getByRole('textbox', { name: new RegExp(name, 'i') })
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
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      expect(screen.getByText('Add New Node')).toBeInTheDocument()
      expect(screen.getByText('Create Node')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('e.g., Production Server 1')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('e.g., 192.168.1.100')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('e.g., us-east-1')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('e.g., production, critical, backend (comma-separated)')).toBeInTheDocument()
    })

    it('validates required fields', async () => {
      const onSubmit = vi.fn()
      render(
        <NodeDialog mode="create" open={true} onSubmit={onSubmit} onCancel={vi.fn()} />
      )

      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Name is required')).toBeInTheDocument()
        expect(screen.getByText('IP address is required')).toBeInTheDocument()
        expect(screen.getByText('Region is required')).toBeInTheDocument()
      })
      expect(onSubmit).not.toHaveBeenCalled()
    })

    it('validates name length', async () => {
      render(
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'A' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.1' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Name must be at least 2 characters')).toBeInTheDocument()
      })

      fireEvent.change(getInputByName('Name'), { target: { value: 'A'.repeat(101) } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Name must be less than 100 characters')).toBeInTheDocument()
      })
    })

    it('validates IP address format', async () => {
      render(
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: 'invalid-ip' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Invalid IP address format')).toBeInTheDocument()
      })
    })

    it('accepts valid IPv4 addresses', async () => {
      render(
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.queryByText('Invalid IP address format')).not.toBeInTheDocument()
      })
    })

    it('accepts valid IPv6 addresses', async () => {
      render(
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '2001:0db8:85a3:0000:0000:8a2e:0370:7334' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.queryByText('Invalid IP address format')).not.toBeInTheDocument()
      })
    })

    it('validates region length', async () => {
      render(
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.1' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'A' } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Region must be at least 2 characters')).toBeInTheDocument()
      })

      fireEvent.change(getInputByName('Region'), { target: { value: 'A'.repeat(51) } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Region must be less than 50 characters')).toBeInTheDocument()
      })
    })

    it('validates tags count', async () => {
      render(
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      const elevenTags = Array.from({ length: 11 }, (_, i) => `tag${i}`).join(', ')

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.1' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.change(getInputByName('Tags'), { target: { value: elevenTags } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Maximum 10 tags allowed')).toBeInTheDocument()
      })
    })

    it('validates individual tag length', async () => {
      render(
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.1' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.change(getInputByName('Tags'), { target: { value: 'A'.repeat(31) } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Each tag must be less than 30 characters')).toBeInTheDocument()
      })
    })

    it('submits form with valid data', async () => {
      const onSubmit = vi.fn().mockResolvedValue(undefined)
      render(
        <NodeDialog mode="create" open={true} onSubmit={onSubmit} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.change(getInputByName('Tags'), { target: { value: 'production, critical' } })
      fireEvent.click(screen.getByText('Create Node'))

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
        <NodeDialog mode="create" open={true} onSubmit={vi.fn()} onCancel={onCancel} />
      )

      fireEvent.click(screen.getByText('Cancel'))
      expect(onCancel).toHaveBeenCalled()
    })

    it('shows loading state while submitting', async () => {
      const onSubmit = vi.fn().mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)))
      render(
        <NodeDialog mode="create" open={true} onSubmit={onSubmit} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(screen.getByText('Creating...')).toBeInTheDocument()
      })
    })
  })

  describe('Edit Mode', () => {
    it('pre-fills form with node data', () => {
      render(
        <NodeDialog mode="edit" node={mockNode} open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      expect((getInputByName('Name') as HTMLInputElement).value).toBe('Test Node')
      expect((getInputByName('IP Address') as HTMLInputElement).value).toBe('192.168.1.100')
      expect((getInputByName('Region') as HTMLInputElement).value).toBe('us-east-1')
      expect((getInputByName('Tags') as HTMLInputElement).value).toBe('production, critical')
      expect(screen.getByText('Edit Node')).toBeInTheDocument()
    })

    it('displays save button in edit mode', () => {
      render(
        <NodeDialog mode="edit" node={mockNode} open={true} onSubmit={vi.fn()} onCancel={vi.fn()} />
      )

      expect(screen.getByText('Save Changes')).toBeInTheDocument()
      expect(screen.queryByText('Create Node')).not.toBeInTheDocument()
    })

    it('submits updated data', async () => {
      const onSubmit = vi.fn().mockResolvedValue(undefined)
      render(
        <NodeDialog mode="edit" node={mockNode} open={true} onSubmit={onSubmit} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Updated Node' } })
      fireEvent.click(screen.getByText('Save Changes'))

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
      render(
        <NodeDialog mode="create" open={true} onSubmit={onSubmit} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.change(getInputByName('Tags'), { target: { value: '  production ,  critical  ' } })
      fireEvent.click(screen.getByText('Create Node'))

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
      render(
        <NodeDialog mode="create" open={true} onSubmit={onSubmit} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.click(screen.getByText('Create Node'))

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
      render(
        <NodeDialog mode="create" open={true} onSubmit={onSubmit} onCancel={vi.fn()} />
      )

      fireEvent.change(getInputByName('Name'), { target: { value: 'Test Node' } })
      fireEvent.change(getInputByName('IP Address'), { target: { value: '192.168.1.100' } })
      fireEvent.change(getInputByName('Region'), { target: { value: 'us-east-1' } })
      fireEvent.click(screen.getByText('Create Node'))

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith('Failed to submit node:', expect.any(Error))
      })

      consoleErrorSpy.mockRestore()
    })
  })
})
