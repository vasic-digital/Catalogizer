import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from '../App'

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  Settings: (props: any) => <span data-testid="icon-settings" {...props} />,
  FileText: (props: any) => <span data-testid="icon-file-text" {...props} />,
  ChevronLeft: (props: any) => <span data-testid="icon-chevron-left" {...props} />,
  ChevronRight: (props: any) => <span data-testid="icon-chevron-right" {...props} />,
  Folder: (props: any) => <span data-testid="icon-folder" {...props} />,
  FolderOpen: (props: any) => <span data-testid="icon-folder-open" {...props} />,
  CheckCircle: (props: any) => <span data-testid="icon-check-circle" {...props} />,
  CheckCircle2: (props: any) => <span data-testid="icon-check-circle-2" {...props} />,
  AlertCircle: (props: any) => <span data-testid="icon-alert-circle" {...props} />,
  AlertTriangle: (props: any) => <span data-testid="icon-alert-triangle" {...props} />,
  Eye: (props: any) => <span data-testid="icon-eye" {...props} />,
  EyeOff: (props: any) => <span data-testid="icon-eye-off" {...props} />,
  TestTube: (props: any) => <span data-testid="icon-test-tube" {...props} />,
  Loader2: (props: any) => <span data-testid="icon-loader" {...props} />,
  Plus: (props: any) => <span data-testid="icon-plus" {...props} />,
  Trash2: (props: any) => <span data-testid="icon-trash" {...props} />,
  Server: (props: any) => <span data-testid="icon-server" {...props} />,
  HardDrive: (props: any) => <span data-testid="icon-hard-drive" {...props} />,
  Globe: (props: any) => <span data-testid="icon-globe" {...props} />,
  Search: (props: any) => <span data-testid="icon-search" {...props} />,
  Wifi: (props: any) => <span data-testid="icon-wifi" {...props} />,
  RefreshCw: (props: any) => <span data-testid="icon-refresh" {...props} />,
  Monitor: (props: any) => <span data-testid="icon-monitor" {...props} />,
  Network: (props: any) => <span data-testid="icon-network" {...props} />,
  Download: (props: any) => <span data-testid="icon-download" {...props} />,
  Upload: (props: any) => <span data-testid="icon-upload" {...props} />,
  ExternalLink: (props: any) => <span data-testid="icon-external-link" {...props} />,
  Save: (props: any) => <span data-testid="icon-save" {...props} />,
  Edit3: (props: any) => <span data-testid="icon-edit" {...props} />,
}))

describe('App', () => {
  it('renders the wizard header title', () => {
    render(<App />)

    expect(screen.getByText('Catalogizer Installation Wizard')).toBeInTheDocument()
  })

  it('renders the wizard header subtitle', () => {
    render(<App />)

    expect(screen.getByText('Configure storage sources for your media collection')).toBeInTheDocument()
  })

  it('renders the Welcome step title in the main content area', () => {
    render(<App />)

    // "Welcome" appears in both the progress bar and card header
    const welcomeElements = screen.getAllByText('Welcome')
    expect(welcomeElements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders navigation buttons (Previous and Next)', () => {
    render(<App />)

    expect(screen.getByText('Previous')).toBeInTheDocument()
    expect(screen.getByText('Next')).toBeInTheDocument()
  })

  it('disables Previous button on the first step', () => {
    render(<App />)

    const previousButton = screen.getByText('Previous').closest('button')
    expect(previousButton).toBeDisabled()
  })

  it('renders step progress indicator', () => {
    render(<App />)

    // Step 1 of N should be visible
    expect(screen.getByText(/Step 1 of/)).toBeInTheDocument()
  })
})
