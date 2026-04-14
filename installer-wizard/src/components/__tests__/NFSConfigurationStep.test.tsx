import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import NFSConfigurationStep from '../wizard/NFSConfigurationStep'
import { TauriService } from '../../services/tauri'
import { TestWrapper, getInputByLabel } from '../../test/test-utils'

describe('NFSConfigurationStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders NFS configuration form', () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('NFS Configuration')).toBeInTheDocument()
    expect(screen.getByText('Configuration Name', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Host/IP Address', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Export Path', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Mount Point', { selector: 'label' })).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Name is required')).toBeInTheDocument()
      expect(screen.getByText('Host is required')).toBeInTheDocument()
      expect(screen.getByText('Path is required')).toBeInTheDocument()
      expect(screen.getByText('Mount point is required')).toBeInTheDocument()
    })
  })

  it('tests NFS connection successfully', async () => {
    const mockTestNFSConnection = vi.spyOn(TauriService, 'testNFSConnection')
      .mockResolvedValue(true)

    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.example.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/export/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/nfs' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(mockTestNFSConnection).toHaveBeenCalledWith('nfs.example.com', '/export/data', '/mnt/nfs', 'vers=3')
      expect(screen.getByText('Connection successful!')).toBeInTheDocument()
    })
  })

  it('adds NFS configuration successfully', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test NFS' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.example.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/export/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/nfs' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Test NFS')).toBeInTheDocument()
      expect(screen.getByText(/nfs\.example\.com/)).toBeInTheDocument()
    })
  })

  it('handles NFS connection test failure', async () => {
    vi.spyOn(TauriService, 'testNFSConnection')
      .mockRejectedValue(new Error('Mount failed'))

    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.example.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/export/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/nfs' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText(/Connection test failed/)).toBeInTheDocument()
    })
  })

  it('requires fields before testing connection', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText('Please fill in all required fields before testing')).toBeInTheDocument()
    })
  })

  it('shows optional mount options field', () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Mount Options (optional)', { selector: 'label' })).toBeInTheDocument()
  })

  it('shows success count when configurations are added', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'NFS Share' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.example.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/export/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/nfs' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('1 NFS source(s) configured')).toBeInTheDocument()
    })
  })

  it('shows empty state in configuration list', () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('No configurations yet')).toBeInTheDocument()
    expect(screen.getByText('Add your first NFS configuration to get started')).toBeInTheDocument()
  })

  it('shows form description text', () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Enter the NFS connection details')).toBeInTheDocument()
  })

  it('shows subtitle text', () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Configure NFS connections for your selected devices')).toBeInTheDocument()
  })

  it('removes a configuration entry', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Removable NFS' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.test.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/data' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('Removable NFS')).toBeInTheDocument()
    })

    const deleteButtons = screen.getAllByRole('button').filter(btn =>
      btn.classList.contains('text-red-600') || btn.className.includes('text-red')
    )
    expect(deleteButtons.length).toBeGreaterThan(0)
    fireEvent.click(deleteButtons[0])

    await waitFor(() => {
      expect(screen.getByText('No configurations yet')).toBeInTheDocument()
    })
  })

  it('shows edit button for existing configurations', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Editable' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.test.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/data' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('Edit')).toBeInTheDocument()
    })
  })

  it('switches to edit mode when Edit is clicked', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'NFS Share' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.test.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/data' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('NFS Share')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Edit'))

    await waitFor(() => {
      expect(screen.getByText('Edit Configuration')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Update Configuration' })).toBeInTheDocument()
      expect(screen.getByText('Cancel')).toBeInTheDocument()
    })
  })

  it('shows mount point in config entry details', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Share' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.test.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/export' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/share' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText(/\/mnt\/share/)).toBeInTheDocument()
    })
  })

  it('shows next step instruction when configs exist', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Share' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.test.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/export' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/share' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText(/Click "Next" to manage your configuration file/)).toBeInTheDocument()
    })
  })

  it('shows Add New button', () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Add New')).toBeInTheDocument()
  })

  it('shows configured sources count in list header', () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Configured Sources \(0\)/)).toBeInTheDocument()
  })

  it('manages NFS source configurations text in list', () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Manage your NFS source configurations')).toBeInTheDocument()
  })
})
