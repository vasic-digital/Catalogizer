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

  it('navigates to SMB configuration on SMB selection', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('SMB/CIFS'))
    fireEvent.click(screen.getByText('Next'))

    expect(mockNavigate).toHaveBeenCalledWith('/configure-smb')
  })

  it('navigates to FTP configuration on FTP selection', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('FTP'))
    fireEvent.click(screen.getByText('Next'))

    expect(mockNavigate).toHaveBeenCalledWith('/configure-ftp')
  })

  it('navigates to WebDAV configuration on WebDAV selection', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('WebDAV'))
    fireEvent.click(screen.getByText('Next'))

    expect(mockNavigate).toHaveBeenCalledWith('/configure-webdav')
  })

  it('navigates to local configuration on Local Files selection', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('Local Files'))
    fireEvent.click(screen.getByText('Next'))

    expect(mockNavigate).toHaveBeenCalledWith('/configure-local')
  })

  it('shows NFS-specific features', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    expect(screen.getByText('Mount point configuration')).toBeInTheDocument()
    expect(screen.getByText('Version specification')).toBeInTheDocument()
    expect(screen.getByText('Options support')).toBeInTheDocument()
    expect(screen.getByText('Host-based access')).toBeInTheDocument()
  })

  it('shows WebDAV-specific features', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    expect(screen.getByText('HTTP/HTTPS support')).toBeInTheDocument()
    expect(screen.getByText('SSL/TLS encryption')).toBeInTheDocument()
  })

  it('shows Local Files specific features', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    expect(screen.getByText('Fast access')).toBeInTheDocument()
    expect(screen.getByText('Full permissions')).toBeInTheDocument()
  })

  it('shows confirmation text with lowercase protocol name', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    fireEvent.click(screen.getByText('NFS'))
    expect(screen.getByText(/Click "Next" to configure your nfs connection/)).toBeInTheDocument()
  })

  it('does not show confirmation text before selection', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    expect(screen.queryByText(/Selected$/)).toBeNull()
  })

  it('does not navigate on Next click without protocol selected', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    const nextButton = screen.getByText('Next')
    fireEvent.click(nextButton)

    expect(mockNavigate).not.toHaveBeenCalledWith(expect.stringContaining('/configure'))
  })

  it('renders Previous button', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    expect(screen.getByText('Previous')).toBeInTheDocument()
  })

  it('renders five protocol cards', () => {
    render(
      <TestWrapper>
        <ProtocolSelectionStep />
      </TestWrapper>
    )

    const grid = screen.getByText('SMB/CIFS').closest('.grid')
    expect(grid).toBeInTheDocument()
    expect(grid!.children.length).toBe(5)
  })
})
