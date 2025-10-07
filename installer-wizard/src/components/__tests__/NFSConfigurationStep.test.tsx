import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import NFSConfigurationStep from '../wizard/NFSConfigurationStep'
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
    expect(screen.getByLabelText('Configuration Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Host/IP Address')).toBeInTheDocument()
    expect(screen.getByLabelText('Export Path')).toBeInTheDocument()
    expect(screen.getByLabelText('Mount Point')).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByText('Add Configuration')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Name is required')).toBeInTheDocument()
      expect(screen.getByText('Host is required')).toBeInTheDocument()
      expect(screen.getByText('Path is required')).toBeInTheDocument()
      expect(screen.getByText('Mount point is required')).toBeInTheDocument()
    })
  })

  it('tests NFS connection successfully', async () => {
    const mockTestNFSConnection = vi.spyOn(TauriService.TauriService, 'testNFSConnection')
      .mockResolvedValue(true)

    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Host/IP Address'), { target: { value: 'nfs.example.com' } })
    fireEvent.change(screen.getByLabelText('Export Path'), { target: { value: '/export/data' } })
    fireEvent.change(screen.getByLabelText('Mount Point'), { target: { value: '/mnt/nfs' } })

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

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Configuration Name'), { target: { value: 'Test NFS' } })
    fireEvent.change(screen.getByLabelText('Host/IP Address'), { target: { value: 'nfs.example.com' } })
    fireEvent.change(screen.getByLabelText('Export Path'), { target: { value: '/export/data' } })
    fireEvent.change(screen.getByLabelText('Mount Point'), { target: { value: '/mnt/nfs' } })

    const submitButton = screen.getByText('Add Configuration')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Test NFS')).toBeInTheDocument()
      expect(screen.getByText('nfs.example.com:/export/data → /mnt/nfs')).toBeInTheDocument()
    })
  })
})