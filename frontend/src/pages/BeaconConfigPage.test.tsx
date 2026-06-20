import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { vi } from 'vitest'
import BeaconConfigPage from './BeaconConfigPage'
import { fetchNodes } from '../api/nodes'
import { fetchBeaconConfig, fetchConfigHistory } from '../api/beaconConfig'

const translations: Record<string, string> = {
  'beaconConfig.title': 'Beacon Configuration',
  'beaconConfig.subtitle': 'Manage probe configurations for beacon agents',
  'beaconConfig.selectNode': 'Select Node',
  'beaconConfig.globalSettings': 'Global Settings',
  'beaconConfig.intervalSeconds': 'Default Interval (seconds)',
  'beaconConfig.timeoutSeconds': 'Default Timeout (seconds)',
  'beaconConfig.version': 'Version',
  'beaconConfig.updated': 'Updated',
  'beaconConfig.probes': 'Probes',
  'beaconConfig.probe': 'Probe',
  'beaconConfig.probeType': 'Type',
  'beaconConfig.probeTarget': 'Target',
  'beaconConfig.probeTargetPlaceholder': 'IP or domain',
  'beaconConfig.probePort': 'Port',
  'beaconConfig.probeIntervalSeconds': 'Interval (s)',
  'beaconConfig.probeTimeoutSeconds': 'Timeout (s)',
  'beaconConfig.probeCount': 'Count',
  'beaconConfig.showHistory': 'History',
  'beaconConfig.hideHistory': 'Hide History',
  'beaconConfig.configHistory': 'Configuration History',
  'beaconConfig.applyStatus': 'Apply Status',
  'beaconConfig.lastAckAt': 'Last Ack',
  'beaconConfig.lastAckVersion': 'Ack Version',
  'beaconConfig.ackStatus.applied': 'Applied',
  'beaconConfig.ackStatus.failed': 'Failed',
  'beaconConfig.ackStatus.pending': 'Pending',
  'beaconConfig.templates': 'Config Templates',
  'beaconConfig.saveAsTemplate': 'Save as Template',
  'beaconConfig.noTemplates': 'No templates saved.',
  'common.save': 'Save',
  'common.saving': 'Saving...',
  'common.none': 'None',
}

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, values?: Record<string, unknown>) => {
      if (key === 'beaconConfig.applyStatusDetail') {
        return `Current v${values?.current}, last acknowledged v${values?.acked}`
      }
      return translations[key] ?? key
    },
  }),
}))

vi.mock('../api/nodes', () => ({
  fetchNodes: vi.fn(),
}))

vi.mock('../api/beaconConfig', () => ({
  fetchBeaconConfig: vi.fn(),
  fetchConfigHistory: vi.fn(),
  updateBeaconConfig: vi.fn(),
}))

vi.mock('../stores/settingsStore', () => ({
  useSettingsStore: () => ({
    configTemplates: [],
    addConfigTemplate: vi.fn(),
    deleteConfigTemplate: vi.fn(),
  }),
}))

const mockFetchNodes = fetchNodes as ReturnType<typeof vi.fn>
const mockFetchBeaconConfig = fetchBeaconConfig as ReturnType<typeof vi.fn>
const mockFetchConfigHistory = fetchConfigHistory as ReturnType<typeof vi.fn>

describe('BeaconConfigPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetchNodes.mockResolvedValue({
      data: {
        nodes: [
          { id: 'node-1', name: 'Primary Node', region: 'us-east', ip: '192.0.2.10', tags: [], status: 'online' },
        ],
      },
    })
    mockFetchBeaconConfig.mockResolvedValue({
      data: {
        probes: [],
        interval_seconds: 60,
        timeout_seconds: 5,
        updated_at: '2024-01-02T12:00:00Z',
        version: 3,
        last_ack_version: 3,
        last_ack_at: '2024-01-02T12:01:00Z',
        last_ack_status: 'applied',
      },
      message: 'ok',
      timestamp: '2024-01-02T12:02:00Z',
    })
    mockFetchConfigHistory.mockResolvedValue({
      data: [
        {
          version: 2,
          changed_at: '2024-01-01T12:00:00Z',
          changed_by: 'system',
          config: {
            probes: [{ id: 'probe-1', type: 'TCP', target: 'example.com', port: 443, interval_seconds: 60, timeout_seconds: 5, count: 3 }],
            interval_seconds: 60,
            timeout_seconds: 5,
            updated_at: '2024-01-01T12:00:00Z',
            version: 2,
            last_ack_version: 2,
            last_ack_at: '2024-01-01T12:01:00Z',
            last_ack_status: 'failed',
            last_ack_error: 'failed to parse config',
          },
        },
      ],
      message: 'ok',
      timestamp: '2024-01-02T12:02:00Z',
    })
  })

  it('renders current config acknowledgement and historical ack status', async () => {
    render(<BeaconConfigPage />)

    await waitFor(() => {
      expect(mockFetchBeaconConfig).toHaveBeenCalledWith('node-1')
    })

    expect(screen.getByText('Apply Status')).toBeInTheDocument()
    expect(screen.getByText('Applied')).toBeInTheDocument()
    expect(screen.getByText('Current v3, last acknowledged v3')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'History' }))

    expect(screen.getByText('Configuration History')).toBeInTheDocument()
    expect(screen.getByText('v2')).toBeInTheDocument()
    expect(screen.getByText('Failed')).toBeInTheDocument()
    expect(screen.getByText('failed to parse config')).toBeInTheDocument()
  })

  it('renders localized probe field labels', async () => {
    mockFetchBeaconConfig.mockResolvedValueOnce({
      data: {
        probes: [{ id: 'probe-1', type: 'TCP', target: 'example.com', port: 443, interval_seconds: 60, timeout_seconds: 5, count: 3 }],
        interval_seconds: 60,
        timeout_seconds: 5,
        updated_at: '2024-01-02T12:00:00Z',
        version: 3,
        last_ack_version: 3,
        last_ack_at: '2024-01-02T12:01:00Z',
        last_ack_status: 'applied',
      },
      message: 'ok',
      timestamp: '2024-01-02T12:02:00Z',
    })

    render(<BeaconConfigPage />)

    await waitFor(() => {
      expect(screen.getByText('Probe #1')).toBeInTheDocument()
    })

    expect(screen.getByText('Type')).toBeInTheDocument()
    expect(screen.getByText('Target')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('IP or domain')).toBeInTheDocument()
    expect(screen.getByText('Port')).toBeInTheDocument()
    expect(screen.getByText('Interval (s)')).toBeInTheDocument()
    expect(screen.getByText('Timeout (s)')).toBeInTheDocument()
    expect(screen.getByText('Count')).toBeInTheDocument()
  })
})
