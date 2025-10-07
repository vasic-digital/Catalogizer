import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import LocalConfigurationStep from '../wizard/LocalConfigurationStep'
import { WizardProvider } from '../../contexts/WizardContext'
import { ConfigurationProvider } from '../../contexts/ConfigurationContext'
import { TauriService } from '../../services/tauri'

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

describe.skip('LocalConfigurationStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders local configuration form', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Local Configuration')).toBeInTheDocument()
    expect(screen.getByLabelText('Configuration Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Base Path')).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByText('Add Configuration')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Name is required')).toBeInTheDocument()
      expect(screen.getByText('Base path is required')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('tests local path successfully', async () => {
    const mockTestLocalConnection = vi.spyOn(TauriService, 'testLocalConnection')
      .mockResolvedValue(true)

    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Base Path'), { target: { value: '/home/user/media' } })

    const testButton = screen.getByText('Test Path')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(mockTestLocalConnection).toHaveBeenCalledWith('/home/user/media')
      expect(screen.getByText('Path accessible!')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('adds local configuration successfully', async () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // Fill in the form
    fireEvent.change(screen.getByLabelText('Configuration Name'), { target: { value: 'Local Media' } })
    fireEvent.change(screen.getByLabelText('Base Path'), { target: { value: '/home/user/media' } })

    const submitButton = screen.getByText('Add Configuration')
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Local Media')).toBeInTheDocument()
      expect(screen.getByText('/home/user/media')).toBeInTheDocument()
    }, { timeout: 3000 })
  })
})