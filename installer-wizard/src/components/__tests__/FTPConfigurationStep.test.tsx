import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import FTPConfigurationStep from '../wizard/FTPConfigurationStep'
import { TauriService } from '../../services/tauri'
import { TestWrapper, getInputByLabel } from '../../test/test-utils'

describe('FTPConfigurationStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders without crashing', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )
    // Article XI: assert the component actually mounted with its identifying
    // heading rendered (not just "no exception").
    expect(screen.getByText('FTP Configuration')).toBeInTheDocument()
  })

  it('renders FTP configuration form', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('FTP Configuration')).toBeInTheDocument()
    expect(screen.getByText('Configuration Name', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Host/IP Address', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Port', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Username', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Password', { selector: 'label' })).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Name is required')).toBeInTheDocument()
      expect(screen.getByText('Host is required')).toBeInTheDocument()
      expect(screen.getByText('Username is required')).toBeInTheDocument()
      expect(screen.getByText('Password is required')).toBeInTheDocument()
    })
  })

  it('tests FTP connection successfully', async () => {
    vi.spyOn(TauriService, 'testFTPConnection').mockResolvedValue(true)

    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.example.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText('Connection successful!')).toBeInTheDocument()
    })
  })

  it('handles FTP connection test failure', async () => {
    vi.spyOn(TauriService, 'testFTPConnection').mockRejectedValue(new Error('Connection failed'))

    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.example.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText(/Connection test failed/)).toBeInTheDocument()
    })
  })

  it('adds FTP configuration successfully', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test FTP' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.example.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Test FTP')).toBeInTheDocument()
      expect(screen.getByText('ftp.example.com:21')).toBeInTheDocument()
    })
  })

  it('shows success message when configurations are added', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test FTP' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.example.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('1 FTP source(s) configured')).toBeInTheDocument()
    })
  })

  it('requires fields before testing connection', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText('Please fill in all required fields before testing')).toBeInTheDocument()
    })
  })

  it('shows optional path field', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Path (optional)', { selector: 'label' })).toBeInTheDocument()
  })

  it('defaults port to 21', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    const portInput = getInputByLabel('Port')
    expect(portInput.value).toBe('21')
  })

  it('shows empty state in configuration list', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('No configurations yet')).toBeInTheDocument()
    expect(screen.getByText('Add your first FTP configuration to get started')).toBeInTheDocument()
  })

  it('shows form description text', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Enter the FTP connection details')).toBeInTheDocument()
  })

  it('shows subtitle text', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Configure FTP connections for your selected devices')).toBeInTheDocument()
  })

  it('shows Add New button', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Add New')).toBeInTheDocument()
  })

  it('removes a configuration entry', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Removable' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.test.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'user' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('Removable')).toBeInTheDocument()
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
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Editable' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.test.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'user' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('Edit')).toBeInTheDocument()
    })
  })

  it('switches to edit mode when Edit is clicked', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'FTP Server' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.test.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'admin' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'secret' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('FTP Server')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Edit'))

    await waitFor(() => {
      expect(screen.getByText('Edit Configuration')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Update Configuration' })).toBeInTheDocument()
      expect(screen.getByText('Cancel')).toBeInTheDocument()
    })
  })

  it('shows username in config entry details', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Server' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.test.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'ftpuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText(/User: ftpuser/)).toBeInTheDocument()
    })
  })

  it('shows next step instruction when configs exist', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Server' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'ftp.test.com' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'user' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText(/Click "Next" to manage your configuration file/)).toBeInTheDocument()
    })
  })

  it('shows configured sources count in list header', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Configured Sources \(0\)/)).toBeInTheDocument()
  })

  it('manages FTP source configurations text in list', () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Manage your FTP source configurations')).toBeInTheDocument()
  })
})
