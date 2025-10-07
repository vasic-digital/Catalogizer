import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import WelcomeStep from '../wizard/WelcomeStep'
import { WizardProvider } from '../../contexts/WizardContext'
import { ConfigurationProvider } from '../../contexts/ConfigurationContext'

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

describe.skip('WelcomeStep', () => {
  it('renders welcome message', () => {
    render(
      <TestWrapper>
        <WelcomeStep />
      </TestWrapper>
    )

    expect(screen.getByText('Welcome to Catalogizer Installation Wizard')).toBeInTheDocument()
    expect(screen.getByText(/This wizard will help you configure storage sources/)).toBeInTheDocument()
  })

  it('displays feature cards', () => {
    render(
      <TestWrapper>
        <WelcomeStep />
      </TestWrapper>
    )

    expect(screen.getByText('Protocol Selection')).toBeInTheDocument()
    expect(screen.getByText('Source Configuration')).toBeInTheDocument()
    expect(screen.getByText('Configuration')).toBeInTheDocument()
  })

  it('shows requirements section', () => {
    render(
      <TestWrapper>
        <WelcomeStep />
      </TestWrapper>
    )

    expect(screen.getByText("What you'll need:")).toBeInTheDocument()
    expect(screen.getByText(/Access to your storage system/)).toBeInTheDocument()
    expect(screen.getByText(/Valid credentials for the storage system/)).toBeInTheDocument()
    expect(screen.getByText(/A location to save your configuration file/)).toBeInTheDocument()
  })
})