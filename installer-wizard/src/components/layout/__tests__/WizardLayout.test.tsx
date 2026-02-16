import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import WizardLayout from '../WizardLayout'
import { WizardProvider } from '../../../contexts/WizardContext'
import { ConfigurationProvider } from '../../../contexts/ConfigurationContext'

// Mock lucide-react
vi.mock('lucide-react', () => ({
  Settings: (props: any) => <span data-testid="icon-settings" {...props} />,
  FileText: (props: any) => <span data-testid="icon-file-text" {...props} />,
  ChevronLeft: (props: any) => <span data-testid="icon-chevron-left" {...props} />,
  ChevronRight: (props: any) => <span data-testid="icon-chevron-right" {...props} />,
}))

const renderWizardLayout = (initialPath = '/') => {
  return render(
    <ConfigurationProvider>
      <WizardProvider>
        <MemoryRouter initialEntries={[initialPath]}>
          <Routes>
            <Route path="/" element={<WizardLayout />}>
              <Route index element={<div data-testid="welcome-step">Welcome Content</div>} />
              <Route path="protocol" element={<div data-testid="protocol-step">Protocol Content</div>} />
              <Route path="summary" element={<div data-testid="summary-step">Summary Content</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      </WizardProvider>
    </ConfigurationProvider>
  )
}

describe('WizardLayout', () => {
  it('renders the wizard header', () => {
    renderWizardLayout()

    expect(screen.getByText('Catalogizer Installation Wizard')).toBeInTheDocument()
  })

  it('renders the wizard subtitle', () => {
    renderWizardLayout()

    expect(screen.getByText('Configure storage sources for your media collection')).toBeInTheDocument()
  })

  it('renders Previous and Next navigation buttons', () => {
    renderWizardLayout()

    expect(screen.getByText('Previous')).toBeInTheDocument()
    expect(screen.getByText('Next')).toBeInTheDocument()
  })

  it('disables Previous button on first step', () => {
    renderWizardLayout('/')

    const previousButton = screen.getByText('Previous').closest('button')
    expect(previousButton).toBeDisabled()
  })

  it('renders the step counter', () => {
    renderWizardLayout()

    expect(screen.getByText(/Step 1 of/)).toBeInTheDocument()
  })

  it('renders the Welcome step title', () => {
    renderWizardLayout('/')

    // "Welcome" appears in the progress bar and in the card header
    const welcomeElements = screen.getAllByText('Welcome')
    expect(welcomeElements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders the step description', () => {
    renderWizardLayout('/')

    expect(screen.getByText('Introduction to the setup wizard')).toBeInTheDocument()
  })

  it('renders the Outlet content for the welcome route', () => {
    renderWizardLayout('/')

    expect(screen.getByTestId('welcome-step')).toBeInTheDocument()
  })

  it('renders the settings icon in the header', () => {
    renderWizardLayout()

    expect(screen.getByTestId('icon-settings')).toBeInTheDocument()
  })

  it('renders the file-text icon in the card header', () => {
    renderWizardLayout()

    expect(screen.getByTestId('icon-file-text')).toBeInTheDocument()
  })
})
