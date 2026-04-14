import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import SMBConfigurationStep from '../wizard/SMBConfigurationStep'
import { TauriService } from '../../services/tauri'
import { TestWrapper, getInputByLabel } from '../../test/test-utils'

describe('SMBConfigurationStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders SMB configuration form', () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('SMB Configuration')).toBeInTheDocument()
    expect(screen.getByText('Configure SMB connections for your selected devices')).toBeInTheDocument()
    expect(screen.getByText('Configuration Name', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Host/IP Address', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Port', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Share Name', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Username', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Password', { selector: 'label' })).toBeInTheDocument()
  })

  it('renders configuration list panel', () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Configured Sources/)).toBeInTheDocument()
    expect(screen.getByText('No configurations yet')).toBeInTheDocument()
    expect(screen.getByText('Add New')).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Name is required')).toBeInTheDocument()
      expect(screen.getByText('Host is required')).toBeInTheDocument()
      expect(screen.getByText('Share name is required')).toBeInTheDocument()
      expect(screen.getByText('Username is required')).toBeInTheDocument()
      expect(screen.getByText('Password is required')).toBeInTheDocument()
    })
  })

  it('shows test connection button', () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Test Connection')).toBeInTheDocument()
  })

  it('tests SMB connection successfully', async () => {
    const mockTestSMBConnection = vi.spyOn(TauriService, 'testSMBConnection')
      .mockResolvedValue(true)

    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '192.168.1.100' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'shared' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(mockTestSMBConnection).toHaveBeenCalled()
      expect(screen.getByText('Connection successful!')).toBeInTheDocument()
    })
  })

  it('handles SMB connection test failure', async () => {
    const mockTestSMBConnection = vi.spyOn(TauriService, 'testSMBConnection')
      .mockRejectedValue(new Error('SMB connection test failed: Connection refused'))

    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '192.168.1.100' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'shared' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText(/Connection test failed/)).toBeInTheDocument()
    })
  })

  it('requires all fields before testing connection', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    // Click test without filling fields
    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText('Please fill in all required fields before testing')).toBeInTheDocument()
    })
  })

  it('adds SMB configuration successfully', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Media Server' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '192.168.1.100' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'shared' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Media Server')).toBeInTheDocument()
      expect(screen.getByText(/192\.168\.1\.100:445/)).toBeInTheDocument()
    })
  })

  it('shows success count when configurations are added', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Media Server' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '192.168.1.100' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'shared' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('1 SMB source(s) configured')).toBeInTheDocument()
    })
  })

  it('shows optional domain field', () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Domain (optional)', { selector: 'label' })).toBeInTheDocument()
  })

  it('shows optional path field', () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Path (optional)', { selector: 'label' })).toBeInTheDocument()
  })

  it('removes a configuration entry', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'To Remove' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '192.168.1.100' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'shared' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('To Remove')).toBeInTheDocument()
    })

    // Find and click the delete button (Trash2 icon button)
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
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Editable' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '192.168.1.100' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'shared' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Edit')).toBeInTheDocument()
    })
  })

  it('switches to edit mode when Edit is clicked', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test Config' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '10.0.0.1' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'files' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'admin' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'secret' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Test Config')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Edit'))

    await waitFor(() => {
      expect(screen.getByText('Edit Configuration')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Update Configuration' })).toBeInTheDocument()
    })
  })

  it('shows cancel button during editing', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '10.0.0.1' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'share' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'user' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('Edit')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Edit'))

    await waitFor(() => {
      expect(screen.getByText('Cancel')).toBeInTheDocument()
    })
  })

  it('shows username in config entry details', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Server' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '10.0.0.1' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'files' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'admin' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText(/User: admin/)).toBeInTheDocument()
    })
  })

  it('shows next step instruction when configs exist', async () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Server' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: '10.0.0.1' } })
    fireEvent.change(getInputByLabel('Share Name'), { target: { value: 'files' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'admin' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText(/Click "Next" to manage your configuration file/)).toBeInTheDocument()
    })
  })

  it('shows form description text', () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Enter the SMB connection details')).toBeInTheDocument()
  })

  it('shows empty state description in config list', () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Add your first SMB configuration to get started')).toBeInTheDocument()
  })

  it('defaults port to 445', () => {
    render(
      <TestWrapper>
        <SMBConfigurationStep />
      </TestWrapper>
    )

    const portInput = getInputByLabel('Port')
    expect(portInput.value).toBe('445')
  })
})
