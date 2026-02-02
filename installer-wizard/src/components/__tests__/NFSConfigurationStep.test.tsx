import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import NFSConfigurationStep from '../wizard/NFSConfigurationStep'
import { TauriService } from '../../services/tauri'
import { TestWrapper, getInputByLabel } from '../../test/test-utils'

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
    expect(screen.getByText('Configuration Name', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Host/IP Address', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Export Path', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Mount Point', { selector: 'label' })).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Name is required')).toBeInTheDocument()
      expect(screen.getByText('Host is required')).toBeInTheDocument()
      expect(screen.getByText('Path is required')).toBeInTheDocument()
      expect(screen.getByText('Mount point is required')).toBeInTheDocument()
    })
  })

  it('tests NFS connection successfully', async () => {
    const mockTestNFSConnection = vi.spyOn(TauriService, 'testNFSConnection')
      .mockResolvedValue(true)

    render(
      <TestWrapper>
        <NFSConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.example.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/export/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/nfs' } })

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

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test NFS' } })
    fireEvent.change(getInputByLabel('Host/IP Address'), { target: { value: 'nfs.example.com' } })
    fireEvent.change(getInputByLabel('Export Path'), { target: { value: '/export/data' } })
    fireEvent.change(getInputByLabel('Mount Point'), { target: { value: '/mnt/nfs' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Test NFS')).toBeInTheDocument()
      expect(screen.getByText(/nfs\.example\.com/)).toBeInTheDocument()
    })
  })
})
