import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import DataExportPage from './DataExportPage'

vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn((selector) => {
    const state = {
      user: { id: '1', username: 'admin', role: 'admin' },
      isAuthenticated: true,
      logout: vi.fn(),
      clearAuth: vi.fn(),
    }
    return selector ? selector(state) : state
  }),
}))

vi.mock('../api/nodes', () => ({
  fetchNodes: vi.fn(() => Promise.resolve({ data: [] })),
}))

describe('DataExportPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders page header with title', () => {
    render(
      <MemoryRouter>
        <DataExportPage />
      </MemoryRouter>
    )

    expect(screen.getByText('Data Export')).toBeInTheDocument()
  })

  it('renders breadcrumb navigation', () => {
    render(
      <MemoryRouter>
        <DataExportPage />
      </MemoryRouter>
    )

    expect(screen.getByRole('navigation', { name: 'Breadcrumb' })).toBeInTheDocument()
  })

  it('renders page title', () => {
    render(
      <MemoryRouter>
        <DataExportPage />
      </MemoryRouter>
    )

    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument()
  })

  it('shows loading state initially', () => {
    render(
      <MemoryRouter>
        <DataExportPage />
      </MemoryRouter>
    )

    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
  })

  it('displays error message when data loading fails', async () => {
    const { fetchNodes } = await import('../api/nodes')
    vi.mocked(fetchNodes).mockRejectedValueOnce(new Error('Failed to load nodes'))

    render(
      <MemoryRouter>
        <DataExportPage />
      </MemoryRouter>
    )

    const errorMessage = await screen.findByText(/Failed to load nodes/)
    expect(errorMessage).toBeInTheDocument()
  })
})
