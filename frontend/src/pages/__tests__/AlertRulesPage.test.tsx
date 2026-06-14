import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AlertRulesPage from '../AlertRulesPage'

vi.mock('../../hooks/useTheme', () => ({
  useTheme: () => ({ isDark: false }),
}))

vi.mock('../../api/nodes', () => ({
  fetchNodes: vi.fn().mockResolvedValue({
    data: { nodes: [] },
  }),
}))

const mockFetchAlertRules = vi.fn().mockResolvedValue(undefined)
const mockAddAlertRule = vi.fn().mockResolvedValue(undefined)
const mockUpdateAlertRule = vi.fn().mockResolvedValue(undefined)
const mockRemoveAlertRule = vi.fn().mockResolvedValue(undefined)

vi.mock('../../stores/alertsStore', () => ({
  useAlertsStore: vi.fn(() => ({
    alertRules: [
      {
        id: 'rule-1',
        name: 'High Latency',
        node_id: 'node-1',
        metric: 'latency_ms',
        condition: '>',
        threshold: 100,
        level: 'P1',
        enabled: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ],
    fetchAlertRules: mockFetchAlertRules,
    addAlertRule: mockAddAlertRule,
    updateAlertRule: mockUpdateAlertRule,
    removeAlertRule: mockRemoveAlertRule,
  })),
}))

vi.mock('../../stores/authStore', () => ({
  useAuthStore: vi.fn(() => ({ user: { id: 'u1', username: 'admin', role: 'admin' } })),
}))

vi.mock('../../components/alerts/AlertRulesTable', () => ({
  AlertRulesTable: ({ onEdit, onDelete }: { onEdit: (id: string) => void; onDelete: (id: string) => void }) => (
    <div data-testid="alert-rules-table">
      <button onClick={() => onEdit('rule-1')}>Edit Rule</button>
      <button onClick={() => onDelete('rule-1')}>Delete Rule</button>
    </div>
  ),
}))

vi.mock('../../components/alerts/AlertRuleDialog', () => ({
  AlertRuleDialog: ({ onCancel, onSubmit }: { onCancel: () => void; onSubmit: (r: unknown) => void }) => (
    <div data-testid="alert-rule-dialog">
      <button onClick={onCancel}>Close Dialog</button>
      <button onClick={() => onSubmit({ name: 'New Rule', metric: 'latency_ms' })}>Save Rule</button>
    </div>
  ),
}))

describe('AlertRulesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads and displays alert rules table', async () => {
    render(<MemoryRouter><AlertRulesPage /></MemoryRouter>)
    await waitFor(() => {
      expect(screen.getByTestId('alert-rules-table')).toBeInTheDocument()
    })
  })

  it('opens create dialog on add button click', async () => {
    render(<MemoryRouter><AlertRulesPage /></MemoryRouter>)
    await waitFor(() => expect(screen.getByTestId('alert-rules-table')).toBeInTheDocument())
    fireEvent.click(screen.getByText('Create Alert Rule'))
    expect(screen.getByTestId('alert-rule-dialog')).toBeInTheDocument()
  })

  it('opens edit dialog from table', async () => {
    render(<MemoryRouter><AlertRulesPage /></MemoryRouter>)
    await waitFor(() => expect(screen.getByTestId('alert-rules-table')).toBeInTheDocument())
    fireEvent.click(screen.getByText('Edit Rule'))
    expect(screen.getByTestId('alert-rule-dialog')).toBeInTheDocument()
  })

  it('opens delete confirm dialog', async () => {
    render(<MemoryRouter><AlertRulesPage /></MemoryRouter>)
    await waitFor(() => expect(screen.getByTestId('alert-rules-table')).toBeInTheDocument())
    fireEvent.click(screen.getByText('Delete Rule'))
    expect(screen.getByText('Delete Alert Rule')).toBeInTheDocument()
  })

  it('saves new rule via dialog', async () => {
    render(<MemoryRouter><AlertRulesPage /></MemoryRouter>)
    await waitFor(() => expect(screen.getByTestId('alert-rules-table')).toBeInTheDocument())
    fireEvent.click(screen.getByText('Create Alert Rule'))
    fireEvent.click(screen.getByText('Save Rule'))
    await waitFor(() => expect(mockAddAlertRule).toHaveBeenCalled())
  })

  it('saves edited rule via dialog', async () => {
    render(<MemoryRouter><AlertRulesPage /></MemoryRouter>)
    await waitFor(() => expect(screen.getByTestId('alert-rules-table')).toBeInTheDocument())
    fireEvent.click(screen.getByText('Edit Rule'))
    fireEvent.click(screen.getByText('Save Rule'))
    await waitFor(() => expect(mockUpdateAlertRule).toHaveBeenCalled())
  })

  it('deletes rule after confirmation', async () => {
    render(<MemoryRouter><AlertRulesPage /></MemoryRouter>)
    await waitFor(() => expect(screen.getByTestId('alert-rules-table')).toBeInTheDocument())
    fireEvent.click(screen.getByText('Delete Rule'))
    // Find and click confirm button in the dialog
    fireEvent.click(screen.getByText('Delete'))
    await waitFor(() => expect(mockRemoveAlertRule).toHaveBeenCalledWith('rule-1'))
  })

  it('handles fetch error', async () => {
    mockFetchAlertRules.mockRejectedValueOnce(new Error('Load failed'))
    render(<MemoryRouter><AlertRulesPage /></MemoryRouter>)
    await waitFor(() => {
      expect(screen.getByText('Load failed')).toBeInTheDocument()
    })
  })
})
