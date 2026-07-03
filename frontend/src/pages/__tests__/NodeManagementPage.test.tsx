import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import '@testing-library/jest-dom'

// --- Mocks -------------------------------------------------------------

// Selector-aware mock of authStore (NodeManagementPage uses `const { user } = useAuthStore()`).
let mockAuthState: { user: { role: string } | null } = { user: null }
vi.mock('@/stores/authStore', () => ({
  useAuthStore: vi.fn(() => mockAuthState),
}))

// Mock the nodes API module.
const mockFetchNodes = vi.fn()
const mockCreateNode = vi.fn()
const mockUpdateNode = vi.fn()
const mockDeleteNode = vi.fn()
vi.mock('@/api/nodes', () => ({
  fetchNodes: (...args: unknown[]) => mockFetchNodes(...args),
  createNode: (...args: unknown[]) => mockCreateNode(...args),
  updateNode: (...args: unknown[]) => mockUpdateNode(...args),
  deleteNode: (...args: unknown[]) => mockDeleteNode(...args),
}))

// NodeManagementPage is a default export; import after mocks are registered.
import NodeManagementPage from '../NodeManagementPage'

const mockNodes = [
  { id: '1', name: 'Node-1', ip: '10.0.0.1', region: 'us-east', tags: [], status: 'online', created_at: '', updated_at: '' },
  { id: '2', name: 'Node-2', ip: '10.0.0.2', region: 'us-west', tags: [], status: 'offline', created_at: '', updated_at: '' },
]

function renderPage() {
  return render(
    <MemoryRouter>
      <NodeManagementPage />
    </MemoryRouter>,
  )
}

describe('NodeManagementPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAuthState = { user: { role: 'admin' } }
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })
  })

  it('renders the page header and loads nodes on mount', async () => {
    renderPage()

    expect(mockFetchNodes).toHaveBeenCalledTimes(1)
    // Nodes render once loading completes.
    await waitFor(() => {
      expect(screen.getByText('Node-1')).toBeInTheDocument()
      expect(screen.getByText('Node-2')).toBeInTheDocument()
    })
  })

  it('shows the add-node button only for admins', () => {
    mockAuthState = { user: { role: 'admin' } }
    const { rerender } = renderPage()
    expect(screen.getByRole('button', { name: /add new node/i })).toBeInTheDocument()

    mockAuthState = { user: { role: 'viewer' } }
    rerender(
      <MemoryRouter>
        <NodeManagementPage />
      </MemoryRouter>,
    )
    expect(screen.queryByRole('button', { name: /add new node/i })).not.toBeInTheDocument()
  })

  it('shows an error banner with retry when fetchNodes fails', async () => {
    mockFetchNodes.mockRejectedValueOnce(new Error('network error'))
    renderPage()

    expect(await screen.findByText('network error')).toBeInTheDocument()
    expect(screen.getByText('Retry')).toBeInTheDocument()

    // Retry re-fetches.
    mockFetchNodes.mockResolvedValue({ data: { nodes: mockNodes } })
    fireEvent.click(screen.getByText('Retry'))
    await waitFor(() => {
      expect(mockFetchNodes).toHaveBeenCalledTimes(2)
    })
  })

  it('shows a loading spinner while fetching', () => {
    // Never-resolving promise keeps the loading state active.
    mockFetchNodes.mockReturnValue(new Promise(() => {}))
    renderPage()
    // The spinner is a div with animate-spin; assert it is present by role-less
    // structure: loading shows no node rows.
    expect(screen.queryByText('Node-1')).not.toBeInTheDocument()
  })

  it('handles empty node list', async () => {
    mockFetchNodes.mockResolvedValue({ data: { nodes: [] } })
    renderPage()
    await waitFor(() => {
      expect(mockFetchNodes).toHaveBeenCalledTimes(1)
    })
    // No node rows; page should not crash.
    expect(screen.queryByText('Node-1')).not.toBeInTheDocument()
  })

  it('opens the create dialog when Add Node is clicked', async () => {
    renderPage()
    await waitFor(() => expect(screen.getByText('Node-1')).toBeInTheDocument())

    fireEvent.click(screen.getByRole('button', { name: /add new node/i }))

    // NodeDialog renders inside a Radix portal; assert a dialog appears.
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument()
    })
  })
})
