import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { TestWrapper } from '../../test/test-utils'

// We must mock react-query's useQuery to control the component state,
// because react-query v4's default behavior with enabled:false starts
// with isLoading:true and refetch() does not settle in jsdom.
const mockRefetch = vi.fn()
let mockQueryReturn = {
  data: undefined as any,
  isLoading: false,
  error: null as any,
  refetch: mockRefetch,
  isError: false,
}

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/react-query')>('@tanstack/react-query')
  return {
    ...actual,
    useQuery: () => mockQueryReturn,
  }
})

// Import after mock setup
import NetworkScanStep from '../wizard/NetworkScanStep'

describe('NetworkScanStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockQueryReturn = {
      data: undefined,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
      isError: false,
    }
    mockRefetch.mockResolvedValue({ data: undefined })
  })

  it('renders network scan heading', () => {
    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('Network Discovery')).toBeInTheDocument()
    expect(screen.getByText(/Scan your local network to discover SMB-enabled devices/)).toBeInTheDocument()
  })

  it('displays scan controls section', () => {
    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('Network Scanning')).toBeInTheDocument()
    expect(screen.getByText(/Click "Start Scan" to discover SMB shares/)).toBeInTheDocument()
  })

  it('has a scan button', () => {
    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    const scanButton = screen.getByRole('button', { name: /Start Scan/ })
    expect(scanButton).toBeInTheDocument()
  })

  it('shows scanning state when isLoading is true', () => {
    mockQueryReturn.isLoading = true

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Scanning network for SMB devices/)).toBeInTheDocument()
    expect(screen.getByText(/This may take a few moments/)).toBeInTheDocument()
  })

  it('renders without crashing', () => {
    const { container } = render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(container).toBeDefined()
    expect(container.querySelector('.space-y-6')).toBeInTheDocument()
  })

  it('displays discovered devices with SMB ports', () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'nas-server',
        open_ports: [445, 139],
        smb_shares: ['shared', 'media'],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Discovered Devices/)).toBeInTheDocument()
    expect(screen.getByText(/1 device\(s\) with SMB shares/)).toBeInTheDocument()
  })

  it('shows hostname and IP for discovered hosts', () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'nas-server',
        open_ports: [445],
        smb_shares: ['shared'],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('nas-server')).toBeInTheDocument()
    expect(screen.getByText(/192\.168\.1\.100/)).toBeInTheDocument()
  })

  it('shows no devices found message when scan returns empty', () => {
    mockQueryReturn.data = []

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('No Devices Found')).toBeInTheDocument()
  })

  it('shows helpful tips when no devices found', () => {
    mockQueryReturn.data = []

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Devices are powered off or not accessible/)).toBeInTheDocument()
    expect(screen.getByText(/Firewall is blocking discovery/)).toBeInTheDocument()
    expect(screen.getByText(/Network configuration prevents scanning/)).toBeInTheDocument()
  })

  it('allows skipping when no devices found', () => {
    mockQueryReturn.data = []

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText(/You can skip this step and configure SMB sources manually/)).toBeInTheDocument()
  })

  it('shows error state when scan fails', () => {
    mockQueryReturn.isError = true
    mockQueryReturn.error = new Error('Network unreachable')

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('Scan Failed')).toBeInTheDocument()
    expect(screen.getByText('Network unreachable')).toBeInTheDocument()
  })

  it('shows select all and deselect all buttons when hosts exist', () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'nas',
        open_ports: [445],
        smb_shares: ['media'],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('Select All')).toBeInTheDocument()
    expect(screen.getByText('Deselect All')).toBeInTheDocument()
  })

  it('shows selected count when hosts are clicked', async () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'nas',
        open_ports: [445],
        smb_shares: ['media'],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('nas'))

    await waitFor(() => {
      expect(screen.getByText(/1 device\(s\) selected/)).toBeInTheDocument()
    })
  })

  it('shows share count per host', () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'nas',
        open_ports: [445],
        smb_shares: ['media', 'backup', 'docs'],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText(/3 share\(s\)/)).toBeInTheDocument()
  })

  it('filters hosts without SMB ports from display', () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'smb-server',
        open_ports: [445],
        smb_shares: ['media'],
      },
      {
        ip: '192.168.1.200',
        hostname: 'web-server',
        open_ports: [80, 443],
        smb_shares: [],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('smb-server')).toBeInTheDocument()
    expect(screen.getByText(/1 device\(s\) with SMB shares/)).toBeInTheDocument()
  })

  it('calls refetch when scan button is clicked', async () => {
    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    const scanButton = screen.getByRole('button', { name: /Start Scan/ })
    fireEvent.click(scanButton)

    await waitFor(() => {
      expect(mockRefetch).toHaveBeenCalled()
    })
  })

  it('shows rescan button when data exists', () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'nas',
        open_ports: [445],
        smb_shares: ['media'],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('Rescan')).toBeInTheDocument()
  })

  it('shows port numbers for SMB hosts', () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'nas',
        open_ports: [445, 139],
        smb_shares: ['media'],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Ports: 139, 445|Ports: 445, 139/)).toBeInTheDocument()
  })

  it('deselects a host when clicked again', async () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.100',
        hostname: 'nas',
        open_ports: [445],
        smb_shares: ['media'],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    // Select
    fireEvent.click(screen.getByText('nas'))

    await waitFor(() => {
      expect(screen.getByText(/1 device\(s\) selected/)).toBeInTheDocument()
    })

    // Deselect
    fireEvent.click(screen.getByText('nas'))

    await waitFor(() => {
      expect(screen.queryByText(/device\(s\) selected/)).toBeNull()
    })
  })

  it('shows no SMB devices message for non-SMB hosts only', () => {
    mockQueryReturn.data = [
      {
        ip: '192.168.1.200',
        hostname: 'web-server',
        open_ports: [80, 443],
        smb_shares: [],
      },
    ]

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('No SMB devices found')).toBeInTheDocument()
    expect(screen.getByText(/Make sure SMB-enabled devices are powered on/)).toBeInTheDocument()
  })

  it('disables scan button during loading', () => {
    mockQueryReturn.isLoading = true

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    const scanButton = screen.getByRole('button', { name: /Scanning/ })
    expect(scanButton).toBeDisabled()
  })

  it('shows error message text from error object', () => {
    mockQueryReturn.isError = true
    mockQueryReturn.error = new Error('Connection timed out')

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('Connection timed out')).toBeInTheDocument()
  })

  it('shows generic error message when error has no message', () => {
    mockQueryReturn.isError = true
    mockQueryReturn.error = null

    render(
      <TestWrapper>
        <NetworkScanStep />
      </TestWrapper>
    )

    expect(screen.getByText('Scan Failed')).toBeInTheDocument()
    expect(screen.getByText(/Failed to scan network/)).toBeInTheDocument()
  })
})
