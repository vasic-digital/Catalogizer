import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ProtocolSelectionStep from '../wizard/ProtocolSelectionStep'
import { TestWrapper } from '../../test/test-utils'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('ProtocolSelectionStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the protocol selection heading', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    expect(screen.getByText('Select Storage Protocol')).toBeInTheDocument()
    expect(screen.getByText(/Choose the protocol for your media storage/)).toBeInTheDocument()
  })

  it('displays all protocol options', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    expect(screen.getByText('SMB/CIFS')).toBeInTheDocument()
    expect(screen.getByText('FTP')).toBeInTheDocument()
    expect(screen.getByText('NFS')).toBeInTheDocument()
    expect(screen.getByText('WebDAV')).toBeInTheDocument()
    expect(screen.getByText('Local Files')).toBeInTheDocument()
  })

  it('displays protocol descriptions', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    expect(screen.getByText('Windows file sharing protocol for network drives')).toBeInTheDocument()
    expect(screen.getByText('File Transfer Protocol for remote file access')).toBeInTheDocument()
    expect(screen.getByText('Network File System for Unix/Linux file sharing')).toBeInTheDocument()
    expect(screen.getByText('Web-based Distributed Authoring and Versioning')).toBeInTheDocument()
    expect(screen.getByText('Direct access to local filesystem paths')).toBeInTheDocument()
  })

  it('displays protocol features', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    // SMB features
    expect(screen.getByText('Network discovery')).toBeInTheDocument()
    expect(screen.getByText('Share browsing')).toBeInTheDocument()
    expect(screen.getByText('Domain support')).toBeInTheDocument()

    // FTP features
    expect(screen.getByText('Passive/Active modes')).toBeInTheDocument()
    expect(screen.getByText('Port configuration')).toBeInTheDocument()

    // Local features
    expect(screen.getByText('Base path configuration')).toBeInTheDocument()
    expect(screen.getByText('No authentication')).toBeInTheDocument()
  })

  it('has a disabled Next button when no protocol is selected', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    const nextButton = screen.getByText('Next')
    expect(nextButton).toBeDisabled()
  })

  it('enables Next button when a protocol is selected', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('SMB/CIFS'))

    const nextButton = screen.getByText('Next')
    expect(nextButton).not.toBeDisabled()
  })

  it('shows selection confirmation when protocol is selected', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('FTP'))

    expect(screen.getByText('FTP Selected')).toBeInTheDocument()
    expect(screen.getByText(/Click "Next" to configure your ftp connection/)).toBeInTheDocument()
  })

  it('navigates to correct route on Next click', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('NFS'))
    fireEvent.click(screen.getByText('Next'))

    expect(mockNavigate).toHaveBeenCalledWith('/configure-nfs')
  })

  it('navigates back on Previous click', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('Previous'))

    expect(mockNavigate).toHaveBeenCalledWith('/')
  })

  it('allows switching protocol selection', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('SMB/CIFS'))
    expect(screen.getByText('SMB/CIFS Selected')).toBeInTheDocument()

    fireEvent.click(screen.getByText('WebDAV'))
    expect(screen.getByText('WebDAV Selected')).toBeInTheDocument()
  })
})
