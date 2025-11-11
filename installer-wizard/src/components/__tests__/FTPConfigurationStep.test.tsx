import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import FTPConfigurationStep from '../wizard/FTPConfigurationStep'
import { WizardProvider } from '../../contexts/WizardContext'
import { ConfigurationProvider } from '../../contexts/ConfigurationContext'
import { TauriService } from '../../services/tauri'

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  return (
    <BrowserRouter>
      <ConfigurationProvider>
        <WizardProvider>
          {children}
        </WizardProvider>
      </ConfigurationProvider>
    </BrowserRouter>
  )
}

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
    expect(screen.getByText('Configure FTP connections for your selected devices')).toBeInTheDocument()
    expect(screen.getByLabelText('Configuration Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Host/IP Address')).toBeInTheDocument()
    expect(screen.getByLabelText('Port')).toBeInTheDocument()
    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByText('Add Configuration')
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

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Host/IP Address'), { target: { value: 'ftp.example.com' } })
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'testpass' } })

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

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Host/IP Address'), { target: { value: 'ftp.example.com' } })
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText('Connection test failed: Connection failed')).toBeInTheDocument()
    })
  })

  it('adds FTP configuration successfully', async () => {
    render(
      <TestWrapper>
        <FTPConfigurationStep />
      </TestWrapper>
    )

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Configuration Name'), { target: { value: 'Test FTP' } })
    fireEvent.change(screen.getByLabelText('Host/IP Address'), { target: { value: 'ftp.example.com' } })
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByText('Add Configuration')
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

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Configuration Name'), { target: { value: 'Test FTP' } })
    fireEvent.change(screen.getByLabelText('Host/IP Address'), { target: { value: 'ftp.example.com' } })
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByText('Add Configuration')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('1 FTP source(s) configured')).toBeInTheDocument()
    })
  })
})