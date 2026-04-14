import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import LocalConfigurationStep from '../wizard/LocalConfigurationStep'
import { TauriService } from '../../services/tauri'
import { TestWrapper, getInputByLabel } from '../../test/test-utils'

describe('LocalConfigurationStep', () => {
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
    expect(screen.getByText('Configuration Name', { selector: 'label' })).toBeInTheDocument()
    expect(screen.getByText('Base Path', { selector: 'label' })).toBeInTheDocument()
  })

  it('pre-populates with default local path', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // LocalConfigurationStep auto-populates with a default config and starts in edit mode
    expect(screen.getByText('Edit Configuration')).toBeInTheDocument()
  })

  it('validates required fields on empty form', async () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // First, click "Cancel" to get out of edit mode with pre-populated data
    const cancelButton = screen.getByText('Cancel')
    fireEvent.click(cancelButton)

    // Clear the fields
    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: '' } })
    fireEvent.change(getInputByLabel('Base Path'), { target: { value: '' } })

    // Now submit the empty form
    const submitButton = screen.getByRole('button', { name: 'Add Configuration' })
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

    fireEvent.change(getInputByLabel('Base Path'), { target: { value: '/home/user/media' } })

    const testButton = screen.getByText('Test Path')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(mockTestLocalConnection).toHaveBeenCalledWith('/home/user/media')
      expect(screen.getByText('Path accessible!')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('updates local configuration successfully', async () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // Since the component starts in edit mode, update the pre-populated config
    fireEvent.change(getInputByLabel('Configuration Name'), { target: { value: 'My Media' } })
    fireEvent.change(getInputByLabel('Base Path'), { target: { value: '/opt/media' } })

    const submitButton = screen.getByRole('button', { name: 'Update Configuration' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('My Media')).toBeInTheDocument()
      expect(screen.getByText('/opt/media')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('shows configured sources count', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // Component auto-creates one default config
    expect(screen.getByText(/1 local source\(s\) configured/)).toBeInTheDocument()
  })

  it('handles local path test failure', async () => {
    vi.spyOn(TauriService, 'testLocalConnection')
      .mockRejectedValue(new Error('Permission denied'))

    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    fireEvent.change(getInputByLabel('Base Path'), { target: { value: '/root/protected' } })

    const testButton = screen.getByText('Test Path')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText(/Path test failed/)).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('requires base path before testing', async () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // Clear the base path
    fireEvent.change(getInputByLabel('Base Path'), { target: { value: '' } })

    const testButton = screen.getByText('Test Path')
    fireEvent.click(testButton)

    await waitFor(() => {
      expect(screen.getByText('Please fill in the base path before testing')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('shows subtitle text', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Configure local filesystem paths for your media')).toBeInTheDocument()
  })

  it('shows form description text', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Enter the local filesystem path details')).toBeInTheDocument()
  })

  it('shows Add New button', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Add New')).toBeInTheDocument()
  })

  it('shows next step instruction when configs exist', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // Auto-created default config exists
    expect(screen.getByText(/Click "Next" to manage your configuration file/)).toBeInTheDocument()
  })

  it('shows edit button for existing configuration', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // The auto-created default config should have an Edit button
    expect(screen.getByText('Edit')).toBeInTheDocument()
  })

  it('shows default config name and path', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Local Media')).toBeInTheDocument()
    expect(screen.getByText('/home/user/media')).toBeInTheDocument()
  })

  it('re-creates default config after removing last entry', async () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    // Component auto-creates one config; remove it
    const deleteButtons = screen.getAllByRole('button').filter(btn =>
      btn.classList.contains('text-red-600') || btn.className.includes('text-red')
    )
    expect(deleteButtons.length).toBeGreaterThan(0)
    fireEvent.click(deleteButtons[0])

    // The useEffect re-populates when localConfigs becomes empty,
    // so a new default config is created automatically
    await waitFor(() => {
      expect(screen.getByText('Local Media')).toBeInTheDocument()
      expect(screen.getByText('/home/user/media')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('shows manage local filesystem text in list description', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText('Manage your local filesystem source configurations')).toBeInTheDocument()
  })

  it('shows configured sources count in list header', () => {
    render(
      <TestWrapper>
        <LocalConfigurationStep />
      </TestWrapper>
    )

    expect(screen.getByText(/Configured Sources \(1\)/)).toBeInTheDocument()
  })
})
