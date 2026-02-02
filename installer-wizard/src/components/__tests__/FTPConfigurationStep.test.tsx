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
    expect(true).toBe(true)
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
    global.mockInvoke.mockResolvedValue(true)

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
      expect(global.mockInvoke).toHaveBeenCalledWith('test_ftp_connection', {
        host: 'ftp.example.com',
        port: 21,
        username: 'testuser',
        password: 'testpass',
        path: undefined,
      })
      expect(screen.getByText('Connection successful!')).toBeInTheDocument()
    })
  })

  it('handles FTP connection test failure', async () => {
    global.mockInvoke.mockRejectedValue(new Error('Connection failed'))

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
})
