import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import SummaryStep from '../wizard/SummaryStep'
import { TestWrapper } from '../../test/test-utils'

describe('SummaryStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders setup complete heading', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    expect(screen.getByText('Setup Complete!')).toBeInTheDocument()
    expect(screen.getByText(/Your Catalogizer installation wizard has completed successfully/)).toBeInTheDocument()
  })

  it('displays configuration summary section', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    expect(screen.getByText('Configuration Summary')).toBeInTheDocument()
    expect(screen.getByText('Access Credentials')).toBeInTheDocument()
    expect(screen.getByText('Media Sources')).toBeInTheDocument()
    expect(screen.getByText('SMB Sources')).toBeInTheDocument()
    expect(screen.getByText('Unique Hosts')).toBeInTheDocument()
  })

  it('shows configured sources section', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    expect(screen.getByText('Configured Sources')).toBeInTheDocument()
  })

  it('displays no sources message when no sources configured', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    expect(screen.getByText('No sources configured')).toBeInTheDocument()
    expect(screen.getByText('Consider going back to add some SMB sources')).toBeInTheDocument()
  })

  it('shows next steps section', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    expect(screen.getByText('Next Steps')).toBeInTheDocument()
    expect(screen.getByText('Deploy your configuration')).toBeInTheDocument()
    expect(screen.getByText('Start Catalogizer server')).toBeInTheDocument()
    expect(screen.getByText('Access the web interface')).toBeInTheDocument()
    expect(screen.getByText('Monitor and enjoy')).toBeInTheDocument()
  })

  it('shows important notes section', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    expect(screen.getByText('Important Notes')).toBeInTheDocument()
    expect(screen.getByText(/Ensure your SMB credentials are secure/)).toBeInTheDocument()
    expect(screen.getByText(/Test your configuration in a development environment/)).toBeInTheDocument()
  })

  it('shows action buttons', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    expect(screen.getByText('Start Over')).toBeInTheDocument()
    expect(screen.getByText('Save Configuration Again')).toBeInTheDocument()
  })

  it('shows final success message', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    expect(screen.getByText('Catalogizer Installation Wizard Complete!')).toBeInTheDocument()
    expect(screen.getByText(/Your SMB sources have been configured successfully/)).toBeInTheDocument()
  })

  it('shows zero counts for empty configuration', () => {
    render(
      <TestWrapper>
        <SummaryStep />
      </TestWrapper>
    )

    // With empty configuration, all counts should be 0
    const zeroes = screen.getAllByText('0')
    expect(zeroes.length).toBeGreaterThanOrEqual(4)
  })
})
