import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import DataExportPage from './DataExportPage'

vi.mock('../i18n', () => ({
  default: {},
  i18nInitPromise: Promise.resolve(),
}))

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next')

  return {
    ...actual,
    useTranslation: () => ({
      t: (key: string) => {
        const translations: Record<string, string> = {
          'dataExport.title': 'Data Export',
          'dataExport.description': 'Export monitoring data for offline analysis and reporting.',
        }

        return translations[key] || key
      },
      i18n: { changeLanguage: vi.fn() },
    }),
  }
})

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
  fetchNodes: vi.fn(() => Promise.resolve({ data: { nodes: [] } })),
}))

describe('DataExportPage', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    vi.clearAllMocks()
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    consoleErrorSpy.mockRestore()
  })

  it('renders page header with title', async () => {
    render(
      <MemoryRouter>
        <DataExportPage />
      </MemoryRouter>
    )

    expect(await screen.findByText('Data Export')).toBeInTheDocument()
  })

  it('renders page title', async () => {
    render(
      <MemoryRouter>
        <DataExportPage />
      </MemoryRouter>
    )

    expect(await screen.findByRole('heading', { level: 1 })).toBeInTheDocument()
  })

  it('shows loading state initially', async () => {
    const { fetchNodes } = await import('../api/nodes')
    vi.mocked(fetchNodes).mockImplementationOnce(() => new Promise(() => {}))

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
    await waitFor(() => {
      expect(consoleErrorSpy).toHaveBeenCalled()
    })
  })
})
