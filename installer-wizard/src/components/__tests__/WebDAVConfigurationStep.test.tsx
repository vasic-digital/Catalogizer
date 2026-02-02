import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import WebDAVConfigurationStep from '../wizard/WebDAVConfigurationStep'
import { TauriService } from '../../services/tauri'
import { TestWrapper, getInputByLabel } from '../../test/test-utils'

describe('WebDAVConfigurationStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders WebDAV configuration form', () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('WebDAV Configuration')).toBeInTheDocument()
    expect(screen.getByText('Configuration Name', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('WebDAV URL', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Username', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Password', { selector: 'label' })).toBeInTheDocument()
  })

  it('validates required fields', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
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

    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'invalid-url' } })
    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'user' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'pass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
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

    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://webdav.example.com/dav' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const testButton = screen.getByText('Test Connection')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(mockTestWebDAVConnection).toHaveBeenCalled()
      expect(screen.getByText('Connection successful!')).toBeInTheDocument()
    })
  })

  it('adds WebDAV configuration successfully', async () => {
    render(
      <TestWrapper>
        <WebDAVConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'Test WebDAV' } })
    fireEvent.change(getInputByLabel('WebDAV URL'), { target: { value: 'https://webdav.example.com/dav' } })
    fireEvent.change(getInputByLabel('Username'), { target: { value: 'testuser' } })
    fireEvent.change(getInputByLabel('Password'), { target: { value: 'testpass' } })

    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Test WebDAV')).toBeInTheDocument()
      expect(screen.getByText('https://webdav.example.com/dav')).toBeInTheDocument()
    })
  })
})
