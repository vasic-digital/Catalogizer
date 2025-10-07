import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import FTPConfigurationStep from '../wizard/FTPConfigurationStep'
import { WizardProvider } from '../../contexts/WizardContext'
import { ConfigurationProvider } from '../../contexts/ConfigurationContext'
import * as TauriService from '../../services/tauri'

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ConfigurationProvider>
          <WizardProvider>
            {children}
          </WizardProvider>
        </ConfigurationProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

describe('FTPConfigurationStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
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
    const mockTestFTPConnection = vi.spyOn(TauriService.TauriService, 'testFTPConnection')
      .mockResolvedValue(true)

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
      expect(mockTestFTPConnection).toHaveBeenCalledWith('ftp.example.com', 21, 'testuser', 'testpass', undefined)
      expect(screen.getByText('Connection successful!')).toBeInTheDocument()
    })
  })

  it('handles FTP connection test failure', async () => {
    const mockTestFTPConnection = vi.spyOn(TauriService.TauriService, 'testFTPConnection')
      .mockRejectedValue(new Error('Connection failed'))

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