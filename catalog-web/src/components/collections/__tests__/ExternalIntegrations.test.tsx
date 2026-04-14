import React from 'react'
import { render, screen } from '@testing-library/react'

// Mock framer-motion
vi.mock('framer-motion', () => {
  const MockDiv = ({ children, ...props }: any) => <div {...props}>{children}</div>
  MockDiv.displayName = 'MockDiv'
  const MockButton = ({ children, ...props }: any) => <button {...props}>{children}</button>
  MockButton.displayName = 'MockButton'
  return {
    motion: { div: MockDiv, button: MockButton },
    AnimatePresence: ({ children }: any) => <>{children}</>,
  }
})

// Mock react-hot-toast
vi.mock('react-hot-toast', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Mock lucide-react icons
vi.mock('lucide-react', () => {
  const icon = (name: string) => {
    const Comp = (props: any) => <svg data-testid={`icon-${name}`} {...props} />
    Comp.displayName = name
    return Comp
  }
  return {
    Globe: icon('Globe'), Cloud: icon('Cloud'), Plus: icon('Plus'),
    Trash2: icon('Trash2'), Edit: icon('Edit'), CheckCircle: icon('CheckCircle'),
    AlertCircle: icon('AlertCircle'), XCircle: icon('XCircle'), Clock: icon('Clock'),
    RefreshCw: icon('RefreshCw'), Link: icon('Link'), Key: icon('Key'),
    Zap: icon('Zap'), Database: icon('Database'), FolderSync: icon('FolderSync'),
    Share2: icon('Share2'), Info: icon('Info'), ExternalLink: icon('ExternalLink'),
    TestTube: icon('TestTube'), Activity: icon('Activity'),
  }
})

// Mock fetch for integration loading
global.fetch = vi.fn().mockResolvedValue({
  ok: false,
  json: async () => ([]),
})

import ExternalIntegrations from '../ExternalIntegrations'

describe('ExternalIntegrations', () => {
  it('renders the heading', () => {
    render(<ExternalIntegrations />)
    expect(screen.getByText('External Integrations')).toBeInTheDocument()
  })

  it('renders the subtitle', () => {
    render(<ExternalIntegrations />)
    expect(screen.getByText('Connect with external services to extend functionality')).toBeInTheDocument()
  })

  it('shows Add Integration button', () => {
    render(<ExternalIntegrations />)
    expect(screen.getByText('Add Integration')).toBeInTheDocument()
  })

  it('renders stats cards', () => {
    render(<ExternalIntegrations />)
    expect(screen.getByText('Total Integrations')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<ExternalIntegrations />)
    expect(screen.getByPlaceholderText('Search integrations...')).toBeInTheDocument()
  })

  it('renders filter buttons', () => {
    render(<ExternalIntegrations />)
    expect(screen.getAllByText('All').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Connected').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Disconnected').length).toBeGreaterThan(0)
  })

  it('shows empty state when no integrations', async () => {
    render(<ExternalIntegrations />)
    expect(await screen.findByText('No integrations found')).toBeInTheDocument()
  })

  it('renders without crashing', () => {
    const { container } = render(<ExternalIntegrations />)
    expect(container.firstChild).toBeTruthy()
  })
})
