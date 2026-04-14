import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import WebDAVConfigurationStep from '../wizard/WebDAVConfigurationStep'
import { TauriService } from '../../services/tauri'
import { TestWrapper, getInputByLabel } from '../../test/test-utils'

describe('WebDAVConfigurationStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders WebDAV configuration form', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('WebDAV Configuration')).toBeInTheDocument()
    expect(screen.getByText('Configuration Name', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('WebDAV URL', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Username', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Password', { selector: 'label' })).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Name is required')).toBeInTheDocument()
      expect(screen.getByText('URL is required')).toBeInTheDocument()
      expect(screen.getByText('Username is required')).toBeInTheDocument()
      expect(screen.getByText('Password is required')).toBeInTheDocument()
    })
  })

  it('validates URL format', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'invalid-url' } })
    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'user' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Invalid URL format')).toBeInTheDocument()
    })
  })

  it('tests WebDAV connection successfully', async () => {
    const mockTestWebDAVConnection = vi.spyOn(TauriService, 'testWebDAVConnection')
      .mockResolvedValue(true)

    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://webdav.example.com/dav' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(mockTestWebDAVConnection).toHaveBeenCalled()
      expect(screen.getByText('Connection successful!')).toBeInTheDocument()
    })
  })

  it('adds WebDAV configuration successfully', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test WebDAV' } })
    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://webdav.example.com/dav' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Test WebDAV')).toBeInTheDocument()
      expect(screen.getByText('https://webdav.example.com/dav')).toBeInTheDocument()
    })
  })

  it('handles WebDAV connection test failure', async () => {
    vi.spyOn(TauriService, 'testWebDAVConnection')
      .mockRejectedValue(new Error('SSL handshake failed'))

    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://webdav.example.com/dav' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText(/Connection test failed/)).toBeInTheDocument()
    })
  })

  it('requires fields before testing connection', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText('Please fill in all required fields before testing')).toBeInTheDocument()
    })
  })

  it('shows success count when configurations are added', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Nextcloud' } })
    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://cloud.example.com/dav' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'user' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('1 WebDAV source(s) configured')).toBeInTheDocument()
    })
  })

  it('shows optional path field', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Path (optional)', { selector: 'label' })).toBeInTheDocument()
  })

  it('shows empty state in configuration list', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('No configurations yet')).toBeInTheDocument()
    expect(screen.getByText('Add your first WebDAV configuration to get started')).toBeInTheDocument()
  })

  it('shows form description text', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Enter the WebDAV connection details')).toBeInTheDocument()
  })

  it('shows subtitle text', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Configure WebDAV connections for your selected devices')).toBeInTheDocument()
  })

  it('removes a configuration entry', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Removable' } })
    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://dav.test.com/files' } })
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
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Editable' } })
    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://dav.test.com/files' } })
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
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Cloud DAV' } })
    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://dav.test.com/files' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'admin' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'secret' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText('Cloud DAV')).toBeInTheDocument()
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
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Server' } })
    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://dav.test.com/files' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'davuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText(/User: davuser/)).toBeInTheDocument()
    })
  })

  it('shows next step instruction when configs exist', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Server' } })
    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://dav.test.com/files' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'user' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    fireEvent.click(screen.getByRole('button', { name: 'Add Configuration' }))

    await waitFor(() => {
      expect(screen.getByText(/Click "Next" to manage your configuration file/)).toBeInTheDocument()
    })
  })

  it('shows Add New button', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Add New')).toBeInTheDocument()
  })

  it('shows configured sources count in list header', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Configured Sources \(0\)/)).toBeInTheDocument()
  })

  it('manages WebDAV source configurations text in list', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Manage your WebDAV source configurations')).toBeInTheDocument()
  })
})
