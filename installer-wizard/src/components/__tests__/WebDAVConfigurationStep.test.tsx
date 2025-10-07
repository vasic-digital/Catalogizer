import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import WebDAVConfigurationStep from '../wizard/WebDAVConfigurationStep'
import { WizardProvider } from '../../contexts/WizardContext'
import { ConfigurationProvider } from '../../contexts/ConfigurationContext'
import { TauriService } from '../../services/tauri'

let queryClient: QueryClient

const TestWrapper = ({ children }: { children: React.ReactNode }) => {
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

describe.skip('WebDAVConfigurationStep', () => {
  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })
    vi.clearAllMocks()
  })

  afterEach(() => {
    queryClient.clear()
    cleanup()
  })

  it('renders WebDAV configuration form', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('WebDAV Configuration')).toBeInTheDocument()
    expect(screen.getByLabelText('Configuration Name')).toBeInTheDocument()
    expect(screen.getByLabelText('WebDAV URL')).toBeInTheDocument()
    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByText('Add Configuration')
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

    fireEvent.change(screen.getByLabelText('WebDAV URL'), { target: { value: 'invalid-url' } })
    fireEvent.change(screen.getByLabelText('Configuration Name'), { target: { value: 'Test' } })
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'user' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'pass' } })

    const submitButton = screen.getByText('Add Configuration')
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

    // Fill in the form
    fireEvent.change(screen.getByLabelText('WebDAV URL'), { target: { value: 'https://webdav.example.com/dav' } })
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(mockTestWebDAVConnection).toHaveBeenCalledWith('https://webdav.example.com/dav', 'testuser', 'testpass', undefined)
      expect(screen.getByText('Connection successful!')).toBeInTheDocument()
    })
  })

  it('adds WebDAV configuration successfully', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Configuration Name'), { target: { value: 'Test WebDAV' } })
    fireEvent.change(screen.getByLabelText('WebDAV URL'), { target: { value: 'https://webdav.example.com/dav' } })
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByText('Add Configuration')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Test WebDAV')).toBeInTheDocument()
      expect(screen.getByText('https://webdav.example.com/dav')).toBeInTheDocument()
    })
  })
})