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

    // First, click "Add New" to get a clean form (since component starts in edit mode with pre-populated data)
    // Cancel editing to get to "Add Configuration" state
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
})
